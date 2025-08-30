package testing

import (
	"context"
	"sync"

	"go-templ-template/internal/modules/auth/domain"
	userDomain "go-templ-template/internal/modules/user/domain"
	userInfra "go-templ-template/internal/modules/user/infrastructure"
	"go-templ-template/internal/shared/audit"
	"go-templ-template/internal/shared/database"
	"go-templ-template/internal/shared/events"

	"github.com/stretchr/testify/mock"
)

// MockSessionRepository provides a mock implementation of SessionRepository for testing
type MockSessionRepository struct {
	mock.Mock
	sessions map[string]*domain.Session
	mu       sync.RWMutex
}

// NewMockSessionRepository creates a new mock session repository
func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[string]*domain.Session),
	}
}

// Create mocks session creation
func (m *MockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	args := m.Called(ctx, session)
	if args.Error(0) == nil {
		m.mu.Lock()
		m.sessions[session.ID] = session
		m.mu.Unlock()
	}
	return args.Error(0)
}

// GetByID mocks session retrieval by ID
func (m *MockSessionRepository) GetByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	// Check if there are any expectations set for this method
	if len(m.ExpectedCalls) > 0 {
		// Use mock framework if expectations are set
		args := m.Called(ctx, sessionID)
		if args.Error(1) != nil {
			return nil, args.Error(1)
		}

		// Return the mocked value if provided
		if args.Get(0) != nil {
			return args.Get(0).(*domain.Session), nil
		}
	}

	// Fallback to internal storage
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return nil, database.ErrNotFound
	}

	return session, nil
}

// GetByUserID mocks session retrieval by user ID
func (m *MockSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	args := m.Called(ctx, userID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	// Return the mocked value if provided
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Session), nil
	}

	// Fallback to internal storage
	m.mu.RLock()
	defer m.mu.RUnlock()

	var userSessions []*domain.Session
	for _, session := range m.sessions {
		if session.UserID == userID && session.IsActive && !session.IsExpired() {
			userSessions = append(userSessions, session)
		}
	}

	return userSessions, nil
}

// Update mocks session update
func (m *MockSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	args := m.Called(ctx, session)
	if args.Error(0) == nil {
		m.mu.Lock()
		if _, exists := m.sessions[session.ID]; exists {
			m.sessions[session.ID] = session
		}
		m.mu.Unlock()
	}
	return args.Error(0)
}

// Delete mocks session deletion by ID
func (m *MockSessionRepository) Delete(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	if args.Error(0) == nil {
		m.mu.Lock()
		delete(m.sessions, sessionID)
		m.mu.Unlock()
	}
	return args.Error(0)
}

// DeleteByUserID mocks session deletion by user ID
func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	if args.Error(0) == nil {
		m.mu.Lock()
		for sessionID, session := range m.sessions {
			if session.UserID == userID {
				delete(m.sessions, sessionID)
			}
		}
		m.mu.Unlock()
	}
	return args.Error(0)
}

// DeleteExpired mocks expired session cleanup
func (m *MockSessionRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.mu.Lock()
		for sessionID, session := range m.sessions {
			if session.IsExpired() || !session.IsActive {
				delete(m.sessions, sessionID)
			}
		}
		m.mu.Unlock()
	}
	return args.Error(0)
}

// ExistsByID mocks session existence check by ID
func (m *MockSessionRepository) ExistsByID(ctx context.Context, sessionID string) (bool, error) {
	args := m.Called(ctx, sessionID)
	if args.Error(1) != nil {
		return false, args.Error(1)
	}

	m.mu.RLock()
	_, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	return exists, nil
}

// MockEventBus provides a mock implementation of EventBus for testing
type MockEventBus struct {
	mock.Mock
	publishedEvents []events.DomainEvent
	handlers        map[string][]events.EventHandler
	mu              sync.RWMutex
}

// NewMockEventBus creates a new mock event bus
func NewMockEventBus() *MockEventBus {
	return &MockEventBus{
		publishedEvents: make([]events.DomainEvent, 0),
		handlers:        make(map[string][]events.EventHandler),
	}
}

// Publish mocks event publishing
func (m *MockEventBus) Publish(ctx context.Context, event events.DomainEvent) error {
	args := m.Called(ctx, event)
	if args.Error(0) == nil {
		m.mu.Lock()
		m.publishedEvents = append(m.publishedEvents, event)
		m.mu.Unlock()
	}
	return args.Error(0)
}

