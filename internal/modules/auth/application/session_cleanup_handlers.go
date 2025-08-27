package application

import (
	"context"
	"fmt"
	"log/slog"

	"go-templ-template/internal/shared/audit"
	"go-templ-template/internal/shared/events"
)

// UserStatusChangedSessionCleanupHandler handles user status changes and cleans up sessions
type UserStatusChangedSessionCleanupHandler struct {
	sessionRepo SessionRepository
	authService AuthService
	logger      *slog.Logger
	auditLogger audit.AuditLogger
}

// NewUserStatusChangedSessionCleanupHandler creates a new session cleanup handler for user status changes
func NewUserStatusChangedSessionCleanupHandler(
	sessionRepo SessionRepository,
	authService AuthService,
	logger *slog.Logger,
	auditLogger audit.AuditLogger,
) *UserStatusChangedSessionCleanupHandler {
	return &UserStatusChangedSessionCleanupHandler{
		sessionRepo: sessionRepo,
		authService: authService,
		logger:      logger,
		auditLogger: auditLogger,
	}
}

// Handle processes user status changed events and cleans up sessions if needed
func (h *UserStatusChangedSessionCleanupHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Processing user status changed event for session cleanup",
		"event_id", event.EventID(),
		"user_id", event.AggregateID(),
		"event_type", event.EventType(),
	)

	// Extract event data
	eventData, ok := event.EventData().(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid event data format for user status changed event",
			"event_id", event.EventID(),
			"user_id", event.AggregateID(),
		)
		return fmt.Errorf("invalid event data format")
	}

	userID := event.AggregateID()
	newStatus, _ := eventData["new_status"].(string)
	previousStatus, _ := eventData["previous_status"].(string)
	changedBy, _ := eventData["changed_by"].(string)
	reason, _ := eventData["reason"].(string)

	// Only cleanup sessions if user is being deactivated or suspended
	if newStatus == "inactive" || newStatus == "suspended" {
		h.logger.Info("User status changed to inactive/suspended, cleaning up sessions",
			"user_id", userID,
			"new_status", newStatus,
			"previous_status", previousStatus,
		)

		// Get all sessions for the user
		sessions, err := h.sessionRepo.GetByUserID(ctx, userID)
		if err != nil {
			h.logger.Error("Failed to get user sessions for cleanup",
				"error", err,
				"user_id", userID,
			)
			return fmt.Errorf("failed to get user sessions: %w", err)
		}

		// Delete all sessions for the user
		if err := h.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
			h.logger.Error("Failed to delete user sessions",
				"error", err,
				"user_id", userID,
			)
			return fmt.Errorf("failed to delete user sessions: %w", err)
		}

		sessionCount := len(sessions)
		h.logger.Info("Successfully cleaned up user sessions",
			"user_id", userID,
			"session_count", sessionCount,
			"reason", "user_status_changed_to_"+newStatus,
		)

		// Log audit event for session cleanup
		auditEvent := &audit.AuditEvent{
			EventID:       event.EventID(),
			EventType:     "auth.session.bulk_cleanup",
			AggregateID:   userID,
			AggregateType: "user",
			UserID:        userID,
			Action:        "sessions_cleaned_up",
			Resource:      "session",
			ResourceID:    userID,
			Details: map[string]interface{}{
				"session_count":   sessionCount,
				"cleanup_reason":  "user_status_changed",
				"new_status":      newStatus,
				"previous_status": previousStatus,
				"changed_by":      changedBy,
				"change_reason":   reason,
				"source":          "auth-module",
			},
			OccurredAt: event.OccurredAt(),
			Metadata:   event.Metadata(),
		}

		if err := h.auditLogger.LogEvent(ctx, auditEvent); err != nil {
			h.logger.Error("Failed to log audit event for session cleanup",
				"error", err,
				"event_id", event.EventID(),
				"user_id", userID,
			)
			// Don't fail the handler if audit logging fails
		}
	}

	return nil
}

// EventType returns the event type this handler processes
func (h *UserStatusChangedSessionCleanupHandler) EventType() string {
	return "user.status_changed"
}

// HandlerName returns a unique name for this handler
func (h *UserStatusChangedSessionCleanupHandler) HandlerName() string {
	return "auth.user_status_changed_session_cleanup_handler"
}

// UserDeletedSessionCleanupHandler handles user deletion events and cleans up sessions
type UserDeletedSessionCleanupHandler struct {
	sessionRepo SessionRepository
	authService AuthService
	logger      *slog.Logger
	auditLogger audit.AuditLogger
}

// NewUserDeletedSessionCleanupHandler creates a new session cleanup handler for user deletion
func NewUserDeletedSessionCleanupHandler(
	sessionRepo SessionRepository,
	authService AuthService,
	logger *slog.Logger,
	auditLogger audit.AuditLogger,
) *UserDeletedSessionCleanupHandler {
	return &UserDeletedSessionCleanupHandler{
		sessionRepo: sessionRepo,
		authService: authService,
		logger:      logger,
		auditLogger: auditLogger,
	}
}

