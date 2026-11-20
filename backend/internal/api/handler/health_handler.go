package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/version"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	ObservatoryReady func() bool
	DBCheck          func(ctx context.Context) error
	RemoteDBCheck    func(ctx context.Context) error
	KeycloakCheck    func(ctx context.Context) error
	Redis            *redis.Client
	startedAt        time.Time
}

func NewHealthHandler(
	dbCheck func(ctx context.Context) error,
	remoteDbCheck func(ctx context.Context) error,
	kcCheck func(ctx context.Context) error,
	redis *redis.Client,
) *HealthHandler {
	return &HealthHandler{
		DBCheck:       dbCheck,
		RemoteDBCheck: remoteDbCheck,
		KeycloakCheck: kcCheck,
		startedAt:     time.Now(),
		Redis:         redis,
	}
}

// --------------------------------------------------
// Liveness Probe
// Used by Kubernetes to check if process is alive
// --------------------------------------------------
func (h *HealthHandler) HandleLive(c *gin.Context) {
	build := version.Get()
	response.OK(c, http.StatusOK, HealthResponse{
		Status:        "alive",
		UptimeSeconds: int(time.Since(h.startedAt).Seconds()),
		Version:       build.Version,
		Build:         &build,
	})
}

func (h *HealthHandler) HandleVersion(c *gin.Context) {
	response.OK(c, http.StatusOK, version.Get())
}

// --------------------------------------------------
// Readiness Probe (STRICT)
// Fails if critical dependencies are down
// --------------------------------------------------
func (h *HealthHandler) HandleReady(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.CheckCore(ctx); err != nil {
		response.Fail(c, http.StatusServiceUnavailable, "UNAVAILABLE", err.Error())
		return
	}

	build := version.Get()
	response.OK(c, http.StatusOK, HealthResponse{
		Status:        "ready",
		UptimeSeconds: int(time.Since(h.startedAt).Seconds()),
		Version:       build.Version,
		Build:         &build,
	})
}

// CheckCore excludes optional provider/worker health from mandatory readiness.
func (h *HealthHandler) CheckCore(ctx context.Context) error {
	if h.DBCheck == nil {
		return errors.New("database check unavailable")
	}
	if err := h.DBCheck(ctx); err != nil {
		return errors.New("database unavailable")
	}
	if h.Redis != nil && h.Redis.Ping(ctx).Err() != nil {
		return errors.New("redis unavailable")
	}
	if h.KeycloakCheck == nil {
		return errors.New("keycloak check unavailable")
	}
	if err := h.KeycloakCheck(ctx); err != nil {
		return errors.New("keycloak unavailable")
	}
	return nil
}

// --------------------------------------------------
// Aggregated Health (DEGRADED allowed)
// Does NOT fail pod unless DB is down
// --------------------------------------------------
func (h *HealthHandler) HandleHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	status := "ok"

	db := "ok"
	if err := h.DBCheck(ctx); err != nil {
		db = "unavailable"
		status = "degraded"
	}

	remoteDb := "disabled"
	if h.RemoteDBCheck != nil {
		remoteDb = "ok"
		if err := h.RemoteDBCheck(ctx); err != nil {
			remoteDb = "unavailable"
			status = "degraded"
		}
	}

	keycloak := "ok"
	if err := h.KeycloakCheck(ctx); err != nil {
		keycloak = "unavailable"
		status = "degraded"
	}

	redis := "ok"
	if h.Redis != nil && h.Redis.Ping(ctx).Err() != nil {
		redis = "unavailable"
		status = "degraded"
	}
	observatory := ""
	if h.ObservatoryReady != nil {
		observatory = "ok"
		if !h.ObservatoryReady() {
			observatory = "unavailable"
			status = "degraded"
		}
	}

	httpStatus := http.StatusOK
	if db == "unavailable" {
		httpStatus = http.StatusServiceUnavailable
	}

	build := version.Get()
	response.OK(c, httpStatus, HealthResponse{
		Status:        status,
		UptimeSeconds: int(time.Since(h.startedAt).Seconds()),
		Components: &HealthComponentStatus{
			Observatory: observatory,
			Database:    db,
			RemoteDB:    remoteDb,
			Keycloak:    keycloak,
			Redis:       redis,
		},
		Version: build.Version,
		Build:   &build,
	})
}
