package events

import (
	"fmt"
	"os"
)

// EventBusFactory creates event bus instances based on configuration
type EventBusFactory struct{}

// NewEventBusFactory creates a new event bus factory
func NewEventBusFactory() *EventBusFactory {
	return &EventBusFactory{}
}

// CreateRabbitMQEventBus creates a RabbitMQ event bus with the given configuration
func (f *EventBusFactory) CreateRabbitMQEventBus(config RabbitMQConfig) EventBus {
	return NewRabbitMQEventBus(config)
}

// CreateRabbitMQEventBusFromEnv creates a RabbitMQ event bus using environment variables
func (f *EventBusFactory) CreateRabbitMQEventBusFromEnv() EventBus {
	config := RabbitMQConfigFromEnv()
	return NewRabbitMQEventBus(config)
}

// RabbitMQConfigFromEnv creates RabbitMQ configuration from environment variables
func RabbitMQConfigFromEnv() RabbitMQConfig {
	config := DefaultRabbitMQConfig()

	if url := os.Getenv("RABBITMQ_URL"); url != "" {
		config.URL = url
	}

	if exchange := os.Getenv("RABBITMQ_EXCHANGE"); exchange != "" {
		config.Exchange = exchange
	}

	if exchangeType := os.Getenv("RABBITMQ_EXCHANGE_TYPE"); exchangeType != "" {
		config.ExchangeType = exchangeType
	}

	if queuePrefix := os.Getenv("RABBITMQ_QUEUE_PREFIX"); queuePrefix != "" {
		config.QueuePrefix = queuePrefix
	}

	// Parse boolean environment variables
	if durable := os.Getenv("RABBITMQ_DURABLE"); durable != "" {
		config.Durable = durable == "true"
	}

	if autoDelete := os.Getenv("RABBITMQ_AUTO_DELETE"); autoDelete != "" {
		config.AutoDelete = autoDelete == "true"
	}

	if exclusive := os.Getenv("RABBITMQ_EXCLUSIVE"); exclusive != "" {
		config.Exclusive = exclusive == "true"
	}

	if noWait := os.Getenv("RABBITMQ_NO_WAIT"); noWait != "" {
		config.NoWait = noWait == "true"
	}

	return config
}

// ValidateRabbitMQConfig validates a RabbitMQ configuration
func ValidateRabbitMQConfig(config RabbitMQConfig) error {
	if config.URL == "" {
		return fmt.Errorf("RabbitMQ URL is required")
	}

	if config.Exchange == "" {
		return fmt.Errorf("RabbitMQ exchange name is required")
	}

	if config.ExchangeType == "" {
		return fmt.Errorf("RabbitMQ exchange type is required")
	}

	if config.QueuePrefix == "" {
		return fmt.Errorf("RabbitMQ queue prefix is required")
	}

	// Validate exchange type
	validExchangeTypes := map[string]bool{
		"direct":  true,
		"fanout":  true,
		"topic":   true,
		"headers": true,
	}

	if !validExchangeTypes[config.ExchangeType] {
		return fmt.Errorf("invalid exchange type: %s. Must be one of: direct, fanout, topic, headers", config.ExchangeType)
	}

	return nil
}
