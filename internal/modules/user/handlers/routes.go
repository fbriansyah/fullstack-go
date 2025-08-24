package handlers

import (
	"go-templ-template/internal/modules/user/application"

	"github.com/labstack/echo/v4"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(e *echo.Echo, userService application.UserService) {
	userHandler := NewUserHandler(userService)

	// API v1 routes
	v1 := e.Group("/api/v1")

	// User routes
	users := v1.Group("/users")
	{
		users.POST("", userHandler.CreateUser)                     // POST /api/v1/users
		users.GET("", userHandler.ListUsers)                       // GET /api/v1/users
		users.GET("/:id", userHandler.GetUser)                     // GET /api/v1/users/:id
		users.GET("/by-email/:email", userHandler.GetUserByEmail)  // GET /api/v1/users/by-email/:email
		users.PUT("/:id", userHandler.UpdateUser)                  // PUT /api/v1/users/:id
		users.PUT("/:id/email", userHandler.UpdateUserEmail)       // PUT /api/v1/users/:id/email
		users.PUT("/:id/password", userHandler.ChangeUserPassword) // PUT /api/v1/users/:id/password
		users.PUT("/:id/status", userHandler.ChangeUserStatus)     // PUT /api/v1/users/:id/status
		users.DELETE("/:id", userHandler.DeleteUser)               // DELETE /api/v1/users/:id
	}
}

// RegisterUserRoutesWithMiddleware registers user routes with custom middleware
func RegisterUserRoutesWithMiddleware(e *echo.Echo, userService application.UserService, middlewares ...echo.MiddlewareFunc) {
	userHandler := NewUserHandler(userService)

	// API v1 routes
	v1 := e.Group("/api/v1")

	// Apply middleware to the group
	for _, middleware := range middlewares {
		v1.Use(middleware)
	}

	// User routes
	users := v1.Group("/users")
	{
		users.POST("", userHandler.CreateUser)                     // POST /api/v1/users
		users.GET("", userHandler.ListUsers)                       // GET /api/v1/users
		users.GET("/:id", userHandler.GetUser)                     // GET /api/v1/users/:id
		users.GET("/by-email/:email", userHandler.GetUserByEmail)  // GET /api/v1/users/by-email/:email
		users.PUT("/:id", userHandler.UpdateUser)                  // PUT /api/v1/users/:id
		users.PUT("/:id/email", userHandler.UpdateUserEmail)       // PUT /api/v1/users/:id/email
		users.PUT("/:id/password", userHandler.ChangeUserPassword) // PUT /api/v1/users/:id/password
		users.PUT("/:id/status", userHandler.ChangeUserStatus)     // PUT /api/v1/users/:id/status
		users.DELETE("/:id", userHandler.DeleteUser)               // DELETE /api/v1/users/:id
	}
}

// RegisterUserRoutesOnGroup registers user routes on a provided group (for module system)
func RegisterUserRoutesOnGroup(group *echo.Group, userService application.UserService) {
	userHandler := NewUserHandler(userService)

	// User routes - group is already /api/v1, so we create /users subgroup
	users := group.Group("/users")
	{
		users.POST("", userHandler.CreateUser)                     // POST /api/v1/users
		users.GET("", userHandler.ListUsers)                       // GET /api/v1/users
		users.GET("/:id", userHandler.GetUser)                     // GET /api/v1/users/:id
		users.GET("/by-email/:email", userHandler.GetUserByEmail)  // GET /api/v1/users/by-email/:email
		users.PUT("/:id", userHandler.UpdateUser)                  // PUT /api/v1/users/:id
		users.PUT("/:id/email", userHandler.UpdateUserEmail)       // PUT /api/v1/users/:id/email
		users.PUT("/:id/password", userHandler.ChangeUserPassword) // PUT /api/v1/users/:id/password
		users.PUT("/:id/status", userHandler.ChangeUserStatus)     // PUT /api/v1/users/:id/status
		users.DELETE("/:id", userHandler.DeleteUser)               // DELETE /api/v1/users/:id
	}
}
