package components

import (
	"context"
	"strings"
	"testing"
)

func TestHeader(t *testing.T) {
	var buf strings.Builder
	err := Header().Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render header component: %v", err)
	}

	output := buf.String()

	// Test basic structure
	expectedElements := []string{
		"<header",
		"Go Templ",
		"Home",
		"About",
		"Dashboard",
		"Login",
		"Sign Up",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected header to contain %q, but it didn't", expected)
		}
	}

	// Test responsive design
	responsiveClasses := []string{
		"hidden md:flex",       // Desktop navigation
		"md:hidden",            // Mobile menu button
		"max-w-7xl mx-auto",    // Container
		"px-4 sm:px-6 lg:px-8", // Responsive padding
	}

	for _, class := range responsiveClasses {
		if !strings.Contains(output, class) {
			t.Errorf("Expected header to contain responsive class %q", class)
		}
	}

	// Test accessibility
	accessibilityFeatures := []string{
		"aria-label=\"Toggle mobile menu\"",
		"onclick=\"toggleMobileMenu()\"",
	}

	for _, feature := range accessibilityFeatures {
		if !strings.Contains(output, feature) {
			t.Errorf("Expected header to contain accessibility feature %q", feature)
		}
	}

	// Test mobile menu
	if !strings.Contains(output, "id=\"mobile-menu\"") {
		t.Error("Expected mobile menu to have proper ID")
	}

	// Test logo
	if !strings.Contains(output, "GT") {
		t.Error("Expected logo to contain 'GT' text")
	}
}

func TestHeaderNavigation(t *testing.T) {
	var buf strings.Builder
	err := Header().Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render header component: %v", err)
	}

	output := buf.String()

	// Test navigation links
	navLinks := []struct {
		href string
		text string
	}{
		{"/", "Home"},
		{"/about", "About"},
		{"/dashboard", "Dashboard"},
		{"/login", "Login"},
		{"/register", "Sign Up"},
	}

	for _, link := range navLinks {
		if !strings.Contains(output, `href="`+link.href+`"`) {
			t.Errorf("Expected navigation to contain link to %q", link.href)
		}
		if !strings.Contains(output, link.text) {
			t.Errorf("Expected navigation to contain text %q", link.text)
		}
	}

	// Test button styling
	if !strings.Contains(output, "btn-primary") {
		t.Error("Expected Sign Up button to have btn-primary class")
	}
}

func TestHeaderMobileMenu(t *testing.T) {
	var buf strings.Builder
	err := Header().Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render header component: %v", err)
	}

	output := buf.String()

	// Test mobile menu structure
	mobileMenuFeatures := []string{
		"id=\"mobile-menu\"",
		"md:hidden hidden",
		"toggleMobileMenu()",
		"border-t border-gray-200",
	}

	for _, feature := range mobileMenuFeatures {
		if !strings.Contains(output, feature) {
			t.Errorf("Expected mobile menu to contain %q", feature)
		}
	}

	// Test hamburger icon
	if !strings.Contains(output, "M4 6h16M4 12h16M4 18h16") {
		t.Error("Expected hamburger menu icon SVG path")
	}
}
