package config

import "fmt"

// GetDatabaseURL returns the complete database connection URL
func (c *Config) GetDatabaseURL() string {
	if c.Database.URL != "" {
		return c.Database.URL
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetRabbitMQURL returns the complete RabbitMQ connection URL
func (c *Config) GetRabbitMQURL() string {
	if c.RabbitMQ.URL != "" {
		return c.RabbitMQ.URL
	}

	return fmt.Sprintf("amqp://%s:%s@%s:%s/",
		c.RabbitMQ.User,
		c.RabbitMQ.Password,
		c.RabbitMQ.Host,
		c.RabbitMQ.Port,
	)
}
