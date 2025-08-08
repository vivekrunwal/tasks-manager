package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v10"
)

// Config represents the application configuration
type Config struct {
	App        AppConfig        `envPrefix:"APP_"`
	HTTP       HTTPConfig       `envPrefix:"HTTP_"`
	DB         DBConfig         `envPrefix:"DB_"`
	Pagination PaginationConfig `envPrefix:"PAGINATION_"`
}

// AppConfig contains general application settings
type AppConfig struct {
	Env string `env:"ENV" envDefault:"dev"` // dev or prod
}

// HTTPConfig contains HTTP server configuration
type HTTPConfig struct {
	Addr         string        `env:"ADDR" envDefault:":8080"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s"`
}

// DBConfig contains database configuration
type DBConfig struct {
	DSN            string        `env:"DSN" envDefault:"postgres://postgres:postgres@localhost:5432/tasks?sslmode=disable"`
	MaxConnections int           `env:"MAX_CONNS" envDefault:"10"`
	MaxIdleConns   int           `env:"MAX_IDLE_CONNS" envDefault:"5"`
	MaxLifetime    time.Duration `env:"MAX_LIFETIME" envDefault:"5m"`
}

// PaginationConfig contains pagination settings
type PaginationConfig struct {
	DefaultSize int `env:"DEFAULT_SIZE" envDefault:"20"`
	MaxSize     int `env:"MAX_SIZE" envDefault:"100"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	
	return cfg, nil
}
