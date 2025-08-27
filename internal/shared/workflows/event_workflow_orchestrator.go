package workflows

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	authApp "go-templ-template/internal/modules/auth/application"
	userApp "go-templ-template/internal/modules/user/application"
	"go-templ-template/internal/shared/audit"
	"go-templ-template/internal/shared/events"
)

// EventWorkflowOrchestrator orchestrates event-driven workflows across modules
type EventWorkflowOrchestrator struct {
	eventBus          events.EventBus
	auditTrailService *audit.AuditTrailService
	logger            *slog.Logger

	// Services
	userService       userApp.UserService
	authService       authApp.AuthService
	activationService userApp.ActivationService

	// Repositories
	sessionRepo authApp.SessionRepository

	// Configuration
	config *WorkflowConfig
}

// WorkflowConfig holds configuration for workflow orchestration
type WorkflowConfig struct {
	// Session cleanup settings
	EnableSessionCleanup  bool
	SessionCleanupTimeout time.Duration

	// Audit trail settings
	EnableAuditTrail bool
	AuditAllEvents   bool
	AuditEventTypes  []string

	// Activation workflow settings
	ActivationTokenDuration time.Duration
	EnableAutoActivation    bool

	// Retry settings
	MaxRetries int
	RetryDelay time.Duration
}

// DefaultWorkflowConfig returns default workflow configuration
func DefaultWorkflowConfig() *WorkflowConfig {
	return &WorkflowConfig{
		EnableSessionCleanup:    true,
		SessionCleanupTimeout:   30 * time.Second,
		EnableAuditTrail:        true,
		AuditAllEvents:          true,
		ActivationTokenDuration: 24 * time.Hour,
		EnableAutoActivation:    false,
		MaxRetries:              3,
		RetryDelay:              1 * time.Second,
	}
}

// NewEventWorkflowOrchestrator creates a new event workflow orchestrator
func NewEventWorkflowOrchestrator(
	eventBus events.EventBus,
	auditTrailService *audit.AuditTrailService,
	userService userApp.UserService,
	authService authApp.AuthService,
	activationService userApp.ActivationService,
	sessionRepo authApp.SessionRepository,
	logger *slog.Logger,
	config *WorkflowConfig,
) *EventWorkflowOrchestrator {
	if config == nil {
		config = DefaultWorkflowConfig()
	}

	return &EventWorkflowOrchestrator{
		eventBus:          eventBus,
		auditTrailService: auditTrailService,
		userService:       userService,
		authService:       authService,
		activationService: activationService,
		sessionRepo:       sessionRepo,
		logger:            logger,
		config:            config,
	}
}

// Initialize sets up all event handlers and workflows
func (o *EventWorkflowOrchestrator) Initialize(ctx context.Context) error {
	o.logger.Info("Initializing event workflow orchestrator")

	// Register audit trail handlers if enabled
	if o.config.EnableAuditTrail {
		if err := o.registerAuditHandlers(ctx); err != nil {
			return fmt.Errorf("failed to register audit handlers: %w", err)
		}
	}

	// Register session cleanup handlers if enabled
	if o.config.EnableSessionCleanup {
		if err := o.registerSessionCleanupHandlers(ctx); err != nil {
			return fmt.Errorf("failed to register session cleanup handlers: %w", err)
		}
	}

	// Register activation workflow handlers
	if err := o.registerActivationWorkflowHandlers(ctx); err != nil {
		return fmt.Errorf("failed to register activation workflow handlers: %w", err)
	}

	// Register cross-module event handlers
	if err := o.registerCrossModuleHandlers(ctx); err != nil {
		return fmt.Errorf("failed to register cross-module handlers: %w", err)
	}

	o.logger.Info("Event workflow orchestrator initialized successfully")
	return nil
}

// registerAuditHandlers registers audit trail handlers
func (o *EventWorkflowOrchestrator) registerAuditHandlers(ctx context.Context) error {
	o.logger.Info("Registering audit trail handlers")

	if err := o.auditTrailService.RegisterAuditHandlers(ctx); err != nil {
		return fmt.Errorf("failed to register audit handlers: %w", err)
	}

	return nil
}

// registerSessionCleanupHandlers registers session cleanup handlers
func (o *EventWorkflowOrchestrator) registerSessionCleanupHandlers(ctx context.Context) error {
	o.logger.Info("Registering session cleanup handlers")

	auditLogger := audit.NewAuditLogger(nil) // This should be injected properly

	// User status changed handler
	statusChangeHandler := authApp.NewUserStatusChangedSessionCleanupHandler(
		o.sessionRepo,
		o.authService,
		o.logger,
		auditLogger,
	)
	if err := o.eventBus.Subscribe("user.status_changed", statusChangeHandler); err != nil {
		return fmt.Errorf("failed to subscribe to user.status_changed: %w", err)
	}

	// User deleted handler
	deletedHandler := authApp.NewUserDeletedSessionCleanupHandler(
		o.sessionRepo,
		o.authService,
		o.logger,
		auditLogger,
	)
	if err := o.eventBus.Subscribe("user.deleted", deletedHandler); err != nil {
		return fmt.Errorf("failed to subscribe to user.deleted: %w", err)
	}

	// User deactivated handler
	deactivatedHandler := authApp.NewUserDeactivatedSessionCleanupHandler(
		o.sessionRepo,
		o.authService,
		o.logger,
		auditLogger,
	)
	if err := o.eventBus.Subscribe("user.deactivated", deactivatedHandler); err != nil {
		return fmt.Errorf("failed to subscribe to user.deactivated: %w", err)
	}

	o.logger.Info("Session cleanup handlers registered successfully")
	return nil
}

