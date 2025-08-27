package audit

import (
	"context"
	"fmt"
	"log/slog"

	"go-templ-template/internal/shared/events"
)

// EventAuditHandler handles all domain events and creates audit trail entries
type EventAuditHandler struct {
	auditLogger AuditLogger
	logger      *slog.Logger
}

// NewEventAuditHandler creates a new event audit handler
func NewEventAuditHandler(auditLogger AuditLogger, logger *slog.Logger) *EventAuditHandler {
	return &EventAuditHandler{
		auditLogger: auditLogger,
		logger:      logger,
	}
}

// Handle processes any domain event and creates an audit trail entry
func (h *EventAuditHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.logger.Debug("Processing event for audit trail",
		"event_id", event.EventID(),
		"event_type", event.EventType(),
		"aggregate_id", event.AggregateID(),
		"aggregate_type", event.AggregateType(),
	)

	// Create audit event based on domain event
	auditEvent := h.createAuditEventFromDomainEvent(event)

	// Log the audit event
	if err := h.auditLogger.LogEvent(ctx, auditEvent); err != nil {
		h.logger.Error("Failed to log audit event",
			"error", err,
			"event_id", event.EventID(),
			"event_type", event.EventType(),
		)
		return fmt.Errorf("failed to log audit event: %w", err)
	}

	h.logger.Debug("Successfully logged audit event",
		"event_id", event.EventID(),
		"event_type", event.EventType(),
		"audit_action", auditEvent.Action,
	)

	return nil
}

// EventType returns the event type this handler processes (all events)
func (h *EventAuditHandler) EventType() string {
	return "*" // Handle all events
}

// HandlerName returns a unique name for this handler
func (h *EventAuditHandler) HandlerName() string {
	return "audit.event_audit_handler"
}

// createAuditEventFromDomainEvent creates an audit event from a domain event
func (h *EventAuditHandler) createAuditEventFromDomainEvent(event events.DomainEvent) *AuditEvent {
	// Extract user ID from metadata if available
	userID := event.Metadata().UserID
	if userID == "" {
		// For some events, the user ID might be the aggregate ID
		if event.AggregateType() == "User" {
			userID = event.AggregateID()
		}
	}

	// Map event types to audit actions and resources
	action, resource := h.mapEventTypeToAuditInfo(event.EventType())

	auditEvent := &AuditEvent{
		EventID:       event.EventID(),
		EventType:     fmt.Sprintf("domain.%s", event.EventType()),
		AggregateID:   event.AggregateID(),
		AggregateType: event.AggregateType(),
		UserID:        userID,
		Action:        action,
		Resource:      resource,
		ResourceID:    event.AggregateID(),
		Details:       h.extractEventDetails(event),
		OccurredAt:    event.OccurredAt(),
		Metadata:      event.Metadata(),
	}

	return auditEvent
}

// mapEventTypeToAuditInfo maps event types to audit actions and resources
func (h *EventAuditHandler) mapEventTypeToAuditInfo(eventType string) (action, resource string) {
	switch eventType {
	// User events
	case "user.created":
		return "user_created", "user"
	case "user.updated":
		return "user_updated", "user"
	case "user.deleted":
		return "user_deleted", "user"
	case "user.status_changed":
		return "user_status_changed", "user"
	case "user.email_changed":
		return "user_email_changed", "user"
	case "user.activated":
		return "user_activated", "user"
	case "user.deactivated":
		return "user_deactivated", "user"
	case "user.activation_requested":
		return "user_activation_requested", "user"
	case "user.activation_token_expired":
		return "user_activation_token_expired", "user"

	// Auth events
	case "auth.user_logged_in":
		return "user_logged_in", "session"
	case "auth.user_logged_out":
		return "user_logged_out", "session"
	case "auth.user_registered":
		return "user_registered", "user"
	case "auth.session_expired":
		return "session_expired", "session"
	case "auth.password_changed":
		return "password_changed", "user"

	// System events
	case "system.started":
		return "system_started", "system"
	case "system.shutting_down":
		return "system_shutting_down", "system"
	case "system.health_check_failed":
		return "health_check_failed", "system"

	// Default mapping
	default:
		return "event_occurred", "unknown"
	}
}

