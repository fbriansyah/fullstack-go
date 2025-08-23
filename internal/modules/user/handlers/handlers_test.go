package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-templ-template/internal/modules/user/application"
	"go-templ-template/internal/modules/user/domain"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, cmd *application.CreateUserCommand) (*domain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, query *application.GetUserQuery) (*domain.User, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, query *application.GetUserByEmailQuery) (*domain.User, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, cmd *application.UpdateUserCommand) (*domain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) UpdateUserEmail(ctx context.Context, cmd *application.UpdateUserEmailCommand) (*domain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) ChangeUserPassword(ctx context.Context, cmd *application.ChangeUserPasswordCommand) (*domain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) ChangeUserStatus(ctx context.Context, cmd *application.ChangeUserStatusCommand) (*domain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, cmd *application.DeleteUserCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(ctx context.Context, query *application.ListUsersQuery) ([]*domain.User, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.User), args.Get(1).(int64), args.Error(2)
}

func TestUserHandler_CreateUser(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful user creation",
			requestBody: CreateUserRequest{
				Email:     "test@example.com",
				Password:  "Password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMock: func(service *MockUserService) {
				user := &domain.User{
					ID:        "user-123",
					Email:     "test@example.com",
					FirstName: "John",
					LastName:  "Doe",
					Status:    domain.UserStatusActive,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Version:   1,
				}
				service.On("CreateUser", mock.Anything, mock.AnythingOfType("*application.CreateUserCommand")).Return(user, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "user already exists",
			requestBody: CreateUserRequest{
				Email:     "existing@example.com",
				Password:  "Password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMock: func(service *MockUserService) {
				service.On("CreateUser", mock.Anything, mock.AnythingOfType("*application.CreateUserCommand")).Return(nil, application.NewUserAlreadyExistsError("existing@example.com"))
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "USER_ALREADY_EXISTS",
		},
		{
			name: "invalid request - missing email",
			requestBody: CreateUserRequest{
				Password:  "Password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMock:      func(service *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockUserService{}
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService)
			e := echo.New()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err := handler.CreateUser(c)

			// Assert
			if tt.expectedStatus < 400 {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*MockUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful user retrieval",
			userID: "user-123",
			setupMock: func(service *MockUserService) {
				user := &domain.User{
					ID:        "user-123",
					Email:     "test@example.com",
					FirstName: "John",
					LastName:  "Doe",
					Status:    domain.UserStatusActive,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Version:   1,
				}
				service.On("GetUser", mock.Anything, mock.AnythingOfType("*application.GetUserQuery")).Return(user, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			setupMock: func(service *MockUserService) {
				service.On("GetUser", mock.Anything, mock.AnythingOfType("*application.GetUserQuery")).Return(nil, application.NewUserNotFoundError("nonexistent"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "USER_NOT_FOUND",
		},
		{
			name:           "missing user ID",
			userID:         "",
			setupMock:      func(service *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockUserService{}
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService)
			e := echo.New()

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+tt.userID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.userID)

			// Execute
			err := handler.GetUser(c)

			// Assert
			if tt.expectedStatus < 400 {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		setupMock      func(*MockUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful user update",
			userID: "user-123",
			requestBody: UpdateUserRequest{
				FirstName: "Jane",
				LastName:  "Smith",
				Version:   1,
			},
			setupMock: func(service *MockUserService) {
				user := &domain.User{
					ID:        "user-123",
					Email:     "test@example.com",
					FirstName: "Jane",
					LastName:  "Smith",
					Status:    domain.UserStatusActive,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Version:   2,
				}
				service.On("UpdateUser", mock.Anything, mock.AnythingOfType("*application.UpdateUserCommand")).Return(user, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "optimistic lock error",
			userID: "user-123",
			requestBody: UpdateUserRequest{
				FirstName: "Jane",
				LastName:  "Smith",
				Version:   1,
			},
			setupMock: func(service *MockUserService) {
				service.On("UpdateUser", mock.Anything, mock.AnythingOfType("*application.UpdateUserCommand")).Return(nil, application.NewOptimisticLockError("user-123"))
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "OPTIMISTIC_LOCK_ERROR",
		},
		{
			name:   "invalid request - missing first name",
			userID: "user-123",
			requestBody: UpdateUserRequest{
				LastName: "Smith",
				Version:  1,
			},
			setupMock:      func(service *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockUserService{}
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService)
			e := echo.New()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+tt.userID, bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.userID)

			// Execute
			err := handler.UpdateUser(c)

			// Assert
			if tt.expectedStatus < 400 {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*MockUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful user deletion",
			userID: "user-123",
			setupMock: func(service *MockUserService) {
				service.On("DeleteUser", mock.Anything, mock.AnythingOfType("*application.DeleteUserCommand")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			setupMock: func(service *MockUserService) {
				service.On("DeleteUser", mock.Anything, mock.AnythingOfType("*application.DeleteUserCommand")).Return(application.NewUserNotFoundError("nonexistent"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "USER_NOT_FOUND",
		},
		{
			name:           "missing user ID",
			userID:         "",
			setupMock:      func(service *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockUserService{}
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService)
			e := echo.New()

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+tt.userID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.userID)

			// Execute
			err := handler.DeleteUser(c)

			// Assert
			if tt.expectedStatus < 400 {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*MockUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "successful user listing",
			queryParams: "?limit=10&offset=0",
			setupMock: func(service *MockUserService) {
				users := []*domain.User{
					{
						ID:        "user-1",
						Email:     "user1@example.com",
						FirstName: "John",
						LastName:  "Doe",
						Status:    domain.UserStatusActive,
					},
					{
						ID:        "user-2",
						Email:     "user2@example.com",
						FirstName: "Jane",
						LastName:  "Smith",
						Status:    domain.UserStatusActive,
					},
				}
				service.On("ListUsers", mock.Anything, mock.AnythingOfType("*application.ListUsersQuery")).Return(users, int64(2), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid limit",
			queryParams:    "?limit=200",
			setupMock:      func(service *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockUserService{}
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService)
			e := echo.New()

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err := handler.ListUsers(c)

			// Assert
			if tt.expectedStatus < 400 {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}

			mockService.AssertExpectations(t)
		})
	}
}
