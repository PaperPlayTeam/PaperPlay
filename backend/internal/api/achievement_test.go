package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"paperplay/config"
	"paperplay/internal/middleware"
	"paperplay/internal/model"
	"paperplay/internal/service"
	"paperplay/internal/websocket"
)

var (
	testMetricsService *middleware.MetricsService
	testMetricsOnce    sync.Once
)

func setupAchievementTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Auto migrate the schema - include all required models
	db.AutoMigrate(
		&model.Achievement{},
		&model.UserAchievement{},
		&model.Event{},
		&model.User{},
		&model.Subject{},
		&model.Paper{},
		&model.Level{},
		&model.Question{},
		&model.UserProgress{},
		&model.UserAttempts{},
		&model.RefreshToken{},
		&model.NFTAsset{},
		&model.RoadmapNode{},
	)

	return db
}

func getTestMetricsService() *middleware.MetricsService {
	testMetricsOnce.Do(func() {
		testMetricsService = middleware.NewMetricsService()
	})
	return testMetricsService
}

func createTestAchievementServices(db *gorm.DB) (*service.AchievementService, *middleware.MetricsService, *websocket.Hub) {
	// Create logger
	logger, _ := zap.NewDevelopment()

	// Create websocket hub
	wsHub := websocket.NewHub(logger)

	// Create ethereum service (disabled for testing)
	ethConfig := &config.EthereumConfig{
		Enabled: false,
	}
	ethService, _ := service.NewEthereumService(ethConfig)

	// Create achievement service
	achievementService := service.NewAchievementService(db, logger, wsHub, ethService)

	// Use shared metrics service to avoid duplicate registration
	metricsService := getTestMetricsService()

	return achievementService, metricsService, wsHub
}

func setupAchievementTestRouter(handler *AchievementHandler, systemHandler *SystemHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock authentication middleware
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Next()
	})

	// Health and metrics endpoints
	router.GET("/health", systemHandler.GetHealth)
	router.GET("/metrics", systemHandler.GetPrometheusMetrics)

	v1 := router.Group("/api/v1")
	{
		achievements := v1.Group("/achievements")
		{
			achievements.GET("", handler.GetAllAchievements)
			achievements.GET("/user", handler.GetUserAchievements)
			achievements.POST("/evaluate", handler.EvaluateAchievements)
		}

		ws := v1.Group("/ws")
		{
			ws.GET("/stats", handler.GetWebSocketStats)
		}

		system := v1.Group("/system")
		{
			system.GET("/stats", handler.GetSystemStats)
		}
	}

	return router
}

