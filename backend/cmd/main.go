package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"paperplay/config"
	"paperplay/internal/api"
	"paperplay/internal/cron"
	"paperplay/internal/middleware"
	"paperplay/internal/model"
	"paperplay/internal/service"
	"paperplay/internal/websocket"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := middleware.NewLoggerService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	logger.GetSugar().Info("Starting PaperPlay backend server...")

	// Check if database exists before connecting
	dbPath := cfg.Database.DSN
	if strings.Contains(dbPath, "?") {
		dbPath = strings.Split(dbPath, "?")[0]
	}

	var isNewDatabase bool
	if stat, err := os.Stat(dbPath); err != nil || stat.Size() == 0 {
		isNewDatabase = true
		logger.GetSugar().Info("No existing database found. Will create new database at: %s", cfg.Database.DSN)
	} else {
		isNewDatabase = false
		logger.GetSugar().Info("Existing database detected at: %s", cfg.Database.DSN)
		logger.GetSugar().Info("Database will be validated but NOT modified to protect existing data")
	}

	// Initialize database
	db, err := model.NewDatabase(cfg.Database.DSN, cfg.Log.Level)
	if err != nil {
		logger.GetSugar().Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations (safe for existing databases)
	if isNewDatabase {
		logger.GetSugar().Info("Running full database migrations for new database...")
		if err := db.Migrate(); err != nil {
			logger.GetSugar().Fatalf("Failed to migrate database: %v", err)
		}
	} else {
		logger.GetSugar().Info("Validating existing database schema...")
		if err := db.ValidateSchema(); err != nil {
			logger.GetSugar().Fatalf("Failed to validate database: %v", err)
		}
	}

	// Create additional indexes
	if err := db.CreateIndexes(); err != nil {
		logger.GetSugar().Fatalf("Failed to create database indexes: %v", err)
	}

	// Seed initial data (only for new databases)
	if isNewDatabase {
		logger.GetSugar().Info("Seeding initial data for new database...")
		if err := db.Seed(); err != nil {
			logger.GetSugar().Fatalf("Failed to seed database: %v", err)
		}
	} else {
		logger.GetSugar().Info("Skipping data seeding for existing database to protect existing data")
	}

	// Initialize services
	userService := service.NewUserService(db.DB)

	// Initialize Ethereum service (optional)
	var ethService *service.EthereumService
	if cfg.Ethereum.Enabled {
		ethService, err = service.NewEthereumService(&cfg.Ethereum)
		if err != nil {
			logger.GetSugar().Warnf("Failed to initialize Ethereum service: %v", err)
			ethService = nil
		} else {
			defer ethService.Close()
		}
	}

	// Initialize JWT service
	jwtService := middleware.NewJWTService(cfg, db.DB)

	// Initialize metrics service
	metricsService := middleware.NewMetricsService()

	// Initialize WebSocket hub
	wsHub := websocket.NewHub(logger.GetLogger())
	go wsHub.Run()
	logger.GetSugar().Info("WebSocket hub started")

	// Initialize achievement service (after WebSocket hub)
	achievementService := service.NewAchievementService(db.DB, logger.GetLogger(), wsHub, ethService)

	// Initialize cron job manager
	jobManager := cron.NewJobManager(
		db.DB,
		logger.GetLogger(),
		&cfg.Cron,
		achievementService,
		userService,
		wsHub,
	)

	// Start cron jobs
	if err := jobManager.Start(); err != nil {
		logger.GetSugar().Fatalf("Failed to start cron jobs: %v", err)
	}
	defer jobManager.Stop()

	// Initialize API handlers
	userHandler := api.NewUserHandler(db.DB, jwtService, userService, ethService)
	levelHandler := api.NewLevelHandler(db.DB)
	achievementHandler := api.NewAchievementHandler(db.DB, achievementService, metricsService, wsHub)
	systemHandler := api.NewSystemHandler(db.DB, metricsService, wsHub)

	// Setup Gin router
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(logger.GinLogger())
	router.Use(logger.Recovery())
	router.Use(CORS())

	if cfg.Prometheus.Enabled {
		router.Use(metricsService.HTTPMetrics())
	}

	// Health check endpoint
	router.GET("/health", systemHandler.GetHealth)

	// Metrics endpoint
	if cfg.Prometheus.Enabled {
		router.GET(cfg.Prometheus.Path, systemHandler.GetPrometheusMetrics)
	}

	// WebSocket endpoint
	router.GET("/ws", func(c *gin.Context) {
		wsHub.ServeWS(c, jwtService)
	})

	// API routes
	setupAPIRoutes(router, userHandler, levelHandler, achievementHandler, jwtService, wsHub)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.GetSugar().Infof("Server starting on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.GetSugar().Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.GetSugar().Info("Server shutting down...")

	// Give the server 30 seconds to finish processing requests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.GetSugar().Fatalf("Server forced to shutdown: %v", err)
	}

	logger.GetSugar().Info("Server stopped")
}

