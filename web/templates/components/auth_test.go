package components

import (
	"context"
	"strings"
	"testing"
)

func TestLoginForm(t *testing.T) {
	tests := []struct {
		name   string
		errors map[string]string
		email  string
	}{
		{
			name:   "empty form",
			errors: map[string]string{},
			email:  "",
		},
		{
			name: "with validation errors",
			errors: map[string]string{
				"email":    "Invalid email format",
				"password": "Password is required",
				"general":  "Login failed",
			},
			email: "invalid-email",
		},
		{
			name:   "with prefilled email",
			errors: map[string]string{},
			email:  "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := LoginForm(tt.errors, tt.email).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render login form: %v", err)
			}

			output := buf.String()

			// Test basic form structure
			expectedElements := []string{
				"<form",
				"method=\"POST\"",
				"action=\"/auth/login\"",
				"Sign In",
				"Welcome back!",
				"type=\"email\"",
				"name=\"email\"",
				"type=\"password\"",
				"name=\"password\"",
				"type=\"checkbox\"",
				"name=\"remember\"",
				"Remember me",
				"type=\"submit\"",
			}

			for _, expected := range expectedElements {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected login form to contain %q", expected)
				}
			}

			// Test email value
			if tt.email != "" {
				if !strings.Contains(output, `value="`+tt.email+`"`) {
					t.Errorf("Expected email field to have value %q", tt.email)
				}
			}

			// Test error display
			for field, errorMsg := range tt.errors {
				if field != "general" && errorMsg != "" {
					if !strings.Contains(output, "border-red-300") {
						t.Errorf("Expected error styling for field %q", field)
					}
				}
				if errorMsg != "" {
					if !strings.Contains(output, errorMsg) {
						t.Errorf("Expected error message %q to be displayed", errorMsg)
					}
				}
			}

			// Test footer links
			footerLinks := []string{
				"/auth/forgot-password",
				"Forgot your password?",
				"/auth/register",
				"Sign up",
			}

			for _, link := range footerLinks {
				if !strings.Contains(output, link) {
					t.Errorf("Expected login form to contain %q", link)
				}
			}

			// Test styling classes
			stylingClasses := []string{
				"card",
				"form-label",
				"form-input",
				"btn-primary",
			}

			for _, class := range stylingClasses {
				if !strings.Contains(output, class) {
					t.Errorf("Expected login form to contain class %q", class)
				}
			}
		})
	}
}

func TestRegisterForm(t *testing.T) {
	tests := []struct {
		name     string
		errors   map[string]string
		formData RegisterFormData
	}{
		{
			name:   "empty form",
			errors: map[string]string{},
			formData: RegisterFormData{
				FirstName: "",
				LastName:  "",
				Email:     "",
			},
		},
		{
			name: "with validation errors",
			errors: map[string]string{
				"first_name":       "First name is required",
				"last_name":        "Last name is required",
				"email":            "Invalid email format",
				"password":         "Password must be at least 8 characters",
				"confirm_password": "Passwords do not match",
				"terms":            "You must accept the terms",
				"general":          "Registration failed",
			},
			formData: RegisterFormData{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
		},
		{
			name:   "with prefilled data",
			errors: map[string]string{},
			formData: RegisterFormData{
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := RegisterForm(tt.errors, tt.formData).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render register form: %v", err)
			}

			output := buf.String()

			// Test basic form structure
			expectedElements := []string{
				"<form",
				"method=\"POST\"",
				"action=\"/auth/register\"",
				"Create Account",
				"Join us today!",
				"name=\"first_name\"",
				"name=\"last_name\"",
				"name=\"email\"",
				"name=\"password\"",
				"name=\"confirm_password\"",
				"name=\"terms\"",
				"type=\"submit\"",
			}

			for _, expected := range expectedElements {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected register form to contain %q", expected)
				}
			}

			// Test form data values
			if tt.formData.FirstName != "" {
				if !strings.Contains(output, `value="`+tt.formData.FirstName+`"`) {
					t.Errorf("Expected first name field to have value %q", tt.formData.FirstName)
				}
			}
			if tt.formData.LastName != "" {
				if !strings.Contains(output, `value="`+tt.formData.LastName+`"`) {
					t.Errorf("Expected last name field to have value %q", tt.formData.LastName)
				}
			}
			if tt.formData.Email != "" {
				if !strings.Contains(output, `value="`+tt.formData.Email+`"`) {
					t.Errorf("Expected email field to have value %q", tt.formData.Email)
				}
			}

			// Test error display
			for field, errorMsg := range tt.errors {
				if field != "general" && errorMsg != "" {
					if !strings.Contains(output, "border-red-300") {
						t.Errorf("Expected error styling for field %q", field)
					}
				}
				if errorMsg != "" {
					if !strings.Contains(output, errorMsg) {
						t.Errorf("Expected error message %q to be displayed", errorMsg)
					}
				}
			}

			// Test terms and conditions links
			termsLinks := []string{
				"/terms",
				"Terms of Service",
				"/privacy",
				"Privacy Policy",
			}

			for _, link := range termsLinks {
				if !strings.Contains(output, link) {
					t.Errorf("Expected register form to contain %q", link)
				}
			}

			// Test footer link
			if !strings.Contains(output, "/auth/login") {
				t.Error("Expected register form to contain login link")
			}
			if !strings.Contains(output, "Sign in") {
				t.Error("Expected register form to contain sign in text")
			}

			// Test password requirements
			if !strings.Contains(output, "Password must be at least 8 characters long") {
				t.Error("Expected password requirements to be displayed")
			}
		})
	}
}

