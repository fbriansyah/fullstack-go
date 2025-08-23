package events

import (
	"context"
	"fmt"
)

// BaseEventHandler provides a common implementation for event handlers
type BaseEventHandler struct {
	eventType   string
	handlerName string
}

// NewBaseEventHandler creates a new base event handler
func NewBaseEventHandler(eventType, handlerName string) *BaseEventHandler {
	return &BaseEventHandler{
		eventType:   eventType,
		handlerName: handlerName,
	}
}

// EventType returns the type of event this handler processes
func (h *BaseEventHandler) EventType() string {
	return h.eventType
}

// HandlerName returns a unique name for this handler
func (h *BaseEventHandler) HandlerName() string {
	return h.handlerName
}

// Handle provides a default implementation that should be overridden
func (h *BaseEventHandler) Handle(ctx context.Context, event DomainEvent) error {
	return fmt.Errorf("handler %s does not implement Handle method for event type %s",
		h.handlerName, h.eventType)
}

// AsyncEventHandler wraps an EventHandler to process events asynchronously
type AsyncEventHandler struct {
	handler EventHandler
	buffer  chan DomainEvent
}

// NewAsyncEventHandler creates a new async event handler with a buffer
func NewAsyncEventHandler(handler EventHandler, bufferSize int) *AsyncEventHandler {
	return &AsyncEventHandler{
		handler: handler,
		buffer:  make(chan DomainEvent, bufferSize),
	}
}

// Handle queues the event for asynchronous processing
func (h *AsyncEventHandler) Handle(ctx context.Context, event DomainEvent) error {
	select {
	case h.buffer <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("event buffer full for handler %s", h.handler.HandlerName())
	}
}

// EventType returns the type of event this handler processes
func (h *AsyncEventHandler) EventType() string {
	return h.handler.EventType()
}

// HandlerName returns a unique name for this handler
func (h *AsyncEventHandler) HandlerName() string {
	return fmt.Sprintf("async_%s", h.handler.HandlerName())
}

// Start begins processing events from the buffer
func (h *AsyncEventHandler) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case event := <-h.buffer:
				if err := h.handler.Handle(ctx, event); err != nil {
					// Log error - in a real implementation, you'd use a proper logger
					fmt.Printf("Error handling event %s: %v\n", event.EventType(), err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop gracefully shuts down the async handler
func (h *AsyncEventHandler) Stop() {
	close(h.buffer)
}
