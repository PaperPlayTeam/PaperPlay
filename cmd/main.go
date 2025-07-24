package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"paperplay/config"
	"paperplay/internal/api"
	"paperplay/internal/middleware"
	"paperplay/internal/model"
	"paperplay/internal/service"
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

	// Initialize database
	db, err := model.NewDatabase(cfg.Database.DSN, cfg.Log.Level)
	if err != nil {
		logger.GetSugar().Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := db.Migrate(); err != nil {
		logger.GetSugar().Fatalf("Failed to migrate database: %v", err)
	}

	// Create additional indexes
	if err := db.CreateIndexes(); err != nil {
		logger.GetSugar().Fatalf("Failed to create database indexes: %v", err)
	}

	// Seed initial data
	if err := db.Seed(); err != nil {
		logger.GetSugar().Fatalf("Failed to seed database: %v", err)
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

	// Initialize API handlers
	userHandler := api.NewUserHandler(db.DB, jwtService, userService, ethService)

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
	router.GET("/health", func(c *gin.Context) {
		healthStatus := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
			"services": map[string]interface{}{
				"database": db.Health() == nil,
				"ethereum": ethService != nil && ethService.HealthCheck()["status"] == "healthy",
			},
		}

		c.JSON(http.StatusOK, healthStatus)
	})

	// Metrics endpoint
	if cfg.Prometheus.Enabled {
		router.GET(cfg.Prometheus.Path, metricsService.PrometheusHandler())
	}

	// API routes
	setupAPIRoutes(router, userHandler, jwtService)

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
func setupAPIRoutes(router *gin.Engine, userHandler *api.UserHandler, jwtService *middleware.JWTService) {
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

		// Subjects and levels (will be implemented later)
		// subjects := protected.Group("/subjects")
		// {
		//     subjects.GET("", subjectHandler.GetSubjects)
		//     subjects.GET("/:id", subjectHandler.GetSubject)
		//     subjects.GET("/:id/roadmap", subjectHandler.GetRoadmap)
		// }

		// levels := protected.Group("/levels")
		// {
		//     levels.GET("/:id", levelHandler.GetLevel)
		//     levels.POST("/:id/start", levelHandler.StartLevel)
		//     levels.POST("/:id/submit", levelHandler.SubmitAnswer)
		//     levels.POST("/:id/complete", levelHandler.CompleteLevel)
		// }

		// achievements := protected.Group("/achievements")
		// {
		//     achievements.GET("", achievementHandler.GetAchievements)
		//     achievements.POST("/:id/claim", achievementHandler.ClaimAchievement)
		// }

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
