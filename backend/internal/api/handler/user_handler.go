package handler

import (
	"log"
	"net/http" // Still needed for HTTP status codes

	"github.com/gin-gonic/gin" // Import the Gin framework

	models "github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/service"
)

// UserResponse is a "safe" version of the User model for JSON responses.
// Notice it omits the HashedPassword.
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	IsActive bool   `json:"is_active"`
}

// toUserResponse is a helper function to convert the internal model
// to the external-facing response DTO.
func toUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
}

// UserHandler holds the user service.
type UserHandler struct {
	service *service.UserService // Use a pointer
}

// NewUserHandler creates a new handler with its service dependency.
func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

// CreateUser is the Gin handler for POST /users.
func (h *UserHandler) CreateUser(c *gin.Context) {
	// 1. Decode the JSON request
	var req service.CreateUserRequest
	
	// Gin's ShouldBindJSON handles decoding and error checking
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 2. Call the service
	newUser, err := h.service.CreateUser(req)
	if err != nil {
		log.Printf("ERROR: Failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 3. Convert to the safe response DTO and write the response
	response := toUserResponse(newUser)
	c.JSON(http.StatusCreated, response)
}

// GetUser is the Gin handler for GET /users/:id.
func (h *UserHandler) GetUser(c *gin.Context) {
	// 1. Get the "id" from the URL path
	id := c.Param("id") // Use c.Param to extract the path variable
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// 2. Call the service
	user, err := h.service.GetUser(id)
	if err != nil {
		log.Printf("ERROR: Failed to get user ID %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 3. Convert to the safe response DTO and write the response
	response := toUserResponse(user)
	c.JSON(http.StatusOK, response)
}

// ListUsers is the Gin handler for GET /users.
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 1. Call the service
	users, err := h.service.ListUsers()
	if err != nil {
		log.Printf("ERROR: Failed to list users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	// 2. Convert the *slice* of models to a *slice* of response DTOs
	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = toUserResponse(&user) // Convert each user
	}

	// 3. Write the successful response
	c.JSON(http.StatusOK, responses)
}

// DeleteUser is the Gin handler for DELETE /users/:id.
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// 1. Get the "id" from the URL path
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// 2. Call the service
	if err := h.service.DeleteUser(id); err != nil {
		log.Printf("ERROR: Failed to delete user ID %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	// 3. Write the successful (empty) response
	c.Status(http.StatusNoContent) // 204 No Content
}