package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordValidator_Validate(t *testing.T) {
	validator := NewPasswordValidator()

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "valid password",
			password: "MySecure123!",
			wantErr:  nil,
		},
		{
			name:     "too short",
			password: "Short1!",
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "too long",
			password: "ThisPasswordIsWayTooLongAndExceedsTheMaximumLengthRequirementSetByTheSystemWhichIs128CharactersInTotalSoItShouldFailValidation123!",
			wantErr:  ErrPasswordTooLong,
		},
		{
			name:     "no uppercase",
			password: "mysecure123!",
			wantErr:  ErrPasswordTooWeak,
		},
		{
			name:     "no lowercase",
			password: "MYSECURE123!",
			wantErr:  ErrPasswordTooWeak,
		},
		{
			name:     "no numbers",
			password: "MySecurePass!",
			wantErr:  ErrPasswordTooWeak,
		},
		{
			name:     "no special characters",
			password: "MySecure123",
			wantErr:  ErrPasswordTooWeak,
		},
		{
			name:     "minimum valid password",
			password: "MyPass1!",
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.password)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPasswordValidator_CustomRules(t *testing.T) {
	validator := &PasswordValidator{
		MinLength:        6,
		MaxLength:        20,
		RequireUppercase: false,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSpecial:   false,
	}

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "valid with custom rules",
			password: "mypass123",
			wantErr:  nil,
		},
		{
			name:     "too short for custom rules",
			password: "my1",
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "no lowercase required",
			password: "MYPASS123",
			wantErr:  ErrPasswordTooWeak,
		},
		{
			name:     "no numbers required",
			password: "mypassword",
			wantErr:  ErrPasswordTooWeak,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.password)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPasswordHasher_Hash(t *testing.T) {
	hasher := NewPasswordHasher()
	password := "MySecurePassword123!"

	hash, err := hasher.Hash(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Hash should start with bcrypt prefix
	assert.Contains(t, hash, "$2a$")
}

func TestPasswordHasher_Verify(t *testing.T) {
	hasher := NewPasswordHasher()
	password := "MySecurePassword123!"

	hash, err := hasher.Hash(password)
	require.NoError(t, err)

	// Correct password should verify
	err = hasher.Verify(password, hash)
	assert.NoError(t, err)

	// Wrong password should not verify
	err = hasher.Verify("WrongPassword123!", hash)
	assert.ErrorIs(t, err, ErrInvalidPassword)

	// Empty password should not verify
	err = hasher.Verify("", hash)
	assert.ErrorIs(t, err, ErrInvalidPassword)
}

func TestPasswordHasher_HashAndVerify(t *testing.T) {
	hasher := NewPasswordHasher()
	passwords := []string{
		"SimplePass123!",
		"ComplexP@ssw0rd!",
		"AnotherSecure456#",
		"MyP@ssw0rd789$",
	}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			hash, err := hasher.Hash(password)
			require.NoError(t, err)

			err = hasher.Verify(password, hash)
			assert.NoError(t, err)
		})
	}
}

func TestIsCommonPassword(t *testing.T) {
	tests := []struct {
		password string
		expected bool
	}{
		{"password", true},
		{"123456", true},
		{"qwerty", true},
		{"admin", true},
		{"MySecurePassword123!", false},
		{"UniquePassword456#", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			result := IsCommonPassword(tt.password)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with numbers",
			email:   "user123@example123.com",
			wantErr: false,
		},
		{
			name:    "valid email with special characters",
			email:   "user.name+tag@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
		},
		{
			name:    "invalid email no @",
			email:   "userexample.com",
			wantErr: true,
		},
		{
			name:    "invalid email no domain",
			email:   "user@",
			wantErr: true,
		},
		{
			name:    "invalid email no user",
			email:   "@example.com",
			wantErr: true,
		},
		{
			name:    "invalid email no TLD",
			email:   "user@example",
			wantErr: true,
		},
		{
			name:    "invalid email multiple @",
			email:   "user@@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPasswordHasher_DifferentHashesForSamePassword(t *testing.T) {
	hasher := NewPasswordHasher()
	password := "MySecurePassword123!"

	hash1, err1 := hasher.Hash(password)
	hash2, err2 := hasher.Hash(password)

	require.NoError(t, err1)
	require.NoError(t, err2)

	// Hashes should be different due to salt
	assert.NotEqual(t, hash1, hash2)

	// Both should verify correctly
	assert.NoError(t, hasher.Verify(password, hash1))
	assert.NoError(t, hasher.Verify(password, hash2))
}
