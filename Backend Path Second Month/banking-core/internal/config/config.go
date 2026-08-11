package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv   string
	Server   ServerConfig
	Database DatabaseConfig
	Worker   WorkerConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Host            string
	Port            string
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	URL string
}

type WorkerConfig struct {
	PoolSize  int
	QueueSize int
}

type JWTConfig struct {
	Secret         string
	AccessTokenTTL time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv: getEnv("APP_ENV", "development"),
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnv("SERVER_PORT", "8080"),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://banking:banking@localhost:5432/banking?sslmode=disable"),
		},
		Worker: WorkerConfig{
			PoolSize:  getIntEnv("WORKER_POOL_SIZE", 5),
			QueueSize: getIntEnv("WORKER_QUEUE_SIZE", 100),
		},
		JWT: JWTConfig{
			Secret:         getEnv("JWT_SECRET", "dev-secret-change-me"),
			AccessTokenTTL: getDurationEnv("JWT_ACCESS_TTL", 15*time.Minute),
		},
	}

	if cfg.AppEnv == "production" && cfg.JWT.Secret == "dev-secret-change-me" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production")
	}

	return cfg, nil
}

func (c *Config) Address() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
