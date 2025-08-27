package integration

import (
	"context"
	"testing"
	"time"

	"go-templ-template/internal/modules/auth/application"
	authDomain "go-templ-template/internal/modules/auth/domain"
	authInfra "go-templ-template/internal/modules/auth/infrastructure"
	userApp "go-templ-template/internal/modules/user/application"
	userDomain "go-templ-template/internal/modules/user/domain"
	userInfra "go-templ-template/internal/modules/user/infrastructure"
	"go-templ-template/internal/shared/audit"
	"go-templ-template/internal/shared/database"
	"go-templ-template/internal/shared/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// EventFlowTestSuite tests end-to-end event flows
type EventFlowTestSuite struct {
	suite.Suite
	db                *database.DB
	eventBus          events.EventBus
	userService       userApp.UserService
	authService       application.AuthService
	activationService userApp.ActivationService
	auditLogger       audit.AuditLogger
	auditTrailService *audit.AuditTrailService

	// Test data
	testUserID    string
	testUserEmail string
}

// SetupSuite sets up the test suite
func (suite *EventFlowTestSuite) SetupSuite() {
	// Initialize test database
	db, err := database.NewTestDB()
	require.NoError(suite.T(), err)
	suite.db = db

	// Create tables
	err = database.CreateTables(context.Background(), suite.db.DB)
	require.NoError(suite.T(), err)

	err = audit.CreateAuditEventsTable(context.Background(), suite.db.DB)
	require.NoError(suite.T(), err)

	err = userInfra.CreateActivationTokensTable(context.Background(), suite.db.DB)
	require.NoError(suite.T(), err)

	// Initialize event bus
	eventBusConfig := events.DefaultRabbitMQConfig()
	eventBusConfig.URL = "amqp://guest:guest@localhost:5672/" // Use test RabbitMQ
	suite.eventBus = events.NewRabbitMQEventBus(eventBusConfig)

	// Start event bus
	ctx := context.Background()
	err = suite.eventBus.Start(ctx)
	require.NoError(suite.T(), err)

	// Initialize repositories
	userRepo := userInfra.NewUserRepository(suite.db)
	sessionRepo := authInfra.NewSessionRepository(suite.db)
	tokenRepo := userInfra.NewActivationTokenRepository(suite.db)

	// Initialize audit logger
	suite.auditLogger = audit.NewAuditLogger(suite.db)

	// Initialize services
	suite.userService = userApp.NewUserService(userRepo, suite.eventBus, suite.db)
	suite.authService = application.NewAuthService(
		sessionRepo,
		suite.userService,
		suite.eventBus,
		suite.db,
		application.NewInMemoryRateLimiter(),
		application.DefaultSessionConfig(),
	)
	suite.activationService = userApp.NewActivationService(
		userRepo,
		tokenRepo,
		suite.eventBus,
		suite.db,
		24*time.Hour,
	)

	// Initialize audit trail service
	suite.auditTrailService = audit.NewAuditTrailService(suite.auditLogger, suite.eventBus, nil)

	// Register audit handlers
	err = suite.auditTrailService.RegisterAuditHandlers(ctx)
	require.NoError(suite.T(), err)

	// Register session cleanup handlers
	suite.registerSessionCleanupHandlers()

	// Set up test data
	suite.testUserEmail = "test@example.com"
}

// TearDownSuite cleans up after the test suite
func (suite *EventFlowTestSuite) TearDownSuite() {
	ctx := context.Background()
	if suite.eventBus != nil {
		suite.eventBus.Stop(ctx)
	}
	if suite.db != nil {
		suite.db.Close()
	}
}

// SetupTest sets up each test
func (suite *EventFlowTestSuite) SetupTest() {
	// Clean up test data
	suite.cleanupTestData()
}

// TestUserRegistrationAndActivationFlow tests the complete user registration and activation flow
func (suite *EventFlowTestSuite) TestUserRegistrationAndActivationFlow() {
	ctx := context.Background()

	// Step 1: Register a new user
	registerCmd := &application.RegisterCommand{
		Email:     suite.testUserEmail,
		Password:  "TestPassword123!",
		FirstName: "Test",
		LastName:  "User",
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	}

	authResult, err := suite.authService.Register(ctx, registerCmd)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), authResult)
	require.NotNil(suite.T(), authResult.User)
	require.NotNil(suite.T(), authResult.Session)

	suite.testUserID = authResult.User.ID

	// Verify user is created and active
	assert.Equal(suite.T(), userDomain.UserStatusActive, authResult.User.Status)
	assert.Equal(suite.T(), suite.testUserEmail, authResult.User.Email)

	// Wait for events to be processed
	time.Sleep(100 * time.Millisecond)

	// Step 2: Verify audit trail for registration
	auditEvents, err := suite.auditTrailService.GetUserAuditTrail(ctx, suite.testUserID, 10)
	require.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(auditEvents), 2) // At least user creation and registration events

	// Find user creation audit event
	var userCreatedEvent *audit.AuditEvent
	for _, event := range auditEvents {
		if event.Action == "user_created" {
			userCreatedEvent = event
			break
		}
	}
	require.NotNil(suite.T(), userCreatedEvent)
	assert.Equal(suite.T(), "user", userCreatedEvent.Resource)
	assert.Equal(suite.T(), suite.testUserID, userCreatedEvent.ResourceID)

	// Step 3: Deactivate user to test activation flow
	deactivateCmd := &userApp.DeactivateUserCommand{
		UserID:        suite.testUserID,
		Version:       authResult.User.Version,
		DeactivatedBy: "test-admin",
		Reason:        "testing activation flow",
	}

	deactivatedUser, err := suite.activationService.DeactivateUser(ctx, deactivateCmd)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), userDomain.UserStatusInactive, deactivatedUser.Status)

	// Wait for events to be processed
	time.Sleep(100 * time.Millisecond)

	// Step 4: Verify session cleanup after deactivation
	sessions, err := suite.authService.(*application.AuthServiceImpl).SessionRepo.GetByUserID(ctx, suite.testUserID)
	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), sessions, "All sessions should be cleaned up after user deactivation")

	// Step 5: Request activation
	requestCmd := &userApp.RequestActivationCommand{
		UserID:      suite.testUserID,
		RequestedBy: "test-admin",
	}

	activationToken, err := suite.activationService.RequestActivation(ctx, requestCmd)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), activationToken)
	assert.False(suite.T(), activationToken.IsExpired())
	assert.False(suite.T(), activationToken.IsUsed())

	// Step 6: Activate user using token
	activateCmd := &userApp.ActivateUserCommand{
		Token:       activationToken.Token,
		ActivatedBy: "test-admin",
	}

	activatedUser, err := suite.activationService.ActivateUser(ctx, activateCmd)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), userDomain.UserStatusActive, activatedUser.Status)

	// Wait for events to be processed
	time.Sleep(100 * time.Millisecond)

	// Step 7: Verify complete audit trail
	finalAuditEvents, err := suite.auditTrailService.GetUserAuditTrail(ctx, suite.testUserID, 20)
	require.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(finalAuditEvents), 5) // Multiple events throughout the flow

	// Verify specific events exist
	eventActions := make(map[string]bool)
	for _, event := range finalAuditEvents {
		eventActions[event.Action] = true
	}

	expectedActions := []string{
		"user_created",
		"user_registered",
		"user_deactivated",
		"user_activation_requested",
		"user_activated",
	}

	for _, action := range expectedActions {
		assert.True(suite.T(), eventActions[action], "Expected audit event action: %s", action)
	}
}

