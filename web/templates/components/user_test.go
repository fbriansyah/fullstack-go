package components

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

// TestUserProfile tests the UserProfile component rendering
func TestUserProfile(t *testing.T) {
	user := User{
		ID:        "user-123",
		Email:     "john.doe@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Status:    "active",
		CreatedAt: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
	}

	var buf bytes.Buffer
	err := UserProfile(user).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render UserProfile: %v", err)
	}

	html := buf.String()

	// Test that user information is displayed
	if !strings.Contains(html, "John Doe") {
		t.Error("Expected user full name to be displayed")
	}

	if !strings.Contains(html, "john.doe@example.com") {
		t.Error("Expected user email to be displayed")
	}

	if !strings.Contains(html, "user-123") {
		t.Error("Expected user ID to be displayed")
	}

	if !strings.Contains(html, "active") {
		t.Error("Expected user status to be displayed")
	}

	// Test that action buttons are present
	if !strings.Contains(html, "Edit Profile") {
		t.Error("Expected Edit Profile button to be present")
	}

	if !strings.Contains(html, "Account Settings") {
		t.Error("Expected Account Settings button to be present")
	}

	// Test that avatar initials are displayed
	if !strings.Contains(html, "JD") {
		t.Error("Expected user initials to be displayed in avatar")
	}
}

// TestUserDashboard tests the UserDashboard component rendering
func TestUserDashboard(t *testing.T) {
	user := User{
		ID:        "user-123",
		Email:     "jane.smith@example.com",
		FirstName: "Jane",
		LastName:  "Smith",
		Status:    "active",
		CreatedAt: time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
	}

	stats := DashboardStats{
		ProfileViews: "42",
		LastLogin:    time.Date(2023, 12, 1, 14, 30, 0, 0, time.UTC),
		RecentActivities: []Activity{
			{
				Type:        "login",
				Description: "Logged in from Chrome",
				Timestamp:   time.Date(2023, 12, 1, 14, 30, 0, 0, time.UTC),
			},
			{
				Type:        "profile_update",
				Description: "Updated profile information",
				Timestamp:   time.Date(2023, 11, 30, 10, 15, 0, 0, time.UTC),
			},
		},
	}

	var buf bytes.Buffer
	err := UserDashboard(user, stats).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render UserDashboard: %v", err)
	}

	html := buf.String()

	// Test welcome message
	if !strings.Contains(html, "Welcome back, Jane!") {
		t.Error("Expected welcome message with user's first name")
	}

	// Test stats display
	if !strings.Contains(html, "42") {
		t.Error("Expected profile views count to be displayed")
	}

	// Test recent activities
	if !strings.Contains(html, "Logged in from Chrome") {
		t.Error("Expected recent activity to be displayed")
	}

	if !strings.Contains(html, "Updated profile information") {
		t.Error("Expected recent activity to be displayed")
	}

	// Test quick links
	if !strings.Contains(html, "Profile Settings") {
		t.Error("Expected Profile Settings quick link")
	}

	if !strings.Contains(html, "Security") {
		t.Error("Expected Security quick link")
	}
}

// TestUserEditForm tests the UserEditForm component rendering
func TestUserEditForm(t *testing.T) {
	user := User{
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
	err := UserEditForm(user, errors).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render UserEditForm: %v", err)
	}

	html := buf.String()

	// Test form fields are pre-populated
	if !strings.Contains(html, `value="Test"`) {
		t.Error("Expected first name field to be pre-populated")
	}

	if !strings.Contains(html, `value="User"`) {
		t.Error("Expected last name field to be pre-populated")
	}

	if !strings.Contains(html, `value="test@example.com"`) {
		t.Error("Expected email field to be pre-populated")
	}

	// Test error message display
	if !strings.Contains(html, "Email is already taken") {
		t.Error("Expected error message to be displayed")
	}

	// Test form action
	if !strings.Contains(html, `action="/profile/edit"`) {
		t.Error("Expected form to have correct action URL")
	}

	// Test form buttons
	if !strings.Contains(html, "Save Changes") {
		t.Error("Expected Save Changes button")
	}

	if !strings.Contains(html, "Cancel") {
		t.Error("Expected Cancel button")
	}
}

// TestUserList tests the UserList component rendering
func TestUserList(t *testing.T) {
	users := []User{
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
	err := UserList(users, 1, 3).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render UserList: %v", err)
	}

	html := buf.String()

	// Test page header
	if !strings.Contains(html, "User Management") {
		t.Error("Expected page title to be displayed")
	}

	if !strings.Contains(html, "Add New User") {
		t.Error("Expected Add New User button")
	}

	// Test user data display
	if !strings.Contains(html, "Alice Johnson") {
		t.Error("Expected first user name to be displayed")
	}

	if !strings.Contains(html, "alice@example.com") {
		t.Error("Expected first user email to be displayed")
	}

	if !strings.Contains(html, "Bob Wilson") {
		t.Error("Expected second user name to be displayed")
	}

	if !strings.Contains(html, "bob@example.com") {
		t.Error("Expected second user email to be displayed")
	}

	// Test table structure
	if !strings.Contains(html, "<table") {
		t.Error("Expected table element")
	}

	if !strings.Contains(html, "<thead") {
		t.Error("Expected table header")
	}

	if !strings.Contains(html, "<tbody") {
		t.Error("Expected table body")
	}

	// Test pagination (should be present since totalPages > 1)
	if !strings.Contains(html, "page") {
		t.Error("Expected pagination to be present")
	}
}

