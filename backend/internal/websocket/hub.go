package websocket

import (
	"encoding/json"
	"net/http"
	"paperplay/internal/middleware"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// User ID to client mapping for targeted messages
	userClients map[string][]*Client

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Logger
	logger *zap.Logger

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// User ID for targeted messaging
	userID string

	// Client ID for identification
	clientID string
}

// Message represents a WebSocket message
type Message struct {
	Type      string    `json:"type"`
	UserID    string    `json:"user_id,omitempty"`
	Data      any       `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

// NotificationMessage represents a notification message
type NotificationMessage struct {
	ID          string `json:"id"`
	Type        string `json:"type"` // "achievement", "level_completed", "system"
	Title       string `json:"title"`
	Message     string `json:"message"`
	Icon        string `json:"icon,omitempty"`
	URL         string `json:"url,omitempty"`
	Achievement *struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Level       int    `json:"level"`
		IconURL     string `json:"icon_url"`
	} `json:"achievement,omitempty"`
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should check the origin properly
		return true
	},
}

// NewHub creates a new WebSocket hub
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		userClients: make(map[string][]*Client),
		broadcast:   make(chan []byte, 256),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		logger:      logger,
	}
}

// Run starts the hub and handles client connections
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true

			// Add to user clients mapping
			if client.userID != "" {
				h.userClients[client.userID] = append(h.userClients[client.userID], client)
			}
			h.mu.Unlock()

			h.logger.Info("Client connected",
				zap.String("client_id", client.clientID),
				zap.String("user_id", client.userID),
			)

			// Send welcome message
			welcomeMsg := Message{
				Type:      "connected",
				Data:      map[string]string{"status": "connected"},
				Timestamp: time.Now(),
			}
			if data, err := json.Marshal(welcomeMsg); err == nil {
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// Remove from user clients mapping
				if client.userID != "" {
					clients := h.userClients[client.userID]
					for i, c := range clients {
						if c == client {
							h.userClients[client.userID] = append(clients[:i], clients[i+1:]...)
							break
						}
					}
					if len(h.userClients[client.userID]) == 0 {
						delete(h.userClients, client.userID)
					}
				}
			}
			h.mu.Unlock()

			h.logger.Info("Client disconnected",
				zap.String("client_id", client.clientID),
				zap.String("user_id", client.userID),
			)

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToAll sends a message to all connected clients
func (h *Hub) BroadcastToAll(messageType string, data any) {
	message := Message{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now(),
	}

	if jsonData, err := json.Marshal(message); err == nil {
		select {
		case h.broadcast <- jsonData:
		default:
			h.logger.Warn("Broadcast channel is full, message dropped")
		}
	} else {
		h.logger.Error("Failed to marshal broadcast message", zap.Error(err))
	}
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID string, messageType string, data any) {
	h.mu.RLock()
	clients, exists := h.userClients[userID]
	h.mu.RUnlock()

	if !exists || len(clients) == 0 {
		h.logger.Debug("No clients found for user", zap.String("user_id", userID))
		return
	}

	message := Message{
		Type:      messageType,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now(),
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal user message", zap.Error(err))
		return
	}

	h.mu.RLock()
	for _, client := range clients {
		select {
		case client.send <- jsonData:
		default:
			// Client's send channel is full, remove it
			close(client.send)
			delete(h.clients, client)
		}
	}
	h.mu.RUnlock()
}

// SendNotification sends a notification to a specific user
func (h *Hub) SendNotification(userID string, notification *NotificationMessage) {
	h.SendToUser(userID, "notification", notification)
	h.logger.Info("Notification sent",
		zap.String("user_id", userID),
		zap.String("notification_type", notification.Type),
		zap.String("title", notification.Title),
	)
}

// GetConnectedUsers returns a list of currently connected user IDs
func (h *Hub) GetConnectedUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var users []string
	for userID := range h.userClients {
		if len(h.userClients[userID]) > 0 {
			users = append(users, userID)
		}
	}
	return users
}

// GetStats returns hub statistics
func (h *Hub) GetStats() map[string]any {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]any{
		"total_clients":     len(h.clients),
		"connected_users":   len(h.userClients),
		"broadcast_backlog": len(h.broadcast),
	}
}

// ServeWS handles websocket requests from the peer.
func (h *Hub) ServeWS(c *gin.Context, jwtService *middleware.JWTService) {
	// Get user from JWT token (optional authentication)
	var userID string
	if user, exists := middleware.GetCurrentUser(c); exists {
		userID = user.ID
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Generate client ID
	clientID := generateClientID()

	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		clientID: clientID,
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}

		// Handle incoming messages (ping, authentication, etc.)
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from clients
func (c *Client) handleMessage(message []byte) {
	var msg map[string]any
	if err := json.Unmarshal(message, &msg); err != nil {
		c.hub.logger.Warn("Failed to unmarshal client message", zap.Error(err))
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "ping":
		// Respond with pong
		pongMsg := Message{
			Type:      "pong",
			Data:      map[string]string{"status": "ok"},
			Timestamp: time.Now(),
		}
		if data, err := json.Marshal(pongMsg); err == nil {
			select {
			case c.send <- data:
			default:
				// Channel full, skip
			}
		}

	case "subscribe":
		// Handle subscription to specific channels
		if channel, ok := msg["channel"].(string); ok {
			c.hub.logger.Info("Client subscribed to channel",
				zap.String("client_id", c.clientID),
				zap.String("channel", channel),
			)
		}

	default:
		c.hub.logger.Debug("Unknown message type",
			zap.String("type", msgType),
			zap.String("client_id", c.clientID),
		)
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
