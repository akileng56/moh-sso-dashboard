package handler

import (
	"log"
	"net/http" // Still needed for HTTP status codes (e.g., http.StatusCreated)

	"github.com/gin-gonic/gin" // Import the Gin framework
	"github.com/moh-sso-dashboard/internal/service"
)

// ClientHandler holds the service it will use to perform business logic.
type ClientHandler struct {
	// We store a pointer to the service, not the struct itself.
	service *service.ClientService
}

// NewClientHandler creates a new handler with its service dependency.
func NewClientHandler(s *service.ClientService) *ClientHandler {
	return &ClientHandler{service: s}
}

// CreateClient is the Gin handler for POST /clients.
// It uses c.ShouldBindJSON for automatic request decoding.
func (h *ClientHandler) CreateClient(c *gin.Context) {
	// 1. Decode the JSON request body into the service's DTO
	var req service.CreateClientRequest
	
	// Gin's ShouldBindJSON handles decoding and basic validation (like required fields)
	if err := c.ShouldBindJSON(&req); err != nil {
		// Use c.JSON to write the error response
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 2. Call the service to create the client
	newApp, err := h.service.CreateClient(req)
	if err != nil {
		// In a real application, you'd inspect the error type for a 400 vs 500.
		log.Printf("ERROR: Failed to create client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client"})
		return
	}

	// 3. Write the successful response. Gin automatically marshals the struct to JSON.
	c.JSON(http.StatusCreated, newApp)
}

// GetClient is the Gin handler for GET /clients/:id.
// It uses c.Param to easily extract path parameters.
func (h *ClientHandler) GetClient(c *gin.Context) {
	// 1. Get the "id" from the URL path.
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Client ID is required"})
		return
	}

	// 2. Call the service
	app, err := h.service.GetClient(id)
	if err != nil {
		// Assume "not found" if the service returns an error
		log.Printf("ERROR: Failed to get client ID %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	// 3. Write the successful response
	c.JSON(http.StatusOK, app)
}

// ListClients is the Gin handler for GET /clients.
func (h *ClientHandler) ListClients(c *gin.Context) {
	// 1. Call the service (no input required)
	apps, err := h.service.ListClients()
	if err != nil {
		log.Printf("ERROR: Failed to list clients: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list clients"})
		return
	}

	// 2. Write the successful response
	c.JSON(http.StatusOK, apps)
}

// DeleteClient is the Gin handler for DELETE /clients/:id.
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	// 1. Get the "id" from the URL path
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Client ID is required"})
		return
	}

	// 2. Call the service
	if err := h.service.DeleteClient(id); err != nil {
		log.Printf("ERROR: Failed to delete client ID %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete client"})
		return
	}

	// 3. Write the successful (empty) response
	c.Status(http.StatusNoContent) // Gin's way to return 204 No Content
}
