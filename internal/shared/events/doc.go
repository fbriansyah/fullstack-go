// Package events provides the core event infrastructure for the application.
//
// This package defines the interfaces and base implementations for domain events,
// event handlers, and the event bus. It supports both synchronous and asynchronous
// event processing with proper error handling and metadata tracking.
//
// Key Components:
//
// - EventBus: Interface for publishing and subscribing to events
// - DomainEvent: Interface representing domain events with metadata
// - EventHandler: Interface for processing domain events
// - BaseEvent: Common implementation for domain events
// - BaseEventHandler: Common implementation for event handlers
//
// Usage:
//
//	// Create a domain event
//	event := NewBaseEvent("user.created", userID, "User", userData)
//
//	// Publish the event
//	err := eventBus.Publish(ctx, event)
//
//	// Create and register a handler
//	handler := NewBaseEventHandler("user.created", "user-notification-handler")
//	err := eventBus.Subscribe("user.created", handler)
//
// The package supports event correlation, causation tracking, and distributed
// tracing through the EventMetadata structure.
package events
