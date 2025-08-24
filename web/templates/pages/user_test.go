package pages

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"go-templ-template/web/templates/components"
)

// TestUserProfilePage tests the UserProfilePage component rendering
func TestUserProfilePage(t *testing.T) {
	user := components.User{
		ID:        "user-123",
		Email:     "john.doe@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Status:    "active",
		CreatedAt: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
	}

	var buf bytes.Buffer
	err := UserProfilePage(user).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render UserProfilePage: %v", err)
	}

	html := buf.String()

	// Test page structure (UserProfilePage should render full HTML document)
	if !strings.Contains(html, "<!doctype html>") && !strings.Contains(html, "<!DOCTYPE html>") {
		t.Errorf("Expected HTML document structure. Got: %s", html[:200])
	}

	// Test page title
	if !strings.Contains(html, "Profile - John Doe") {
		t.Error("Expected page title with user name")
	}

	// Test breadcrumb navigation
	if !strings.Contains(html, "Dashboard") {
		t.Error("Expected Dashboard breadcrumb")
	}

	if !strings.Contains(html, "Profile") {
		t.Error("Expected Profile breadcrumb")
	}

	// Test user profile content
	if !strings.Contains(html, "John Doe") {
		t.Error("Expected user name to be displayed")
	}

	if !strings.Contains(html, "john.doe@example.com") {
		t.Error("Expected user email to be displayed")
	}
}

// TestUserDashboardPage tests the UserDashboardPage component rendering
func TestUserDashboardPage(t *testing.T) {
	user := components.User{
		ID:        "user-123",
		Email:     "jane.smith@example.com",
		FirstName: "Jane",
		LastName:  "Smith",
		Status:    "active",
		CreatedAt: time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
	}

	stats := components.DashboardStats{
		ProfileViews: "42",
		LastLogin:    time.Date(2023, 12, 1, 14, 30, 0, 0, time.UTC),
		RecentActivities: []components.Activity{
			{
				Type:        "login",
				Description: "Logged in from Chrome",
				Timestamp:   time.Date(2023, 12, 1, 14, 30, 0, 0, time.UTC),
			},
		},
	}

	var buf bytes.Buffer
	err := UserDashboardPage(user, stats).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render UserDashboardPage: %v", err)
	}

	html := buf.String()

	// Test page title
	if !strings.Contains(html, "Dashboard - Jane") {
		t.Error("Expected page title with user first name")
	}

	// Test dashboard content
	if !strings.Contains(html, "Welcome back, Jane!") {
		t.Error("Expected welcome message")
	}

	// Test stats display
	if !strings.Contains(html, "42") {
		t.Error("Expected profile views to be displayed")
	}

	// Test recent activities
	if !strings.Contains(html, "Logged in from Chrome") {
		t.Error("Expected recent activity to be displayed")
	}
}

// TestUserEditPage tests the UserEditPage component rendering
func TestUserEditPage(t *testing.T) {
	user := components.User{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	errors := map[string]string{
		"email": "Email is already taken",
	}

	var buf bytes.Buffer
	err := UserEditPage(user, errors).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render UserEditPage: %v", err)
	}

	html := buf.String()

	// Test page title
	if !strings.Contains(html, "Edit Profile - Test") {
		t.Error("Expected page title with user first name")
	}

	// Test breadcrumb navigation
	if !strings.Contains(html, "Dashboard") {
		t.Error("Expected Dashboard breadcrumb")
	}

	if !strings.Contains(html, "Profile") {
		t.Error("Expected Profile breadcrumb")
	}

	if !strings.Contains(html, "Edit") {
		t.Error("Expected Edit breadcrumb")
	}

	// Test form content
	if !strings.Contains(html, `value="Test"`) {
		t.Error("Expected first name field to be pre-populated")
	}

	if !strings.Contains(html, `value="test@example.com"`) {
		t.Error("Expected email field to be pre-populated")
	}

	// Test error display
	if !strings.Contains(html, "Email is already taken") {
		t.Error("Expected error message to be displayed")
	}
}

