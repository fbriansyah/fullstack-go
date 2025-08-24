package components

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// TestIcon tests the Icon component rendering
func TestIcon(t *testing.T) {
	testCases := []struct {
		name     string
		color    string
		expected string
	}{
		{"user", "blue", "text-blue-600"},
		{"calendar", "green", "text-green-600"},
		{"eye", "purple", "text-purple-600"},
		{"clock", "yellow", "text-yellow-600"},
		{"cog", "gray", "text-gray-600"},
		{"shield", "red", "text-red-600"},
	}

	for _, tc := range testCases {
		t.Run(tc.name+"_"+tc.color, func(t *testing.T) {
			var buf bytes.Buffer
			err := Icon(tc.name, tc.color).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render Icon: %v", err)
			}

			html := buf.String()

			// Test SVG element
			if !strings.Contains(html, "<svg") {
				t.Error("Expected SVG element")
			}

			// Test color class
			if !strings.Contains(html, tc.expected) {
				t.Errorf("Expected color class '%s' to be present", tc.expected)
			}

			// Test viewBox
			if !strings.Contains(html, `viewBox="0 0 20 20"`) {
				t.Error("Expected correct viewBox attribute")
			}
		})
	}
}

// TestCard tests the Card component rendering
func TestCard(t *testing.T) {
	// Create a simple content component for testing
	content := SimpleContent("Test content")

	var buf bytes.Buffer
	err := Card("Test Title", content).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render Card: %v", err)
	}

	html := buf.String()

	// Test card styling
	if !strings.Contains(html, "card") {
		t.Error("Expected card CSS class")
	}

	// Test title display
	if !strings.Contains(html, "Test Title") {
		t.Error("Expected title to be displayed")
	}

	// Test content display
	if !strings.Contains(html, "Test content") {
		t.Error("Expected content to be displayed")
	}
}

// TestCardWithoutTitle tests Card component without title
func TestCardWithoutTitle(t *testing.T) {
	content := SimpleContent("Content only")

	var buf bytes.Buffer
	err := Card("", content).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render Card without title: %v", err)
	}

	html := buf.String()

	// Test that content is still displayed
	if !strings.Contains(html, "Content only") {
		t.Error("Expected content to be displayed")
	}

	// Test that no title element is rendered
	if strings.Contains(html, "<h3") {
		t.Error("Expected no title element when title is empty")
	}
}

// TestButton tests the Button component rendering
func TestButton(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		variant  string
		href     string
		expected string
	}{
		{"primary_link", "Click Me", "primary", "/test", "btn-primary"},
		{"secondary_button", "Submit", "secondary", "", "btn-secondary"},
		{"danger_button", "Delete", "danger", "", "bg-red-600"},
		{"outline_button", "Cancel", "outline", "", "border-gray-300"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Button(tc.text, tc.variant, tc.href).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render Button: %v", err)
			}

			html := buf.String()

			// Test button text
			if !strings.Contains(html, tc.text) {
				t.Errorf("Expected button text '%s' to be displayed", tc.text)
			}

			// Test styling
			if !strings.Contains(html, tc.expected) {
				t.Errorf("Expected CSS class '%s' to be present", tc.expected)
			}

			// Test href vs button
			if tc.href != "" {
				if !strings.Contains(html, "<a") {
					t.Error("Expected anchor element for button with href")
				}
				if !strings.Contains(html, tc.href) {
					t.Errorf("Expected href '%s' to be present", tc.href)
				}
			} else {
				if !strings.Contains(html, "<button") {
					t.Error("Expected button element")
				}
			}
		})
	}
}

// TestModal tests the Modal component rendering
func TestModal(t *testing.T) {
	content := SimpleContent("Modal content")

	var buf bytes.Buffer
	err := Modal("test-modal", "Test Modal", content, true).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render Modal: %v", err)
	}

	html := buf.String()

	// Test modal structure
	if !strings.Contains(html, `id="test-modal"`) {
		t.Error("Expected modal ID to be set")
	}

	if !strings.Contains(html, "Test Modal") {
		t.Error("Expected modal title to be displayed")
	}

	if !strings.Contains(html, "Modal content") {
		t.Error("Expected modal content to be displayed")
	}

	// Test modal classes
	if !strings.Contains(html, "fixed inset-0") {
		t.Error("Expected modal overlay classes")
	}

	if !strings.Contains(html, "hidden") {
		t.Error("Expected modal to be hidden by default")
	}

	// Test footer (should be present since showFooter is true)
	if !strings.Contains(html, "Confirm") {
		t.Error("Expected Confirm button in footer")
	}

	if !strings.Contains(html, "Cancel") {
		t.Error("Expected Cancel button in footer")
	}
}

