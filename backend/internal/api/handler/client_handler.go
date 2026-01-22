package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/http/apierror"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/keycloak"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

type ClientHandler struct {
	service      *service.ClientService
	auditService *service.AuditService
}

func NewClientHandler(
	s *service.ClientService,
	audit *service.AuditService,
) *ClientHandler {
	return &ClientHandler{
		service:      s,
		auditService: audit,
	}
}

/* =========================================================
 * Create Client
 * ========================================================= */
func (h *ClientHandler) CreateClient(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, _ := uuid.Parse(userIDStr)

	var req service.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.audit(
			c,
			"client.create_failed",
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

	newClient, err := h.service.CreateClient(
		c.Request.Context(),
		req,
		userID,
	)
	if err != nil {
		h.audit(
			c,
			"client.create_failed",
			map[string]interface{}{
				"client_id": req.ClientID,
				"reason":    err.Error(),
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
			"Failed to create client",
		)
		return
	}

	h.audit(
		c,
		"client.create_success",
		map[string]interface{}{
			"client_id": newClient.ClientID,
		},
	)

	response.OK(c, http.StatusCreated, newClient)
}

/* =========================================================
 * Get Client
 * ========================================================= */
func (h *ClientHandler) GetClient(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.audit(
			c,
			"client.get_failed",
			map[string]interface{}{
				"reason": "missing_id",
			},
		)

		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Client ID is required",
		)
		return
	}

	uid, _ := uuid.Parse(id)

	client, err := h.service.GetClient(uid)
	if err != nil {
		h.audit(
			c,
			"client.get_failed",
			map[string]interface{}{
				"client_id": id,
			},
		)

		response.Fail(
			c,
			http.StatusNotFound,
			"CLIENT_NOT_FOUND",
			"Client not found",
		)
		return
	}

	h.audit(
		c,
		"client.get_success",
		map[string]interface{}{
			"client_id": id,
		},
	)

	response.OK(c, http.StatusOK, client)
}

/* =========================================================
 * List Clients
 * ========================================================= */
func (h *ClientHandler) ListClients(c *gin.Context) {
	h.audit(
		c,
		"client.list",
		nil,
	)

	clients, err := h.service.ListClients()
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to list clients",
		)
		return
	}

	clientRoles := c.MustGet("client_roles").(map[string][]string)
	isAdmin := c.GetBool("is_admin")

	filtered := make([]model.Client, 0)

	for _, client := range clients {

		if client.Attributes == nil || client.Attributes["ui.icon"] == "" {
			continue
		}

		if isAdmin {
			filtered = append(filtered, client)
			continue
		}

		roles := clientRoles[client.ClientID]
		expectedRole := client.ClientID + "_access"

		for _, r := range roles {
			if r == expectedRole {
				filtered = append(filtered, client)
				break
			}
		}
	}

	response.OK(c, http.StatusOK, filtered)
}

