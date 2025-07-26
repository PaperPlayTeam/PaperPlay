package api

import (
	"net/http"
	"paperplay/internal/middleware"
	"paperplay/internal/model"
	"paperplay/internal/service"
	"paperplay/internal/websocket"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// AchievementHandler handles achievement-related HTTP requests
type AchievementHandler struct {
	db                 *gorm.DB
	achievementService *service.AchievementService
	metricsService     *middleware.MetricsService
	wsHub              *websocket.Hub
	validator          *validator.Validate
}

// NewAchievementHandler creates a new achievement handler
func NewAchievementHandler(
	db *gorm.DB,
	achievementService *service.AchievementService,
	metricsService *middleware.MetricsService,
	wsHub *websocket.Hub,
) *AchievementHandler {
	return &AchievementHandler{
		db:                 db,
		achievementService: achievementService,
		metricsService:     metricsService,
		wsHub:              wsHub,
		validator:          validator.New(),
	}
}

// AchievementResponse represents achievement data for API responses
type AchievementResponse struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	IconURL     string         `json:"icon_url"`
	Level       int            `json:"level"`
	Category    string         `json:"category"`
	IsActive    bool           `json:"is_active"`
	Rules       map[string]any `json:"rules"`
	NFTMetadata map[string]any `json:"nft_metadata,omitempty"`
}

// UserAchievementResponse represents user achievement data for API responses
type UserAchievementResponse struct {
	ID            string              `json:"id"`
	UserID        string              `json:"user_id"`
	AchievementID string              `json:"achievement_id"`
	EarnedAt      string              `json:"earned_at"`
	EventData     map[string]any      `json:"event_data,omitempty"`
	Achievement   *AchievementSummary `json:"achievement"`
}

// AchievementSummary represents a summary of achievement info
type AchievementSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       int    `json:"level"`
	IconURL     string `json:"icon_url"`
}

// GetAllAchievements handles GET /api/v1/achievements
func (h *AchievementHandler) GetAllAchievements(c *gin.Context) {
	achievements, err := h.achievementService.GetAllAchievements()
	if err != nil {
		h.metricsService.RecordError("database_error", "achievements")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "failed_to_get_achievements",
			Message: "Failed to retrieve achievements",
			Details: err.Error(),
		})
		return
	}

	// Convert to response format
	var response []AchievementResponse
	for _, achievement := range achievements {
		// Parse rule JSON
		rule, _ := achievement.GetRule()
		var rules map[string]any
		if rule != nil {
			rules = map[string]any{
				"type":       rule.Type,
				"conditions": rule.Conditions,
			}
		}

		// Parse NFT metadata JSON if available
		var nftMetadata map[string]any
		if achievement.NFTEnabled && achievement.NFTMetadata != "" {
			metadata, _ := achievement.GetNFTMetadata()
			if metadata != nil {
				nftMetadata = map[string]any{
					"name":        metadata.Name,
					"description": metadata.Description,
					"image":       metadata.Image,
					"attributes":  metadata.Attributes,
				}
			}
		}

		response = append(response, AchievementResponse{
			ID:          achievement.ID,
			Name:        achievement.Name,
			Description: achievement.Description,
			IconURL:     achievement.IconURL,
			Level:       achievement.Level,
			Category:    achievement.BadgeType,
			IsActive:    achievement.IsActive,
			Rules:       rules,
			NFTMetadata: nftMetadata,
		})
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// GetUserAchievements handles GET /api/v1/achievements/user
func (h *AchievementHandler) GetUserAchievements(c *gin.Context) {
	userID := middleware.MustGetCurrentUserID(c)

	userAchievements, err := h.achievementService.GetUserAchievements(userID)
	if err != nil {
		h.metricsService.RecordError("database_error", "user_achievements")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "failed_to_get_user_achievements",
			Message: "Failed to retrieve user achievements",
			Details: err.Error(),
		})
		return
	}

	// Convert to response format
	var response []UserAchievementResponse
	for _, userAchievement := range userAchievements {
		// Get event data from related events if available
		var eventData map[string]interface{}
		var event model.Event
		if err := h.db.Where("user_id = ? AND event_type = ? AND created_at BETWEEN ? AND ?",
			userID, model.EventAchievementEarned,
			userAchievement.EarnedAt.Add(-5*time.Second),
			userAchievement.EarnedAt.Add(5*time.Second)).
			First(&event).Error; err == nil {
			eventData, _ = event.GetData()
		}

		achievementSummary := &AchievementSummary{
			ID:          userAchievement.Achievement.ID,
			Name:        userAchievement.Achievement.Name,
			Description: userAchievement.Achievement.Description,
			Level:       userAchievement.Achievement.Level,
			IconURL:     userAchievement.Achievement.IconURL,
		}

		response = append(response, UserAchievementResponse{
			ID:            userAchievement.ID,
			UserID:        userAchievement.UserID,
			AchievementID: userAchievement.AchievementID,
			EarnedAt:      userAchievement.EarnedAt.Format("2006-01-02T15:04:05Z"),
			EventData:     eventData,
			Achievement:   achievementSummary,
		})
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// EvaluateAchievements handles POST /api/v1/achievements/evaluate
func (h *AchievementHandler) EvaluateAchievements(c *gin.Context) {
	userID := middleware.MustGetCurrentUserID(c)

	// Manually trigger achievement evaluation
	err := h.achievementService.EvaluateUserAchievements(userID)
	if err != nil {
		h.metricsService.RecordError("evaluation_failed", "achievements")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "evaluation_failed",
			Message: "Failed to evaluate achievements",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "成就评估完成",
	})
}

