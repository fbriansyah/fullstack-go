package components

import (
	"context"
	"strings"
	"testing"
)

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "simple error",
			message: "This field is required",
		},
		{
			name:    "validation error",
			message: "Invalid email format",
		},
		{
			name:    "empty message",
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := ErrorMessage(tt.message).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render error message: %v", err)
			}

			output := buf.String()

			// Test basic structure
			expectedElements := []string{
				"text-red-600",
				"flex items-center",
				"<svg",
				"<span>",
			}

			for _, expected := range expectedElements {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected error message to contain %q", expected)
				}
			}

			// Test message content
			if tt.message != "" && !strings.Contains(output, tt.message) {
				t.Errorf("Expected error message to contain %q", tt.message)
			}

			// Test icon
			if !strings.Contains(output, "viewBox=\"0 0 20 20\"") {
				t.Error("Expected error message to contain SVG icon")
			}
		})
	}
}

func TestSuccessMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "success notification",
			message: "Account created successfully",
		},
		{
			name:    "operation completed",
			message: "Password updated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := SuccessMessage(tt.message).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render success message: %v", err)
			}

			output := buf.String()

			// Test basic structure
			expectedElements := []string{
				"bg-green-50",
				"text-green-800",
				"text-green-400",
				"<svg",
				tt.message,
			}

			for _, expected := range expectedElements {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected success message to contain %q", expected)
				}
			}

			// Test checkmark icon
			if !strings.Contains(output, "fill-rule=\"evenodd\"") {
				t.Error("Expected success message to contain checkmark icon")
			}
		})
	}
}

func TestWarningMessage(t *testing.T) {
	message := "Your session will expire soon"
	var buf strings.Builder
	err := WarningMessage(message).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render warning message: %v", err)
	}

	output := buf.String()

	// Test basic structure
	expectedElements := []string{
		"bg-yellow-50",
		"text-yellow-800",
		"text-yellow-400",
		message,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected warning message to contain %q", expected)
		}
	}
}

func TestInfoMessage(t *testing.T) {
	message := "Check your email for verification link"
	var buf strings.Builder
	err := InfoMessage(message).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render info message: %v", err)
	}

	output := buf.String()

	// Test basic structure
	expectedElements := []string{
		"bg-blue-50",
		"text-blue-800",
		"text-blue-400",
		message,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected info message to contain %q", expected)
		}
	}
}

func TestErrorAlert(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		message     string
		dismissible bool
	}{
		{
			name:        "dismissible error",
			title:       "Authentication Failed",
			message:     "Invalid username or password",
			dismissible: true,
		},
		{
			name:        "non-dismissible error",
			title:       "System Error",
			message:     "Please contact support",
			dismissible: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := ErrorAlert(tt.title, tt.message, tt.dismissible).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render error alert: %v", err)
			}

			output := buf.String()

			// Test basic structure
			expectedElements := []string{
				"bg-red-50",
				"text-red-800",
				"text-red-400",
				tt.title,
				tt.message,
				"id=\"error-alert\"",
			}

			for _, expected := range expectedElements {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected error alert to contain %q", expected)
				}
			}

			// Test dismissible functionality
			if tt.dismissible {
				dismissElements := []string{
					"<button",
					"onclick=\"document.getElementById('error-alert').remove()\"",
					"Dismiss",
				}
				for _, element := range dismissElements {
					if !strings.Contains(output, element) {
						t.Errorf("Expected dismissible error alert to contain %q", element)
					}
				}
			} else {
				if strings.Contains(output, "onclick=\"document.getElementById('error-alert').remove()\"") {
					t.Error("Expected non-dismissible error alert to not contain dismiss button")
				}
			}
		})
	}
}

func TestSuccessAlert(t *testing.T) {
	title := "Account Created"
	message := "Welcome to our platform!"
	dismissible := true

	var buf strings.Builder
	err := SuccessAlert(title, message, dismissible).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render success alert: %v", err)
	}

	output := buf.String()

	// Test basic structure
	expectedElements := []string{
		"bg-green-50",
		"text-green-800",
		"text-green-400",
		title,
		message,
		"id=\"success-alert\"",
		"onclick=\"document.getElementById('success-alert').remove()\"",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected success alert to contain %q", expected)
		}
	}
}

func TestValidationSummary(t *testing.T) {
	tests := []struct {
		name   string
		errors []string
	}{
		{
			name:   "no errors",
			errors: []string{},
		},
		{
			name: "multiple errors",
			errors: []string{
				"Email is required",
				"Password must be at least 8 characters",
				"Terms must be accepted",
			},
		},
		{
			name: "single error",
			errors: []string{
				"Invalid email format",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := ValidationSummary(tt.errors).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render validation summary: %v", err)
			}

			output := buf.String()

			if len(tt.errors) == 0 {
				// Should render nothing for empty errors
				if strings.TrimSpace(output) != "" {
					t.Error("Expected empty output for no errors")
				}
				return
			}

			// Test basic structure
			expectedElements := []string{
				"bg-red-50",
				"text-red-800",
				"Please correct the following errors:",
				"<ul",
				"list-disc",
			}

			for _, expected := range expectedElements {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected validation summary to contain %q", expected)
				}
			}

			// Test all errors are displayed
			for _, errorMsg := range tt.errors {
				if !strings.Contains(output, errorMsg) {
					t.Errorf("Expected validation summary to contain error %q", errorMsg)
				}
			}
		})
	}
}

func TestFlashMessage(t *testing.T) {
	tests := []struct {
		name        string
		messageType string
		message     string
	}{
		{
			name:        "success flash",
			messageType: "success",
			message:     "Operation completed successfully",
		},
		{
			name:        "error flash",
			messageType: "error",
			message:     "Something went wrong",
		},
		{
			name:        "warning flash",
			messageType: "warning",
			message:     "Please verify your email",
		},
		{
			name:        "info flash",
			messageType: "info",
			message:     "New features available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := FlashMessage(tt.messageType, tt.message).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render flash message: %v", err)
			}

			output := buf.String()

			// Test basic structure
			expectedElements := []string{
				"flash-message",
				"id=\"flash-message\"",
				tt.message,
				"onclick=\"document.getElementById('flash-message').remove()\"",
				"<svg",
			}

			for _, expected := range expectedElements {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected flash message to contain %q", expected)
				}
			}

			// Test type-specific styling
			switch tt.messageType {
			case "success":
				if !strings.Contains(output, "bg-green-50") {
					t.Error("Expected success flash message to have green background")
				}
				if !strings.Contains(output, "text-green-800") {
					t.Error("Expected success flash message to have green text")
				}
			case "error":
				if !strings.Contains(output, "bg-red-50") {
					t.Error("Expected error flash message to have red background")
				}
				if !strings.Contains(output, "text-red-800") {
					t.Error("Expected error flash message to have red text")
				}
			case "warning":
				if !strings.Contains(output, "bg-yellow-50") {
					t.Error("Expected warning flash message to have yellow background")
				}
				if !strings.Contains(output, "text-yellow-800") {
					t.Error("Expected warning flash message to have yellow text")
				}
			case "info":
				if !strings.Contains(output, "bg-blue-50") {
					t.Error("Expected info flash message to have blue background")
				}
				if !strings.Contains(output, "text-blue-800") {
					t.Error("Expected info flash message to have blue text")
				}
			}
		})
	}
}