// registerActivationWorkflowHandlers registers activation workflow handlers
func (o *EventWorkflowOrchestrator) registerActivationWorkflowHandlers(ctx context.Context) error {
	o.logger.Info("Registering activation workflow handlers")

	// Activation notification handler
	activationNotificationHandler := NewActivationNotificationHandler(
		o.activationService,
		o.logger,
	)
	if err := o.eventBus.Subscribe("user.activation_requested", activationNotificationHandler); err != nil {
		return fmt.Errorf("failed to subscribe to user.activation_requested: %w", err)
	}

	// Token cleanup handler
	tokenCleanupHandler := NewActivationTokenCleanupHandler(
		o.activationService,
		o.logger,
	)
	if err := o.eventBus.Subscribe("user.activated", tokenCleanupHandler); err != nil {
		return fmt.Errorf("failed to subscribe to user.activated: %w", err)
	}
	if err := o.eventBus.Subscribe("user.activation_token_expired", tokenCleanupHandler); err != nil {
		return fmt.Errorf("failed to subscribe to user.activation_token_expired: %w", err)
	}

	o.logger.Info("Activation workflow handlers registered successfully")
	return nil
}

// registerCrossModuleHandlers registers handlers that coordinate between modules
func (o *EventWorkflowOrchestrator) registerCrossModuleHandlers(ctx context.Context) error {
	o.logger.Info("Registering cross-module handlers")

	// User lifecycle handler (coordinates between user and auth modules)
	lifecycleHandler := NewUserLifecycleHandler(
		o.userService,
		o.authService,
		o.activationService,
		o.logger,
	)

	// Subscribe to relevant events
	eventTypes := []string{
		"user.created",
		"user.updated",
		"user.deleted",
		"user.activated",
		"user.deactivated",
		"auth.user_registered",
	}

	for _, eventType := range eventTypes {
		if err := o.eventBus.Subscribe(eventType, lifecycleHandler); err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", eventType, err)
		}
	}

	o.logger.Info("Cross-module handlers registered successfully")
	return nil
}

// Shutdown gracefully shuts down the workflow orchestrator
func (o *EventWorkflowOrchestrator) Shutdown(ctx context.Context) error {
	o.logger.Info("Shutting down event workflow orchestrator")

	// The event bus will handle unsubscribing handlers when it shuts down
	// Any cleanup specific to workflows can be added here

	o.logger.Info("Event workflow orchestrator shut down successfully")
	return nil
}

// GetWorkflowStatus returns the current status of workflows
func (o *EventWorkflowOrchestrator) GetWorkflowStatus(ctx context.Context) (*WorkflowStatus, error) {
	status := &WorkflowStatus{
		EventBusHealth:        o.eventBus.Health() == nil,
		AuditTrailEnabled:     o.config.EnableAuditTrail,
		SessionCleanupEnabled: o.config.EnableSessionCleanup,
		ActivationEnabled:     true,
		LastHealthCheck:       time.Now(),
	}

	return status, nil
}

// WorkflowStatus represents the current status of workflows
type WorkflowStatus struct {
	EventBusHealth        bool      `json:"event_bus_health"`
	AuditTrailEnabled     bool      `json:"audit_trail_enabled"`
	SessionCleanupEnabled bool      `json:"session_cleanup_enabled"`
	ActivationEnabled     bool      `json:"activation_enabled"`
	LastHealthCheck       time.Time `json:"last_health_check"`
}

// ActivationNotificationHandler handles activation request notifications
type ActivationNotificationHandler struct {
	activationService userApp.ActivationService
	logger            *slog.Logger
}

// NewActivationNotificationHandler creates a new activation notification handler
func NewActivationNotificationHandler(
	activationService userApp.ActivationService,
	logger *slog.Logger,
) *ActivationNotificationHandler {
	return &ActivationNotificationHandler{
		activationService: activationService,
		logger:            logger,
	}
}

// Handle processes activation request events
func (h *ActivationNotificationHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Processing activation request notification",
		"event_id", event.EventID(),
		"user_id", event.AggregateID(),
	)

	// In a real implementation, this would:
	// - Send activation email to user
	// - Notify administrators
	// - Log the activation request
	// - Update external systems

	h.logger.Info("Activation notification processed",
		"event_id", event.EventID(),
		"user_id", event.AggregateID(),
	)

	return nil
}

