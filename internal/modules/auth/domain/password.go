package domain

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// MinPasswordLength defines the minimum password length
	MinPasswordLength = 8
	// MaxPasswordLength defines the maximum password length
	MaxPasswordLength = 128
	// BcryptCost defines the cost for bcrypt hashing
	BcryptCost = 12
)

var (
	// ErrPasswordTooShort indicates password is too short
	ErrPasswordTooShort = errors.New("password must be at least 8 characters long")
	// ErrPasswordTooLong indicates password is too long
	ErrPasswordTooLong = errors.New("password must be no more than 128 characters long")
	// ErrPasswordTooWeak indicates password doesn't meet complexity requirements
	ErrPasswordTooWeak = errors.New("password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	// ErrInvalidPassword indicates password validation failed
	ErrInvalidPassword = errors.New("invalid password")
)

// PasswordValidator provides password validation functionality
type PasswordValidator struct {
	MinLength        int
	MaxLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumbers   bool
	RequireSpecial   bool
}

// NewPasswordValidator creates a new password validator with default rules
func NewPasswordValidator() *PasswordValidator {
	return &PasswordValidator{
		MinLength:        MinPasswordLength,
		MaxLength:        MaxPasswordLength,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSpecial:   true,
	}
}

// Validate validates a password against the configured rules
func (pv *PasswordValidator) Validate(password string) error {
	if len(password) < pv.MinLength {
		return ErrPasswordTooShort
	}

	if len(password) > pv.MaxLength {
		return ErrPasswordTooLong
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if pv.RequireUppercase && !hasUpper {
		return ErrPasswordTooWeak
	}
	if pv.RequireLowercase && !hasLower {
		return ErrPasswordTooWeak
	}
	if pv.RequireNumbers && !hasNumber {
		return ErrPasswordTooWeak
	}
	if pv.RequireSpecial && !hasSpecial {
		return ErrPasswordTooWeak
	}

	return nil
}

// PasswordHasher provides password hashing functionality
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a new password hasher
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		cost: BcryptCost,
	}
}

// Hash hashes a password using bcrypt
func (ph *PasswordHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), ph.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// Verify verifies a password against its hash
func (ph *PasswordHasher) Verify(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return fmt.Errorf("failed to verify password: %w", err)
	}
	return nil
}

// IsCommonPassword checks if the password is in a list of common passwords
func IsCommonPassword(password string) bool {
	// List of common passwords to reject
	commonPasswords := []string{
		"password", "123456", "123456789", "12345678", "12345",
		"1234567", "password123", "admin", "qwerty", "abc123",
		"letmein", "monkey", "1234567890", "dragon", "111111",
		"baseball", "iloveyou", "trustno1", "1234", "sunshine",
	}

	return slices.Contains(commonPasswords, password)
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	// Basic email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}