// Subscribe mocks event handler subscription
func (m *MockEventBus) Subscribe(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	if args.Error(0) == nil {
		m.mu.Lock()
		m.handlers[eventType] = append(m.handlers[eventType], handler)
		m.mu.Unlock()
	}
	return args.Error(0)
}

// Unsubscribe mocks event handler unsubscription
func (m *MockEventBus) Unsubscribe(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

// Start mocks event bus startup
func (m *MockEventBus) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Stop mocks event bus shutdown
func (m *MockEventBus) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Health mocks event bus health check
func (m *MockEventBus) Health() error {
	args := m.Called()
	return args.Error(0)
}

// GetPublishedEvents returns all published events for testing verification
func (m *MockEventBus) GetPublishedEvents() []events.DomainEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := make([]events.DomainEvent, len(m.publishedEvents))
	copy(events, m.publishedEvents)
	return events
}

// ClearPublishedEvents clears the published events list
func (m *MockEventBus) ClearPublishedEvents() {
	m.mu.Lock()
	m.publishedEvents = make([]events.DomainEvent, 0)
	m.mu.Unlock()
}

// MockRateLimiter provides a mock implementation of RateLimiter for testing
type MockRateLimiter struct {
	mock.Mock
	attempts map[string]int
	mu       sync.RWMutex
}

// NewMockRateLimiter creates a new mock rate limiter
func NewMockRateLimiter() *MockRateLimiter {
	return &MockRateLimiter{
		attempts: make(map[string]int),
	}
}

// Allow mocks rate limit checking
func (m *MockRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	if args.Error(1) == nil && args.Bool(0) {
		m.mu.Lock()
		m.attempts[key]++
		m.mu.Unlock()
	}
	return args.Bool(0), args.Error(1)
}

// Reset mocks rate limit reset
func (m *MockRateLimiter) Reset(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	if args.Error(0) == nil {
		m.mu.Lock()
		delete(m.attempts, key)
		m.mu.Unlock()
	}
	return args.Error(0)
}

// GetAttempts mocks attempt count retrieval
func (m *MockRateLimiter) GetAttempts(ctx context.Context, key string) (int, error) {
	args := m.Called(ctx, key)
	if args.Error(1) != nil {
		return 0, args.Error(1)
	}

	m.mu.RLock()
	count := m.attempts[key]
	m.mu.RUnlock()

	return count, nil
}

// SetAttempts sets the attempt count for a key (for testing setup)
func (m *MockRateLimiter) SetAttempts(key string, count int) {
	m.mu.Lock()
	m.attempts[key] = count
	m.mu.Unlock()
}

// MockUserRepository provides a mock implementation of UserRepository for testing
type MockUserRepository struct {
	mock.Mock
	users map[string]*userDomain.User
	mu    sync.RWMutex
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*userDomain.User),
	}
}

// Create mocks user creation
func (m *MockUserRepository) Create(ctx context.Context, user *userDomain.User) error {
	args := m.Called(ctx, user)
	if args.Error(0) == nil {
		m.mu.Lock()
		m.users[user.ID] = user
		m.mu.Unlock()
	}
	return args.Error(0)
}

// GetByID mocks user retrieval by ID
func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*userDomain.User, error) {
	// Check if there are any expectations set for this method
	if len(m.ExpectedCalls) > 0 {
		// Use mock framework if expectations are set
		args := m.Called(ctx, id)
		if args.Error(1) != nil {
			return nil, args.Error(1)
		}

		// Return the mocked value if provided
		if args.Get(0) != nil {
			return args.Get(0).(*userDomain.User), nil
		}
	}

	// Fallback to internal storage
	m.mu.RLock()
	user, exists := m.users[id]
	m.mu.RUnlock()

	if !exists {
		return nil, database.ErrNotFound
	}

	return user, nil
}

// GetByEmail mocks user retrieval by email
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	// Check if there are any expectations set for this method
	if len(m.ExpectedCalls) > 0 {
		// Use mock framework if expectations are set
		args := m.Called(ctx, email)
		if args.Error(1) != nil {
			return nil, args.Error(1)
		}

		// Return the mocked value if provided
		if args.Get(0) != nil {
			return args.Get(0).(*userDomain.User), nil
		}
	}

	// Fallback to internal storage
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, database.ErrNotFound
}

// Update mocks user update
func (m *MockUserRepository) Update(ctx context.Context, user *userDomain.User) error {
	args := m.Called(ctx, user)
	if args.Error(0) == nil {
		m.mu.Lock()
		if _, exists := m.users[user.ID]; exists {
			m.users[user.ID] = user
		}
		m.mu.Unlock()
	}
	return args.Error(0)
}

