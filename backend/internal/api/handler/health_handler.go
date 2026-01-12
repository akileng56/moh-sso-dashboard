package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/http/response"
)

type HealthHandler struct {
	DBCheck       func() error
	KeycloakCheck func() error
	startedAt     time.Time
}

func NewHealthHandler(
	dbCheck func() error,
	kcCheck func() error,
) *HealthHandler {
	return &HealthHandler{
		DBCheck:       dbCheck,
		KeycloakCheck: kcCheck,
		startedAt:     time.Now(),
	}
}

func (h *HealthHandler) HandleHealth(c *gin.Context) {
	status := "ok"

	db := "ok"
	if err := h.DBCheck(); err != nil {
		db = "unavailable"
		status = "degraded"
	}

	keycloak := "ok"
	if err := h.KeycloakCheck(); err != nil {
		keycloak = "unavailable"
		status = "degraded"
	}

	response.OK(c, http.StatusOK, gin.H{
		"status": status,
		"components": gin.H{
			"database": db,
			"keycloak": keycloak,
		},
		"uptime_seconds": int(time.Since(h.startedAt).Seconds()),
	})
}