func TestForgotPasswordForm(t *testing.T) {
	tests := []struct {
		name   string
		errors map[string]string
		email  string
	}{
		{
			name:   "empty form",
			errors: map[string]string{},
			email:  "",
		},
		{
			name: "with validation errors",
			errors: map[string]string{
				"email":   "Invalid email format",
				"general": "Email not found",
			},
			email: "invalid-email",
		},
		{
			name:   "with prefilled email",
			errors: map[string]string{},
			email:  "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := ForgotPasswordForm(tt.errors, tt.email).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render forgot password form: %v", err)
			}

			output := buf.String()

			// Test basic form structure
			expectedElements := []string{
				"<form",
				"method=\"POST\"",
				"action=\"/auth/forgot-password\"",
				"Forgot Password",
				"Enter your email address",
				"type=\"email\"",
				"name=\"email\"",
				"Send Reset Link",
			}

			for _, expected := range expectedElements {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected forgot password form to contain %q", expected)
				}
			}

			// Test email value
			if tt.email != "" {
				if !strings.Contains(output, `value="`+tt.email+`"`) {
					t.Errorf("Expected email field to have value %q", tt.email)
				}
			}

			// Test error display
			for _, errorMsg := range tt.errors {
				if errorMsg != "" {
					if !strings.Contains(output, errorMsg) {
						t.Errorf("Expected error message %q to be displayed", errorMsg)
					}
				}
			}

			// Test footer links
			footerLinks := []string{
				"/auth/login",
				"Sign in",
				"/auth/register",
				"Sign up",
			}

			for _, link := range footerLinks {
				if !strings.Contains(output, link) {
					t.Errorf("Expected forgot password form to contain %q", link)
				}
			}
		})
	}
}

func TestResetPasswordForm(t *testing.T) {
	tests := []struct {
		name   string
		errors map[string]string
		token  string
	}{
		{
			name:   "valid token",
			errors: map[string]string{},
			token:  "valid-reset-token",
		},
		{
			name: "with validation errors",
			errors: map[string]string{
				"password":         "Password must be at least 8 characters",
				"confirm_password": "Passwords do not match",
				"general":          "Invalid or expired token",
			},
			token: "invalid-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := ResetPasswordForm(tt.errors, tt.token).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("Failed to render reset password form: %v", err)
			}

			output := buf.String()

			// Test basic form structure
			expectedElements := []string{
				"<form",
				"method=\"POST\"",
				"action=\"/auth/reset-password\"",
				"Reset Password",
				"Enter your new password",
				"type=\"hidden\"",
				"name=\"token\"",
				"name=\"password\"",
				"name=\"confirm_password\"",
				"Reset Password",
			}

			for _, expected := range expectedElements {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected reset password form to contain %q", expected)
				}
			}

			// Test token value
			if !strings.Contains(output, `value="`+tt.token+`"`) {
				t.Errorf("Expected token field to have value %q", tt.token)
			}

			// Test error display
			for _, errorMsg := range tt.errors {
				if errorMsg != "" {
					if !strings.Contains(output, errorMsg) {
						t.Errorf("Expected error message %q to be displayed", errorMsg)
					}
				}
			}

			// Test password requirements
			if !strings.Contains(output, "Password must be at least 8 characters long") {
				t.Error("Expected password requirements to be displayed")
			}

			// Test footer link
			if !strings.Contains(output, "/auth/login") {
				t.Error("Expected reset password form to contain login link")
			}
		})
	}
}

func TestRegisterFormData(t *testing.T) {
	// Test RegisterFormData struct
	formData := RegisterFormData{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	if formData.FirstName != "John" {
		t.Errorf("Expected FirstName to be 'John', got %q", formData.FirstName)
	}
	if formData.LastName != "Doe" {
		t.Errorf("Expected LastName to be 'Doe', got %q", formData.LastName)
	}
	if formData.Email != "john@example.com" {
		t.Errorf("Expected Email to be 'john@example.com', got %q", formData.Email)
	}
}
