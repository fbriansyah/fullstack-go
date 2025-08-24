package pages

import (
	"context"
	"strings"
	"testing"
)

func TestHome(t *testing.T) {
	var buf strings.Builder
	err := Home().Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render home page: %v", err)
	}

	output := buf.String()

	// Test that the page includes the base layout
	if !strings.Contains(output, "<!doctype html>") {
		t.Error("Expected home page to include HTML doctype")
	}

	// Test page title
	if !strings.Contains(output, "<title>Home - Go Templ Template</title>") {
		t.Error("Expected home page to have correct title")
	}

	// Test hero section content
	heroContent := []string{
		"Welcome to",
		"Go Templ",
		"A modern fullstack Go template",
		"Get Started",
		"View Docs",
	}

	for _, content := range heroContent {
		if !strings.Contains(output, content) {
			t.Errorf("Expected home page to contain %q", content)
		}
	}

	// Test features section
	featureContent := []string{
		"Fast & Efficient",
		"Type Safe",
		"Modern Stack",
		"Built with Go's performance",
		"compile-time safety",
		"authentication, database integration",
	}

	for _, content := range featureContent {
		if !strings.Contains(output, content) {
			t.Errorf("Expected home page to contain feature %q", content)
		}
	}

	// Test CTA section
	ctaContent := []string{
		"Ready to build something amazing?",
		"Start Building Now",
	}

	for _, content := range ctaContent {
		if !strings.Contains(output, content) {
			t.Errorf("Expected home page to contain CTA %q", content)
		}
	}

	// Test responsive classes
	responsiveClasses := []string{
		"max-w-7xl mx-auto",
		"px-4 sm:px-6 lg:px-8",
		"grid grid-cols-1 md:grid-cols-3",
		"flex flex-col sm:flex-row",
	}

	for _, class := range responsiveClasses {
		if !strings.Contains(output, class) {
			t.Errorf("Expected home page to contain responsive class %q", class)
		}
	}
}

func TestHomeContent(t *testing.T) {
	var buf strings.Builder
	err := homeContent().Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render home content: %v", err)
	}

	output := buf.String()

	// Test that content doesn't include full HTML structure (just the content)
	if strings.Contains(output, "<!doctype html>") {
		t.Error("Expected home content to not include full HTML structure")
	}

	// Test that it contains the main content div
	if !strings.Contains(output, "max-w-7xl mx-auto") {
		t.Error("Expected home content to contain main container")
	}

	// Test button styling
	if !strings.Contains(output, "btn-primary") {
		t.Error("Expected home content to contain primary button styling")
	}

	if !strings.Contains(output, "btn-secondary") {
		t.Error("Expected home content to contain secondary button styling")
	}
}
