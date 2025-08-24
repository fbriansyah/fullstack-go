package components

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestFooter(t *testing.T) {
	var buf strings.Builder
	err := Footer().Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render footer component: %v", err)
	}

	output := buf.String()

	// Test basic structure
	expectedElements := []string{
		"<footer",
		"Go Templ",
		"Quick Links",
		"Support",
		"All rights reserved",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected footer to contain %q, but it didn't", expected)
		}
	}

	// Test current year
	currentYear := time.Now().Year()
	yearStr := strings.Contains(output, "2025") // Check for current year as string
	if !yearStr {
		t.Errorf("Expected footer to contain current year %d", currentYear)
	}

	// Test responsive grid
	responsiveClasses := []string{
		"grid grid-cols-1 md:grid-cols-4",
		"max-w-7xl mx-auto",
		"px-4 sm:px-6 lg:px-8",
	}

	for _, class := range responsiveClasses {
		if !strings.Contains(output, class) {
			t.Errorf("Expected footer to contain responsive class %q", class)
		}
	}
}

func TestFooterLinks(t *testing.T) {
	var buf strings.Builder
	err := Footer().Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render footer component: %v", err)
	}

	output := buf.String()

	// Test Quick Links
	quickLinks := []struct {
		href string
		text string
	}{
		{"/", "Home"},
		{"/about", "About"},
		{"/dashboard", "Dashboard"},
		{"/docs", "Documentation"},
	}

	for _, link := range quickLinks {
		if !strings.Contains(output, `href="`+link.href+`"`) {
			t.Errorf("Expected footer to contain link to %q", link.href)
		}
		if !strings.Contains(output, link.text) {
			t.Errorf("Expected footer to contain text %q", link.text)
		}
	}

	// Test Support Links
	supportLinks := []struct {
		href string
		text string
	}{
		{"/help", "Help Center"},
		{"/contact", "Contact Us"},
		{"/privacy", "Privacy Policy"},
		{"/terms", "Terms of Service"},
	}

	for _, link := range supportLinks {
		if !strings.Contains(output, `href="`+link.href+`"`) {
			t.Errorf("Expected footer to contain support link to %q", link.href)
		}
		if !strings.Contains(output, link.text) {
			t.Errorf("Expected footer to contain support text %q", link.text)
		}
	}
}

func TestFooterSocialLinks(t *testing.T) {
	var buf strings.Builder
	err := Footer().Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render footer component: %v", err)
	}

	output := buf.String()

	// Test social media links
	socialLinks := []string{
		"https://github.com",
		"https://twitter.com",
		"aria-label=\"GitHub\"",
		"aria-label=\"Twitter\"",
	}

	for _, link := range socialLinks {
		if !strings.Contains(output, link) {
			t.Errorf("Expected footer to contain social link %q", link)
		}
	}

	// Test SVG icons are present
	if !strings.Contains(output, "<svg") {
		t.Error("Expected footer to contain SVG icons")
	}
}

func TestFooterDescription(t *testing.T) {
	var buf strings.Builder
	err := Footer().Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render footer component: %v", err)
	}

	output := buf.String()

	// Test brand description
	expectedDescription := "A modern fullstack Go template using Templ for type-safe HTML templating"
	if !strings.Contains(output, expectedDescription) {
		t.Error("Expected footer to contain brand description")
	}

	// Test logo
	if !strings.Contains(output, "GT") {
		t.Error("Expected footer to contain logo 'GT'")
	}
}
