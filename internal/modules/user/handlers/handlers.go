package handlers

import (
	"net/http"
	"strconv"

	"go-templ-template/internal/modules/user/application"
	"go-templ-template/internal/modules/user/domain"

	"github.com/labstack/echo/v4"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService application.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService application.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser handles POST /api/v1/users
func (h *UserHandler) CreateUser(c echo.Context) error {
	var req CreateUserRequest
	if err := BindAndValidate(c, &req); err != nil {
		return h.handleValidationError(c, err)
	}

	cmd := &application.CreateUserCommand{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	user, err := h.userService.CreateUser(c.Request().Context(), cmd)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	response := ToUserResponse(user)
	return c.JSON(http.StatusCreated, SuccessResponse{
		Message: "User created successfully",
		Data:    response,
	})
}

// GetUser handles GET /api/v1/users/:id
func (h *UserHandler) GetUser(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "User ID is required",
			Field:   "id",
		})
	}

	query := &application.GetUserQuery{ID: id}
	user, err := h.userService.GetUser(c.Request().Context(), query)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	response := ToUserResponse(user)
	return c.JSON(http.StatusOK, response)
}

// GetUserByEmail handles GET /api/v1/users/by-email/:email
func (h *UserHandler) GetUserByEmail(c echo.Context) error {
	email := c.Param("email")
	if email == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Email is required",
			Field:   "email",
		})
	}

	query := &application.GetUserByEmailQuery{Email: email}
	user, err := h.userService.GetUserByEmail(c.Request().Context(), query)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	response := ToUserResponse(user)
	return c.JSON(http.StatusOK, response)
}

// UpdateUser handles PUT /api/v1/users/:id
func (h *UserHandler) UpdateUser(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "User ID is required",
			Field:   "id",
		})
	}

	var req UpdateUserRequest
	if err := BindAndValidate(c, &req); err != nil {
		return h.handleValidationError(c, err)
	}

	cmd := &application.UpdateUserCommand{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Version:   req.Version,
	}

	user, err := h.userService.UpdateUser(c.Request().Context(), cmd)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	response := ToUserResponse(user)
	return c.JSON(http.StatusOK, SuccessResponse{
		Message: "User updated successfully",
		Data:    response,
	})
}

// UpdateUserEmail handles PUT /api/v1/users/:id/email
func (h *UserHandler) UpdateUserEmail(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "User ID is required",
			Field:   "id",
		})
	}

	var req UpdateUserEmailRequest
	if err := BindAndValidate(c, &req); err != nil {
		return h.handleValidationError(c, err)
	}

	cmd := &application.UpdateUserEmailCommand{
		ID:      id,
		Email:   req.Email,
		Version: req.Version,
	}

	user, err := h.userService.UpdateUserEmail(c.Request().Context(), cmd)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	response := ToUserResponse(user)
	return c.JSON(http.StatusOK, SuccessResponse{
		Message: "User email updated successfully",
		Data:    response,
	})
}

// ChangeUserPassword handles PUT /api/v1/users/:id/password
func (h *UserHandler) ChangeUserPassword(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "User ID is required",
			Field:   "id",
		})
	}

	var req ChangeUserPasswordRequest
	if err := BindAndValidate(c, &req); err != nil {
		return h.handleValidationError(c, err)
	}

	cmd := &application.ChangeUserPasswordCommand{
		ID:          id,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
		Version:     req.Version,
	}

	user, err := h.userService.ChangeUserPassword(c.Request().Context(), cmd)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	response := ToUserResponse(user)
	return c.JSON(http.StatusOK, SuccessResponse{
		Message: "Password changed successfully",
		Data:    response,
	})
}

// ChangeUserStatus handles PUT /api/v1/users/:id/status
func (h *UserHandler) ChangeUserStatus(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "User ID is required",
			Field:   "id",
		})
	}

	var req ChangeUserStatusRequest
	if err := BindAndValidate(c, &req); err != nil {
		return h.handleValidationError(c, err)
	}

	cmd := &application.ChangeUserStatusCommand{
		ID:        id,
		Status:    req.Status,
		ChangedBy: req.ChangedBy,
		Reason:    req.Reason,
		Version:   req.Version,
	}

	user, err := h.userService.ChangeUserStatus(c.Request().Context(), cmd)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	response := ToUserResponse(user)
	return c.JSON(http.StatusOK, SuccessResponse{
		Message: "User status changed successfully",
		Data:    response,
	})
}

// DeleteUser handles DELETE /api/v1/users/:id
func (h *UserHandler) DeleteUser(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "User ID is required",
			Field:   "id",
		})
	}

	// Get optional query parameters
	deletedBy := c.QueryParam("deleted_by")
	reason := c.QueryParam("reason")

	cmd := &application.DeleteUserCommand{
		ID:        id,
		DeletedBy: deletedBy,
		Reason:    reason,
	}

	err := h.userService.DeleteUser(c.Request().Context(), cmd)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	return c.JSON(http.StatusOK, SuccessResponse{
		Message: "User deleted successfully",
	})
}

// ListUsers handles GET /api/v1/users
func (h *UserHandler) ListUsers(c echo.Context) error {
	var req ListUsersRequest

	// Parse query parameters
	if status := c.QueryParam("status"); status != "" {
		userStatus := domain.UserStatus(status)
		req.Status = &userStatus
	}
	if email := c.QueryParam("email"); email != "" {
		req.Email = &email
	}
	if firstName := c.QueryParam("first_name"); firstName != "" {
		req.FirstName = &firstName
	}
	if lastName := c.QueryParam("last_name"); lastName != "" {
		req.LastName = &lastName
	}
	if createdAfter := c.QueryParam("created_after"); createdAfter != "" {
		req.CreatedAfter = &createdAfter
	}
	if createdBefore := c.QueryParam("created_before"); createdBefore != "" {
		req.CreatedBefore = &createdBefore
	}

	// Parse limit and offset
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = offset
		}
	}

	// Validate request
	if err := ValidateListUsersRequest(&req); err != nil {
		return h.handleValidationError(c, err)
	}

	query := &application.ListUsersQuery{
		Status:        req.Status,
		Email:         req.Email,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		CreatedAfter:  req.CreatedAfter,
		CreatedBefore: req.CreatedBefore,
		Limit:         req.Limit,
		Offset:        req.Offset,
	}

	users, total, err := h.userService.ListUsers(c.Request().Context(), query)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	response := &ListUsersResponse{
		Users:   ToUserResponseList(users),
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: int64(req.Offset+req.Limit) < total,
	}

	return c.JSON(http.StatusOK, response)
}

// handleValidationError handles validation errors
func (h *UserHandler) handleValidationError(c echo.Context, err error) error {
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
func (h *UserHandler) handleApplicationError(c echo.Context, err error) error {
	if appErr, ok := err.(*application.ApplicationError); ok {
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
func (h *UserHandler) getStatusCodeForError(errorCode string) int {
	switch errorCode {
	case application.ErrCodeValidation:
		return http.StatusBadRequest
	case application.ErrCodeUserNotFound:
		return http.StatusNotFound
	case application.ErrCodeUserAlreadyExists:
		return http.StatusConflict
	case application.ErrCodeInvalidPassword:
		return http.StatusUnauthorized
	case application.ErrCodeOptimisticLock:
		return http.StatusConflict
	case application.ErrCodeBusinessRule:
		return http.StatusBadRequest
	case application.ErrCodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
