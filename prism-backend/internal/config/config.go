package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Port         string
	Env          string
	DatabaseURL  string
	JWTSecret    string
	JWTExpiresIn int
}

func Load() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("JWT_EXPIRES_IN", 86400)

	cfg := &Config{
		Port:         viper.GetString("PORT"),
		Env:          viper.GetString("ENV"),
		DatabaseURL:  viper.GetString("DATABASE_URL"),
		JWTSecret:    viper.GetString("JWT_SECRET"),
		JWTExpiresIn: viper.GetInt("JWT_EXPIRES_IN"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}