// TestUserTableRow tests the UserTableRow component rendering
func TestUserTableRow(t *testing.T) {
	user := User{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Status:    "suspended",
		CreatedAt: time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
	}

	var buf bytes.Buffer
	err := UserTableRow(user).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render UserTableRow: %v", err)
	}

	html := buf.String()

	// Test user information display
	if !strings.Contains(html, "Test User") {
		t.Error("Expected user full name to be displayed")
	}

	if !strings.Contains(html, "test@example.com") {
		t.Error("Expected user email to be displayed")
	}

	// Test status badge
	if !strings.Contains(html, "suspended") {
		t.Error("Expected user status to be displayed")
	}

	// Test action links
	if !strings.Contains(html, "View") {
		t.Error("Expected View action link")
	}

	if !strings.Contains(html, "Edit") {
		t.Error("Expected Edit action link")
	}

	if !strings.Contains(html, "Delete") {
		t.Error("Expected Delete action link")
	}

	// Test action URLs
	if !strings.Contains(html, "/admin/users/user-123") {
		t.Error("Expected correct view URL")
	}

	if !strings.Contains(html, "/admin/users/user-123/edit") {
		t.Error("Expected correct edit URL")
	}

	// Test avatar initials
	if !strings.Contains(html, "TU") {
		t.Error("Expected user initials in avatar")
	}
}

// TestUserStatusBadge tests the UserStatusBadge component rendering
func TestUserStatusBadge(t *testing.T) {
	testCases := []struct {
		status   string
		expected string
	}{
		{"active", "bg-green-100 text-green-800"},
		{"inactive", "bg-yellow-100 text-yellow-800"},
		{"suspended", "bg-red-100 text-red-800"},
	}

	for _, tc := range testCases {
		t.Run(tc.status, func(t *testing.T) {
			var buf bytes.Buffer
			err := UserStatusBadge(tc.status).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render UserStatusBadge: %v", err)
			}

			html := buf.String()

			if !strings.Contains(html, tc.status) {
				t.Errorf("Expected status text '%s' to be displayed", tc.status)
			}

			if !strings.Contains(html, tc.expected) {
				t.Errorf("Expected CSS classes '%s' to be present", tc.expected)
			}
		})
	}
}

// TestStatsCard tests the StatsCard component rendering
func TestStatsCard(t *testing.T) {
	var buf bytes.Buffer
	err := StatsCard("Total Users", "150", "user", "blue").Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render StatsCard: %v", err)
	}

	html := buf.String()

	// Test content display
	if !strings.Contains(html, "Total Users") {
		t.Error("Expected title to be displayed")
	}

	if !strings.Contains(html, "150") {
		t.Error("Expected value to be displayed")
	}

	// Test styling
	if !strings.Contains(html, "bg-blue-100") {
		t.Error("Expected blue color styling")
	}
}

// TestActivityItem tests the ActivityItem component rendering
func TestActivityItem(t *testing.T) {
	activity := Activity{
		Type:        "login",
		Description: "Logged in from mobile device",
		Timestamp:   time.Date(2023, 12, 1, 15, 30, 0, 0, time.UTC),
	}

	var buf bytes.Buffer
	err := ActivityItem(activity).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render ActivityItem: %v", err)
	}

	html := buf.String()

	// Test activity description
	if !strings.Contains(html, "Logged in from mobile device") {
		t.Error("Expected activity description to be displayed")
	}

	// Test timestamp formatting
	if !strings.Contains(html, "Dec 1") {
		t.Error("Expected formatted timestamp to be displayed")
	}
}

// TestQuickLinkCard tests the QuickLinkCard component rendering
func TestQuickLinkCard(t *testing.T) {
	var buf bytes.Buffer
	err := QuickLinkCard("Account Settings", "/settings", "cog").Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render QuickLinkCard: %v", err)
	}

	html := buf.String()

	// Test link content
	if !strings.Contains(html, "Account Settings") {
		t.Error("Expected link title to be displayed")
	}

	// Test link URL
	if !strings.Contains(html, `href="/settings"`) {
		t.Error("Expected correct href attribute")
	}

	// Test hover effects
	if !strings.Contains(html, "hover:border-blue-300") {
		t.Error("Expected hover styling classes")
	}
}

// TestPagination tests the Pagination component rendering
func TestPagination(t *testing.T) {
	var buf bytes.Buffer
	err := Pagination(2, 5, "/admin/users").Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render Pagination: %v", err)
	}

	html := buf.String()

	// Test page information
	if !strings.Contains(html, "Showing page") && !strings.Contains(html, "2") && !strings.Contains(html, "5") {
		t.Error("Expected current page information to be displayed")
	}

	// Test navigation links
	if !strings.Contains(html, "Previous") {
		t.Error("Expected Previous link")
	}

	if !strings.Contains(html, "Next") {
		t.Error("Expected Next link")
	}

	// Test page number links
	if !strings.Contains(html, "/admin/users?page=1") {
		t.Error("Expected page 1 link")
	}

	if !strings.Contains(html, "/admin/users?page=3") {
		t.Error("Expected page 3 link")
	}
}

// TestEmptyUserList tests UserList component with no users
func TestEmptyUserList(t *testing.T) {
	users := []User{}

	var buf bytes.Buffer
	err := UserList(users, 1, 1).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render empty UserList: %v", err)
	}

	html := buf.String()

	// Should still show header
	if !strings.Contains(html, "User Management") {
		t.Error("Expected page title even with no users")
	}

	// Should show add user button
	if !strings.Contains(html, "Add New User") {
		t.Error("Expected Add New User button even with no users")
	}
}

// Benchmark tests
func BenchmarkUserProfile(b *testing.B) {
	user := User{
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
		UserProfile(user).Render(context.Background(), &buf)
	}
}

func BenchmarkUserList(b *testing.B) {
	users := make([]User, 50)
	for i := range users {
		users[i] = User{
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
		UserList(users, 1, 5).Render(context.Background(), &buf)
	}
}
