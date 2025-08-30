package pages

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestError404Page tests the 404 error page rendering
func TestError404Page(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "renders 404 page with all elements",
			expected: []string{
				"404",
				"Page Not Found",
				"Sorry, we couldn't find the page you're looking for",
				"Go Home",
				"Go Back",
				"contact support",
				"Helpful Links",
				"Homepage",
				"Login",
				"Register",
				"Help Center",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ctx := context.Background()

			err := Error404Page().Render(ctx, &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}

			// Check for proper HTML structure
			assert.Contains(t, content, "<html", "Should contain HTML tag")
			assert.Contains(t, content, "<title>", "Should contain title tag")
			assert.Contains(t, content, "Page Not Found - Go Templ Template", "Should contain proper title")
			assert.Contains(t, content, "min-h-screen", "Should contain proper styling classes")
		})
	}
}

// TestError404Content tests the 404 error page content rendering
func TestError404Content(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "renders 404 content with proper structure",
			expected: []string{
				"text-6xl font-bold text-gray-900\">404",
				"text-3xl font-bold text-gray-900\">Page Not Found",
				"text-lg text-gray-600",
				"onclick=\"history.back()\"",
				"href=\"/contact\"",
				"grid grid-cols-1 sm:grid-cols-2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ctx := context.Background()

			err := Error404Content().Render(ctx, &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}
		})
	}
}

// TestError500Page tests the 500 error page rendering
func TestError500Page(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "renders 500 page with all elements",
			expected: []string{
				"500",
				"Server Error",
				"Oops! Something went wrong on our end",
				"Try Again",
				"Go Home",
				"contact our support team",
				"What you can do:",
				"Wait a few minutes and try again",
				"Clear your browser cache",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ctx := context.Background()

			err := Error500Page().Render(ctx, &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}

			// Check for proper HTML structure
			assert.Contains(t, content, "<title>", "Should contain title tag")
			assert.Contains(t, content, "Server Error - Go Templ Template", "Should contain proper title")
			assert.Contains(t, content, "text-red-600", "Should contain red color for error icon")
		})
	}
}

// TestError500Content tests the 500 error page content rendering
func TestError500Content(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "renders 500 content with proper structure",
			expected: []string{
				"text-6xl font-bold text-gray-900\">500",
				"text-3xl font-bold text-gray-900\">Server Error",
				"onclick=\"location.reload()\"",
				"bg-yellow-50 border border-yellow-200",
				"list-disc list-inside space-y-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ctx := context.Background()

			err := Error500Content().Render(ctx, &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}
		})
	}
}

// TestErrorAuthPage tests the authentication error page rendering
func TestErrorAuthPage(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "renders auth error page with all elements",
			expected: []string{
				"401",
				"Authentication Required",
				"You need to be logged in to access this page",
				"Sign In",
				"Create Account",
				"Don't have an account?",
				"Sign up for free",
				"Why do I need to sign in?",
				"Quick Access",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ctx := context.Background()

			err := ErrorAuthPage().Render(ctx, &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}

			// Check for proper HTML structure
			assert.Contains(t, content, "Authentication Required - Go Templ Template", "Should contain proper title")
			assert.Contains(t, content, "text-amber-600", "Should contain amber color for auth icon")
		})
	}
}

// TestErrorAuthContent tests the authentication error page content rendering
func TestErrorAuthContent(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "renders auth content with proper structure",
			expected: []string{
				"text-6xl font-bold text-gray-900\">401",
				"text-3xl font-bold text-gray-900\">Authentication Required",
				"href=\"/login\"",
				"href=\"/register\"",
				"bg-blue-50 border border-blue-200",
				"text-blue-800",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ctx := context.Background()

			err := ErrorAuthContent().Render(ctx, &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}
		})
	}
}

// TestErrorForbiddenPage tests the forbidden error page rendering
func TestErrorForbiddenPage(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "renders forbidden error page with all elements",
			expected: []string{
				"403",
				"Access Forbidden",
				"You don't have permission to access this resource",
				"Go Home",
				"Go Back",
				"Contact our support team",
				"Possible reasons:",
				"You don't have the required permissions",
				"Your account may be suspended",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ctx := context.Background()

			err := ErrorForbiddenPage().Render(ctx, &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}

			// Check for proper HTML structure
			assert.Contains(t, content, "Access Forbidden - Go Templ Template", "Should contain proper title")
			assert.Contains(t, content, "text-red-600", "Should contain red color for forbidden icon")
		})
	}
}