// TestUserListPage tests the UserListPage component rendering
func TestUserListPage(t *testing.T) {
	users := []components.User{
		{
			ID:        "user-1",
			Email:     "alice@example.com",
			FirstName: "Alice",
			LastName:  "Johnson",
			Status:    "active",
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:        "user-2",
			Email:     "bob@example.com",
			FirstName: "Bob",
			LastName:  "Wilson",
			Status:    "inactive",
			CreatedAt: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	var buf bytes.Buffer
	err := UserListPage(users, 1, 3, "alice").Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render UserListPage: %v", err)
	}

	html := buf.String()

	// Test page title
	if !strings.Contains(html, "User Management") {
		t.Error("Expected page title")
	}

	// Test breadcrumb navigation
	if !strings.Contains(html, "Dashboard") {
		t.Error("Expected Dashboard breadcrumb")
	}

	if !strings.Contains(html, "Administration") {
		t.Error("Expected Administration breadcrumb")
	}

	if !strings.Contains(html, "Users") {
		t.Error("Expected Users breadcrumb")
	}

	// Test search functionality
	if !strings.Contains(html, "Search users...") {
		t.Error("Expected search box")
	}

	if !strings.Contains(html, `value="alice"`) {
		t.Error("Expected search query to be pre-populated")
	}

	// Test add user button
	if !strings.Contains(html, "Add User") {
		t.Error("Expected Add User button")
	}

	// Test user data
	if !strings.Contains(html, "Alice Johnson") {
		t.Error("Expected user name to be displayed")
	}

	if !strings.Contains(html, "alice@example.com") {
		t.Error("Expected user email to be displayed")
	}

	// Test confirmation modal
	if !strings.Contains(html, "confirmation-modal") {
		t.Error("Expected confirmation modal")
	}
}

// TestUserListPageEmpty tests UserListPage with no users
func TestUserListPageEmpty(t *testing.T) {
	users := []components.User{}

	var buf bytes.Buffer
	err := UserListPage(users, 1, 1, "").Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render empty UserListPage: %v", err)
	}

	html := buf.String()

	// Test page title is still present
	if !strings.Contains(html, "User Management") {
		t.Error("Expected page title even with no users")
	}

	// Test empty state
	if !strings.Contains(html, "No users found") {
		t.Error("Expected empty state message")
	}

	if !strings.Contains(html, "Add New User") {
		t.Error("Expected empty state action button")
	}
}

// TestUserSettingsPage tests the UserSettingsPage component rendering
func TestUserSettingsPage(t *testing.T) {
	user := components.User{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	var buf bytes.Buffer
	err := UserSettingsPage(user, "profile").Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render UserSettingsPage: %v", err)
	}

	html := buf.String()

	// Test page title
	if !strings.Contains(html, "Account Settings - Test") {
		t.Error("Expected page title with user first name")
	}

	// Test breadcrumb navigation
	if !strings.Contains(html, "Dashboard") {
		t.Error("Expected Dashboard breadcrumb")
	}

	if !strings.Contains(html, "Profile") {
		t.Error("Expected Profile breadcrumb")
	}

	if !strings.Contains(html, "Settings") {
		t.Error("Expected Settings breadcrumb")
	}

	// Test tabs
	if !strings.Contains(html, "Profile") {
		t.Error("Expected Profile tab")
	}

	if !strings.Contains(html, "Security") {
		t.Error("Expected Security tab")
	}

	if !strings.Contains(html, "Notifications") {
		t.Error("Expected Notifications tab")
	}

	if !strings.Contains(html, "Privacy") {
		t.Error("Expected Privacy tab")
	}

	// Test active tab styling (profile should be active)
	if !strings.Contains(html, "border-blue-500 text-blue-600") {
		t.Error("Expected active tab styling for profile tab")
	}
}

// TestProfileSettingsTab tests the ProfileSettingsTab component rendering
func TestProfileSettingsTab(t *testing.T) {
	user := components.User{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	var buf bytes.Buffer
	err := ProfileSettingsTab(user).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render ProfileSettingsTab: %v", err)
	}

	html := buf.String()

	// Test form fields
	if !strings.Contains(html, "Profile Information") {
		t.Error("Expected Profile Information section")
	}

	if !strings.Contains(html, `value="Test"`) {
		t.Error("Expected first name field to be pre-populated")
	}

	if !strings.Contains(html, `value="User"`) {
		t.Error("Expected last name field to be pre-populated")
	}

	if !strings.Contains(html, `value="test@example.com"`) {
		t.Error("Expected email field to be pre-populated")
	}

	// Test profile picture section
	if !strings.Contains(html, "Profile Picture") {
		t.Error("Expected Profile Picture section")
	}

	if !strings.Contains(html, "Change Picture") {
		t.Error("Expected Change Picture button")
	}

	// Test avatar initials
	if !strings.Contains(html, "TU") {
		t.Error("Expected user initials in avatar")
	}

	// Test form action
	if !strings.Contains(html, `action="/profile/settings/profile"`) {
		t.Error("Expected correct form action")
	}
}

