package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	models "github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

func toUserResponse(user *models.User) models.UserResponse {
	return models.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
}

type UserHandler struct {
	service      *service.UserService
	auditService *service.AuditService
}

func NewUserHandler(s *service.UserService, audit *service.AuditService) *UserHandler {
	return &UserHandler{service: s, auditService: audit}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {

		h.auditService.Log(
			c.Request.Context(),
			utils.ToNullUUID(c.GetString("user_id")),
			"create_user_failed",
			map[string]interface{}{
				"reason":     "invalid_body",
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			},
		)

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	newUser, err := h.service.CreateUser(req)
	if err != nil {
		log.Printf("ERROR: Failed to create user: %v", err)

		h.auditService.Log(
			c.Request.Context(),
			utils.ToNullUUID(c.GetString("user_id")),
			"create_user_failed",
			map[string]interface{}{
				"username":   req.Username,
				"email":      req.Email,
				"reason":     err.Error(),
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			},
		)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	h.auditService.Log(
		c.Request.Context(),
		utils.ToNullUUID(c.GetString("user_id")),
		"create_user_success",
		map[string]interface{}{
			"user_id":    newUser.ID,
			"username":   newUser.Username,
			"email":      newUser.Email,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
	)

	response := toUserResponse(newUser)
	c.JSON(http.StatusCreated, response)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {

		h.auditService.Log(
			c.Request.Context(),
			utils.ToNullUUID(c.GetString("user_id")),
			"get_user_failed",
			map[string]interface{}{
				"reason":     "missing_id",
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			},
		)

		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}
	user, err := h.service.GetUser(id)
	if err != nil {
		log.Printf("ERROR: Failed to get user ID %s: %v", id, err)

		h.auditService.Log(
			c.Request.Context(),
			utils.ToNullUUID(c.GetString("user_id")),
			"get_user_not_found",
			map[string]interface{}{
				"user_id":    id,
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			},
		)

		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// AUDIT: success <-- ADDED
	h.auditService.Log(
		c.Request.Context(),
		utils.ToNullUUID(c.GetString("user_id")),
		"get_user_success",
		map[string]interface{}{
			"user_id":    id,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
	)

	response := toUserResponse(user)
	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) ListUsers(c *gin.Context) {

	h.auditService.Log(
		c.Request.Context(),
		utils.ToNullUUID(c.GetString("user_id")),
		"list_users",
		map[string]interface{}{
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
	)

	users, err := h.service.ListUsers()
	if err != nil {
		log.Printf("ERROR: Failed to list users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}
	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = toUserResponse(&user)
	}
	c.JSON(http.StatusOK, responses)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {

		h.auditService.Log(
			c.Request.Context(),
			utils.ToNullUUID(c.GetString("user_id")),
			"delete_user_failed",
			map[string]interface{}{
				"reason":     "missing_id",
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			},
		)

		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}
	if err := h.service.DeleteUser(id); err != nil {
		log.Printf("ERROR: Failed to delete user ID %s: %v", id, err)

		h.auditService.Log(
			c.Request.Context(),
			utils.ToNullUUID(c.GetString("user_id")),
			"delete_user_failed",
			map[string]interface{}{
				"user_id":    id,
				"reason":     err.Error(),
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			},
		)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	h.auditService.Log(
		c.Request.Context(),
		utils.ToNullUUID(c.GetString("user_id")),
		"delete_user_success",
		map[string]interface{}{
			"user_id":    id,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
	)

	c.Status(http.StatusNoContent)
}
