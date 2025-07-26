package model

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database holds the database connection and provides methods for initialization
type Database struct {
	DB     *gorm.DB
	DBPath string
}

// NewDatabase creates a new database instance
func NewDatabase(dsn string, logLevel string) (*Database, error) {
	// Configure GORM logger
	var gormLogLevel logger.LogLevel
	switch logLevel {
	case "debug":
		gormLogLevel = logger.Info
	case "info":
		gormLogLevel = logger.Warn
	case "warn":
		gormLogLevel = logger.Error
	default:
		gormLogLevel = logger.Silent
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &Database{
		DB:     db,
		DBPath: dsn,
	}, nil
}

// DatabaseExists checks if the database file already exists and has content
func (d *Database) DatabaseExists() bool {
	// Extract the actual file path from DSN
	dbPath := d.DBPath
	if strings.Contains(dbPath, "?") {
		dbPath = strings.Split(dbPath, "?")[0]
	}

	if stat, err := os.Stat(dbPath); err == nil {
		// File exists, check if it has content (more than 0 bytes)
		// SQLite creates empty files when connecting, so we need to check size
		return stat.Size() > 0
	}
	return false
}

// ValidateSchema validates that all required tables and columns exist
func (d *Database) ValidateSchema() error {
	log.Println("Validating database schema...")

	// List of required tables and their critical columns
	requiredSchema := map[string][]string{
		"subjects":          {"id", "name", "description", "created_at", "updated_at"},
		"papers":            {"id", "subject_id", "title", "paper_author", "created_at", "updated_at"},
		"levels":            {"id", "paper_id", "name", "pass_condition", "created_at", "updated_at"},
		"questions":         {"id", "level_id", "stem", "content_json", "answer_json", "created_at"},
		"roadmap_nodes":     {"id", "subject_id", "level_id", "path", "sort_order"},
		"users":             {"id", "email", "password_hash", "display_name", "created_at", "updated_at"},
		"refresh_tokens":    {"token", "user_id", "expires_at", "created_at"},
		"user_progresses":   {"id", "user_id", "level_id", "status", "score", "created_at", "updated_at"},
		"user_attempts":     {"stat_date", "user_id", "attempts_total", "attempts_correct", "attempts_first_try_correct", "updated_at"},
		"achievements":      {"id", "name", "description", "level", "badge_type", "is_active"},
		"user_achievements": {"id", "user_id", "achievement_id", "earned_at", "progress"},
		"events":            {"id", "user_id", "event_type", "data_json", "created_at"},
		"nft_assets":        {"id", "user_id", "token_id", "metadata_uri", "status"},
	}

	for tableName, columns := range requiredSchema {
		// Check if table exists
		if !d.DB.Migrator().HasTable(tableName) {
			return fmt.Errorf("CRITICAL ERROR: Required table '%s' does not exist in database. "+
				"Database schema is incomplete. Please check your database file or allow initialization of a new database", tableName)
		}

		// Check if all required columns exist
		for _, column := range columns {
			if !d.DB.Migrator().HasColumn(tableName, column) {
				return fmt.Errorf("CRITICAL ERROR: Required column '%s' does not exist in table '%s'. "+
					"Database schema is incomplete or outdated. To protect existing data, the server will not start. "+
					"Please manually update your database schema or backup your data and allow re-initialization", column, tableName)
			}
		}
	}

	log.Println("Database schema validation completed successfully")
	return nil
}

// Migrate runs database migrations
func (d *Database) Migrate() error {
	log.Println("Running database migrations...")

	// Enable foreign key constraints for SQLite
	if err := d.DB.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Auto migrate all models
	if err := d.DB.AutoMigrate(
		&Subject{},
		&Paper{},
		&Level{},
		&Question{},
		&RoadmapNode{},
		&User{},
		&RefreshToken{},
		&UserProgress{},
		&UserAttempts{},
		&Achievement{},
		&UserAchievement{},
		&Event{},
		&NFTAsset{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// MigrateIfNew runs migrations only if this is a new database
func (d *Database) MigrateIfNew() error {
	if d.DatabaseExists() {
		log.Println("Existing database detected. Skipping migrations to protect existing data.")
		// Validate schema instead of migrating
		return d.ValidateSchema()
	}

	log.Println("New database detected. Running full migrations...")
	return d.Migrate()
}

// CreateIndexes creates additional indexes for performance
func (d *Database) CreateIndexes() error {
	log.Println("Creating additional database indexes...")

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_events_user_type_created ON events(user_id, event_type, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_user_attempts_date_user ON user_attempts(stat_date, user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_progress_user_status ON user_progresses(user_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_roadmap_nodes_subject_path ON roadmap_nodes(subject_id, path)",
		"CREATE INDEX IF NOT EXISTS idx_achievements_active ON achievements(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_nft_assets_status ON nft_assets(status)",
	}

	for _, index := range indexes {
		if err := d.DB.Exec(index).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	log.Println("Database indexes created successfully")
	return nil
}

// Seed populates the database with initial data
func (d *Database) Seed() error {
	log.Println("Seeding database with initial data...")

	// Check if data already exists
	var count int64
	d.DB.Model(&Achievement{}).Count(&count)
	if count > 0 {
		log.Println("Database already seeded, skipping...")
		return nil
	}

	// Seed achievements
	if err := d.seedAchievements(); err != nil {
		return fmt.Errorf("failed to seed achievements: %w", err)
	}

	log.Println("Database seeding completed successfully")
	return nil
}

// SeedIfNew runs seeding only if this is a new database
func (d *Database) SeedIfNew() error {
	if d.DatabaseExists() {
		log.Println("Existing database detected. Skipping seeding to protect existing data.")
		return nil
	}

	log.Println("New database detected. Running initial data seeding...")
	return d.Seed()
}

// seedAchievements creates initial achievement definitions
func (d *Database) seedAchievements() error {
	achievements := []Achievement{
		{
			Name:        "首战告捷",
			Description: "第一次作答即答对任意一道题目",
			Level:       1,
			BadgeType:   "learning",
			RuleJSON:    `{"type":"first_try","conditions":[{"field":"attempts_first_try_correct","operator":">=","value":1}]}`,
			NFTEnabled:  true,
			NFTMetadata: `{"name":"First Victory Badge","description":"Awarded for getting first question right on first try","image":"","attributes":[{"trait_type":"Achievement Type","value":"Learning"},{"trait_type":"Level","value":"Bronze"}]}`,
			IsActive:    true,
		},
		{
			Name:        "连击达人",
			Description: "连续 7 天都有学习记录",
			Level:       2,
			BadgeType:   "streak",
			RuleJSON:    `{"type":"streak","conditions":[{"field":"streak_days","operator":">=","value":7}]}`,
			NFTEnabled:  true,
			NFTMetadata: `{"name":"Streak Master Badge","description":"Awarded for 7-day learning streak","image":"","attributes":[{"trait_type":"Achievement Type","value":"Consistency"},{"trait_type":"Level","value":"Silver"}]}`,
			IsActive:    true,
		},
		{
			Name:        "智速双全",
			Description: "平均每题用时 ≤ 30秒 且正确率 ≥ 90%",
			Level:       3,
			BadgeType:   "speed",
			RuleJSON:    `{"type":"speed_accuracy","conditions":[{"field":"avg_duration_ms","operator":"<=","value":30000},{"field":"correct_rate","operator":">=","value":0.9}]}`,
			NFTEnabled:  true,
			NFTMetadata: `{"name":"Speed & Accuracy Master","description":"Awarded for high speed and accuracy","image":"","attributes":[{"trait_type":"Achievement Type","value":"Performance"},{"trait_type":"Level","value":"Gold"}]}`,
			IsActive:    true,
		},
		{
			Name:        "持久战士",
			Description: "单日学习总时长 ≥ 1 小时",
			Level:       2,
			BadgeType:   "endurance",
			RuleJSON:    `{"type":"endurance","conditions":[{"field":"total_time_ms","operator":">=","value":3600000}]}`,
			NFTEnabled:  true,
			NFTMetadata: `{"name":"Endurance Warrior Badge","description":"Awarded for studying 1+ hours in a day","image":"","attributes":[{"trait_type":"Achievement Type","value":"Dedication"},{"trait_type":"Level","value":"Silver"}]}`,
			IsActive:    true,
		},
		{
			Name:        "记忆大师",
			Description: "保留度 ≥ 85% 且当天无需复习推荐",
			Level:       3,
			BadgeType:   "memory",
			RuleJSON:    `{"type":"memory","conditions":[{"field":"retention_score","operator":">=","value":0.85},{"field":"review_due_count","operator":"=","value":0}]}`,
			NFTEnabled:  true,
			NFTMetadata: `{"name":"Memory Master Badge","description":"Awarded for excellent retention without review","image":"","attributes":[{"trait_type":"Achievement Type","value":"Mastery"},{"trait_type":"Level","value":"Gold"}]}`,
			IsActive:    true,
		},
	}

	for _, achievement := range achievements {
		if err := d.DB.Create(&achievement).Error; err != nil {
			return fmt.Errorf("failed to create achievement %s: %w", achievement.Name, err)
		}
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health checks database health
func (d *Database) Health() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// BeginTransaction starts a new database transaction
func (d *Database) BeginTransaction() *gorm.DB {
	return d.DB.Begin()
}

// GetStats returns database statistics
func (d *Database) GetStats() map[string]any {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return map[string]any{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]any{
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}