// GetWebSocketStats handles GET /api/v1/ws/stats
func (h *AchievementHandler) GetWebSocketStats(c *gin.Context) {
	if h.wsHub == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "websocket_unavailable",
			Message: "WebSocket service is not available",
		})
		return
	}

	stats := h.wsHub.GetStats()

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    stats,
	})
}

// GetSystemStats handles GET /api/v1/system/stats
func (h *AchievementHandler) GetSystemStats(c *gin.Context) {
	stats := make(map[string]interface{})

	// WebSocket stats
	if h.wsHub != nil {
		stats["websocket"] = h.wsHub.GetStats()
	} else {
		stats["websocket"] = map[string]interface{}{
			"status": "unavailable",
		}
	}

	// Database health
	dbHealth, err := h.db.DB()
	if err == nil && dbHealth != nil {
		pingErr := dbHealth.Ping()
		stats["database"] = map[string]interface{}{
			"healthy": pingErr == nil,
		}
	} else {
		stats["database"] = map[string]interface{}{
			"healthy": false,
		}
	}

	// Metrics health (if available)
	if h.metricsService != nil {
		stats["metrics"] = h.metricsService.HealthCheck()
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    stats,
	})
}

// SystemHandler handles system-level HTTP requests
type SystemHandler struct {
	db             *gorm.DB
	metricsService *middleware.MetricsService
	wsHub          *websocket.Hub
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(
	db *gorm.DB,
	metricsService *middleware.MetricsService,
	wsHub *websocket.Hub,
) *SystemHandler {
	return &SystemHandler{
		db:             db,
		metricsService: metricsService,
		wsHub:          wsHub,
	}
}

// GetHealth handles GET /health - comprehensive health check
func (h *SystemHandler) GetHealth(c *gin.Context) {
	healthStatus := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"services":  make(map[string]interface{}),
	}

	services := healthStatus["services"].(map[string]interface{})

	// Database health
	if sqlDB, err := h.db.DB(); err == nil {
		services["database"] = sqlDB.Ping() == nil
	} else {
		services["database"] = false
	}

	// WebSocket health
	if h.wsHub != nil {
		wsStats := h.wsHub.GetStats()
		services["websocket"] = true
		if cronJobsData, exists := wsStats["cron_jobs"]; exists {
			services["cron_jobs"] = cronJobsData
		}
	} else {
		services["websocket"] = false
	}

	// Achievements service health
	var achievementCount int64
	result := h.db.Model(&model.Achievement{}).Count(&achievementCount)
	if result.Error == nil {
		services["achievements"] = true
	} else {
		services["achievements"] = false
	}

	// Ethereum service health (placeholder - would be implemented if Ethereum service is active)
	services["ethereum"] = false

	c.JSON(http.StatusOK, healthStatus)
}

// GetPrometheusMetrics handles GET /metrics - Prometheus metrics endpoint
func (h *SystemHandler) GetPrometheusMetrics(c *gin.Context) {
	if h.metricsService == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "metrics_unavailable",
			Message: "Metrics service is not available",
		})
		return
	}

	// Use the Prometheus handler from metrics service
	h.metricsService.PrometheusHandler()(c)
}
