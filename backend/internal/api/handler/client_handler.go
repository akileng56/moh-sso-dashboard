package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin" // Import the Gin framework
	"github.com/moh-sso-dashboard/internal/service"
)

type ClientHandler struct {
	service *service.ClientService
}

func NewClientHandler(s *service.ClientService) *ClientHandler {
	return &ClientHandler{service: s}
}

func (h *ClientHandler) CreateClient(c *gin.Context) {
	var req service.CreateClientRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	newApp, err := h.service.CreateClient(req)
	if err != nil {
		log.Printf("ERROR: Failed to create client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client"})
		return
	}
	c.JSON(http.StatusCreated, newApp)
}

func (h *ClientHandler) GetClient(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Client ID is required"})
		return
	}
	app, err := h.service.GetClient(id)
	if err != nil {
		log.Printf("ERROR: Failed to get client ID %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	c.JSON(http.StatusOK, app)
}

func (h *ClientHandler) ListClients(c *gin.Context) {
	apps, err := h.service.ListClients()
	if err != nil {
		log.Printf("ERROR: Failed to list clients: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list clients"})
		return
	}
	c.JSON(http.StatusOK, apps)
}

func (h *ClientHandler) DeleteClient(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Client ID is required"})
		return
	}
	if err := h.service.DeleteClient(id); err != nil {
		log.Printf("ERROR: Failed to delete client ID %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete client"})
		return
	}
	c.Status(http.StatusNoContent)
}
