package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/http/apierror"
	"github.com/moh-sso-dashboard/internal/http/response"
	models "github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

/* =========================================================
 * Helpers
 * ========================================================= */

func toUserResponse(u *models.User) models.UserResponse {
	fullName := strings.TrimSpace(
		strings.Join([]string{u.FirstName, u.LastName}, " "),
	)

	return models.UserResponse{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		FullName: fullName,

		IsAdmin:     u.IsAdmin,
		RealmRoles:  u.RealmRoles,
		ClientRoles: u.ClientRoles,

		IsActive:         u.Enabled,
		EmailVerified:    u.EmailVerified,
		RequirePwdChange: u.RequirePwdChange,

		LastLoginAt: u.LastLoginAt,
		CreatedAt:   &u.CreatedAt,
	}
}

type UserHandler struct {
	service      *service.UserService
	auditService *service.AuditService
}

func NewUserHandler(
	s *service.UserService,
	audit *service.AuditService,
) *UserHandler {
	return &UserHandler{
		service:      s,
		auditService: audit,
	}
}

/* =========================================================
 * Create User
 * ========================================================= */

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req service.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.audit(
			c,
			"user.create_failed",
			map[string]interface{}{
				"reason": "invalid_body",
			},
		)

		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Invalid request payload",
		)
		return
	}

	actorID, _ := uuid.Parse(c.GetString("user_id"))

	user, err := h.service.CreateUser(
		c.Request.Context(),
		req,
		actorID,
	)
	if err != nil {
		h.audit(
			c,
			"user.create_failed",
			map[string]interface{}{
				"username": req.Username,
				"email":    req.Email,
				"reason":   err.Error(),
			},
		)

		var apiErr *apierror.APIError
		if errors.As(err, &apiErr) {
			response.Fail(
				c,
				apiErr.HTTPStatus,
				apiErr.Code,
				apiErr.Message,
			)
			return
		}

		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to create user",
		)
		return
	}

	h.audit(
		c,
		"user.create_success",
		map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	)

	response.OK(
		c,
		http.StatusCreated,
		toUserResponse(user),
	)
}

/* =========================================================
 * Get User
 * ========================================================= */

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.audit(
			c,
			"user.get_failed",
			map[string]interface{}{
				"reason": "missing_id",
			},
		)

		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"User ID is required",
		)
		return
	}

	userID, err := uuid.Parse(id)
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_UUID",
			"Invalid user ID format",
		)
		return
	}

	user, err := h.service.GetUser(userID)
	if err != nil {
		h.audit(
			c,
			"user.get_failed",
			map[string]interface{}{
				"user_id": id,
			},
		)

		response.Fail(
			c,
			http.StatusNotFound,
			"USER_NOT_FOUND",
			"User not found",
		)
		return
	}

	h.audit(
		c,
		"user.get_success",
		map[string]interface{}{
			"user_id": id,
		},
	)

	response.OK(
		c,
		http.StatusOK,
		toUserResponse(user),
	)
}

/* =========================================================
 * List Users
 * ========================================================= */

func (h *UserHandler) ListUsers(c *gin.Context) {
	h.audit(
		c,
		"user.list",
		nil,
	)

	users, err := h.service.ListUsers()
	if err != nil {

		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to list users",
		)
		return
	}

	responses := make([]models.UserResponse, len(users))
	for i := range users {
		responses[i] = toUserResponse(&users[i])
	}

	response.OK(
		c,
		http.StatusOK,
		responses,
	)
}

/* =========================================================
 * Delete User
 * ========================================================= */

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.audit(
			c,
			"user.delete_failed",
			map[string]interface{}{
				"reason": "missing_id",
			},
		)

		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"User ID is required",
		)
		return
	}

	userID, err := uuid.Parse(id)
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_UUID",
			"Invalid user ID format",
		)
		return
	}

	actorID, _ := uuid.Parse(c.GetString("user_id"))

	if err := h.service.DeleteUser(
		c.Request.Context(),
		userID,
		actorID,
	); err != nil {
		h.audit(
			c,
			"user.delete_failed",
			map[string]interface{}{
				"user_id": id,
				"reason":  err.Error(),
			},
		)

		var apiErr *apierror.APIError
		if errors.As(err, &apiErr) {
			response.Fail(
				c,
				apiErr.HTTPStatus,
				apiErr.Code,
				apiErr.Message,
			)
			return
		}

		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to delete user",
		)
		return
	}

	h.audit(
		c,
		"user.delete_success",
		map[string]interface{}{
			"user_id": id,
		},
	)

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * Audit helper
 * ========================================================= */

func (h *UserHandler) audit(
	c *gin.Context,
	action string,
	meta map[string]interface{},
) {
	if meta == nil {
		meta = map[string]interface{}{}
	}

	meta["ip"] = c.ClientIP()
	meta["user_agent"] = c.Request.UserAgent()

	_ = h.auditService.Log(
		c.Request.Context(),
		utils.ToNullUUID(c.GetString("user_id")),
		action,
		meta,
	)
}
