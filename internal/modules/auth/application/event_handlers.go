package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go-templ-template/internal/shared/audit"
	"go-templ-template/internal/shared/events"
)

// UserCreatedEventHandler handles user created events from the user module
type UserCreatedEventHandler struct {
	authService AuthService
	logger      *slog.Logger
	auditLogger audit.AuditLogger
}

// NewUserCreatedEventHandler creates a new user created event handler
func NewUserCreatedEventHandler(authService AuthService, logger *slog.Logger, auditLogger audit.AuditLogger) *UserCreatedEventHandler {
	return &UserCreatedEventHandler{
		authService: authService,
		logger:      logger,
		auditLogger: auditLogger,
	}
}

// Handle processes user created events
func (h *UserCreatedEventHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Processing user created event",
		"event_id", event.EventID(),
		"user_id", event.AggregateID(),
		"event_type", event.EventType(),
	)

	// Extract event data
	eventData, ok := event.EventData().(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid event data format for user created event",
			"event_id", event.EventID(),
			"user_id", event.AggregateID(),
		)
		return fmt.Errorf("invalid event data format")
	}

	userID := event.AggregateID()
	email, _ := eventData["email"].(string)
	firstName, _ := eventData["first_name"].(string)
	lastName, _ := eventData["last_name"].(string)
	status, _ := eventData["status"].(string)

	// Log audit event for user creation
	auditEvent := &audit.AuditEvent{
		EventID:       event.EventID(),
		EventType:     "user.lifecycle.created",
		AggregateID:   userID,
		AggregateType: "user",
		UserID:        userID,
		Action:        "user_created",
		Resource:      "user",
		ResourceID:    userID,
		Details: map[string]interface{}{
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
			"status":     status,
			"source":     "user-module",
		},
		OccurredAt: event.OccurredAt(),
		Metadata:   event.Metadata(),
	}

	if err := h.auditLogger.LogEvent(ctx, auditEvent); err != nil {
		h.logger.Error("Failed to log audit event for user creation",
			"error", err,
			"event_id", event.EventID(),
			"user_id", userID,
		)
		// Don't fail the handler if audit logging fails
	}

	// Trigger notification for user creation (future extensibility)
	notificationEvent := &NotificationTrigger{
		EventID:     event.EventID(),
		TriggerType: "user_created",
		UserID:      userID,
		Data: map[string]interface{}{
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
		},
		OccurredAt: event.OccurredAt(),
	}

	if err := h.triggerNotification(ctx, notificationEvent); err != nil {
		h.logger.Error("Failed to trigger notification for user creation",
			"error", err,
			"event_id", event.EventID(),
			"user_id", userID,
		)
		// Don't fail the handler if notification trigger fails
	}

	h.logger.Info("Successfully processed user created event",
		"event_id", event.EventID(),
		"user_id", userID,
	)

	return nil
}

// EventType returns the event type this handler processes
func (h *UserCreatedEventHandler) EventType() string {
	return "user.created"
}

// HandlerName returns a unique name for this handler
func (h *UserCreatedEventHandler) HandlerName() string {
	return "auth.user_created_handler"
}

// UserUpdatedEventHandler handles user updated events from the user module
type UserUpdatedEventHandler struct {
	authService AuthService
	logger      *slog.Logger
	auditLogger audit.AuditLogger
}

// NewUserUpdatedEventHandler creates a new user updated event handler
func NewUserUpdatedEventHandler(authService AuthService, logger *slog.Logger, auditLogger audit.AuditLogger) *UserUpdatedEventHandler {
	return &UserUpdatedEventHandler{
		authService: authService,
		logger:      logger,
		auditLogger: auditLogger,
	}
}

// Handle processes user updated events
func (h *UserUpdatedEventHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Processing user updated event",
		"event_id", event.EventID(),
		"user_id", event.AggregateID(),
		"event_type", event.EventType(),
	)

	// Extract event data
	eventData, ok := event.EventData().(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid event data format for user updated event",
			"event_id", event.EventID(),
			"user_id", event.AggregateID(),
		)
		return fmt.Errorf("invalid event data format")
	}

	userID := event.AggregateID()
	changes, _ := eventData["changes"].(map[string]interface{})
	previousVersion, _ := eventData["previous_version"].(int)

	// Log audit event for user update
	auditEvent := &audit.AuditEvent{
		EventID:       event.EventID(),
		EventType:     "user.lifecycle.updated",
		AggregateID:   userID,
		AggregateType: "user",
		UserID:        userID,
		Action:        "user_updated",
		Resource:      "user",
		ResourceID:    userID,
		Details: map[string]interface{}{
			"changes":          changes,
			"previous_version": previousVersion,
			"source":           "user-module",
		},
		OccurredAt: event.OccurredAt(),
		Metadata:   event.Metadata(),
	}

	if err := h.auditLogger.LogEvent(ctx, auditEvent); err != nil {
		h.logger.Error("Failed to log audit event for user update",
			"error", err,
			"event_id", event.EventID(),
			"user_id", userID,
		)
		// Don't fail the handler if audit logging fails
	}

	// Check if status changed - might need to invalidate sessions
	if statusChange, exists := changes["status"]; exists {
		if err := h.handleUserStatusChange(ctx, userID, statusChange); err != nil {
			h.logger.Error("Failed to handle user status change",
				"error", err,
				"user_id", userID,
				"status_change", statusChange,
			)
			// Don't fail the handler if status change handling fails
		}
	}

	h.logger.Info("Successfully processed user updated event",
		"event_id", event.EventID(),
		"user_id", userID,
	)

	return nil
}

