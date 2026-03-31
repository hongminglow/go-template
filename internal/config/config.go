package config

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	AppName string
	AppEnv  string

	HTTPAddr string

	SeedAdminName  string
	SeedAdminEmail string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string

	RedisHost     string
	RedisPort     string
	RedisPassword string

	JWTSecret string
}

func FromEnv() Config {
	return Config{
		AppName: getEnv("APP_NAME", "go-template"),
		AppEnv:  getEnv("APP_ENV", "development"),

		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),

		SeedAdminName:  getEnv("SEED_ADMIN_NAME", "Admin"),
		SeedAdminEmail: getEnv("SEED_ADMIN_EMAIL", "admin@email.com"),

		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:       getEnv("POSTGRES_DB", "app_db"),
		PostgresSSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		JWTSecret: getEnv("JWT_SECRET", "super-secret-default-key-change-me"),
	}
}

func (c Config) PostgresDSN() string {
	dsn := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.PostgresUser, c.PostgresPassword),
		Host:   fmt.Sprintf("%s:%s", c.PostgresHost, c.PostgresPort),
		Path:   c.PostgresDB,
	}

	query := dsn.Query()
	query.Set("sslmode", c.PostgresSSLMode)
	dsn.RawQuery = query.Encode()

	return dsn.String()
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
