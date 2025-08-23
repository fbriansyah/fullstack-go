package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQEventBus implements EventBus using RabbitMQ
type RabbitMQEventBus struct {
	connection   *amqp.Connection
	channel      *amqp.Channel
	config       RabbitMQConfig
	handlers     map[string][]EventHandler
	handlersMux  sync.RWMutex
	consumers    map[string]*amqp.Channel
	consumersMux sync.RWMutex
	done         chan bool
	wg           sync.WaitGroup
	stopped      bool
	stopMux      sync.Mutex
}

// RabbitMQConfig holds configuration for RabbitMQ connection
type RabbitMQConfig struct {
	URL          string
	Exchange     string
	ExchangeType string
	QueuePrefix  string
	Durable      bool
	AutoDelete   bool
	Exclusive    bool
	NoWait       bool
}

// NewRabbitMQEventBus creates a new RabbitMQ event bus
func NewRabbitMQEventBus(config RabbitMQConfig) *RabbitMQEventBus {
	return &RabbitMQEventBus{
		config:    config,
		handlers:  make(map[string][]EventHandler),
		consumers: make(map[string]*amqp.Channel),
		done:      make(chan bool),
	}
}

// Start initializes the RabbitMQ connection and sets up the exchange
func (r *RabbitMQEventBus) Start(ctx context.Context) error {
	var err error

	// Establish connection
	r.connection, err = amqp.Dial(r.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel for publishing
	r.channel, err = r.connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = r.channel.ExchangeDeclare(
		r.config.Exchange,     // name
		r.config.ExchangeType, // type
		r.config.Durable,      // durable
		r.config.AutoDelete,   // auto-deleted
		false,                 // internal
		r.config.NoWait,       // no-wait
		nil,                   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Start consumers for existing handlers
	r.handlersMux.RLock()
	for eventType := range r.handlers {
		if err := r.startConsumer(eventType); err != nil {
			r.handlersMux.RUnlock()
			return fmt.Errorf("failed to start consumer for %s: %w", eventType, err)
		}
	}
	r.handlersMux.RUnlock()

	log.Printf("RabbitMQ EventBus started with exchange: %s", r.config.Exchange)
	return nil
}

// Stop gracefully shuts down the event bus
func (r *RabbitMQEventBus) Stop(ctx context.Context) error {
	r.stopMux.Lock()
	defer r.stopMux.Unlock()

	if r.stopped {
		return nil // Already stopped
	}

	close(r.done)
	r.wg.Wait()
	r.stopped = true

	// Close consumer channels
	r.consumersMux.Lock()
	for _, ch := range r.consumers {
		if err := ch.Close(); err != nil {
			log.Printf("Error closing consumer channel: %v", err)
		}
	}
	r.consumers = make(map[string]*amqp.Channel)
	r.consumersMux.Unlock()

	// Close main channel
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			log.Printf("Error closing main channel: %v", err)
		}
	}

	// Close connection
	if r.connection != nil {
		if err := r.connection.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}

	log.Println("RabbitMQ EventBus stopped")
	return nil
}

