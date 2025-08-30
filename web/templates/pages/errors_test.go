package pages

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError404Page(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "renders 404 error page",
			expected: []string{
				"Page Not Found - Go Templ Template",
				"404",
				"Page Not Found",
				"Sorry, we couldn't find the page you're looking for",
				"Go Home",
				"Go Back",
				"contact support",
				"Helpful Links",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Error404Page().Render(context.Background(), &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}

			// Check for proper HTML structure
			assert.Contains(t, content, "<!doctype html>")
			assert.Contains(t, content, "<html")
			assert.Contains(t, content, "</html>")
			assert.Contains(t, content, "<head>")
			assert.Contains(t, content, "</head>")
			assert.Contains(t, content, "<body")
			assert.Contains(t, content, "</body>")

			// Check for responsive design classes
			assert.Contains(t, content, "min-h-screen")
			assert.Contains(t, content, "flex")
			assert.Contains(t, content, "justify-center")
			assert.Contains(t, content, "items-center")
		})
	}
}

func TestError500Page(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "renders 500 error page",
			expected: []string{
				"Server Error - Go Templ Template",
				"500",
				"Server Error",
				"Oops! Something went wrong on our end",
				"Try Again",
				"Go Home",
				"contact our support team",
				"What you can do:",
				"Wait a few minutes and try again",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Error500Page().Render(context.Background(), &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}

			// Check for proper HTML structure
			assert.Contains(t, content, "<!doctype html>")
			assert.Contains(t, content, "<html")
			assert.Contains(t, content, "</html>")

			// Check for error styling (red colors)
			assert.Contains(t, content, "text-red-600")

			// Check for helpful information
			assert.Contains(t, content, "location.reload()")
		})
	}
}

func TestErrorAuthPage(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "renders authentication error page",
			expected: []string{
				"Authentication Required - Go Templ Template",
				"401",
				"Authentication Required",
				"You need to be logged in to access this page",
				"Sign In",
				"Create Account",
				"/login",
				"/register",
				"Why do I need to sign in?",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := ErrorAuthPage().Render(context.Background(), &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}

			// Check for proper HTML structure
			assert.Contains(t, content, "<!doctype html>")

			// Check for authentication-specific styling (amber colors)
			assert.Contains(t, content, "text-amber-600")

			// Check for login/register links
			assert.Contains(t, content, `href="/login"`)
			assert.Contains(t, content, `href="/register"`)
		})
	}
}

func TestErrorForbiddenPage(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "renders forbidden error page",
			expected: []string{
				"Access Forbidden - Go Templ Template",
				"403",
				"Access Forbidden",
				"You don't have permission to access this resource",
				"Go Home",
				"Go Back",
				"Possible reasons:",
				"You don't have the required permissions",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := ErrorForbiddenPage().Render(context.Background(), &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}

			// Check for proper HTML structure
			assert.Contains(t, content, "<!doctype html>")

			// Check for forbidden-specific styling (red colors)
			assert.Contains(t, content, "text-red-600")
		})
	}
}

func TestGenericErrorPage(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		message  string
		code     string
		expected []string
	}{
		{
			name:    "renders generic error page with all parameters",
			title:   "Custom Error",
			message: "This is a custom error message",
			code:    "999",
			expected: []string{
				"Custom Error - Go Templ Template",
				"999",
				"Custom Error",
				"This is a custom error message",
				"Go Home",
				"Go Back",
				"contact support",
			},
		},
		{
			name:    "renders generic error page without code",
			title:   "Another Error",
			message: "Another error message",
			code:    "",
			expected: []string{
				"Another Error - Go Templ Template",
				"Another Error",
				"Another error message",
				"Go Home",
				"Go Back",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := GenericErrorPage(tt.title, tt.message, tt.code).Render(context.Background(), &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}

			// Check for proper HTML structure
			assert.Contains(t, content, "<!doctype html>")

			// If code is provided, it should be displayed
			if tt.code != "" {
				assert.Contains(t, content, tt.code)
			}
		})
	}
}

func TestErrorPagesAccessibility(t *testing.T) {
	errorPages := []struct {
		name     string
		renderer func() templ.Component
	}{
		{"404", func() templ.Component { return Error404Page() }},
		{"500", func() templ.Component { return Error500Page() }},
		{"Auth", func() templ.Component { return ErrorAuthPage() }},
		{"Forbidden", func() templ.Component { return ErrorForbiddenPage() }},
		{"Generic", func() templ.Component { return GenericErrorPage("Test", "Test message", "123") }},
	}

	for _, page := range errorPages {
		t.Run(page.name+" accessibility", func(t *testing.T) {
			var buf bytes.Buffer
			err := page.renderer().Render(context.Background(), &buf)
			require.NoError(t, err)

			content := buf.String()

			// Check for semantic HTML
			assert.Contains(t, content, "<h1")
			assert.Contains(t, content, "<h2")
			assert.Contains(t, content, "<p")

			// Check for proper language attribute
			assert.Contains(t, content, `lang="en"`)

			// Check for viewport meta tag
			assert.Contains(t, content, `name="viewport"`)

			// Check for charset
			assert.Contains(t, content, `charset="UTF-8"`)

			// Check for title tag
			assert.Contains(t, content, "<title>")
		})
	}
}