/* =========================================================
 * Delete Client
 * ========================================================= */
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.audit(
			c,
			"client.delete_failed",
			map[string]interface{}{
				"reason": "missing_id",
			},
		)

		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Client ID is required",
		)
		return
	}

	clientID, err := uuid.Parse(id)
	if err != nil {
		h.audit(
			c,
			"client.delete_failed",
			map[string]interface{}{
				"client_id": id,
				"reason":    "invalid_uuid",
			},
		)

		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_UUID",
			"Invalid client ID format",
		)
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))

	if err := h.service.DeleteClient(
		c.Request.Context(),
		clientID,
		userID,
	); err != nil {
		h.audit(
			c,
			"client.delete_failed",
			map[string]interface{}{
				"client_id": id,
				"reason":    err.Error(),
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
			"Failed to delete client",
		)
		return
	}

	h.audit(
		c,
		"client.delete_success",
		map[string]interface{}{
			"client_id": id,
		},
	)

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * CLIENT ROLES
 * ========================================================= */

// POST /clients/:id/roles
func (h *ClientHandler) CreateClientRole(c *gin.Context) {
	// ------------------------------------------------
	// 1) Parse client ID
	// ------------------------------------------------
	clientIDStr := c.Param("id")
	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid client ID")
		return
	}

	// ------------------------------------------------
	// 2) Bind request body (FIXED)
	// ------------------------------------------------
	var body *model.CreateClientRoleRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.Role == "" {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Role is required",
		)
		return
	}

	// ------------------------------------------------
	// 3) Get cached user from context (SOURCE OF TRUTH)
	// ------------------------------------------------
	userAny, exists := c.Get("user")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user context")
		return
	}

	user, ok := userAny.(*keycloak.AuthUser)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user context")
		return
	}

	// ------------------------------------------------
	// 4) Extract admin ID from cached profile
	// ------------------------------------------------
	adminID, err := uuid.Parse(user.ID)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user identity")
		return
	}

	// ------------------------------------------------
	// 5) Create role
	// ------------------------------------------------
	if err := h.service.CreateClientRole(
		c.Request.Context(),
		clientID,
		body,
		adminID,
	); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to create client role",
		)
		log.Printf("err", err)

		return
	}

	// ------------------------------------------------
	// 6) Audit
	// ------------------------------------------------
	h.audit(c, "client.role_created", map[string]interface{}{
		"client_id": clientID.String(),
		"role":      body.Role,
		"admin_id":  user.ID,
	})

	c.Status(http.StatusCreated)
}

// GET /clients/:id/roles
func (h *ClientHandler) ListClientRoles(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid client ID")
		return
	}

	roles, err := h.service.ListClientRoles(clientID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to list client roles",
		)
		return
	}

	h.audit(c, "client.roles_listed", map[string]interface{}{
		"client_id": clientID.String(),
	})

	response.OK(c, http.StatusOK, roles)
}

// DELETE /clients/:id/roles/:role
func (h *ClientHandler) DeleteClientRole(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid client ID")
		return
	}

	role := c.Param("role")
	if role == "" {
		response.Fail(c, http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Role is required",
		)
		return
	}

	adminID, _ := uuid.Parse(c.GetString("user_id"))

	if err := h.service.DeleteClientRole(
		c.Request.Context(),
		clientID,
		role,
		adminID,
	); err != nil {
		response.Fail(c, http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to delete client role",
		)

		return
	}

	h.audit(c, "client.role_deleted", map[string]interface{}{
		"client_id": clientID.String(),
		"role":      role,
	})

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * Audit helper
 * ========================================================= */
func (h *ClientHandler) audit(
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

/* =========================================================
 * Assign client role to user (ADMIN)
 * ========================================================= */
func (h *ClientHandler) AssignClientRoleToUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid user ID")
		return
	}

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid client ID")
		return
	}

	var body struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Role == "" {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Role is required",
		)
		return
	}

	adminID, _ := uuid.Parse(c.GetString("user_id"))

	if err := h.service.AssignClientRoleToUser(
		c.Request.Context(),
		userID,
		clientID,
		body.Role,
		adminID,
	); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to assign role to user",
		)
		return
	}

	h.audit(c, "client.role_assigned", map[string]interface{}{
		"user_id":   userID.String(),
		"client_id": clientID.String(),
		"role":      body.Role,
	})

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * Remove client role from user (ADMIN)
 * ========================================================= */
func (h *ClientHandler) RemoveClientRoleFromUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid user ID")
		return
	}

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid client ID")
		return
	}

	role := c.Param("role")
	if role == "" {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Role is required",
		)
		return
	}

	adminID, _ := uuid.Parse(c.GetString("user_id"))

	if err := h.service.RemoveClientRoleFromUser(
		c.Request.Context(),
		userID,
		clientID,
		role,
		adminID,
	); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to remove role from user",
		)
		return
	}

	h.audit(c, "client.role_removed", map[string]interface{}{
		"user_id":   userID.String(),
		"client_id": clientID.String(),
		"role":      role,
	})

	c.Status(http.StatusNoContent)
}
