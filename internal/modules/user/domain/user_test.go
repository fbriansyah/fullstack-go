package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status UserStatus
		want   bool
	}{
		{"Active status is valid", UserStatusActive, true},
		{"Inactive status is valid", UserStatusInactive, true},
		{"Suspended status is valid", UserStatusSuspended, true},
		{"Invalid status", UserStatus("invalid"), false},
		{"Empty status", UserStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsValid())
		})
	}
}

func TestUserStatus_String(t *testing.T) {
	assert.Equal(t, "active", UserStatusActive.String())
	assert.Equal(t, "inactive", UserStatusInactive.String())
	assert.Equal(t, "suspended", UserStatusSuspended.String())
}

func TestNewUser(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		email     string
		password  string
		firstName string
		lastName  string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Valid user creation",
			id:        "user-123",
			email:     "john.doe@example.com",
			password:  "Password123",
			firstName: "John",
			lastName:  "Doe",
			wantErr:   false,
		},
		{
			name:      "Empty ID",
			id:        "",
			email:     "john.doe@example.com",
			password:  "Password123",
			firstName: "John",
			lastName:  "Doe",
			wantErr:   true,
			errMsg:    "user ID cannot be empty",
		},
		{
			name:      "Invalid email",
			id:        "user-123",
			email:     "invalid-email",
			password:  "Password123",
			firstName: "John",
			lastName:  "Doe",
			wantErr:   true,
			errMsg:    "invalid email format",
		},
		{
			name:      "Weak password",
			id:        "user-123",
			email:     "john.doe@example.com",
			password:  "weak",
			firstName: "John",
			lastName:  "Doe",
			wantErr:   true,
			errMsg:    "password must be at least 8 characters long",
		},
		{
			name:      "Empty first name",
			id:        "user-123",
			email:     "john.doe@example.com",
			password:  "Password123",
			firstName: "",
			lastName:  "Doe",
			wantErr:   true,
			errMsg:    "first name cannot be empty",
		},
		{
			name:      "Empty last name",
			id:        "user-123",
			email:     "john.doe@example.com",
			password:  "Password123",
			firstName: "John",
			lastName:  "",
			wantErr:   true,
			errMsg:    "last name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.id, tt.email, tt.password, tt.firstName, tt.lastName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.id, user.ID)
				assert.Equal(t, strings.ToLower(tt.email), user.Email)
				assert.Equal(t, tt.firstName, user.FirstName)
				assert.Equal(t, tt.lastName, user.LastName)
				assert.Equal(t, UserStatusActive, user.Status)
				assert.Equal(t, 1, user.Version)
				assert.NotEmpty(t, user.Password)
				assert.NotEqual(t, tt.password, user.Password) // Password should be hashed
			}
		})
	}
}