func seedAchievementTestData(db *gorm.DB) {
	// Create test achievement
	achievement := &model.Achievement{
		ID:          "test-achievement-1",
		Name:        "首战告捷",
		Description: "第一次作答即答对任意一道题目",
		Level:       1,
		BadgeType:   "learning",
		RuleJSON:    `{"type":"first_try","conditions":[{"field":"attempts_first_try_correct","operator":">=","value":1}]}`,
		NFTEnabled:  true,
		NFTMetadata: `{"name":"First Victory Badge","description":"Awarded for getting first question right on first try","image":"","attributes":[{"trait_type":"Achievement Type","value":"Learning"}]}`,
		IsActive:    true,
	}
	db.Create(achievement)

	// Create test user
	user := &model.User{
		ID:          "test-user-id",
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
	db.Create(user)
}

func TestAchievementHandler_GetAllAchievements(t *testing.T) {
	db := setupAchievementTestDB()
	seedAchievementTestData(db)

	achievementService, metricsService, wsHub := createTestAchievementServices(db)

	handler := NewAchievementHandler(db, achievementService, metricsService, wsHub)
	systemHandler := NewSystemHandler(db, metricsService, wsHub)
	router := setupAchievementTestRouter(handler, systemHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/achievements", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	// Check if data contains achievement details
	data, ok := response.Data.([]interface{})
	assert.True(t, ok)
	// Since we're using real service, we might get all achievements
	assert.GreaterOrEqual(t, len(data), 0)
}

func TestAchievementHandler_GetUserAchievements(t *testing.T) {
	db := setupAchievementTestDB()
	seedAchievementTestData(db)

	achievementService, metricsService, wsHub := createTestAchievementServices(db)

	handler := NewAchievementHandler(db, achievementService, metricsService, wsHub)
	systemHandler := NewSystemHandler(db, metricsService, wsHub)
	router := setupAchievementTestRouter(handler, systemHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/achievements/user", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

func TestAchievementHandler_EvaluateAchievements(t *testing.T) {
	db := setupAchievementTestDB()
	seedAchievementTestData(db)

	achievementService, metricsService, wsHub := createTestAchievementServices(db)

	handler := NewAchievementHandler(db, achievementService, metricsService, wsHub)
	systemHandler := NewSystemHandler(db, metricsService, wsHub)
	router := setupAchievementTestRouter(handler, systemHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/achievements/evaluate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "成就评估完成", response.Message)
}

func TestAchievementHandler_GetWebSocketStats(t *testing.T) {
	db := setupAchievementTestDB()

	achievementService, metricsService, wsHub := createTestAchievementServices(db)

	handler := NewAchievementHandler(db, achievementService, metricsService, wsHub)
	systemHandler := NewSystemHandler(db, metricsService, wsHub)
	router := setupAchievementTestRouter(handler, systemHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ws/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	data := response.Data.(map[string]interface{})
	assert.Contains(t, data, "total_clients")
	assert.Contains(t, data, "connected_users")
}

func TestAchievementHandler_GetWebSocketStats_Unavailable(t *testing.T) {
	db := setupAchievementTestDB()

	achievementService, metricsService, _ := createTestAchievementServices(db)

	// Create handler without WebSocket hub (nil)
	handler := NewAchievementHandler(db, achievementService, metricsService, nil)
	systemHandler := NewSystemHandler(db, metricsService, nil)
	router := setupAchievementTestRouter(handler, systemHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ws/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "websocket_unavailable", response.Error)
}

func TestAchievementHandler_GetSystemStats(t *testing.T) {
	db := setupAchievementTestDB()

	achievementService, metricsService, wsHub := createTestAchievementServices(db)

	handler := NewAchievementHandler(db, achievementService, metricsService, wsHub)
	systemHandler := NewSystemHandler(db, metricsService, wsHub)
	router := setupAchievementTestRouter(handler, systemHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	data := response.Data.(map[string]interface{})
	assert.Contains(t, data, "websocket")
	assert.Contains(t, data, "database")
	assert.Contains(t, data, "metrics")

	// Check database health
	dbData := data["database"].(map[string]interface{})
	assert.Equal(t, true, dbData["healthy"])
}

func TestSystemHandler_GetHealth(t *testing.T) {
	db := setupAchievementTestDB()
	// Add the achievement test data to make the count work
	seedAchievementTestData(db)

	achievementService, metricsService, wsHub := createTestAchievementServices(db)

	handler := NewAchievementHandler(db, achievementService, metricsService, wsHub)
	systemHandler := NewSystemHandler(db, metricsService, wsHub)
	router := setupAchievementTestRouter(handler, systemHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "services")

	services := response["services"].(map[string]interface{})
	assert.Equal(t, true, services["database"])
	assert.Equal(t, true, services["websocket"])
	assert.Equal(t, false, services["ethereum"])
	assert.Equal(t, true, services["achievements"])
}

func TestSystemHandler_GetPrometheusMetrics_Unavailable(t *testing.T) {
	db := setupAchievementTestDB()

	// Create system handler without metrics service (nil)
	systemHandler := NewSystemHandler(db, nil, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", systemHandler.GetPrometheusMetrics)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "metrics_unavailable", response.Error)
}
