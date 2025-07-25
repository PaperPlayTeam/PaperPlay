package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for our application
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Ethereum   EthereumConfig   `mapstructure:"ethereum"`
	WebSocket  WebSocketConfig  `mapstructure:"websocket"`
	Log        LogConfig        `mapstructure:"log"`
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	Cron       CronConfig       `mapstructure:"cron"`
}

type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Mode         string `mapstructure:"mode"` // gin mode: debug, release, test
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	DSN             string `mapstructure:"dsn"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type JWTConfig struct {
	SecretKey            string `mapstructure:"secret_key"`
	AccessTokenDuration  int    `mapstructure:"access_token_duration"`  // minutes
	RefreshTokenDuration int    `mapstructure:"refresh_token_duration"` // days
}

type EthereumConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	NetworkURL      string `mapstructure:"network_url"`
	ChainID         int64  `mapstructure:"chain_id"`
	ContractAddress string `mapstructure:"contract_address"`
	PrivateKey      string `mapstructure:"private_key"`
	GasLimit        uint64 `mapstructure:"gas_limit"`
	GasPrice        int64  `mapstructure:"gas_price"`
}

type WebSocketConfig struct {
	ReadBufferSize  int `mapstructure:"read_buffer_size"`
	WriteBufferSize int `mapstructure:"write_buffer_size"`
	PingPeriod      int `mapstructure:"ping_period"` // seconds
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	OutputPath string `mapstructure:"output_path"`
	MaxSize    int    `mapstructure:"max_size"`    // MB
	MaxAge     int    `mapstructure:"max_age"`     // days
	MaxBackups int    `mapstructure:"max_backups"` // files
}

type PrometheusConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

type CronConfig struct {
	Enabled              bool   `mapstructure:"enabled"`
	StatsUpdateSpec      string `mapstructure:"stats_update_spec"`
	ReportGenerationSpec string `mapstructure:"report_generation_spec"`
	AchievementCheckSpec string `mapstructure:"achievement_check_spec"`
}

var globalConfig *Config

// Load reads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set configuration file path
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	// Read from environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Try to read the config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Config file not found, using defaults and environment variables")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	globalConfig = &config
	return &config, nil
}

// Get returns the global configuration instance
func Get() *Config {
	if globalConfig == nil {
		panic("Configuration not loaded. Call Load() first.")
	}
	return globalConfig
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.read_timeout", 30)
	v.SetDefault("server.write_timeout", 30)
	v.SetDefault("server.idle_timeout", 120)

	// Database defaults
	v.SetDefault("database.driver", "sqlite3")
	v.SetDefault("database.dsn", "./data/paperplay.db")
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.conn_max_lifetime", 3600)

	// JWT defaults
	v.SetDefault("jwt.secret_key", "your-secret-key-change-in-production")
	v.SetDefault("jwt.access_token_duration", 15)
	v.SetDefault("jwt.refresh_token_duration", 7)

	// Ethereum defaults
	v.SetDefault("ethereum.enabled", false)
	v.SetDefault("ethereum.network_url", "https://sepolia.infura.io/v3/your-project-id")
	v.SetDefault("ethereum.chain_id", 11155111) // Sepolia testnet
	v.SetDefault("ethereum.gas_limit", 100000)
	v.SetDefault("ethereum.gas_price", 20000000000) // 20 Gwei

	// WebSocket defaults
	v.SetDefault("websocket.read_buffer_size", 1024)
	v.SetDefault("websocket.write_buffer_size", 1024)
	v.SetDefault("websocket.ping_period", 54)

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.output_path", "./logs/app.log")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_age", 7)
	v.SetDefault("log.max_backups", 3)

	// Prometheus defaults
	v.SetDefault("prometheus.enabled", true)
	v.SetDefault("prometheus.path", "/metrics")

	// Cron defaults
	v.SetDefault("cron.enabled", true)
	v.SetDefault("cron.stats_update_spec", "0 2 * * *")        // Daily at 2 AM
	v.SetDefault("cron.report_generation_spec", "0 3 * * 0")   // Weekly on Sunday at 3 AM
	v.SetDefault("cron.achievement_check_spec", "*/5 * * * *") // Every 5 minutes
}

// validateConfig performs basic validation on the configuration
func validateConfig(config *Config) error {
	if config.JWT.AccessTokenDuration <= 0 {
		return fmt.Errorf("JWT access token duration must be positive")
	}

	if config.JWT.RefreshTokenDuration <= 0 {
		return fmt.Errorf("JWT refresh token duration must be positive")
	}

	if config.Database.DSN == "" {
		return fmt.Errorf("database DSN cannot be empty")
	}

	if config.Ethereum.Enabled {
		if config.Ethereum.NetworkURL == "" {
			return fmt.Errorf("ethereum network URL must be set when ethereum is enabled")
		}
		if config.Ethereum.ChainID <= 0 {
			return fmt.Errorf("ethereum chain ID must be positive")
		}
	}

	return nil
}