// Delete mocks user deletion by ID
func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if args.Error(0) == nil {
		m.mu.Lock()
		delete(m.users, id)
		m.mu.Unlock()
	}
	return args.Error(0)
}

// List mocks user listing with pagination and filtering
func (m *MockUserRepository) List(ctx context.Context, filter userInfra.UserFilter, limit, offset int) ([]*userDomain.User, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	// Return the mocked value if provided
	if args.Get(0) != nil {
		return args.Get(0).([]*userDomain.User), nil
	}

	// Fallback to internal storage with basic filtering
	m.mu.RLock()
	defer m.mu.RUnlock()

	var users []*userDomain.User
	for _, user := range m.users {
		// Apply basic filtering
		if filter.Status != nil && user.Status != *filter.Status {
			continue
		}
		if filter.Email != nil && user.Email != *filter.Email {
			continue
		}
		users = append(users, user)
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(users) {
		return []*userDomain.User{}, nil
	}
	if end > len(users) {
		end = len(users)
	}

	return users[start:end], nil
}

// Count mocks user count with filtering
func (m *MockUserRepository) Count(ctx context.Context, filter userInfra.UserFilter) (int64, error) {
	args := m.Called(ctx, filter)
	if args.Error(1) != nil {
		return 0, args.Error(1)
	}

	// Return the mocked value if provided
	if args.Get(0) != nil {
		return args.Get(0).(int64), nil
	}

	// Fallback to internal storage with basic filtering
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := int64(0)
	for _, user := range m.users {
		// Apply basic filtering
		if filter.Status != nil && user.Status != *filter.Status {
			continue
		}
		if filter.Email != nil && user.Email != *filter.Email {
			continue
		}
		count++
	}

	return count, nil
}

// Exists mocks user existence check by ID
func (m *MockUserRepository) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	if args.Error(1) != nil {
		return false, args.Error(1)
	}

	// Return the mocked value if provided
	if len(args) > 0 {
		return args.Bool(0), nil
	}

	// Fallback to internal storage
	m.mu.RLock()
	_, exists := m.users[id]
	m.mu.RUnlock()

	return exists, nil
}

// ExistsByEmail mocks user existence check by email
func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	if args.Error(1) != nil {
		return false, args.Error(1)
	}

	// Return the mocked value if provided
	if len(args) > 0 {
		return args.Bool(0), nil
	}

	// Fallback to internal storage
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.users {
		if user.Email == email {
			return true, nil
		}
	}

	return false, nil
}

// MockAuditLogger provides a mock implementation of AuditLogger for testing
type MockAuditLogger struct {
	mock.Mock
	events []audit.AuditEvent
	mu     sync.RWMutex
}

// NewMockAuditLogger creates a new mock audit logger
func NewMockAuditLogger() *MockAuditLogger {
	return &MockAuditLogger{
		events: make([]audit.AuditEvent, 0),
	}
}

// LogEvent mocks audit event logging
func (m *MockAuditLogger) LogEvent(ctx context.Context, event *audit.AuditEvent) error {
	args := m.Called(ctx, event)
	if args.Error(0) == nil {
		m.mu.Lock()
		m.events = append(m.events, *event)
		m.mu.Unlock()
	}
	return args.Error(0)
}

// GetEvents mocks audit event retrieval
func (m *MockAuditLogger) GetEvents(ctx context.Context, filter *audit.AuditFilter) ([]*audit.AuditEvent, error) {
	args := m.Called(ctx, filter)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	// Return the mocked value if provided
	if args.Get(0) != nil {
		return args.Get(0).([]*audit.AuditEvent), nil
	}

	// Fallback to internal storage
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filteredEvents []*audit.AuditEvent
	for i := range m.events {
		// Basic filtering - in a real implementation, you'd apply the filter criteria
		filteredEvents = append(filteredEvents, &m.events[i])
	}

	return filteredEvents, nil
}

// GetLoggedEvents returns all logged events for testing verification
func (m *MockAuditLogger) GetLoggedEvents() []audit.AuditEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := make([]audit.AuditEvent, len(m.events))
	copy(events, m.events)
	return events
}

// ClearLoggedEvents clears the logged events list
func (m *MockAuditLogger) ClearLoggedEvents() {
	m.mu.Lock()
	m.events = make([]audit.AuditEvent, 0)
	m.mu.Unlock()
}