// TestModalWithoutFooter tests Modal component without footer
func TestModalWithoutFooter(t *testing.T) {
	content := SimpleContent("Simple modal")

	var buf bytes.Buffer
	err := Modal("simple-modal", "Simple Modal", content, false).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render Modal without footer: %v", err)
	}

	html := buf.String()

	// Test that content is displayed
	if !strings.Contains(html, "Simple modal") {
		t.Error("Expected modal content to be displayed")
	}

	// Test that footer is not present
	if strings.Contains(html, "Confirm") {
		t.Error("Expected no Confirm button when showFooter is false")
	}
}

// TestSearchBox tests the SearchBox component rendering
func TestSearchBox(t *testing.T) {
	var buf bytes.Buffer
	err := SearchBox("Search users...", "john", "search").Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render SearchBox: %v", err)
	}

	html := buf.String()

	// Test input element
	if !strings.Contains(html, `type="text"`) {
		t.Error("Expected text input element")
	}

	// Test placeholder
	if !strings.Contains(html, `placeholder="Search users..."`) {
		t.Error("Expected placeholder attribute")
	}

	// Test value
	if !strings.Contains(html, `value="john"`) {
		t.Error("Expected value attribute")
	}

	// Test name attribute
	if !strings.Contains(html, `name="search"`) {
		t.Error("Expected name attribute")
	}

	// Test search icon
	if !strings.Contains(html, "search") {
		t.Error("Expected search icon")
	}

	// Test styling
	if !strings.Contains(html, "form-input") {
		t.Error("Expected form-input CSS class")
	}
}

// TestEmptyState tests the EmptyState component rendering
func TestEmptyState(t *testing.T) {
	var buf bytes.Buffer
	err := EmptyState("No Data", "There's nothing to show here.", "Add Item", "/add").Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render EmptyState: %v", err)
	}

	html := buf.String()

	// Test title and description
	if !strings.Contains(html, "No Data") {
		t.Error("Expected title to be displayed")
	}

	if !strings.Contains(html, "There&#39;s nothing to show here.") {
		t.Error("Expected description to be displayed")
	}

	// Test action button
	if !strings.Contains(html, "Add Item") {
		t.Error("Expected action text to be displayed")
	}

	if !strings.Contains(html, `href="/add"`) {
		t.Error("Expected action href to be present")
	}

	// Test styling
	if !strings.Contains(html, "text-center") {
		t.Error("Expected centered styling")
	}
}

// TestEmptyStateWithoutAction tests EmptyState component without action
func TestEmptyStateWithoutAction(t *testing.T) {
	var buf bytes.Buffer
	err := EmptyState("No Data", "Nothing here.", "", "").Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render EmptyState without action: %v", err)
	}

	html := buf.String()

	// Test that title and description are displayed
	if !strings.Contains(html, "No Data") {
		t.Error("Expected title to be displayed")
	}

	// Test that no action button is rendered
	if strings.Contains(html, "btn-primary") {
		t.Error("Expected no action button when action text is empty")
	}
}

// TestLoadingSpinner tests the LoadingSpinner component rendering
func TestLoadingSpinner(t *testing.T) {
	testCases := []struct {
		size     string
		expected string
	}{
		{"sm", "h-4 w-4"},
		{"md", "h-8 w-8"},
		{"lg", "h-12 w-12"},
	}

	for _, tc := range testCases {
		t.Run(tc.size, func(t *testing.T) {
			var buf bytes.Buffer
			err := LoadingSpinner(tc.size).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render LoadingSpinner: %v", err)
			}

			html := buf.String()

			// Test size classes
			if !strings.Contains(html, tc.expected) {
				t.Errorf("Expected size classes '%s' to be present", tc.expected)
			}

			// Test animation class
			if !strings.Contains(html, "animate-spin") {
				t.Error("Expected animate-spin class")
			}

			// Test border styling
			if !strings.Contains(html, "border-blue-600") {
				t.Error("Expected border color class")
			}
		})
	}
}

