package storage

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
	SSLMode      string
	TimeZone     string
}

// DefaultDatabaseConfig returns a default database configuration
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:         getEnv("DB_HOST", "localhost"),
		Port:         getEnvAsInt("DB_PORT", 5432),
		User:         getEnv("DB_USER", "postgres"),
		Password:     getEnv("DB_PASSWORD", "postgres"),
		DatabaseName: getEnv("DB_NAME", "abac_system"),
		SSLMode:      getEnv("DB_SSL_MODE", "disable"),
		TimeZone:     getEnv("DB_TIMEZONE", "UTC"),
	}
}

// DSN returns the database connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		c.Host, c.User, c.Password, c.DatabaseName, c.Port, c.SSLMode, c.TimeZone)
}

// NewDatabaseConnection creates a new database connection
func NewDatabaseConnection(config *DatabaseConfig) (*gorm.DB, error) {
	if config == nil {
		config = DefaultDatabaseConfig()
	}

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if getEnv("DB_LOG_LEVEL", "info") == "silent" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(config.DSN()), &gorm.Config{
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
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(getEnvAsInt("DB_MAX_IDLE_CONNS", 10))
	sqlDB.SetMaxOpenConns(getEnvAsInt("DB_MAX_OPEN_CONNS", 100))
	sqlDB.SetConnMaxLifetime(time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME", 3600)) * time.Second)

	return db, nil
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
