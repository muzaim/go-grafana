package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Telegram TelegramConfig `mapstructure:"telegram"`
	Worker   WorkerConfig   `mapstructure:"worker"`
	Log      LogConfig      `mapstructure:"log"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type JWTConfig struct {
	Secret                string        `mapstructure:"secret"`
	AccessTokenExpiryMin  time.Duration `mapstructure:"access_token_expiry_min"`
	RefreshTokenExpiryDay time.Duration `mapstructure:"refresh_token_expiry_day"`
}

type TelegramConfig struct {
	BotToken      string `mapstructure:"bot_token"`
	DefaultChatID string `mapstructure:"default_chat_id"`
}

type WorkerConfig struct {
	Concurrency int `mapstructure:"concurrency"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"` // debug, info, warn, error
	Filename   string `mapstructure:"filename"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
	Compress   bool   `mapstructure:"compress"`
}

func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	// Default configurations
	v.SetDefault("app.name", "go-uptime-monitor")
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.mode", "debug")

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "uptime_monitor")
	v.SetDefault("database.sslmode", "disable")

	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	v.SetDefault("jwt.secret", "supersecret-jwt-key-change-in-production")
	v.SetDefault("jwt.access_token_expiry_min", 15)
	v.SetDefault("jwt.refresh_token_expiry_day", 7)

	v.SetDefault("telegram.bot_token", "")
	v.SetDefault("telegram.default_chat_id", "")

	v.SetDefault("worker.concurrency", 10)

	v.SetDefault("log.level", "info")
	v.SetDefault("log.filename", "logs/app.log")
	v.SetDefault("log.max_size_mb", 10)
	v.SetDefault("log.max_backups", 5)
	v.SetDefault("log.max_age_days", 30)
	v.SetDefault("log.compress", true)

	// Environment variable settings
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Optionally load from file if specified
	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			// If file doesn't exist, log warning or allow fallbacks
			if !strings.Contains(err.Error(), "no such file") {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		}
	} else {
		v.AddConfigPath(".")
		v.SetConfigName(".env")
		v.SetConfigType("env")
		_ = v.ReadInConfig()
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &cfg, nil
}