// TestBadge tests the Badge component rendering
func TestBadge(t *testing.T) {
	testCases := []struct {
		variant  string
		expected string
	}{
		{"success", "bg-green-100 text-green-800"},
		{"error", "bg-red-100 text-red-800"},
		{"warning", "bg-yellow-100 text-yellow-800"},
		{"info", "bg-blue-100 text-blue-800"},
		{"default", "bg-gray-100 text-gray-800"},
	}

	for _, tc := range testCases {
		t.Run(tc.variant, func(t *testing.T) {
			var buf bytes.Buffer
			err := Badge("Test Badge", tc.variant).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render Badge: %v", err)
			}

			html := buf.String()

			// Test badge text
			if !strings.Contains(html, "Test Badge") {
				t.Error("Expected badge text to be displayed")
			}

			// Test variant styling
			if !strings.Contains(html, tc.expected) {
				t.Errorf("Expected variant classes '%s' to be present", tc.expected)
			}

			// Test base styling
			if !strings.Contains(html, "rounded-full") {
				t.Error("Expected rounded-full class")
			}
		})
	}
}

// TestBreadcrumb tests the Breadcrumb component rendering
func TestBreadcrumb(t *testing.T) {
	items := []BreadcrumbItem{
		{Text: "Home", Href: "/"},
		{Text: "Users", Href: "/users"},
		{Text: "Profile", Href: ""},
	}

	var buf bytes.Buffer
	err := Breadcrumb(items).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render Breadcrumb: %v", err)
	}

	html := buf.String()

	// Test breadcrumb structure
	if !strings.Contains(html, "<nav") {
		t.Error("Expected nav element")
	}

	if !strings.Contains(html, `aria-label="Breadcrumb"`) {
		t.Error("Expected aria-label attribute")
	}

	// Test breadcrumb items
	if !strings.Contains(html, "Home") {
		t.Error("Expected Home breadcrumb item")
	}

	if !strings.Contains(html, "Users") {
		t.Error("Expected Users breadcrumb item")
	}

	if !strings.Contains(html, "Profile") {
		t.Error("Expected Profile breadcrumb item")
	}

	// Test links vs text
	if !strings.Contains(html, `href="/"`) {
		t.Error("Expected Home link")
	}

	if !strings.Contains(html, `href="/users"`) {
		t.Error("Expected Users link")
	}

	// Profile should not be a link (empty href)
	if strings.Contains(html, `href=""`) {
		t.Error("Expected Profile to not be a link")
	}
}

// TestTabs tests the Tabs component rendering
func TestTabs(t *testing.T) {
	tabs := []Tab{
		{ID: "profile", Text: "Profile", Href: "/profile"},
		{ID: "settings", Text: "Settings", Href: "/settings"},
		{ID: "security", Text: "Security", Href: "/security"},
	}

	var buf bytes.Buffer
	err := Tabs(tabs, "settings").Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render Tabs: %v", err)
	}

	html := buf.String()

	// Test tab structure
	if !strings.Contains(html, `aria-label="Tabs"`) {
		t.Error("Expected aria-label attribute")
	}

	// Test tab items
	if !strings.Contains(html, "Profile") {
		t.Error("Expected Profile tab")
	}

	if !strings.Contains(html, "Settings") {
		t.Error("Expected Settings tab")
	}

	if !strings.Contains(html, "Security") {
		t.Error("Expected Security tab")
	}

	// Test active tab styling (settings should be active)
	if !strings.Contains(html, "border-blue-500 text-blue-600") {
		t.Error("Expected active tab styling")
	}

	// Test tab links
	if !strings.Contains(html, `href="/profile"`) {
		t.Error("Expected Profile tab link")
	}

	if !strings.Contains(html, `href="/settings"`) {
		t.Error("Expected Settings tab link")
	}
}

// TestConfirmationDialog tests the ConfirmationDialog component rendering
func TestConfirmationDialog(t *testing.T) {
	var buf bytes.Buffer
	err := ConfirmationDialog().Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render ConfirmationDialog: %v", err)
	}

	html := buf.String()

	// Test modal structure
	if !strings.Contains(html, `id="confirmation-modal"`) {
		t.Error("Expected confirmation modal ID")
	}

	if !strings.Contains(html, "Confirm Action") {
		t.Error("Expected modal title")
	}

	// Test confirmation content
	if !strings.Contains(html, "Are you sure") {
		t.Error("Expected confirmation message")
	}

	// Test buttons
	if !strings.Contains(html, "Confirm") {
		t.Error("Expected Confirm button")
	}

	if !strings.Contains(html, "Cancel") {
		t.Error("Expected Cancel button")
	}

	// Test trash icon (check for the SVG path that represents trash icon)
	if !strings.Contains(html, "M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9z") {
		t.Error("Expected trash icon SVG path")
	}
}

