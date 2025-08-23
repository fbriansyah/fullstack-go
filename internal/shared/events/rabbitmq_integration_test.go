package events

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Integration tests for RabbitMQ EventBus
// These tests require a running RabbitMQ instance
// Run with: go test -tags=integration

func TestRabbitMQEventBus_Integration_StartStop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping integration test - requires RabbitMQ")

	config := DefaultRabbitMQConfig()
	bus := NewRabbitMQEventBus(config)

	ctx := context.Background()

	// Test starting the bus
	err := bus.Start(ctx)
	if err != nil {
		t.Skipf("Skipping integration test - RabbitMQ not available: %v", err)
	}
	defer bus.Stop(ctx)

	// Test health check
	err = bus.Health()
	if err != nil {
		t.Errorf("Expected healthy bus, got error: %v", err)
	}

	// Test stopping the bus
	err = bus.Stop(ctx)
	if err != nil {
		t.Errorf("Expected no error stopping bus, got: %v", err)
	}
}

func TestRabbitMQEventBus_Integration_PublishSubscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping integration test - requires RabbitMQ")

	config := DefaultRabbitMQConfig()
	config.QueuePrefix = fmt.Sprintf("test-%d", time.Now().UnixNano())
	bus := NewRabbitMQEventBus(config)

	ctx := context.Background()

	// Start the bus
	err := bus.Start(ctx)
	if err != nil {
		t.Skipf("Skipping integration test - RabbitMQ not available: %v", err)
	}
	defer bus.Stop(ctx)

	// Create a test handler
	handler := NewMockEventHandler("integration-handler", "integration.test")

	// Subscribe the handler
	err = bus.Subscribe("integration.test", handler)
	if err != nil {
		t.Fatalf("Expected no error subscribing, got: %v", err)
	}

	// Give some time for the consumer to be set up
	time.Sleep(100 * time.Millisecond)

	// Create and publish a test event
	testEvent := NewTestEvent("integration-test-id", "integration-test-data")

	err = bus.Publish(ctx, testEvent)
	if err != nil {
		t.Fatalf("Expected no error publishing, got: %v", err)
	}

	// Wait for the event to be processed
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for event to be processed")
		case <-ticker.C:
			handledEvents := handler.GetHandledEvents()
			if len(handledEvents) > 0 {
				// Event was processed
				if len(handledEvents) != 1 {
					t.Errorf("Expected 1 handled event, got %d", len(handledEvents))
				}

				handledEvent := handledEvents[0]
				if handledEvent.EventID() != testEvent.EventID() {
					t.Errorf("Expected event ID %s, got %s",
						testEvent.EventID(), handledEvent.EventID())
				}

				if handledEvent.EventType() != testEvent.EventType() {
					t.Errorf("Expected event type %s, got %s",
						testEvent.EventType(), handledEvent.EventType())
				}

				return // Test passed
			}
		}
	}
}