func TestUser_ValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		{"Valid email", "test@example.com", false, ""},
		{"Valid email with subdomain", "user@mail.example.com", false, ""},
		{"Valid email with numbers", "user123@example.com", false, ""},
		{"Valid email with special chars", "user.name+tag@example.com", false, ""},
		{"Empty email", "", true, "email cannot be empty"},
		{"Invalid format - no @", "testexample.com", true, "invalid email format"},
		{"Invalid format - no domain", "test@", true, "invalid email format"},
		{"Invalid format - no TLD", "test@example", true, "invalid email format"},
		{"Too long email", strings.Repeat("a", 250) + "@example.com", true, "email cannot exceed 255 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Email: tt.email}
			err := user.validateEmail()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_ValidateName(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		wantErr   bool
		errMsg    string
	}{
		{"Valid names", "John", "Doe", false, ""},
		{"Names with spaces", "Mary Jane", "Smith-Jones", false, ""},
		{"Names with apostrophes", "O'Connor", "D'Angelo", false, ""},
		{"Names with hyphens", "Jean-Pierre", "Smith-Wilson", false, ""},
		{"Empty first name", "", "Doe", true, "first name cannot be empty"},
		{"Empty last name", "John", "", true, "last name cannot be empty"},
		{"First name too long", strings.Repeat("a", 101), "Doe", true, "first name cannot exceed 100 characters"},
		{"Last name too long", "John", strings.Repeat("a", 101), true, "last name cannot exceed 100 characters"},
		{"Invalid characters in first name", "John123", "Doe", true, "first name contains invalid characters"},
		{"Invalid characters in last name", "John", "Doe@", true, "last name contains invalid characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{FirstName: tt.firstName, LastName: tt.lastName}
			err := user.validateName()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_SetPassword(t *testing.T) {
	user := &User{ID: "test-user"}

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{"Valid password", "Password123", false, ""},
		{"Empty password", "", true, "password cannot be empty"},
		{"Too short password", "Pass1", true, "password must be at least 8 characters long"},
		{"Too long password", strings.Repeat("a", 129), true, "password cannot exceed 128 characters"},
		{"No uppercase", "password123", true, "password must contain at least one uppercase letter"},
		{"No lowercase", "PASSWORD123", true, "password must contain at least one lowercase letter"},
		{"No digit", "Password", true, "password must contain at least one digit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.SetPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, user.Password)
				assert.NotEqual(t, tt.password, user.Password) // Should be hashed
			}
		})
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{ID: "test-user"}
	password := "Password123"

	err := user.SetPassword(password)
	require.NoError(t, err)

	// Test correct password
	assert.True(t, user.CheckPassword(password))

	// Test incorrect password
	assert.False(t, user.CheckPassword("WrongPassword"))
	assert.False(t, user.CheckPassword(""))
}

func TestUser_UpdateProfile(t *testing.T) {
	user, err := NewUser("test-user", "test@example.com", "Password123", "John", "Doe")
	require.NoError(t, err)

	originalVersion := user.Version
	originalUpdatedAt := user.UpdatedAt

	// Wait a bit to ensure UpdatedAt changes
	time.Sleep(time.Millisecond)

	err = user.UpdateProfile("Jane", "Smith")
	assert.NoError(t, err)
	assert.Equal(t, "Jane", user.FirstName)
	assert.Equal(t, "Smith", user.LastName)
	assert.Equal(t, originalVersion+1, user.Version)
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))

	// Test invalid profile update
	err = user.UpdateProfile("", "Smith")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "first name cannot be empty")
}

func TestUser_UpdateEmail(t *testing.T) {
	user, err := NewUser("test-user", "test@example.com", "Password123", "John", "Doe")
	require.NoError(t, err)

	originalVersion := user.Version
	originalUpdatedAt := user.UpdatedAt

	// Wait a bit to ensure UpdatedAt changes
	time.Sleep(time.Millisecond)

	err = user.UpdateEmail("newemail@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "newemail@example.com", user.Email)
	assert.Equal(t, originalVersion+1, user.Version)
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))

	// Test invalid email update
	err = user.UpdateEmail("invalid-email")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestUser_ChangeStatus(t *testing.T) {
	user, err := NewUser("test-user", "test@example.com", "Password123", "John", "Doe")
	require.NoError(t, err)

	originalVersion := user.Version
	originalUpdatedAt := user.UpdatedAt

	// Wait a bit to ensure UpdatedAt changes
	time.Sleep(time.Millisecond)

	err = user.ChangeStatus(UserStatusSuspended)
	assert.NoError(t, err)
	assert.Equal(t, UserStatusSuspended, user.Status)
	assert.Equal(t, originalVersion+1, user.Version)
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))

	// Test invalid status
	err = user.ChangeStatus(UserStatus("invalid"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user status")
}

func TestUser_StatusMethods(t *testing.T) {
	user, err := NewUser("test-user", "test@example.com", "Password123", "John", "Doe")
	require.NoError(t, err)

	// Test Activate
	err = user.Activate()
	assert.NoError(t, err)
	assert.Equal(t, UserStatusActive, user.Status)
	assert.True(t, user.IsActive())

	// Test Deactivate
	err = user.Deactivate()
	assert.NoError(t, err)
	assert.Equal(t, UserStatusInactive, user.Status)
	assert.False(t, user.IsActive())

	// Test Suspend
	err = user.Suspend()
	assert.NoError(t, err)
	assert.Equal(t, UserStatusSuspended, user.Status)
	assert.False(t, user.IsActive())
}

func TestUser_FullName(t *testing.T) {
	user, err := NewUser("test-user", "test@example.com", "Password123", "John", "Doe")
	require.NoError(t, err)

	assert.Equal(t, "John Doe", user.FullName())
}

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid user",
			user: &User{
				ID:        "test-user",
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Status:    UserStatusActive,
			},
			wantErr: false,
		},
		{
			name: "Empty ID",
			user: &User{
				ID:        "",
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Status:    UserStatusActive,
			},
			wantErr: true,
			errMsg:  "user ID cannot be empty",
		},
		{
			name: "Invalid status",
			user: &User{
				ID:        "test-user",
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Status:    UserStatus("invalid"),
			},
			wantErr: true,
			errMsg:  "invalid user status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