// Benchmark tests
func BenchmarkIcon(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		Icon("user", "blue").Render(context.Background(), &buf)
	}
}

func BenchmarkButton(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		Button("Click Me", "primary", "/test").Render(context.Background(), &buf)
	}
}

// TestResponsiveGrid tests the ResponsiveGrid component rendering
func TestResponsiveGrid(t *testing.T) {
	content := SimpleContent("Grid content")

	testCases := []struct {
		columns  string
		expected string
	}{
		{"2", "grid-cols-1 md:grid-cols-2"},
		{"3", "grid-cols-1 md:grid-cols-2 lg:grid-cols-3"},
		{"4", "grid-cols-1 md:grid-cols-2 lg:grid-cols-4"},
		{"6", "grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6"},
	}

	for _, tc := range testCases {
		t.Run(tc.columns+"_columns", func(t *testing.T) {
			var buf bytes.Buffer
			err := ResponsiveGrid(tc.columns, "4", content).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render ResponsiveGrid: %v", err)
			}

			html := buf.String()

			// Test grid classes
			if !strings.Contains(html, tc.expected) {
				t.Errorf("Expected grid classes '%s' to be present", tc.expected)
			}

			// Test gap class
			if !strings.Contains(html, "gap-4") {
				t.Error("Expected gap class to be present")
			}

			// Test content
			if !strings.Contains(html, "Grid content") {
				t.Error("Expected grid content to be displayed")
			}
		})
	}
}

// TestActionButton tests the ActionButton component rendering
func TestActionButton(t *testing.T) {
	testCases := []struct {
		variant  string
		expected string
	}{
		{"primary", "bg-blue-600 text-white"},
		{"secondary", "bg-white text-gray-700"},
		{"danger", "bg-red-600 text-white"},
		{"success", "bg-green-600 text-white"},
	}

	for _, tc := range testCases {
		t.Run(tc.variant, func(t *testing.T) {
			var buf bytes.Buffer
			err := ActionButton("Test Action", "testAction()", tc.variant, "user").Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render ActionButton: %v", err)
			}

			html := buf.String()

			// Test button text
			if !strings.Contains(html, "Test Action") {
				t.Error("Expected button text to be displayed")
			}

			// Test variant styling
			if !strings.Contains(html, tc.expected) {
				t.Errorf("Expected variant classes '%s' to be present", tc.expected)
			}

			// Test data-action attribute
			if !strings.Contains(html, "testAction()") {
				t.Error("Expected data-action attribute to be present")
			}

			// Test icon
			if !strings.Contains(html, "text-white") {
				t.Error("Expected icon with white color to be present")
			}
		})
	}
}

// TestStatusIndicator tests the StatusIndicator component rendering
func TestStatusIndicator(t *testing.T) {
	testCases := []struct {
		status   string
		expected string
	}{
		{"online", "bg-green-400"},
		{"offline", "bg-red-400"},
		{"away", "bg-yellow-400"},
		{"unknown", "bg-gray-400"},
	}

	for _, tc := range testCases {
		t.Run(tc.status, func(t *testing.T) {
			var buf bytes.Buffer
			err := StatusIndicator(tc.status, "User is "+tc.status).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render StatusIndicator: %v", err)
			}

			html := buf.String()

			// Test status color
			if !strings.Contains(html, tc.expected) {
				t.Errorf("Expected status color '%s' to be present", tc.expected)
			}

			// Test status text
			if !strings.Contains(html, "User is "+tc.status) {
				t.Error("Expected status text to be displayed")
			}

			// Test indicator structure
			if !strings.Contains(html, "w-2 h-2 rounded-full") {
				t.Error("Expected status indicator styling")
			}
		})
	}
}

// TestProgressBar tests the ProgressBar component rendering
func TestProgressBar(t *testing.T) {
	var buf bytes.Buffer
	err := ProgressBar(75, "blue", true).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render ProgressBar: %v", err)
	}

	html := buf.String()

	// Test progress bar structure
	if !strings.Contains(html, "bg-gray-200 rounded-full") {
		t.Error("Expected progress bar background styling")
	}

	// Test progress color
	if !strings.Contains(html, "bg-blue-600") {
		t.Error("Expected blue progress color")
	}

	// Test progress width
	if !strings.Contains(html, "width: 75%") {
		t.Error("Expected 75% width styling")
	}

	// Test progress text
	if !strings.Contains(html, "75% complete") {
		t.Error("Expected progress text to be displayed")
	}
}