// extractEventDetails extracts relevant details from the event data
func (h *EventAuditHandler) extractEventDetails(event events.DomainEvent) map[string]interface{} {
	details := make(map[string]interface{})

	// Add basic event information
	details["event_version"] = event.Version()
	details["source"] = event.Metadata().Source

	// Add correlation information if available
	if event.Metadata().CorrelationID != "" {
		details["correlation_id"] = event.Metadata().CorrelationID
	}
	if event.Metadata().CausationID != "" {
		details["causation_id"] = event.Metadata().CausationID
	}
	if event.Metadata().TraceID != "" {
		details["trace_id"] = event.Metadata().TraceID
	}

	// Add event-specific data
	if eventData := event.EventData(); eventData != nil {
		if dataMap, ok := eventData.(map[string]interface{}); ok {
			for key, value := range dataMap {
				// Avoid overwriting basic details
				if key != "event_version" && key != "source" {
					details[key] = value
				}
			}
		} else {
			details["event_data"] = eventData
		}
	}

	// Add custom metadata if available
	if event.Metadata().Custom != nil {
		for key, value := range event.Metadata().Custom {
			details[fmt.Sprintf("custom_%s", key)] = value
		}
	}

	return details
}

// UniversalEventAuditHandler is a specialized handler that can be registered for all event types
type UniversalEventAuditHandler struct {
	*EventAuditHandler
	eventTypes []string
}

// NewUniversalEventAuditHandler creates a handler that processes multiple event types
func NewUniversalEventAuditHandler(auditLogger AuditLogger, logger *slog.Logger, eventTypes []string) *UniversalEventAuditHandler {
	return &UniversalEventAuditHandler{
		EventAuditHandler: NewEventAuditHandler(auditLogger, logger),
		eventTypes:        eventTypes,
	}
}

// GetEventTypes returns all event types this handler should process
func (h *UniversalEventAuditHandler) GetEventTypes() []string {
	return h.eventTypes
}

// AuditTrailService provides high-level audit trail functionality
type AuditTrailService struct {
	auditLogger AuditLogger
	eventBus    events.EventBus
	logger      *slog.Logger
}

// NewAuditTrailService creates a new audit trail service
func NewAuditTrailService(auditLogger AuditLogger, eventBus events.EventBus, logger *slog.Logger) *AuditTrailService {
	return &AuditTrailService{
		auditLogger: auditLogger,
		eventBus:    eventBus,
		logger:      logger,
	}
}

// RegisterAuditHandlers registers audit handlers for all relevant event types
func (s *AuditTrailService) RegisterAuditHandlers(ctx context.Context) error {
	// Define event types to audit
	eventTypes := []string{
		// User events
		"user.created",
		"user.updated",
		"user.deleted",
		"user.status_changed",
		"user.email_changed",
		"user.activated",
		"user.deactivated",
		"user.activation_requested",
		"user.activation_token_expired",

		// Auth events
		"auth.user_logged_in",
		"auth.user_logged_out",
		"auth.user_registered",
		"auth.session_expired",
		"auth.password_changed",

		// System events
		"system.started",
		"system.shutting_down",
		"system.health_check_failed",
	}

	// Create and register audit handler for each event type
	for _, eventType := range eventTypes {
		handler := NewEventAuditHandler(s.auditLogger, s.logger)
		if err := s.eventBus.Subscribe(eventType, handler); err != nil {
			s.logger.Error("Failed to register audit handler",
				"event_type", eventType,
				"error", err,
			)
			return fmt.Errorf("failed to register audit handler for %s: %w", eventType, err)
		}

		s.logger.Info("Registered audit handler",
			"event_type", eventType,
			"handler", handler.HandlerName(),
		)
	}

	return nil
}

// GetAuditTrail retrieves audit events with filtering
func (s *AuditTrailService) GetAuditTrail(ctx context.Context, filter *AuditFilter) ([]*AuditEvent, error) {
	return s.auditLogger.GetEvents(ctx, filter)
}

// GetUserAuditTrail retrieves audit events for a specific user
func (s *AuditTrailService) GetUserAuditTrail(ctx context.Context, userID string, limit int) ([]*AuditEvent, error) {
	filter := &AuditFilter{
		UserID: userID,
		Limit:  limit,
	}
	return s.auditLogger.GetEvents(ctx, filter)
}

// GetResourceAuditTrail retrieves audit events for a specific resource
func (s *AuditTrailService) GetResourceAuditTrail(ctx context.Context, resource, resourceID string, limit int) ([]*AuditEvent, error) {
	filter := &AuditFilter{
		Resource:   resource,
		ResourceID: resourceID,
		Limit:      limit,
	}
	return s.auditLogger.GetEvents(ctx, filter)
}
