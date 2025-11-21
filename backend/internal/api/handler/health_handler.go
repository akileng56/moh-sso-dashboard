package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	DBCheck       func() error
	KeycloakCheck func() error
}

func NewHealthHandler(dbCheck func() error, kcCheck func() error) *HealthHandler {
	return &HealthHandler{
		DBCheck:       dbCheck,
		KeycloakCheck: kcCheck,
	}
}

func (h *HealthHandler) HandleHealth(c *gin.Context) {
	status := "ok"
	dbStatus := "ok"
	kcStatus := "ok"

	// DB health
	if err := h.DBCheck(); err != nil {
		dbStatus = err.Error()
		status = "degraded"
	}

	// Keycloak health
	if err := h.KeycloakCheck(); err != nil {
		kcStatus = err.Error()
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         status,
		"database":       dbStatus,
		"keycloak":       kcStatus,
		"uptime_seconds": 0,
	})
}
