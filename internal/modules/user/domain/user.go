package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserStatus represents the status of a user account
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

// IsValid checks if the UserStatus is valid
func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusActive, UserStatusInactive, UserStatusSuspended:
		return true
	default:
		return false
	}
}

// String returns the string representation of UserStatus
func (s UserStatus) String() string {
	return string(s)
}

// User represents the user aggregate root
type User struct {
	ID        string     `db:"id" json:"id"`
	Email     string     `db:"email" json:"email"`
	Password  string     `db:"password" json:"-"`
	FirstName string     `db:"first_name" json:"first_name"`
	LastName  string     `db:"last_name" json:"last_name"`
	Status    UserStatus `db:"status" json:"status"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	Version   int        `db:"version" json:"version"` // Optimistic locking
}

// NewUser creates a new User aggregate with validation
func NewUser(id, email, password, firstName, lastName string) (*User, error) {
	user := &User{
		ID:        id,
		Email:     strings.ToLower(strings.TrimSpace(email)),
		FirstName: strings.TrimSpace(firstName),
		LastName:  strings.TrimSpace(lastName),
		Status:    UserStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Version:   1,
	}

	if err := user.SetPassword(password); err != nil {
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	return user, nil
}

// Validate performs comprehensive validation of the User aggregate
func (u *User) Validate() error {
	if u.ID == "" {
		return errors.New("user ID cannot be empty")
	}

	if err := u.validateEmail(); err != nil {
		return err
	}

	if err := u.validateName(); err != nil {
		return err
	}

	if !u.Status.IsValid() {
		return fmt.Errorf("invalid user status: %s", u.Status)
	}

	return nil
}

// validateEmail validates the email format and requirements
func (u *User) validateEmail() error {
	if u.Email == "" {
		return errors.New("email cannot be empty")
	}

	// Basic email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return errors.New("invalid email format")
	}

	if len(u.Email) > 255 {
		return errors.New("email cannot exceed 255 characters")
	}

	return nil
}

// validateName validates first and last name requirements
func (u *User) validateName() error {
	if u.FirstName == "" {
		return errors.New("first name cannot be empty")
	}

	if u.LastName == "" {
		return errors.New("last name cannot be empty")
	}

	if len(u.FirstName) > 100 {
		return errors.New("first name cannot exceed 100 characters")
	}

	if len(u.LastName) > 100 {
		return errors.New("last name cannot exceed 100 characters")
	}

	// Check for valid characters (letters, spaces, hyphens, apostrophes)
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !nameRegex.MatchString(u.FirstName) {
		return errors.New("first name contains invalid characters")
	}

	if !nameRegex.MatchString(u.LastName) {
		return errors.New("last name contains invalid characters")
	}

	return nil
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	if err := validatePassword(password); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u.Password = string(hashedPassword)
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// CheckPassword verifies if the provided password matches the user's password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// UpdateProfile updates the user's profile information
func (u *User) UpdateProfile(firstName, lastName string) error {
	u.FirstName = strings.TrimSpace(firstName)
	u.LastName = strings.TrimSpace(lastName)
	u.UpdatedAt = time.Now().UTC()
	u.Version++

	return u.validateName()
}

// UpdateEmail updates the user's email address
func (u *User) UpdateEmail(email string) error {
	u.Email = strings.ToLower(strings.TrimSpace(email))
	u.UpdatedAt = time.Now().UTC()
	u.Version++

	return u.validateEmail()
}

// ChangeStatus changes the user's status
func (u *User) ChangeStatus(status UserStatus) error {
	if !status.IsValid() {
		return fmt.Errorf("invalid user status: %s", status)
	}

	u.Status = status
	u.UpdatedAt = time.Now().UTC()
	u.Version++

	return nil
}

// Activate sets the user status to active
func (u *User) Activate() error {
	return u.ChangeStatus(UserStatusActive)
}

// Deactivate sets the user status to inactive
func (u *User) Deactivate() error {
	return u.ChangeStatus(UserStatusInactive)
}

// Suspend sets the user status to suspended
func (u *User) Suspend() error {
	return u.ChangeStatus(UserStatusSuspended)
}

// IsActive returns true if the user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// FullName returns the user's full name
func (u *User) FullName() string {
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}

// validatePassword validates password requirements
func validatePassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return errors.New("password cannot exceed 128 characters")
	}

	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	return nil
}
