package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/http/apierror"
	"github.com/moh-sso-dashboard/internal/http/response"
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

	client, err := h.service.GetClient(id)
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
		if client.Attributes == nil || client.Attributes["icon"] == "" {
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