// Handle processes user deleted events and cleans up all sessions
func (h *UserDeletedSessionCleanupHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Processing user deleted event for session cleanup",
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

	h.logger.Info("User deleted, cleaning up all sessions",
		"user_id", userID,
		"email", email,
	)

	// Get all sessions for the user before deletion
	sessions, err := h.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user sessions for cleanup",
			"error", err,
			"user_id", userID,
		)
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Delete all sessions for the user
	if err := h.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		h.logger.Error("Failed to delete user sessions",
			"error", err,
			"user_id", userID,
		)
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	sessionCount := len(sessions)
	h.logger.Info("Successfully cleaned up user sessions after deletion",
		"user_id", userID,
		"session_count", sessionCount,
	)

	// Log audit event for session cleanup
	auditEvent := &audit.AuditEvent{
		EventID:       event.EventID(),
		EventType:     "auth.session.bulk_cleanup",
		AggregateID:   userID,
		AggregateType: "user",
		UserID:        userID,
		Action:        "sessions_cleaned_up",
		Resource:      "session",
		ResourceID:    userID,
		Details: map[string]interface{}{
			"session_count":  sessionCount,
			"cleanup_reason": "user_deleted",
			"email":          email,
			"deleted_by":     deletedBy,
			"delete_reason":  reason,
			"source":         "auth-module",
		},
		OccurredAt: event.OccurredAt(),
		Metadata:   event.Metadata(),
	}

	if err := h.auditLogger.LogEvent(ctx, auditEvent); err != nil {
		h.logger.Error("Failed to log audit event for session cleanup",
			"error", err,
			"event_id", event.EventID(),
			"user_id", userID,
		)
		// Don't fail the handler if audit logging fails
	}

	return nil
}

// EventType returns the event type this handler processes
func (h *UserDeletedSessionCleanupHandler) EventType() string {
	return "user.deleted"
}

// HandlerName returns a unique name for this handler
func (h *UserDeletedSessionCleanupHandler) HandlerName() string {
	return "auth.user_deleted_session_cleanup_handler"
}

// UserDeactivatedSessionCleanupHandler handles user deactivation events and cleans up sessions
type UserDeactivatedSessionCleanupHandler struct {
	sessionRepo SessionRepository
	authService AuthService
	logger      *slog.Logger
	auditLogger audit.AuditLogger
}

// NewUserDeactivatedSessionCleanupHandler creates a new session cleanup handler for user deactivation
func NewUserDeactivatedSessionCleanupHandler(
	sessionRepo SessionRepository,
	authService AuthService,
	logger *slog.Logger,
	auditLogger audit.AuditLogger,
) *UserDeactivatedSessionCleanupHandler {
	return &UserDeactivatedSessionCleanupHandler{
		sessionRepo: sessionRepo,
		authService: authService,
		logger:      logger,
		auditLogger: auditLogger,
	}
}

// Handle processes user deactivated events and cleans up sessions
func (h *UserDeactivatedSessionCleanupHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Processing user deactivated event for session cleanup",
		"event_id", event.EventID(),
		"user_id", event.AggregateID(),
		"event_type", event.EventType(),
	)

	// Extract event data
	eventData, ok := event.EventData().(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid event data format for user deactivated event",
			"event_id", event.EventID(),
			"user_id", event.AggregateID(),
		)
		return fmt.Errorf("invalid event data format")
	}

	userID := event.AggregateID()
	email, _ := eventData["email"].(string)
	deactivatedBy, _ := eventData["deactivated_by"].(string)
	reason, _ := eventData["reason"].(string)

	h.logger.Info("User deactivated, cleaning up sessions",
		"user_id", userID,
		"email", email,
	)

	// Get all sessions for the user
	sessions, err := h.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user sessions for cleanup",
			"error", err,
			"user_id", userID,
		)
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Delete all sessions for the user
	if err := h.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		h.logger.Error("Failed to delete user sessions",
			"error", err,
			"user_id", userID,
		)
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	sessionCount := len(sessions)
	h.logger.Info("Successfully cleaned up user sessions after deactivation",
		"user_id", userID,
		"session_count", sessionCount,
	)

	// Log audit event for session cleanup
	auditEvent := &audit.AuditEvent{
		EventID:       event.EventID(),
		EventType:     "auth.session.bulk_cleanup",
		AggregateID:   userID,
		AggregateType: "user",
		UserID:        userID,
		Action:        "sessions_cleaned_up",
		Resource:      "session",
		ResourceID:    userID,
		Details: map[string]interface{}{
			"session_count":       sessionCount,
			"cleanup_reason":      "user_deactivated",
			"email":               email,
			"deactivated_by":      deactivatedBy,
			"deactivation_reason": reason,
			"source":              "auth-module",
		},
		OccurredAt: event.OccurredAt(),
		Metadata:   event.Metadata(),
	}

	if err := h.auditLogger.LogEvent(ctx, auditEvent); err != nil {
		h.logger.Error("Failed to log audit event for session cleanup",
			"error", err,
			"event_id", event.EventID(),
			"user_id", userID,
		)
		// Don't fail the handler if audit logging fails
	}

	return nil
}

// EventType returns the event type this handler processes
func (h *UserDeactivatedSessionCleanupHandler) EventType() string {
	return "user.deactivated"
}

// HandlerName returns a unique name for this handler
func (h *UserDeactivatedSessionCleanupHandler) HandlerName() string {
	return "auth.user_deactivated_session_cleanup_handler"
}