// TestSecuritySettingsTab tests the SecuritySettingsTab component rendering
func TestSecuritySettingsTab(t *testing.T) {
	user := components.User{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	var buf bytes.Buffer
	err := SecuritySettingsTab(user).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render SecuritySettingsTab: %v", err)
	}

	html := buf.String()

	// Test password change section
	if !strings.Contains(html, "Change Password") {
		t.Error("Expected Change Password section")
	}

	if !strings.Contains(html, "Current Password") {
		t.Error("Expected Current Password field")
	}

	if !strings.Contains(html, "New Password") {
		t.Error("Expected New Password field")
	}

	if !strings.Contains(html, "Confirm New Password") {
		t.Error("Expected Confirm New Password field")
	}

	// Test 2FA section
	if !strings.Contains(html, "Two-Factor Authentication") {
		t.Error("Expected Two-Factor Authentication section")
	}

	if !strings.Contains(html, "Enable 2FA") {
		t.Error("Expected Enable 2FA button")
	}

	// Test active sessions section
	if !strings.Contains(html, "Active Sessions") {
		t.Error("Expected Active Sessions section")
	}

	if !strings.Contains(html, "Current Session") {
		t.Error("Expected Current Session display")
	}

	// Test form actions
	if !strings.Contains(html, `action="/profile/settings/password"`) {
		t.Error("Expected correct password form action")
	}
}

// TestNotificationSettingsTab tests the NotificationSettingsTab component rendering
func TestNotificationSettingsTab(t *testing.T) {
	user := components.User{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	var buf bytes.Buffer
	err := NotificationSettingsTab(user).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render NotificationSettingsTab: %v", err)
	}

	html := buf.String()

	// Test notification preferences
	if !strings.Contains(html, "Notification Preferences") {
		t.Error("Expected Notification Preferences title")
	}

	// Test email notifications section
	if !strings.Contains(html, "Email Notifications") {
		t.Error("Expected Email Notifications section")
	}

	if !strings.Contains(html, "Account updates and security alerts") {
		t.Error("Expected account updates checkbox")
	}

	if !strings.Contains(html, "Marketing and promotional emails") {
		t.Error("Expected marketing emails checkbox")
	}

	// Test push notifications section
	if !strings.Contains(html, "Push Notifications") {
		t.Error("Expected Push Notifications section")
	}

	if !strings.Contains(html, "Important account updates") {
		t.Error("Expected important updates checkbox")
	}

	// Test form action
	if !strings.Contains(html, `action="/profile/settings/notifications"`) {
		t.Error("Expected correct form action")
	}

	// Test save button
	if !strings.Contains(html, "Save Preferences") {
		t.Error("Expected Save Preferences button")
	}
}

// TestPrivacySettingsTab tests the PrivacySettingsTab component rendering
func TestPrivacySettingsTab(t *testing.T) {
	user := components.User{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	var buf bytes.Buffer
	err := PrivacySettingsTab(user).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render PrivacySettingsTab: %v", err)
	}

	html := buf.String()

	// Test privacy settings section
	if !strings.Contains(html, "Privacy Settings") {
		t.Error("Expected Privacy Settings section")
	}

	if !strings.Contains(html, "Make my profile visible") {
		t.Error("Expected profile visibility checkbox")
	}

	if !strings.Contains(html, "Allow activity tracking") {
		t.Error("Expected activity tracking checkbox")
	}

	// Test data export section
	if !strings.Contains(html, "Data Export") {
		t.Error("Expected Data Export section")
	}

	if !strings.Contains(html, "Request Data Export") {
		t.Error("Expected Request Data Export button")
	}

	// Test account deletion section
	if !strings.Contains(html, "Delete Account") {
		t.Error("Expected Delete Account section")
	}

	if !strings.Contains(html, "Permanently delete your account") {
		t.Error("Expected account deletion warning")
	}

	if !strings.Contains(html, "btn-danger") {
		t.Error("Expected danger button styling for delete account")
	}

	// Test form actions
	if !strings.Contains(html, `action="/profile/settings/privacy"`) {
		t.Error("Expected correct privacy form action")
	}

	// Test delete account button ID
	if !strings.Contains(html, "delete-account-btn") {
		t.Error("Expected delete account button with ID")
	}
}

// Benchmark tests
func BenchmarkUserProfilePage(b *testing.B) {
	user := components.User{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		UserProfilePage(user).Render(context.Background(), &buf)
	}
}

func BenchmarkUserDashboardPage(b *testing.B) {
	user := components.User{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	stats := components.DashboardStats{
		ProfileViews: "42",
		LastLogin:    time.Now(),
		RecentActivities: []components.Activity{
			{
				Type:        "login",
				Description: "Test activity",
				Timestamp:   time.Now(),
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		UserDashboardPage(user, stats).Render(context.Background(), &buf)
	}
}

func BenchmarkUserListPage(b *testing.B) {
	users := make([]components.User, 50)
	for i := range users {
		users[i] = components.User{
			ID:        "user-" + string(rune(i)),
			Email:     "user" + string(rune(i)) + "@example.com",
			FirstName: "User",
			LastName:  string(rune(i)),
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		UserListPage(users, 1, 5, "").Render(context.Background(), &buf)
	}
}
