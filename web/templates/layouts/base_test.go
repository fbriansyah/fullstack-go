package layouts

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestBase(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		content  templ.Component
		expected []string
	}{
		{
			name:  "renders basic layout with title",
			title: "Test Page",
			content: templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
				_, err := w.Write([]byte("<div>Test Content</div>"))
				return err
			}),
			expected: []string{
				"<!doctype html>",
				"<title>Test Page</title>",
				"<div>Test Content</div>",
				"<link href=\"/static/css/tailwind.css\" rel=\"stylesheet\">",
				"<script src=\"/static/js/main.js\"></script>",
			},
		},
		{
			name:  "includes responsive meta tags",
			title: "Mobile Test",
			content: templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
				_, err := w.Write([]byte("<p>Mobile content</p>"))
				return err
			}),
			expected: []string{
				"<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">",
				"<meta charset=\"UTF-8\">",
				"<meta name=\"description\" content=\"Go Templ Template - Modern fullstack web application\">",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Render the component
			var buf strings.Builder
			err := Base(tt.title, tt.content).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render component: %v", err)
			}

			output := buf.String()

			// Check that all expected strings are present
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput: %s", expected, output)
				}
			}

			// Verify HTML structure
			if !strings.Contains(output, "<html") {
				t.Error("Expected HTML tag")
			}
			if !strings.Contains(output, "<head>") {
				t.Error("Expected head tag")
			}
			if !strings.Contains(output, "<body") {
				t.Error("Expected body tag")
			}
			if !strings.Contains(output, "<main") {
				t.Error("Expected main tag")
			}
		})
	}
}

func TestBaseStructure(t *testing.T) {
	content := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte("<div>Test</div>"))
		return err
	})

	var buf strings.Builder
	err := Base("Test", content).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render component: %v", err)
	}

	output := buf.String()

	// Test responsive classes
	if !strings.Contains(output, "class=\"h-full\"") {
		t.Error("Expected html tag to have h-full class")
	}

	// Test flexbox layout
	if !strings.Contains(output, "class=\"h-full bg-gray-50 flex flex-col\"") {
		t.Error("Expected body to have proper flexbox classes")
	}

	// Test main content area
	if !strings.Contains(output, "class=\"flex-1\"") {
		t.Error("Expected main to have flex-1 class")
	}
}
