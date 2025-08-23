package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
}

type ServerConfig struct {
	Port string
	Host string
	Env  string
}

type DatabaseConfig struct {
	URL      string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RabbitMQConfig struct {
	URL         string
	Host        string
	Port        string
	User        string
	Password    string
	Exchange    string
	QueuePrefix string
	Durable     bool
}

func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Env:  getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", ""),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "go_templ_template"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:         getEnv("RABBITMQ_URL", ""),
			Host:        getEnv("RABBITMQ_HOST", "localhost"),
			Port:        getEnv("RABBITMQ_PORT", "5672"),
			User:        getEnv("RABBITMQ_USER", "guest"),
			Password:    getEnv("RABBITMQ_PASSWORD", "guest"),
			Exchange:    getEnv("RABBITMQ_EXCHANGE", "go_templ_template"),
			QueuePrefix: getEnv("RABBITMQ_QUEUE_PREFIX", "go_templ_template"),
			Durable:     getEnvBool("RABBITMQ_DURABLE", true),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
