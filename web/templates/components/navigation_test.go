package components

import (
	"context"
	"strings"
	"testing"
)

func TestNavigation(t *testing.T) {
	items := []NavItem{
		{Label: "Home", URL: "/", Active: true},
		{Label: "About", URL: "/about", Active: false},
		{Label: "Contact", URL: "/contact", Active: false},
	}

	var buf strings.Builder
	err := Navigation(items).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render navigation component: %v", err)
	}

	output := buf.String()

	// Test basic structure
	if !strings.Contains(output, "<nav") {
		t.Error("Expected navigation to contain nav tag")
	}

	// Test all items are rendered
	for _, item := range items {
		if !strings.Contains(output, item.Label) {
			t.Errorf("Expected navigation to contain label %q", item.Label)
		}
		if !strings.Contains(output, `href="`+item.URL+`"`) {
			t.Errorf("Expected navigation to contain URL %q", item.URL)
		}
	}

	// Test active state styling
	if !strings.Contains(output, "text-blue-600 border-b-2 border-blue-600") {
		t.Error("Expected active item to have proper styling")
	}

	// Test inactive state styling
	if !strings.Contains(output, "text-gray-700 hover:text-blue-600") {
		t.Error("Expected inactive items to have proper styling")
	}
}

func TestMobileNavigation(t *testing.T) {
	items := []NavItem{
		{Label: "Home", URL: "/", Active: true},
		{Label: "About", URL: "/about", Active: false},
		{Label: "Services", URL: "/services", Active: false},
	}

	var buf strings.Builder
	err := MobileNavigation(items).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render mobile navigation component: %v", err)
	}

	output := buf.String()

	// Test mobile-specific structure
	if !strings.Contains(output, "px-2 pt-2 pb-3 space-y-1") {
		t.Error("Expected mobile navigation to have proper container classes")
	}

	// Test all items are rendered
	for _, item := range items {
		if !strings.Contains(output, item.Label) {
			t.Errorf("Expected mobile navigation to contain label %q", item.Label)
		}
		if !strings.Contains(output, `href="`+item.URL+`"`) {
			t.Errorf("Expected mobile navigation to contain URL %q", item.URL)
		}
	}

	// Test active state styling for mobile
	if !strings.Contains(output, "text-blue-600 bg-blue-50") {
		t.Error("Expected active item to have proper mobile styling")
	}

	// Test inactive state styling for mobile
	if !strings.Contains(output, "hover:bg-gray-50") {
		t.Error("Expected inactive items to have proper mobile hover styling")
	}

	// Test block display for mobile
	if !strings.Contains(output, "block") {
		t.Error("Expected mobile navigation items to have block display")
	}
}

func TestNavigationWithEmptyItems(t *testing.T) {
	items := []NavItem{}

	var buf strings.Builder
	err := Navigation(items).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render navigation with empty items: %v", err)
	}

	output := buf.String()

	// Should still render nav container
	if !strings.Contains(output, "<nav") {
		t.Error("Expected navigation to contain nav tag even with empty items")
	}

	// Should not contain any links
	if strings.Contains(output, "<a") {
		t.Error("Expected no links with empty items")
	}
}

func TestNavigationActiveStates(t *testing.T) {
	tests := []struct {
		name   string
		items  []NavItem
		active string
	}{
		{
			name: "first item active",
			items: []NavItem{
				{Label: "Home", URL: "/", Active: true},
				{Label: "About", URL: "/about", Active: false},
			},
			active: "Home",
		},
		{
			name: "second item active",
			items: []NavItem{
				{Label: "Home", URL: "/", Active: false},
				{Label: "About", URL: "/about", Active: true},
			},
			active: "About",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := Navigation(tt.items).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render navigation: %v", err)
			}

			output := buf.String()

			// Count active styling occurrences
			activeCount := strings.Count(output, "text-blue-600 border-b-2 border-blue-600")
			if activeCount != 1 {
				t.Errorf("Expected exactly 1 active item, got %d", activeCount)
			}
		})
	}
}