// TestUserStatusChangeSessionCleanup tests session cleanup when user status changes
func (suite *EventFlowTestSuite) TestUserStatusChangeSessionCleanup() {
	ctx := context.Background()

	// Step 1: Create and login user
	user := suite.createTestUser()
	session := suite.createTestSession(user.ID)

	// Verify session exists
	sessions, err := suite.authService.(*application.AuthServiceImpl).SessionRepo.GetByUserID(ctx, user.ID)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 1)

	// Step 2: Change user status to suspended
	changeStatusCmd := &userApp.ChangeUserStatusCommand{
		ID:        user.ID,
		Status:    userDomain.UserStatusSuspended,
		Version:   user.Version,
		ChangedBy: "test-admin",
		Reason:    "testing session cleanup",
	}

	_, err = suite.userService.ChangeUserStatus(ctx, changeStatusCmd)
	require.NoError(suite.T(), err)

	// Wait for events to be processed
	time.Sleep(100 * time.Millisecond)

	// Step 3: Verify sessions are cleaned up
	sessions, err = suite.authService.(*application.AuthServiceImpl).SessionRepo.GetByUserID(ctx, user.ID)
	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), sessions, "Sessions should be cleaned up when user is suspended")

	// Step 4: Verify audit trail includes session cleanup
	auditEvents, err := suite.auditTrailService.GetUserAuditTrail(ctx, user.ID, 10)
	require.NoError(suite.T(), err)

	var sessionCleanupEvent *audit.AuditEvent
	for _, event := range auditEvents {
		if event.Action == "sessions_cleaned_up" {
			sessionCleanupEvent = event
			break
		}
	}
	require.NotNil(suite.T(), sessionCleanupEvent)
	assert.Equal(suite.T(), "session", sessionCleanupEvent.Resource)
	assert.Contains(suite.T(), sessionCleanupEvent.Details, "cleanup_reason")
	assert.Equal(suite.T(), "user_status_changed", sessionCleanupEvent.Details["cleanup_reason"])
}