func TestErrorPagesResponsiveDesign(t *testing.T) {
	errorPages := []struct {
		name     string
		renderer func() templ.Component
	}{
		{"404", func() templ.Component { return Error404Page() }},
		{"500", func() templ.Component { return Error500Page() }},
		{"Auth", func() templ.Component { return ErrorAuthPage() }},
		{"Forbidden", func() templ.Component { return ErrorForbiddenPage() }},
	}

	for _, page := range errorPages {
		t.Run(page.name+" responsive design", func(t *testing.T) {
			var buf bytes.Buffer
			err := page.renderer().Render(context.Background(), &buf)
			require.NoError(t, err)

			content := buf.String()

			// Check for responsive classes
			responsiveClasses := []string{
				"min-h-screen",
				"flex",
				"items-center",
				"justify-center",
				"px-4",
				"sm:px-6",
				"lg:px-8",
				"max-w-md",
				"w-full",
			}

			for _, class := range responsiveClasses {
				assert.Contains(t, content, class, "Missing responsive class: %s", class)
			}

			// Check for responsive button layout
			assert.Contains(t, content, "flex-col")
			assert.Contains(t, content, "sm:flex-row")
		})
	}
}

func TestErrorPagesUserFriendlyMessages(t *testing.T) {
	tests := []struct {
		name     string
		renderer func() templ.Component
		messages []string
	}{
		{
			name:     "404 user-friendly messages",
			renderer: func() templ.Component { return Error404Page() },
			messages: []string{
				"Sorry, we couldn't find the page you're looking for",
				"The page might have been moved, deleted, or you might have entered the wrong URL",
				"contact support",
			},
		},
		{
			name:     "500 user-friendly messages",
			renderer: func() templ.Component { return Error500Page() },
			messages: []string{
				"Oops! Something went wrong on our end",
				"We're working to fix this issue as quickly as possible",
				"contact our support team",
				"Wait a few minutes and try again",
				"Clear your browser cache and cookies",
			},
		},
		{
			name:     "Auth user-friendly messages",
			renderer: func() templ.Component { return ErrorAuthPage() },
			messages: []string{
				"You need to be logged in to access this page",
				"Please sign in to continue",
				"Sign up for free",
				"Why do I need to sign in?",
				"This page contains personalized content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tt.renderer().Render(context.Background(), &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, message := range tt.messages {
				assert.Contains(t, content, message, "Missing user-friendly message: %s", message)
			}
		})
	}
}

func TestErrorPagesRecoverySuggestions(t *testing.T) {
	tests := []struct {
		name        string
		renderer    func() templ.Component
		suggestions []string
	}{
		{
			name:     "404 recovery suggestions",
			renderer: func() templ.Component { return Error404Page() },
			suggestions: []string{
				"Go Home",
				"Go Back",
				"Homepage",
				"Login",
				"Register",
				"Help Center",
			},
		},
		{
			name:     "500 recovery suggestions",
			renderer: func() templ.Component { return Error500Page() },
			suggestions: []string{
				"Try Again",
				"Go Home",
				"Wait a few minutes and try again",
				"Check if the issue persists on other pages",
				"Clear your browser cache and cookies",
				"Contact support if the problem continues",
			},
		},
		{
			name:     "Auth recovery suggestions",
			renderer: func() templ.Component { return ErrorAuthPage() },
			suggestions: []string{
				"Sign In",
				"Create Account",
				"Sign up for free",
				"Homepage",
				"About Us",
				"Help Center",
				"Contact Support",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tt.renderer().Render(context.Background(), &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, suggestion := range tt.suggestions {
				assert.Contains(t, content, suggestion, "Missing recovery suggestion: %s", suggestion)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkError404Page(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		Error404Page().Render(context.Background(), &buf)
	}
}

func BenchmarkError500Page(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		Error500Page().Render(context.Background(), &buf)
	}
}

func BenchmarkErrorAuthPage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		ErrorAuthPage().Render(context.Background(), &buf)
	}
}

// Helper function to check if content contains all expected strings
func containsAll(content string, expected []string) bool {
	for _, exp := range expected {
		if !strings.Contains(content, exp) {
			return false
		}
	}
	return true
}