// setupAPIRoutes configures all API routes
func setupAPIRoutes(
	router *gin.Engine,
	userHandler *api.UserHandler,
	levelHandler *api.LevelHandler,
	achievementHandler *api.AchievementHandler,
	jwtService *middleware.JWTService,
	wsHub *websocket.Hub,
) {
	// API v1 group
	v1 := router.Group("/api/v1")

	// Public routes (no authentication required)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", userHandler.Register)
		auth.POST("/login", userHandler.Login)
		auth.POST("/refresh", userHandler.RefreshToken)
	}

	// Protected routes (authentication required)
	protected := v1.Group("")
	protected.Use(jwtService.AuthMiddleware())
	{
		// User management
		users := protected.Group("/users")
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.GET("/progress", userHandler.GetUserProgress)
			users.GET("/achievements", userHandler.GetUserAchievements)
			users.POST("/logout", userHandler.Logout)
		}

		// WebSocket connection info
		ws := protected.Group("/ws")
		{
			ws.GET("/stats", achievementHandler.GetWebSocketStats)
		}

		// Achievement system endpoints
		achievements := protected.Group("/achievements")
		{
			achievements.GET("", achievementHandler.GetAllAchievements)
			achievements.GET("/user", achievementHandler.GetUserAchievements)
			achievements.POST("/evaluate", achievementHandler.EvaluateAchievements)
		}

		// System stats (admin endpoints)
		system := protected.Group("/system")
		{
			system.GET("/stats", achievementHandler.GetSystemStats)
		}

		// Level System API endpoints
		subjects := protected.Group("/subjects")
		{
			subjects.GET("", levelHandler.GetAllSubjects)
			subjects.GET("/:subject_id", levelHandler.GetSubject)
			subjects.GET("/:subject_id/papers", levelHandler.GetSubjectPapers)
			subjects.GET("/:subject_id/roadmap", levelHandler.GetSubjectRoadmap)
		}

		papers := protected.Group("/papers")
		{
			papers.GET("/:paper_id", levelHandler.GetPaper)
			papers.GET("/:paper_id/level", levelHandler.GetPaperLevel)
		}

		levels := protected.Group("/levels")
		{
			levels.GET("/:level_id", levelHandler.GetLevel)
			levels.GET("/:level_id/questions", levelHandler.GetLevelQuestions)
			levels.POST("/:level_id/start", levelHandler.StartLevel)
			levels.POST("/:level_id/submit", levelHandler.SubmitAnswer)
			levels.POST("/:level_id/complete", levelHandler.CompleteLevel)
		}

		questions := protected.Group("/questions")
		{
			questions.GET("/:question_id", levelHandler.GetQuestion)
		}

		// Future stats endpoints (to be implemented later)
		// stats := protected.Group("/stats")
		// {
		//     stats.GET("/dashboard", statsHandler.GetDashboard)
		//     stats.GET("/progress", statsHandler.GetProgress)
		// }
	}
}

// CORS middleware for handling cross-origin requests
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