// TestUserDeletionFlow tests the complete user deletion flow
func (suite *EventFlowTestSuite) TestUserDeletionFlow() {
	ctx := context.Background()

	// Step 1: Create user and session
	user := suite.createTestUser()
	session := suite.createTestSession(user.ID)

	// Verify session exists
	sessions, err := suite.authService.(*application.AuthServiceImpl).SessionRepo.GetByUserID(ctx, user.ID)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 1)

	// Step 2: Delete user
	deleteCmd := &userApp.DeleteUserCommand{
		ID:        user.ID,
		DeletedBy: "test-admin",
		Reason:    "testing deletion flow",
	}

	err = suite.userService.DeleteUser(ctx, deleteCmd)
	require.NoError(suite.T(), err)

	// Wait for events to be processed
	time.Sleep(100 * time.Millisecond)

	// Step 3: Verify sessions are cleaned up
	sessions, err = suite.authService.(*application.AuthServiceImpl).SessionRepo.GetByUserID(ctx, user.ID)
	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), sessions, "Sessions should be cleaned up when user is deleted")

	// Step 4: Verify audit trail
	auditEvents, err := suite.auditTrailService.GetUserAuditTrail(ctx, user.ID, 10)
	require.NoError(suite.T(), err)

	eventActions := make(map[string]bool)
	for _, event := range auditEvents {
		eventActions[event.Action] = true
	}

	assert.True(suite.T(), eventActions["user_deleted"])
	assert.True(suite.T(), eventActions["sessions_cleaned_up"])
}