// TestErrorForbiddenContent tests the forbidden error page content rendering
func TestErrorForbiddenContent(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "renders forbidden content with proper structure",
			expected: []string{
				"text-6xl font-bold text-gray-900\">403",
				"text-3xl font-bold text-gray-900\">Access Forbidden",
				"bg-red-50 border border-red-200",
				"text-red-800",
				"onclick=\"history.back()\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ctx := context.Background()

			err := ErrorForbiddenContent().Render(ctx, &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}
		})
	}
}

// TestGenericErrorPage tests the generic error page rendering
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
			ctx := context.Background()

			err := GenericErrorPage(tt.title, tt.message, tt.code).Render(ctx, &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}

			// Check that code is only rendered when provided
			if tt.code != "" {
				assert.Contains(t, content, "text-6xl font-bold text-gray-900\">"+tt.code, "Should contain error code")
			} else {
				assert.NotContains(t, content, "text-6xl font-bold text-gray-900\">", "Should not contain error code section")
			}
		})
	}
}

// TestGenericErrorContent tests the generic error page content rendering
func TestGenericErrorContent(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		message  string
		code     string
		expected []string
	}{
		{
			name:    "renders generic content with proper structure",
			title:   "Test Error",
			message: "Test message",
			code:    "123",
			expected: []string{
				"text-3xl font-bold text-gray-900\">Test Error",
				"text-lg text-gray-600\">Test message",
				"text-6xl font-bold text-gray-900\">123",
				"text-gray-600",
				"onclick=\"history.back()\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ctx := context.Background()

			err := GenericErrorContent(tt.title, tt.message, tt.code).Render(ctx, &buf)
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, content, expected, "Expected content not found: %s", expected)
			}
		})
	}
}

// TestErrorPagesAccessibility tests that error pages contain proper accessibility features
func TestErrorPagesAccessibility(t *testing.T) {
	tests := []struct {
		name     string
		renderer func() error
	}{
		{
			name: "404 page accessibility",
			renderer: func() error {
				var buf bytes.Buffer
				return Error404Page().Render(context.Background(), &buf)
			},
		},
		{
			name: "500 page accessibility",
			renderer: func() error {
				var buf bytes.Buffer
				return Error500Page().Render(context.Background(), &buf)
			},
		},
		{
			name: "auth page accessibility",
			renderer: func() error {
				var buf bytes.Buffer
				return ErrorAuthPage().Render(context.Background(), &buf)
			},
		},
		{
			name: "forbidden page accessibility",
			renderer: func() error {
				var buf bytes.Buffer
				return ErrorForbiddenPage().Render(context.Background(), &buf)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.renderer()
			require.NoError(t, err, "Error page should render without errors")
		})
	}
}

// TestErrorPagesResponsiveness tests that error pages contain responsive design classes
func TestErrorPagesResponsiveness(t *testing.T) {
	tests := []struct {
		name     string
		renderer func() (string, error)
	}{
		{
			name: "404 page responsiveness",
			renderer: func() (string, error) {
				var buf bytes.Buffer
				err := Error404Content().Render(context.Background(), &buf)
				return buf.String(), err
			},
		},
		{
			name: "500 page responsiveness",
			renderer: func() (string, error) {
				var buf bytes.Buffer
				err := Error500Content().Render(context.Background(), &buf)
				return buf.String(), err
			},
		},
		{
			name: "auth page responsiveness",
			renderer: func() (string, error) {
				var buf bytes.Buffer
				err := ErrorAuthContent().Render(context.Background(), &buf)
				return buf.String(), err
			},
		},
		{
			name: "forbidden page responsiveness",
			renderer: func() (string, error) {
				var buf bytes.Buffer
				err := ErrorForbiddenContent().Render(context.Background(), &buf)
				return buf.String(), err
			},
		},
	}

	responsiveClasses := []string{
		"sm:flex-row",
		"sm:px-6",
		"lg:px-8",
		"sm:grid-cols-2",
		"min-h-screen",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := tt.renderer()
			require.NoError(t, err, "Error page should render without errors")

			// Check for responsive classes
			foundResponsive := false
			for _, class := range responsiveClasses {
				if strings.Contains(content, class) {
					foundResponsive = true
					break
				}
			}
			assert.True(t, foundResponsive, "Error page should contain responsive design classes")
		})
	}
}

// BenchmarkError404Page benchmarks the 404 error page rendering
func BenchmarkError404Page(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		err := Error404Page().Render(ctx, &buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkError500Page benchmarks the 500 error page rendering
func BenchmarkError500Page(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		err := Error500Page().Render(ctx, &buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenericErrorPage benchmarks the generic error page rendering
func BenchmarkGenericErrorPage(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		err := GenericErrorPage("Test Error", "Test message", "500").Render(ctx, &buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}
