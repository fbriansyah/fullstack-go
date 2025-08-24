package pages

import (
	"context"
	"strings"
	"testing"

	"go-templ-template/web/templates/components"

	"github.com/a-h/templ"
)

func TestLoginPage(t *testing.T) {
	tests := []struct {
		name         string
		errors       map[string]string
		email        string
		flashMessage string
		flashType    string
	}{
		{
			name:         "empty login page",
			errors:       map[string]string{},
			email:        "",
			flashMessage: "",
			flashType:    "",
		},
		{
			name: "login page with errors",
			errors: map[string]string{
				"email":    "Invalid email format",
				"password": "Password is required",
			},
			email:        "invalid@email",
			flashMessage: "",
			flashType:    "",
		},
		{
			name:         "login page with flash message",
			errors:       map[string]string{},
			email:        "",
			flashMessage: "Please log in to continue",
			flashType:    "info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := LoginPage(tt.errors, tt.email, tt.flashMessage, tt.flashType).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render login page: %v", err)
			}

			output := buf.String()

			// Test page structure
			expectedElements := []string{
				"<!doctype html>",
				"<html",
				"<head>",
				"<title>Sign In - Go Templ Template</title>",
				"<body",
				"<main",
				"min-h-screen",
				"flex items-center justify-center",
			}

			for _, expected := range expectedElements {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected login page to contain %q", expected)
				}
			}

			// Test login form presence
			loginFormElements := []string{
				"Sign In",
				"Welcome back!",
				"action=\"/auth/login\"",
				"type=\"email\"",
				"type=\"password\"",
			}

			for _, element := range loginFormElements {
				if !strings.Contains(output, element) {
					t.Errorf("Expected login page to contain login form element %q", element)
				}
			}

			// Test flash message
			if tt.flashMessage != "" {
				if !strings.Contains(output, tt.flashMessage) {
					t.Errorf("Expected login page to contain flash message %q", tt.flashMessage)
				}
				if !strings.Contains(output, "flash-message") {
					t.Error("Expected login page to contain flash message component")
				}
			}

			// Test error display
			for _, errorMsg := range tt.errors {
				if errorMsg != "" {
					if !strings.Contains(output, errorMsg) {
						t.Errorf("Expected login page to contain error %q", errorMsg)
					}
				}
			}

			// Test email prefill
			if tt.email != "" {
				if !strings.Contains(output, `value="`+tt.email+`"`) {
					t.Errorf("Expected login page to prefill email %q", tt.email)
				}
			}
		})
	}
}

func TestRegisterPage(t *testing.T) {
	tests := []struct {
		name         string
		errors       map[string]string
		formData     components.RegisterFormData
		flashMessage string
		flashType    string
	}{
		{
			name:   "empty register page",
			errors: map[string]string{},
			formData: components.RegisterFormData{
				FirstName: "",
				LastName:  "",
				Email:     "",
			},
			flashMessage: "",
			flashType:    "",
		},
		{
			name: "register page with errors",
			errors: map[string]string{
				"first_name": "First name is required",
				"email":      "Email already exists",
			},
			formData: components.RegisterFormData{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			flashMessage: "",
			flashType:    "",
		},
		{
			name:   "register page with flash message",
			errors: map[string]string{},
			formData: components.RegisterFormData{
				FirstName: "",
				LastName:  "",
				Email:     "",
			},
			flashMessage: "Registration is currently disabled",
			flashType:    "warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := RegisterPage(tt.errors, tt.formData, tt.flashMessage, tt.flashType).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render register page: %v", err)
			}

			output := buf.String()

			// Test page structure
			expectedElements := []string{
				"<!doctype html>",
				"<title>Create Account - Go Templ Template</title>",
				"min-h-screen",
				"flex items-center justify-center",
			}

			for _, expected := range expectedElements {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected register page to contain %q", expected)
				}
			}

			// Test register form presence
			registerFormElements := []string{
				"Create Account",
				"Join us today!",
				"action=\"/auth/register\"",
				"name=\"first_name\"",
				"name=\"last_name\"",
				"name=\"email\"",
				"name=\"password\"",
				"name=\"confirm_password\"",
				"name=\"terms\"",
			}

			for _, element := range registerFormElements {
				if !strings.Contains(output, element) {
					t.Errorf("Expected register page to contain register form element %q", element)
				}
			}

			// Test form data prefill
			if tt.formData.FirstName != "" {
				if !strings.Contains(output, `value="`+tt.formData.FirstName+`"`) {
					t.Errorf("Expected register page to prefill first name %q", tt.formData.FirstName)
				}
			}
			if tt.formData.LastName != "" {
				if !strings.Contains(output, `value="`+tt.formData.LastName+`"`) {
					t.Errorf("Expected register page to prefill last name %q", tt.formData.LastName)
				}
			}
			if tt.formData.Email != "" {
				if !strings.Contains(output, `value="`+tt.formData.Email+`"`) {
					t.Errorf("Expected register page to prefill email %q", tt.formData.Email)
				}
			}

			// Test flash message
			if tt.flashMessage != "" {
				if !strings.Contains(output, tt.flashMessage) {
					t.Errorf("Expected register page to contain flash message %q", tt.flashMessage)
				}
			}

			// Test error display
			for _, errorMsg := range tt.errors {
				if errorMsg != "" {
					if !strings.Contains(output, errorMsg) {
						t.Errorf("Expected register page to contain error %q", errorMsg)
					}
				}
			}
		})
	}
}