// TestActivationTokenExpiration tests activation token expiration handling
func (suite *EventFlowTestSuite) TestActivationTokenExpiration() {
	ctx := context.Background()

	// Create inactive user
	user := suite.createInactiveTestUser()

	// Create activation service with very short token duration for testing
	shortDurationService := userApp.NewActivationService(
		userInfra.NewUserRepository(suite.db),
		userInfra.NewActivationTokenRepository(suite.db),
		suite.eventBus,
		suite.db,
		1*time.Millisecond, // Very short duration
	)

	// Request activation
	requestCmd := &userApp.RequestActivationCommand{
		UserID:      user.ID,
		RequestedBy: "test-admin",
	}

	activationToken, err := shortDurationService.RequestActivation(ctx, requestCmd)
	require.NoError(suite.T(), err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Try to activate with expired token
	activateCmd := &userApp.ActivateUserCommand{
		Token:       activationToken.Token,
		ActivatedBy: "test-admin",
	}

	_, err = shortDurationService.ActivateUser(ctx, activateCmd)
	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "expired")

	// Wait for events to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify audit trail includes token expiration
	auditEvents, err := suite.auditTrailService.GetUserAuditTrail(ctx, user.ID, 10)
	require.NoError(suite.T(), err)

	var tokenExpiredEvent *audit.AuditEvent
	for _, event := range auditEvents {
		if event.Action == "user_activation_token_expired" {
			tokenExpiredEvent = event
			break
		}
	}
	require.NotNil(suite.T(), tokenExpiredEvent)
}

// Helper methods

func (suite *EventFlowTestSuite) registerSessionCleanupHandlers() {
	ctx := context.Background()

	// Create session cleanup handlers
	sessionRepo := authInfra.NewSessionRepository(suite.db)

	statusChangeHandler := application.NewUserStatusChangedSessionCleanupHandler(
		sessionRepo, suite.authService, nil, suite.auditLogger,
	)
	deletedHandler := application.NewUserDeletedSessionCleanupHandler(
		sessionRepo, suite.authService, nil, suite.auditLogger,
	)
	deactivatedHandler := application.NewUserDeactivatedSessionCleanupHandler(
		sessionRepo, suite.authService, nil, suite.auditLogger,
	)

	// Register handlers
	suite.eventBus.Subscribe("user.status_changed", statusChangeHandler)
	suite.eventBus.Subscribe("user.deleted", deletedHandler)
	suite.eventBus.Subscribe("user.deactivated", deactivatedHandler)
}

func (suite *EventFlowTestSuite) createTestUser() *userDomain.User {
	ctx := context.Background()

	createCmd := &userApp.CreateUserCommand{
		Email:     "testuser@example.com",
		Password:  "TestPassword123!",
		FirstName: "Test",
		LastName:  "User",
	}

	user, err := suite.userService.CreateUser(ctx, createCmd)
	require.NoError(suite.T(), err)

	return user
}

func (suite *EventFlowTestSuite) createInactiveTestUser() *userDomain.User {
	user := suite.createTestUser()

	// Deactivate the user
	ctx := context.Background()
	deactivateCmd := &userApp.DeactivateUserCommand{
		UserID:        user.ID,
		Version:       user.Version,
		DeactivatedBy: "test",
		Reason:        "testing",
	}

	deactivatedUser, err := suite.activationService.DeactivateUser(ctx, deactivateCmd)
	require.NoError(suite.T(), err)

	return deactivatedUser
}

func (suite *EventFlowTestSuite) createTestSession(userID string) *authDomain.Session {
	ctx := context.Background()

	session, err := authDomain.NewSession(
		userID,
		"127.0.0.1",
		"test-agent",
		application.DefaultSessionConfig(),
	)
	require.NoError(suite.T(), err)

	sessionRepo := authInfra.NewSessionRepository(suite.db)
	err = sessionRepo.Create(ctx, session)
	require.NoError(suite.T(), err)

	return session
}

func (suite *EventFlowTestSuite) cleanupTestData() {
	ctx := context.Background()

	// Clean up test data
	suite.db.ExecContext(ctx, "DELETE FROM audit_events")
	suite.db.ExecContext(ctx, "DELETE FROM activation_tokens")
	suite.db.ExecContext(ctx, "DELETE FROM sessions")
	suite.db.ExecContext(ctx, "DELETE FROM users")
}

// TestEventFlowTestSuite runs the test suite
func TestEventFlowTestSuite(t *testing.T) {
	suite.Run(t, new(EventFlowTestSuite))
}
