package handlers

import (
	"net/http"
	"time"

	"go-templ-template/internal/modules/auth/application"
	"go-templ-template/internal/modules/auth/domain"
	userDomain "go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/shared/middleware"

	"github.com/labstack/echo/v4"
)

// AuthHandler handles HTTP requests for authentication operations
type AuthHandler struct {
	authService application.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService application.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := BindAndValidate(c, &req); err != nil {
		return h.handleValidationError(c, err)
	}

	cmd := &application.LoginCommand{
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	}

	result, err := h.authService.Login(c.Request().Context(), cmd)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	// Set session cookie
	sessionDuration := int(24 * time.Hour / time.Second) // 24 hours in seconds
	middleware.SetSessionCookie(c, result.Session.ID, sessionDuration)

	response := ToAuthResponse(result.User, result.Session, "Login successful")
	return c.JSON(http.StatusOK, response)
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := BindAndValidate(c, &req); err != nil {
		return h.handleValidationError(c, err)
	}

	cmd := &application.RegisterCommand{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	}

	result, err := h.authService.Register(c.Request().Context(), cmd)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	// Set session cookie
	sessionDuration := int(24 * time.Hour / time.Second) // 24 hours in seconds
	middleware.SetSessionCookie(c, result.Session.ID, sessionDuration)

	response := ToAuthResponse(result.User, result.Session, "Registration successful")
	return c.JSON(http.StatusCreated, response)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c echo.Context) error {
	// Get session from context (set by auth middleware)
	sessionData := middleware.GetSessionFromContext(c)
	if sessionData == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "No active session found",
		})
	}

	// Get user from context (set by auth middleware)
	userData := middleware.GetUserFromContext(c)
	if userData == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "No active user found",
		})
	}

	// Type assert to get session and user
	session, ok := sessionData.(*domain.Session)
	if !ok {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Invalid session data",
		})
	}

	user, ok := userData.(*userDomain.User)
	if !ok {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Invalid user data",
		})
	}

	cmd := &application.LogoutCommand{
		SessionID: session.ID,
		UserID:    user.ID,
	}

	err := h.authService.Logout(c.Request().Context(), cmd)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	// Clear session cookie
	middleware.SetSessionCookie(c, "", -1)

	return c.JSON(http.StatusOK, SuccessResponse{
		Message: "Logout successful",
	})
}

// Me handles GET /api/v1/auth/me
func (h *AuthHandler) Me(c echo.Context) error {
	// Get user from context (set by auth middleware)
	userData := middleware.GetUserFromContext(c)
	if userData == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "Authentication required",
		})
	}

	// Type assert to get user
	user, ok := userData.(*userDomain.User)
	if !ok {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Invalid user data",
		})
	}

	response := ToUserResponse(user)
	return c.JSON(http.StatusOK, response)
}

// RefreshSession handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshSession(c echo.Context) error {
	// Get session from context (set by auth middleware)
	sessionData := middleware.GetSessionFromContext(c)
	if sessionData == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "No active session found",
		})
	}

	// Type assert to get session
	session, ok := sessionData.(*domain.Session)
	if !ok {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Invalid session data",
		})
	}

	cmd := &application.RefreshSessionCommand{
		SessionID: session.ID,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	}

	refreshedSession, err := h.authService.RefreshSession(c.Request().Context(), cmd)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	// Update session cookie with new expiration
	sessionDuration := int(24 * time.Hour / time.Second) // 24 hours in seconds
	middleware.SetSessionCookie(c, refreshedSession.ID, sessionDuration)

	response := ToSessionResponse(refreshedSession)
	return c.JSON(http.StatusOK, SuccessResponse{
		Message: "Session refreshed successfully",
		Data:    response,
	})
}

// ChangePassword handles PUT /api/v1/auth/password
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	// Get user from context (set by auth middleware)
	userData := middleware.GetUserFromContext(c)
	if userData == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "Authentication required",
		})
	}

	// Type assert to get user
	user, ok := userData.(*userDomain.User)
	if !ok {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Invalid user data",
		})
	}

	var req ChangePasswordRequest
	if err := BindAndValidate(c, &req); err != nil {
		return h.handleValidationError(c, err)
	}

	cmd := &application.ChangePasswordCommand{
		UserID:      user.ID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
		IPAddress:   c.RealIP(),
		UserAgent:   c.Request().UserAgent(),
	}

	err := h.authService.ChangePassword(c.Request().Context(), cmd)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	// Clear session cookie to force re-login
	middleware.SetSessionCookie(c, "", -1)

	return c.JSON(http.StatusOK, SuccessResponse{
		Message: "Password changed successfully. Please log in again.",
	})
}

// ValidateSession handles GET /api/v1/auth/validate
func (h *AuthHandler) ValidateSession(c echo.Context) error {
	sessionID := c.QueryParam("session_id")
	if sessionID == "" {
		// Try to get from cookie or header
		cookie, err := c.Cookie(middleware.SessionCookieName)
		if err == nil {
			sessionID = cookie.Value
		}
	}

	if sessionID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Session ID is required",
			Field:   "session_id",
		})
	}

	query := &application.ValidateSessionQuery{
		SessionID: sessionID,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	}

	result, err := h.authService.ValidateSession(c.Request().Context(), query)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	if !result.Valid {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "INVALID_SESSION",
			Message: "Session is invalid or expired",
		})
	}

	response := map[string]interface{}{
		"valid":   true,
		"user":    ToUserResponse(result.User),
		"session": ToSessionResponse(result.Session),
	}

	return c.JSON(http.StatusOK, response)
}

// handleValidationError handles validation errors
func (h *AuthHandler) handleValidationError(c echo.Context, err error) error {
	if validationErrs, ok := err.(ValidationErrors); ok {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "VALIDATION_ERROR",
			"message": "Validation failed",
			"details": validationErrs.Errors,
		})
	}

	return c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   "VALIDATION_ERROR",
		Message: err.Error(),
	})
}

// handleApplicationError handles application layer errors
func (h *AuthHandler) handleApplicationError(c echo.Context, err error) error {
	if appErr, ok := err.(*application.AuthError); ok {
		statusCode := h.getStatusCodeForError(appErr.Code)
		return c.JSON(statusCode, ErrorResponse{
			Error:   appErr.Code,
			Message: appErr.Message,
			Field:   appErr.Field,
		})
	}

	// Handle unknown errors
	return c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   "INTERNAL_ERROR",
		Message: "An unexpected error occurred",
	})
}

// getStatusCodeForError maps application error codes to HTTP status codes
func (h *AuthHandler) getStatusCodeForError(errorCode string) int {
	switch errorCode {
	case "VALIDATION_ERROR":
		return http.StatusBadRequest
	case "INVALID_CREDENTIALS":
		return http.StatusUnauthorized
	case "USER_ALREADY_EXISTS":
		return http.StatusConflict
	case "USER_NOT_FOUND":
		return http.StatusNotFound
	case "SESSION_NOT_FOUND":
		return http.StatusNotFound
	case "SESSION_EXPIRED":
		return http.StatusUnauthorized
	case "SESSION_INVALID":
		return http.StatusUnauthorized
	case "ACCOUNT_SUSPENDED":
		return http.StatusForbidden
	case "RATE_LIMIT_EXCEEDED":
		return http.StatusTooManyRequests
	case "INTERNAL_ERROR":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
