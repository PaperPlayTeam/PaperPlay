package middleware

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"paperplay/config"
)

// LoggerService handles structured logging
type LoggerService struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// NewLoggerService creates a new logger service
func NewLoggerService(config *config.Config) (*LoggerService, error) {
	// Create log directory if it doesn't exist
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Configure log level
	var level zapcore.Level
	switch config.Log.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Configure log rotation
	logWriter := &lumberjack.Logger{
		Filename:   config.Log.OutputPath,
		MaxSize:    config.Log.MaxSize,    // MB
		MaxAge:     config.Log.MaxAge,     // days
		MaxBackups: config.Log.MaxBackups, // files
		LocalTime:  true,
		Compress:   true,
	}

	// Configure encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.LevelKey = "level"
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderConfig.CallerKey = "caller"
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Create cores for different outputs
	var cores []zapcore.Core

	// File output
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(logWriter), level)
	cores = append(cores, fileCore)

	// Console output for development
	if gin.Mode() == gin.DebugMode {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, consoleCore)
	}

	// Combine cores
	core := zapcore.NewTee(cores...)

	// Create logger with caller information
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &LoggerService{
		logger: logger,
		sugar:  logger.Sugar(),
	}, nil
}

// GetLogger returns the zap logger instance
func (l *LoggerService) GetLogger() *zap.Logger {
	return l.logger
}

// GetSugar returns the sugared logger instance
func (l *LoggerService) GetSugar() *zap.SugaredLogger {
	return l.sugar
}

// GinLogger creates a Gin middleware for request logging
func (l *LoggerService) GinLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log using structured logging
		fields := []zap.Field{
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.String("ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("time", param.TimeStamp.Format("2006-01-02 15:04:05")),
		}

		if param.ErrorMessage != "" {
			fields = append(fields, zap.String("error", param.ErrorMessage))
		}

		// Log level based on status code
		if param.StatusCode >= 500 {
			l.logger.Error("HTTP Request", fields...)
		} else if param.StatusCode >= 400 {
			l.logger.Warn("HTTP Request", fields...)
		} else {
			l.logger.Info("HTTP Request", fields...)
		}

		return "" // Return empty string since we're using structured logging
	})
}

// Recovery creates a Gin middleware for panic recovery with logging
func (l *LoggerService) Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		l.logger.Error("Panic recovered",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.Any("panic", recovered),
			zap.Stack("stack"),
		)

		// Return 500 Internal Server Error
		c.JSON(500, gin.H{
			"error":   "Internal server error",
			"message": "An unexpected error occurred",
		})
	})
}

// LogUserAction logs user actions for audit trail
func (l *LoggerService) LogUserAction(userID, action, resource string, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.String("user_id", userID),
		zap.String("action", action),
		zap.String("resource", resource),
		zap.Time("timestamp", time.Now()),
	}

	if metadata != nil {
		for key, value := range metadata {
			fields = append(fields, zap.Any(key, value))
		}
	}

	l.logger.Info("User Action", fields...)
}

// LogSecurityEvent logs security-related events
func (l *LoggerService) LogSecurityEvent(eventType, userID, description string, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("user_id", userID),
		zap.String("description", description),
		zap.Time("timestamp", time.Now()),
	}

	if metadata != nil {
		for key, value := range metadata {
			fields = append(fields, zap.Any(key, value))
		}
	}

	l.logger.Warn("Security Event", fields...)
}

// LogError logs application errors
func (l *LoggerService) LogError(err error, context string, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.Error(err),
		zap.String("context", context),
		zap.Time("timestamp", time.Now()),
	}

	if metadata != nil {
		for key, value := range metadata {
			fields = append(fields, zap.Any(key, value))
		}
	}

	l.logger.Error("Application Error", fields...)
}

// Close gracefully closes the logger
func (l *LoggerService) Close() error {
	return l.logger.Sync()
}
