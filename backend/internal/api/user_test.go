package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"paperplay/config"
	"paperplay/internal/middleware"
	"paperplay/internal/model"
	"paperplay/internal/service"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Auto migrate the schema
	db.AutoMigrate(
		&model.User{},
		&model.RefreshToken{},
		&model.UserProgress{},
		&model.UserAttempts{},
		&model.Achievement{},
		&model.UserAchievement{},
		&model.Event{},
		&model.NFTAsset{},
	)

	return db
}

func setupTestRouter(handler *UserHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.RefreshToken)
	}

	protected := v1.Group("/users")
	// For testing, we'll add a simple middleware that sets user ID
	protected.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Set("user", &model.User{ID: "test-user-id", Email: "test@example.com"})
		c.Next()
	})
	{
		protected.GET("/profile", handler.GetProfile)
		protected.PUT("/profile", handler.UpdateProfile)
		protected.GET("/progress", handler.GetUserProgress)
		protected.GET("/achievements", handler.GetUserAchievements)
		protected.POST("/logout", handler.Logout)
	}

	return router
}

func createTestServices(db *gorm.DB) (*middleware.JWTService, *service.UserService, *service.EthereumService) {
	// Create JWT service
	jwtService := middleware.NewJWTService(&config.Config{
		JWT: config.JWTConfig{
			SecretKey:            "test-secret",
			AccessTokenDuration:  15,
			RefreshTokenDuration: 7,
		},
	}, db)

	// Create user service
	userService := service.NewUserService(db)

	// Create ethereum service (disabled for testing)
	ethConfig := &config.EthereumConfig{
		Enabled: false,
	}
	ethService, _ := service.NewEthereumService(ethConfig)

	return jwtService, userService, ethService
}

func TestUserHandler_Register(t *testing.T) {
	db := setupTestDB()
	jwtService, userService, ethService := createTestServices(db)

	handler := NewUserHandler(db, jwtService, userService, ethService)
	router := setupTestRouter(handler)

	tests := []struct {
		name           string
		payload        RegisterRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Valid registration",
			payload: RegisterRequest{
				Email:       "test@example.com",
				Password:    "password123",
				DisplayName: "Test User",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid email",
			payload: RegisterRequest{
				Email:       "invalid-email",
				Password:    "password123",
				DisplayName: "Test User",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
		},
		{
			name: "Short password",
			payload: RegisterRequest{
				Email:       "test2@example.com",
				Password:    "123",
				DisplayName: "Test User",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
		},
		{
			name: "Missing display name",
			payload: RegisterRequest{
				Email:    "test3@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	db := setupTestDB()
	jwtService, userService, ethService := createTestServices(db)

	handler := NewUserHandler(db, jwtService, userService, ethService)
	router := setupTestRouter(handler)

	// Create a test user first
	user := &model.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
	user.SetPassword("password123")
	db.Create(user)

	tests := []struct {
		name           string
		payload        LoginRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Valid login",
			payload: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid email",
			payload: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid_credentials",
		},
		{
			name: "Invalid password",
			payload: LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid_credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			} else if tt.expectedStatus == http.StatusOK {
				var response AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Equal(t, "Bearer", response.TokenType)
			}
		})
	}
}

func TestUserHandler_GetProfile(t *testing.T) {
	db := setupTestDB()
	jwtService, userService, ethService := createTestServices(db)

	handler := NewUserHandler(db, jwtService, userService, ethService)
	router := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "user")
	assert.Contains(t, response, "stats")
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	db := setupTestDB()
	jwtService, userService, ethService := createTestServices(db)

	handler := NewUserHandler(db, jwtService, userService, ethService)
	router := setupTestRouter(handler)

	// Create a test user
	user := &model.User{
		ID:          "test-user-id",
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
	db.Create(user)

	payload := map[string]string{
		"display_name": "Updated Name",
		"avatar_url":   "https://example.com/avatar.jpg",
	}

	jsonPayload, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/profile", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Profile updated successfully", response.Message)
}

func TestUserHandler_GetUserProgress(t *testing.T) {
	db := setupTestDB()
	jwtService, userService, ethService := createTestServices(db)

	handler := NewUserHandler(db, jwtService, userService, ethService)
	router := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/progress", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

func TestUserHandler_GetUserAchievements(t *testing.T) {
	db := setupTestDB()
	jwtService, userService, ethService := createTestServices(db)

	handler := NewUserHandler(db, jwtService, userService, ethService)
	router := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/achievements", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

func TestUserHandler_Logout(t *testing.T) {
	db := setupTestDB()
	jwtService, userService, ethService := createTestServices(db)

	handler := NewUserHandler(db, jwtService, userService, ethService)
	router := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Successfully logged out", response.Message)
}