func TestForgotPasswordPage(t *testing.T) {
	tests := []struct {
		name         string
		errors       map[string]string
		email        string
		flashMessage string
		flashType    string
	}{
		{
			name:         "empty forgot password page",
			errors:       map[string]string{},
			email:        "",
			flashMessage: "",
			flashType:    "",
		},
		{
			name: "forgot password page with errors",
			errors: map[string]string{
				"email":   "Email not found",
				"general": "Too many requests",
			},
			email:        "notfound@example.com",
			flashMessage: "",
			flashType:    "",
		},
		{
			name:         "forgot password page with success message",
			errors:       map[string]string{},
			email:        "",
			flashMessage: "Reset link sent to your email",
			flashType:    "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := ForgotPasswordPage(tt.errors, tt.email, tt.flashMessage, tt.flashType).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render forgot password page: %v", err)
			}

			output := buf.String()

			// Test page structure
			expectedElements := []string{
				"<title>Forgot Password - Go Templ Template</title>",
				"Forgot Password",
				"Enter your email address",
				"action=\"/auth/forgot-password\"",
				"Send Reset Link",
			}

			for _, expected := range expectedElements {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected forgot password page to contain %q", expected)
				}
			}

			// Test email prefill
			if tt.email != "" {
				if !strings.Contains(output, `value="`+tt.email+`"`) {
					t.Errorf("Expected forgot password page to prefill email %q", tt.email)
				}
			}

			// Test flash message
			if tt.flashMessage != "" {
				if !strings.Contains(output, tt.flashMessage) {
					t.Errorf("Expected forgot password page to contain flash message %q", tt.flashMessage)
				}
			}

			// Test error display
			for _, errorMsg := range tt.errors {
				if errorMsg != "" {
					if !strings.Contains(output, errorMsg) {
						t.Errorf("Expected forgot password page to contain error %q", errorMsg)
					}
				}
			}
		})
	}
}

func TestResetPasswordPage(t *testing.T) {
	tests := []struct {
		name         string
		errors       map[string]string
		token        string
		flashMessage string
		flashType    string
	}{
		{
			name:         "valid reset password page",
			errors:       map[string]string{},
			token:        "valid-token-123",
			flashMessage: "",
			flashType:    "",
		},
		{
			name: "reset password page with errors",
			errors: map[string]string{
				"password":         "Password too weak",
				"confirm_password": "Passwords do not match",
				"general":          "Token expired",
			},
			token:        "expired-token",
			flashMessage: "",
			flashType:    "",
		},
		{
			name:         "reset password page with warning",
			errors:       map[string]string{},
			token:        "valid-token",
			flashMessage: "This link will expire in 1 hour",
			flashType:    "warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := ResetPasswordPage(tt.errors, tt.token, tt.flashMessage, tt.flashType).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render reset password page: %v", err)
			}

			output := buf.String()

			// Test page structure
			expectedElements := []string{
				"<title>Reset Password - Go Templ Template</title>",
				"Reset Password",
				"Enter your new password",
				"action=\"/auth/reset-password\"",
				"type=\"hidden\"",
				"name=\"token\"",
				"name=\"password\"",
				"name=\"confirm_password\"",
			}

			for _, expected := range expectedElements {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected reset password page to contain %q", expected)
				}
			}

			// Test token value
			if !strings.Contains(output, `value="`+tt.token+`"`) {
				t.Errorf("Expected reset password page to contain token %q", tt.token)
			}

			// Test flash message
			if tt.flashMessage != "" {
				if !strings.Contains(output, tt.flashMessage) {
					t.Errorf("Expected reset password page to contain flash message %q", tt.flashMessage)
				}
			}

			// Test error display
			for _, errorMsg := range tt.errors {
				if errorMsg != "" {
					if !strings.Contains(output, errorMsg) {
						t.Errorf("Expected reset password page to contain error %q", errorMsg)
					}
				}
			}

			// Test password requirements
			if !strings.Contains(output, "Password must be at least 8 characters long") {
				t.Error("Expected reset password page to show password requirements")
			}
		})
	}
}

func TestAuthPageResponsiveness(t *testing.T) {
	// Test that all auth pages have responsive design
	pages := []struct {
		name     string
		renderer func() templ.Component
	}{
		{
			name: "login page",
			renderer: func() templ.Component {
				return LoginPage(map[string]string{}, "", "", "")
			},
		},
		{
			name: "register page",
			renderer: func() templ.Component {
				return RegisterPage(map[string]string{}, components.RegisterFormData{}, "", "")
			},
		},
		{
			name: "forgot password page",
			renderer: func() templ.Component {
				return ForgotPasswordPage(map[string]string{}, "", "", "")
			},
		},
		{
			name: "reset password page",
			renderer: func() templ.Component {
				return ResetPasswordPage(map[string]string{}, "token", "", "")
			},
		},
	}

	for _, page := range pages {
		t.Run(page.name, func(t *testing.T) {
			var buf strings.Builder
			err := page.renderer().Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render %s: %v", page.name, err)
			}

			output := buf.String()

			// Test responsive classes
			responsiveClasses := []string{
				"min-h-screen",
				"flex items-center justify-center",
				"py-12 px-4 sm:px-6 lg:px-8",
				"max-w-md w-full",
			}

			for _, class := range responsiveClasses {
				if !strings.Contains(output, class) {
					t.Errorf("Expected %s to contain responsive class %q", page.name, class)
				}
			}

			// Test viewport meta tag
			if !strings.Contains(output, `name="viewport"`) {
				t.Errorf("Expected %s to contain viewport meta tag", page.name)
			}
		})
	}
}