// Publish sends an event to the exchange
func (r *RabbitMQEventBus) Publish(ctx context.Context, event DomainEvent) error {
	if r.channel == nil {
		return fmt.Errorf("event bus not started")
	}

	// Create serializable event envelope
	envelope, err := NewSerializableEventEnvelope(event)
	if err != nil {
		return fmt.Errorf("failed to create serializable envelope: %w", err)
	}

	// Serialize event
	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Publish message
	err = r.channel.PublishWithContext(
		ctx,
		r.config.Exchange, // exchange
		event.EventType(), // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Make message persistent
			Timestamp:    time.Now(),
			MessageId:    event.EventID(),
			Headers: amqp.Table{
				"event_type":     event.EventType(),
				"aggregate_id":   event.AggregateID(),
				"aggregate_type": event.AggregateType(),
				"correlation_id": event.Metadata().CorrelationID,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Published event: %s for aggregate: %s", event.EventType(), event.AggregateID())
	return nil
}

// Subscribe registers an event handler for a specific event type
func (r *RabbitMQEventBus) Subscribe(eventType string, handler EventHandler) error {
	r.handlersMux.Lock()
	defer r.handlersMux.Unlock()

	// Add handler to the list
	r.handlers[eventType] = append(r.handlers[eventType], handler)

	// If the bus is already started, start a consumer for this event type
	if r.connection != nil && !r.connection.IsClosed() {
		if err := r.startConsumer(eventType); err != nil {
			return fmt.Errorf("failed to start consumer for %s: %w", eventType, err)
		}
	}

	log.Printf("Subscribed handler %s to event type: %s", handler.HandlerName(), eventType)
	return nil
}

// Unsubscribe removes an event handler for a specific event type
func (r *RabbitMQEventBus) Unsubscribe(eventType string, handler EventHandler) error {
	r.handlersMux.Lock()
	defer r.handlersMux.Unlock()

	handlers, exists := r.handlers[eventType]
	if !exists {
		return nil // Nothing to unsubscribe
	}

	// Remove the handler from the list
	for i, h := range handlers {
		if h.HandlerName() == handler.HandlerName() {
			r.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	// If no more handlers for this event type, we could stop the consumer
	// For simplicity, we'll leave the consumer running

	log.Printf("Unsubscribed handler %s from event type: %s", handler.HandlerName(), eventType)
	return nil
}

// Health checks the health of the RabbitMQ connection
func (r *RabbitMQEventBus) Health() error {
	if r.connection == nil || r.connection.IsClosed() {
		return fmt.Errorf("RabbitMQ connection is closed")
	}
	if r.channel == nil || r.channel.IsClosed() {
		return fmt.Errorf("RabbitMQ channel is closed")
	}
	return nil
}

// startConsumer creates a consumer for a specific event type
func (r *RabbitMQEventBus) startConsumer(eventType string) error {
	// Create a new channel for this consumer
	ch, err := r.connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open consumer channel: %w", err)
	}

	// Set QoS to process one message at a time
	err = ch.Qos(1, 0, false)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Declare queue for this event type
	queueName := fmt.Sprintf("%s.%s", r.config.QueuePrefix, eventType)
	queue, err := ch.QueueDeclare(
		queueName,           // name
		r.config.Durable,    // durable
		r.config.AutoDelete, // delete when unused
		r.config.Exclusive,  // exclusive
		r.config.NoWait,     // no-wait
		nil,                 // arguments
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = ch.QueueBind(
		queue.Name,        // queue name
		eventType,         // routing key
		r.config.Exchange, // exchange
		r.config.NoWait,   // no-wait
		nil,               // arguments
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Start consuming messages
	msgs, err := ch.Consume(
		queue.Name,         // queue
		"",                 // consumer
		false,              // auto-ack
		r.config.Exclusive, // exclusive
		false,              // no-local
		r.config.NoWait,    // no-wait
		nil,                // args
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Store consumer channel
	r.consumersMux.Lock()
	r.consumers[eventType] = ch
	r.consumersMux.Unlock()

	// Start goroutine to process messages
	r.wg.Add(1)
	go r.processMessages(eventType, msgs, ch)

	log.Printf("Started consumer for event type: %s on queue: %s", eventType, queueName)
	return nil
}

// processMessages processes incoming messages for a specific event type
func (r *RabbitMQEventBus) processMessages(eventType string, msgs <-chan amqp.Delivery, ch *amqp.Channel) {
	defer r.wg.Done()
	defer ch.Close()

	for {
		select {
		case <-r.done:
			log.Printf("Stopping consumer for event type: %s", eventType)
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Printf("Consumer channel closed for event type: %s", eventType)
				return
			}

			// Process the message
			if err := r.handleMessage(eventType, msg); err != nil {
				log.Printf("Error handling message for event type %s: %v", eventType, err)

				// Check if we should retry
				envelope := &SerializableEventEnvelope{}
				if json.Unmarshal(msg.Body, envelope) == nil && envelope.ShouldRetry() {
					// Reject and requeue for retry
					msg.Nack(false, true)
				} else {
					// Max retries exceeded or unmarshal error, reject without requeue
					msg.Nack(false, false)
				}
			} else {
				// Acknowledge successful processing
				msg.Ack(false)
			}
		}
	}
}

// handleMessage processes a single message
func (r *RabbitMQEventBus) handleMessage(eventType string, msg amqp.Delivery) error {
	// Deserialize event envelope
	var envelope SerializableEventEnvelope
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		return fmt.Errorf("failed to deserialize event envelope: %w", err)
	}

	// Get handlers for this event type
	r.handlersMux.RLock()
	handlers, exists := r.handlers[eventType]
	if !exists {
		r.handlersMux.RUnlock()
		log.Printf("No handlers registered for event type: %s", eventType)
		return nil
	}

	// Create a copy of handlers to avoid holding the lock during processing
	handlersCopy := make([]EventHandler, len(handlers))
	copy(handlersCopy, handlers)
	r.handlersMux.RUnlock()

	// Process with each handler
	ctx := context.Background()
	for _, handler := range handlersCopy {
		if err := handler.Handle(ctx, envelope.Event); err != nil {
			log.Printf("Handler %s failed to process event %s: %v",
				handler.HandlerName(), eventType, err)
			return err
		}
	}

	return nil
}

// DefaultRabbitMQConfig returns a default configuration for RabbitMQ
func DefaultRabbitMQConfig() RabbitMQConfig {
	return RabbitMQConfig{
		URL:          "amqp://guest:guest@localhost:5672/",
		Exchange:     "events",
		ExchangeType: "topic",
		QueuePrefix:  "go-templ-template",
		Durable:      true,
		AutoDelete:   false,
		Exclusive:    false,
		NoWait:       false,
	}
}
