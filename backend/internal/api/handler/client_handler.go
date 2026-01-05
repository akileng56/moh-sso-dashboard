package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

type ClientHandler struct {
	service      *service.ClientService
	auditService *service.AuditService
}

func NewClientHandler(s *service.ClientService, audit *service.AuditService) *ClientHandler {
	return &ClientHandler{service: s, auditService: audit}
}

func (h *ClientHandler) CreateClient(c *gin.Context) {
	var req service.CreateClientRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		_ = h.auditService.Log(
			c.Request.Context(),
			utils.ToNullUUID(c.GetString("user_id")),
			"client.create_failed",
			map[string]interface{}{
				"reason":     "invalid_body",
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			},
		)

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	newApp, err := h.service.CreateClient(req)
	if err != nil {
		log.Printf("ERROR: Failed to create client: %v", err)

		_ = h.auditService.Log(
			c.Request.Context(),
			utils.ToNullUUID(c.GetString("user_id")),
			"client.create_failed",
			map[string]interface{}{
				"client_id":  req.ClientID,
				"reason":     err.Error(),
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			},
		)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client"})
		return
	}

	_ = h.auditService.Log(
		c.Request.Context(),
		utils.ToNullUUID(c.GetString("user_id")),
		"client.create_success",
		map[string]interface{}{
			"client_id":  newApp.ClientID,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
	)

	c.JSON(http.StatusCreated, newApp)
}

func (h *ClientHandler) GetClient(c *gin.Context) {
	id := c.Param("id")
	if id == "" {

		_ = h.auditService.Log(
			c.Request.Context(),
			utils.ToNullUUID(c.GetString("user_id")),
			"client.get_failed",
			map[string]interface{}{
				"reason":     "missing_id",
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			},
		)

		c.JSON(http.StatusBadRequest, gin.H{"error": "Client ID is required"})
		return
	}

	app, err := h.service.GetClient(id)
	if err != nil {
		log.Printf("ERROR: Failed to get client ID %s: %v", id, err)

		_ = h.auditService.Log(
			c.Request.Context(),
			utils.ToNullUUID(c.GetString("user_id")),
			"client.get_failed",
			map[string]interface{}{
				"client_id":  id,
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			},
		)

		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	_ = h.auditService.Log(
		c.Request.Context(),
		utils.ToNullUUID(c.GetString("user_id")),
		"client.success",
		map[string]interface{}{
			"client_id":  id,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
	)

	c.JSON(http.StatusOK, app)
}

func (h *ClientHandler) ListClients(c *gin.Context) {

	_ = h.auditService.Log(
		c.Request.Context(),
		utils.ToNullUUID(c.GetString("user_id")),
		"client.lists",
		map[string]interface{}{
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
	)

	apps, err := h.service.ListClients()
	if err != nil {
		log.Printf("ERROR: Failed to list clients: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list clients"})
		return
	}

	clientRoles := c.MustGet("client_roles").(map[string][]string)
	isAdmin := c.GetBool("is_admin")

	filtered := []model.Client{}

	for _, client := range apps {

		if client.Attributes == nil || client.Attributes["icon"] == "" {
			continue
		}

		if isAdmin {
			filtered = append(filtered, client)
			continue
		}

		clientID := client.ClientID

		roles, ok := clientRoles[clientID]
		if !ok || len(roles) == 0 {
			continue
		}

		expectedRole := clientID + "_access"

		hasAccess := false
		for _, r := range roles {
			if r == expectedRole {
				hasAccess = true
				break
			}
		}

		if hasAccess {
			filtered = append(filtered, client)
		}
	}

	c.JSON(http.StatusOK, filtered)
}

func (h *ClientHandler) DeleteClient(c *gin.Context) {
	id := c.Param("id")
	if id == "" {

		_ = h.auditService.Log(
			c.Request.Context(),
			utils.ToNullUUID(c.GetString("user_id")),
			"client.delete_failed",
			map[string]interface{}{
				"reason":     "missing_id",
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			},
		)

		c.JSON(http.StatusBadRequest, gin.H{"error": "Client ID is required"})
		return
	}

	uid, err := uuid.Parse(id)
	if err != nil {

		_ = h.auditService.Log(
			c.Request.Context(),
			utils.ToNullUUID(c.GetString("user_id")),
			"client.delete_failed",
			map[string]interface{}{
				"client_id":  id,
				"reason":     "invalid_uuid",
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			},
		)

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	if err := h.service.DeleteClient(uid); err != nil {
		log.Printf("ERROR: Failed to delete client ID %s: %v", id, err)

		_ = h.auditService.Log(
			c.Request.Context(),
			utils.ToNullUUID(c.GetString("user_id")),
			"client.delete_failed",
			map[string]interface{}{
				"client_id":  id,
				"reason":     err.Error(),
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			},
		)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete client"})
		return
	}

	_ = h.auditService.Log(
		c.Request.Context(),
		utils.ToNullUUID(c.GetString("user_id")),
		"client.delete_success",
		map[string]interface{}{
			"client_id":  id,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
	)

	c.Status(http.StatusNoContent)
}