// EventType returns the event type this handler processes
func (h *ActivationNotificationHandler) EventType() string {
	return "user.activation_requested"
}

// HandlerName returns a unique name for this handler
func (h *ActivationNotificationHandler) HandlerName() string {
	return "workflows.activation_notification_handler"
}

// ActivationTokenCleanupHandler handles cleanup of activation tokens
type ActivationTokenCleanupHandler struct {
	activationService userApp.ActivationService
	logger            *slog.Logger
}

// NewActivationTokenCleanupHandler creates a new activation token cleanup handler
func NewActivationTokenCleanupHandler(
	activationService userApp.ActivationService,
	logger *slog.Logger,
) *ActivationTokenCleanupHandler {
	return &ActivationTokenCleanupHandler{
		activationService: activationService,
		logger:            logger,
	}
}

// Handle processes token cleanup events
func (h *ActivationTokenCleanupHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Processing activation token cleanup",
		"event_id", event.EventID(),
		"event_type", event.EventType(),
		"user_id", event.AggregateID(),
	)

	// Clean up expired tokens periodically
	if err := h.activationService.CleanupExpiredTokens(ctx); err != nil {
		h.logger.Error("Failed to cleanup expired tokens",
			"error", err,
			"event_id", event.EventID(),
		)
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	h.logger.Info("Activation token cleanup completed",
		"event_id", event.EventID(),
	)

	return nil
}

// EventType returns the event type this handler processes
func (h *ActivationTokenCleanupHandler) EventType() string {
	return "*" // Handle multiple event types
}

// HandlerName returns a unique name for this handler
func (h *ActivationTokenCleanupHandler) HandlerName() string {
	return "workflows.activation_token_cleanup_handler"
}

// UserLifecycleHandler coordinates user lifecycle events across modules
type UserLifecycleHandler struct {
	userService       userApp.UserService
	authService       authApp.AuthService
	activationService userApp.ActivationService
	logger            *slog.Logger
}

// NewUserLifecycleHandler creates a new user lifecycle handler
func NewUserLifecycleHandler(
	userService userApp.UserService,
	authService authApp.AuthService,
	activationService userApp.ActivationService,
	logger *slog.Logger,
) *UserLifecycleHandler {
	return &UserLifecycleHandler{
		userService:       userService,
		authService:       authService,
		activationService: activationService,
		logger:            logger,
	}
}

// Handle processes user lifecycle events
func (h *UserLifecycleHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Processing user lifecycle event",
		"event_id", event.EventID(),
		"event_type", event.EventType(),
		"user_id", event.AggregateID(),
	)

	switch event.EventType() {
	case "user.created":
		return h.handleUserCreated(ctx, event)
	case "user.activated":
		return h.handleUserActivated(ctx, event)
	case "user.deactivated":
		return h.handleUserDeactivated(ctx, event)
	case "user.deleted":
		return h.handleUserDeleted(ctx, event)
	case "auth.user_registered":
		return h.handleUserRegistered(ctx, event)
	default:
		h.logger.Debug("No specific handling for event type",
			"event_type", event.EventType(),
			"event_id", event.EventID(),
		)
	}

	return nil
}

// EventType returns the event type this handler processes
func (h *UserLifecycleHandler) EventType() string {
	return "*" // Handle multiple event types
}

// HandlerName returns a unique name for this handler
func (h *UserLifecycleHandler) HandlerName() string {
	return "workflows.user_lifecycle_handler"
}

func (h *UserLifecycleHandler) handleUserCreated(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Handling user created event",
		"event_id", event.EventID(),
		"user_id", event.AggregateID(),
	)

	// Coordinate any cross-module actions needed when a user is created
	// This could include:
	// - Setting up default preferences
	// - Creating related entities in other modules
	// - Triggering welcome workflows

	return nil
}

func (h *UserLifecycleHandler) handleUserActivated(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Handling user activated event",
		"event_id", event.EventID(),
		"user_id", event.AggregateID(),
	)

	// Actions when user is activated:
	// - Enable full access to the system
	// - Send welcome notifications
	// - Update external systems

	return nil
}

func (h *UserLifecycleHandler) handleUserDeactivated(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Handling user deactivated event",
		"event_id", event.EventID(),
		"user_id", event.AggregateID(),
	)

	// Actions when user is deactivated:
	// - Disable system access
	// - Clean up temporary data
	// - Notify relevant parties

	return nil
}

func (h *UserLifecycleHandler) handleUserDeleted(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Handling user deleted event",
		"event_id", event.EventID(),
		"user_id", event.AggregateID(),
	)

	// Actions when user is deleted:
	// - Clean up all related data
	// - Archive important information
	// - Update external systems

	return nil
}

func (h *UserLifecycleHandler) handleUserRegistered(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Handling user registered event",
		"event_id", event.EventID(),
		"user_id", event.AggregateID(),
	)

	// Actions when user registers:
	// - Send welcome email
	// - Set up onboarding workflow
	// - Create default settings

	return nil
}