// TestProgressBarWithoutText tests ProgressBar component without text
func TestProgressBarWithoutText(t *testing.T) {
	var buf bytes.Buffer
	err := ProgressBar(50, "green", false).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render ProgressBar without text: %v", err)
	}

	html := buf.String()

	// Test that progress text is not displayed
	if strings.Contains(html, "50% complete") {
		t.Error("Expected no progress text when showText is false")
	}

	// Test progress color
	if !strings.Contains(html, "bg-green-600") {
		t.Error("Expected green progress color")
	}
}

// TestTooltip tests the Tooltip component rendering
func TestTooltip(t *testing.T) {
	var buf bytes.Buffer
	err := Tooltip("Hover me", "This is a tooltip").Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render Tooltip: %v", err)
	}

	html := buf.String()

	// Test tooltip content
	if !strings.Contains(html, "Hover me") {
		t.Error("Expected tooltip trigger content to be displayed")
	}

	if !strings.Contains(html, "This is a tooltip") {
		t.Error("Expected tooltip text to be displayed")
	}

	// Test tooltip styling
	if !strings.Contains(html, "group-hover:opacity-100") {
		t.Error("Expected tooltip hover styling")
	}

	if !strings.Contains(html, "bg-gray-900") {
		t.Error("Expected tooltip background styling")
	}
}

// TestAlertBanner tests the AlertBanner component rendering
func TestAlertBanner(t *testing.T) {
	testCases := []struct {
		variant  string
		expected string
	}{
		{"info", "bg-blue-50 border-blue-200"},
		{"success", "bg-green-50 border-green-200"},
		{"warning", "bg-yellow-50 border-yellow-200"},
		{"error", "bg-red-50 border-red-200"},
	}

	for _, tc := range testCases {
		t.Run(tc.variant, func(t *testing.T) {
			var buf bytes.Buffer
			err := AlertBanner("Test message", tc.variant, true).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render AlertBanner: %v", err)
			}

			html := buf.String()

			// Test alert message
			if !strings.Contains(html, "Test message") {
				t.Error("Expected alert message to be displayed")
			}

			// Test variant styling - check for individual classes since they might be separated
			expectedClasses := strings.Split(tc.expected, " ")
			for _, class := range expectedClasses {
				if !strings.Contains(html, class) {
					t.Errorf("Expected class '%s' to be present in HTML: %s", class, html)
				}
			}

			// Test dismissible functionality
			if !strings.Contains(html, "this.parentElement.parentElement.parentElement.remove()") {
				t.Error("Expected dismiss functionality")
			}
		})
	}
}

// TestAlertBannerNotDismissible tests AlertBanner component without dismiss button
func TestAlertBannerNotDismissible(t *testing.T) {
	var buf bytes.Buffer
	err := AlertBanner("Test message", "info", false).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render non-dismissible AlertBanner: %v", err)
	}

	html := buf.String()

	// Test that dismiss button is not present
	if strings.Contains(html, "Dismiss") {
		t.Error("Expected no dismiss button when dismissible is false")
	}

	// Test message is still displayed
	if !strings.Contains(html, "Test message") {
		t.Error("Expected alert message to be displayed")
	}
}

// TestNewIcons tests the newly added icons
func TestNewIcons(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
	}{
		{"check", "M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"},
		{"warning", "M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92z"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Icon(tc.name, "blue").Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render Icon: %v", err)
			}

			html := buf.String()

			// Test icon path
			if !strings.Contains(html, tc.expected) {
				t.Errorf("Expected icon path '%s' to be present", tc.expected)
			}
		})
	}
}

func BenchmarkModal(b *testing.B) {
	content := SimpleContent("Content")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		Modal("test", "Title", content, true).Render(context.Background(), &buf)
	}
}

func BenchmarkResponsiveGrid(b *testing.B) {
	content := SimpleContent("Grid content")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		ResponsiveGrid("3", "4", content).Render(context.Background(), &buf)
	}
}

func BenchmarkActionButton(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		ActionButton("Test", "test()", "primary", "user").Render(context.Background(), &buf)
	}
}