// EventType returns the event type this handler processes
func (h *UserUpdatedEventHandler) EventType() string {
	return "user.updated"
}

// HandlerName returns a unique name for this handler
func (h *UserUpdatedEventHandler) HandlerName() string {
	return "auth.user_updated_handler"
}

// handleUserStatusChange handles user status changes that might affect authentication
func (h *UserUpdatedEventHandler) handleUserStatusChange(ctx context.Context, userID string, statusChange interface{}) error {
	statusMap, ok := statusChange.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid status change format")
	}

	newStatus, _ := statusMap["new"].(string)

	// If user is suspended or deactivated, invalidate all their sessions
	if newStatus == "suspended" || newStatus == "inactive" {
		h.logger.Info("User status changed to inactive/suspended, invalidating sessions",
			"user_id", userID,
			"new_status", newStatus,
		)

		// We would need to add a method to the auth service to invalidate sessions by user ID
		// For now, we'll log the action that should be taken
		h.logger.Info("Would invalidate all sessions for user",
			"user_id", userID,
			"reason", "status_change_to_"+newStatus,
		)
	}

	return nil
}

// UserDeletedEventHandler handles user deleted events from the user module
type UserDeletedEventHandler struct {
	authService AuthService
	logger      *slog.Logger
	auditLogger audit.AuditLogger
}

// NewUserDeletedEventHandler creates a new user deleted event handler
func NewUserDeletedEventHandler(authService AuthService, logger *slog.Logger, auditLogger audit.AuditLogger) *UserDeletedEventHandler {
	return &UserDeletedEventHandler{
		authService: authService,
		logger:      logger,
		auditLogger: auditLogger,
	}
}

// Handle processes user deleted events
func (h *UserDeletedEventHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Processing user deleted event",
		"event_id", event.EventID(),
		"user_id", event.AggregateID(),
		"event_type", event.EventType(),
	)

	// Extract event data
	eventData, ok := event.EventData().(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid event data format for user deleted event",
			"event_id", event.EventID(),
			"user_id", event.AggregateID(),
		)
		return fmt.Errorf("invalid event data format")
	}

	userID := event.AggregateID()
	email, _ := eventData["email"].(string)
	deletedBy, _ := eventData["deleted_by"].(string)
	reason, _ := eventData["reason"].(string)

	// Log audit event for user deletion
	auditEvent := &audit.AuditEvent{
		EventID:       event.EventID(),
		EventType:     "user.lifecycle.deleted",
		AggregateID:   userID,
		AggregateType: "user",
		UserID:        userID,
		Action:        "user_deleted",
		Resource:      "user",
		ResourceID:    userID,
		Details: map[string]interface{}{
			"email":      email,
			"deleted_by": deletedBy,
			"reason":     reason,
			"source":     "user-module",
		},
		OccurredAt: event.OccurredAt(),
		Metadata:   event.Metadata(),
	}

	if err := h.auditLogger.LogEvent(ctx, auditEvent); err != nil {
		h.logger.Error("Failed to log audit event for user deletion",
			"error", err,
			"event_id", event.EventID(),
			"user_id", userID,
		)
		// Don't fail the handler if audit logging fails
	}

	// Clean up all sessions for the deleted user
	h.logger.Info("Cleaning up sessions for deleted user",
		"user_id", userID,
	)

	// We would need to add a method to clean up sessions by user ID
	// For now, we'll log the action that should be taken
	h.logger.Info("Would clean up all sessions for deleted user",
		"user_id", userID,
		"reason", "user_deleted",
	)

	h.logger.Info("Successfully processed user deleted event",
		"event_id", event.EventID(),
		"user_id", userID,
	)

	return nil
}

// EventType returns the event type this handler processes
func (h *UserDeletedEventHandler) EventType() string {
	return "user.deleted"
}

// HandlerName returns a unique name for this handler
func (h *UserDeletedEventHandler) HandlerName() string {
	return "auth.user_deleted_handler"
}

// triggerNotification triggers a notification for user events (future extensibility)
func (h *UserCreatedEventHandler) triggerNotification(ctx context.Context, trigger *NotificationTrigger) error {
	// This is a placeholder for future notification system integration
	// In a real implementation, this might:
	// - Send welcome emails
	// - Trigger onboarding workflows
	// - Send notifications to administrators
	// - Update external systems

	h.logger.Info("Notification trigger created",
		"trigger_type", trigger.TriggerType,
		"user_id", trigger.UserID,
		"event_id", trigger.EventID,
	)

	// For now, we'll just log that a notification would be sent
	// In the future, this could publish to a notification service
	return nil
}

// NotificationTrigger represents a trigger for the notification system
type NotificationTrigger struct {
	EventID     string                 `json:"event_id"`
	TriggerType string                 `json:"trigger_type"`
	UserID      string                 `json:"user_id"`
	Data        map[string]interface{} `json:"data"`
	OccurredAt  time.Time              `json:"occurred_at"`
}
