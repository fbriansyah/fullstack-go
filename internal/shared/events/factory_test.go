package events

import (
	"os"
	"testing"
)

func TestEventBusFactory_CreateRabbitMQEventBus(t *testing.T) {
	factory := NewEventBusFactory()
	config := DefaultRabbitMQConfig()

	bus := factory.CreateRabbitMQEventBus(config)

	if bus == nil {
		t.Fatal("Expected non-nil event bus")
	}

	// Type assertion to check it's actually a RabbitMQ event bus
	rabbitBus, ok := bus.(*RabbitMQEventBus)
	if !ok {
		t.Fatal("Expected RabbitMQEventBus type")
	}

	if rabbitBus.config.URL != config.URL {
		t.Errorf("Expected URL %s, got %s", config.URL, rabbitBus.config.URL)
	}
}

func TestRabbitMQConfigFromEnv_Defaults(t *testing.T) {
	// Clear environment variables
	envVars := []string{
		"RABBITMQ_URL", "RABBITMQ_EXCHANGE", "RABBITMQ_EXCHANGE_TYPE",
		"RABBITMQ_QUEUE_PREFIX", "RABBITMQ_DURABLE", "RABBITMQ_AUTO_DELETE",
		"RABBITMQ_EXCLUSIVE", "RABBITMQ_NO_WAIT",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	config := RabbitMQConfigFromEnv()
	defaultConfig := DefaultRabbitMQConfig()

	if config.URL != defaultConfig.URL {
		t.Errorf("Expected default URL %s, got %s", defaultConfig.URL, config.URL)
	}

	if config.Exchange != defaultConfig.Exchange {
		t.Errorf("Expected default exchange %s, got %s", defaultConfig.Exchange, config.Exchange)
	}
}

func TestRabbitMQConfigFromEnv_CustomValues(t *testing.T) {
	// Set custom environment variables
	os.Setenv("RABBITMQ_URL", "amqp://custom:custom@localhost:5672/")
	os.Setenv("RABBITMQ_EXCHANGE", "custom-exchange")
	os.Setenv("RABBITMQ_EXCHANGE_TYPE", "direct")
	os.Setenv("RABBITMQ_QUEUE_PREFIX", "custom-prefix")
	os.Setenv("RABBITMQ_DURABLE", "false")
	os.Setenv("RABBITMQ_AUTO_DELETE", "true")
	os.Setenv("RABBITMQ_EXCLUSIVE", "true")
	os.Setenv("RABBITMQ_NO_WAIT", "true")

	defer func() {
		// Clean up
		envVars := []string{
			"RABBITMQ_URL", "RABBITMQ_EXCHANGE", "RABBITMQ_EXCHANGE_TYPE",
			"RABBITMQ_QUEUE_PREFIX", "RABBITMQ_DURABLE", "RABBITMQ_AUTO_DELETE",
			"RABBITMQ_EXCLUSIVE", "RABBITMQ_NO_WAIT",
		}
		for _, envVar := range envVars {
			os.Unsetenv(envVar)
		}
	}()

	config := RabbitMQConfigFromEnv()

	if config.URL != "amqp://custom:custom@localhost:5672/" {
		t.Errorf("Expected custom URL, got %s", config.URL)
	}

	if config.Exchange != "custom-exchange" {
		t.Errorf("Expected custom exchange, got %s", config.Exchange)
	}

	if config.ExchangeType != "direct" {
		t.Errorf("Expected direct exchange type, got %s", config.ExchangeType)
	}

	if config.QueuePrefix != "custom-prefix" {
		t.Errorf("Expected custom queue prefix, got %s", config.QueuePrefix)
	}

	if config.Durable {
		t.Error("Expected durable to be false")
	}

	if !config.AutoDelete {
		t.Error("Expected auto delete to be true")
	}

	if !config.Exclusive {
		t.Error("Expected exclusive to be true")
	}

	if !config.NoWait {
		t.Error("Expected no wait to be true")
	}
}

func TestValidateRabbitMQConfig_Valid(t *testing.T) {
	config := DefaultRabbitMQConfig()

	err := ValidateRabbitMQConfig(config)
	if err != nil {
		t.Errorf("Expected valid config, got error: %v", err)
	}
}

func TestValidateRabbitMQConfig_InvalidURL(t *testing.T) {
	config := DefaultRabbitMQConfig()
	config.URL = ""

	err := ValidateRabbitMQConfig(config)
	if err == nil {
		t.Error("Expected error for empty URL")
	}
}

func TestValidateRabbitMQConfig_InvalidExchange(t *testing.T) {
	config := DefaultRabbitMQConfig()
	config.Exchange = ""

	err := ValidateRabbitMQConfig(config)
	if err == nil {
		t.Error("Expected error for empty exchange")
	}
}

func TestValidateRabbitMQConfig_InvalidExchangeType(t *testing.T) {
	config := DefaultRabbitMQConfig()
	config.ExchangeType = "invalid"

	err := ValidateRabbitMQConfig(config)
	if err == nil {
		t.Error("Expected error for invalid exchange type")
	}
}

func TestValidateRabbitMQConfig_InvalidQueuePrefix(t *testing.T) {
	config := DefaultRabbitMQConfig()
	config.QueuePrefix = ""

	err := ValidateRabbitMQConfig(config)
	if err == nil {
		t.Error("Expected error for empty queue prefix")
	}
}

func TestValidateRabbitMQConfig_ValidExchangeTypes(t *testing.T) {
	validTypes := []string{"direct", "fanout", "topic", "headers"}

	for _, exchangeType := range validTypes {
		config := DefaultRabbitMQConfig()
		config.ExchangeType = exchangeType

		err := ValidateRabbitMQConfig(config)
		if err != nil {
			t.Errorf("Expected %s to be valid exchange type, got error: %v", exchangeType, err)
		}
	}
}
