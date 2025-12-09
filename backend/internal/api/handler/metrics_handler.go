package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/service"
)

type MetricsHandler struct {
	service *service.MetricsService
}

func NewMetricsHandler(s *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{service: s}
}

// ===== Helper: Parse time range =====
func (h *MetricsHandler) parseRange(c *gin.Context) (time.Time, time.Time, bool) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing start or end parameter"})
		return time.Time{}, time.Time{}, false
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start format (must be RFC3339)"})
		return time.Time{}, time.Time{}, false
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end format (must be RFC3339)"})
		return time.Time{}, time.Time{}, false
	}

	return start, end, true
}

//
// ===== SYSTEM METRICS =====
//

func (h *MetricsHandler) CountUsers(c *gin.Context) {
	v, err := h.service.CountUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"total_users": v})
}

func (h *MetricsHandler) CountDisabledUsers(c *gin.Context) {
	v, err := h.service.CountDisabledUsers(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"disabled_users": v})
}

func (h *MetricsHandler) ActiveUsersToday(c *gin.Context) {
	v, err := h.service.ActiveUsersToday(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"active_users_today": v})
}

func (h *MetricsHandler) ActiveUsersThisWeek(c *gin.Context) {
	v, err := h.service.ActiveUsersThisWeek(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"active_users_this_week": v})
}

//
// ===== LOGIN TREND =====
//

func (h *MetricsHandler) LoginTrend(c *gin.Context) {
	rows, err := h.service.LoginTrend(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, rows)
}

func (h *MetricsHandler) LoginTrendByDay(c *gin.Context) {
	start, end, ok := h.parseRange(c)
	if !ok {
		return
	}

	rows, err := h.service.LoginTrendByDay(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, rows)
}

//
// ===== SECURITY METRICS =====
//

func (h *MetricsHandler) CountFailedLogins(c *gin.Context) {
	v, err := h.service.CountFailedLogins(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"failed_logins": v})
}

func (h *MetricsHandler) CountFailedLoginsInRange(c *gin.Context) {
	start, end, ok := h.parseRange(c)
	if !ok {
		return
	}

	v, err := h.service.CountFailedLoginsInRange(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"failed_logins": v})
}

func (h *MetricsHandler) SuspiciousLogins(c *gin.Context) {
	start, end, ok := h.parseRange(c)
	if !ok {
		return
	}

	homeCountry := c.Query("home")

	rows, err := h.service.SuspiciousLogins(c.Request.Context(), start, end, homeCountry)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, rows)
}

//
// ===== CLIENT METRICS =====
//

func (h *MetricsHandler) CountClients(c *gin.Context) {
	v, err := h.service.CountClients(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"total_clients": v})
}

func (h *MetricsHandler) MostAccessedClients(c *gin.Context) {
	start, end, ok := h.parseRange(c)
	if !ok {
		return
	}

	limit := int32(10)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = int32(parsed)
		}
	}

	rows, err := h.service.MostAccessedClients(c.Request.Context(), start, end, limit)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, rows)
}

func (h *MetricsHandler) ActiveUsersPerClientToday(c *gin.Context) {
	rows, err := h.service.ActiveUsersPerClientToday(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, rows)
}

func (h *MetricsHandler) LoginCountForClient(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		c.JSON(400, gin.H{"error": "client_id required"})
		return
	}

	start, end, ok := h.parseRange(c)
	if !ok {
		return
	}

	v, err := h.service.LoginCountForClient(c.Request.Context(), clientID, start, end)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"login_count": v})
}

//
// ===== USER METRICS =====
//

func (h *MetricsHandler) LastLoginForUser(c *gin.Context) {
	uid := c.Param("userID")

	userID, err := uuid.Parse(uid)

	if err != nil {

		c.JSON(500, gin.H{"error": err.Error()})

	}
	row, err := h.service.LastLoginForUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, row)
}

func (h *MetricsHandler) UserClientUsage(c *gin.Context) {
	uid := c.Param("userID")

	userID, err := uuid.Parse(uid)

	if err != nil {

		c.JSON(500, gin.H{"error": err.Error()})

	}

	start, end, ok := h.parseRange(c)
	if !ok {
		return
	}

	rows, err := h.service.ClientUsageForUser(c.Request.Context(), userID, start, end)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, rows)
}

func (h *MetricsHandler) NewUsersInRange(c *gin.Context) {
	start, end, ok := h.parseRange(c)
	if !ok {
		return
	}

	rows, err := h.service.NewUsersInRange(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, rows)
}

func (h *MetricsHandler) NewUsersTrend(c *gin.Context) {
	start, end, ok := h.parseRange(c)
	if !ok {
		return
	}

	rows, err := h.service.NewUsersTrend(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, rows)
}

func (h *MetricsHandler) NeverLoggedInUsers(c *gin.Context) {
	rows, err := h.service.NeverLoggedInUsers(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, rows)
}

func (h *MetricsHandler) Overview(c *gin.Context) {
	ctx := c.Request.Context()

	// ===== System Metrics =====
	totalUsers, _ := h.service.CountUsers(ctx)
	disabledUsers, _ := h.service.CountDisabledUsers(ctx)
	activeToday, _ := h.service.ActiveUsersToday(ctx)
	activeWeek, _ := h.service.ActiveUsersThisWeek(ctx)

	// ===== Client Metrics =====
	totalClients, _ := h.service.CountClients(ctx)
	enabledClients, _ := h.service.CountEnabledClients(ctx)
	activePerClientToday, _ := h.service.ActiveUsersPerClientToday(ctx)

	// ===== Security Metrics =====
	failedLogins, _ := h.service.CountFailedLogins(ctx)
	activeSessions, _ := h.service.ApproximateActiveSessions(ctx)

	// ===== Recent Trends (Last 30 Days) =====
	now := time.Now()
	start := now.AddDate(0, 0, -30)

	loginTrend, _ := h.service.LoginTrendByDay(ctx, start, now)
	newUsersTrend, _ := h.service.NewUsersTrend(ctx, start, now)

	// ===== Recently Created =====
	recentClients, _ := h.service.RecentlyCreatedClients(ctx, 10)
	recentUsers, _ := h.service.NewUsersInRange(ctx, start, now)

	// ===== Never Logged In =====
	neverLogged, _ := h.service.NeverLoggedInUsers(ctx)

	// Build Response
	response := gin.H{
		"system": gin.H{
			"total_users":            totalUsers,
			"disabled_users":         disabledUsers,
			"active_users_today":     activeToday,
			"active_users_this_week": activeWeek,
		},
		"clients": gin.H{
			"total_clients":   totalClients,
			"enabled_clients": enabledClients,
			"active_today":    activePerClientToday,
			"recent_clients":  recentClients,
		},
		"security": gin.H{
			"failed_logins":   failedLogins,
			"active_sessions": activeSessions,
		},
		"trends": gin.H{
			"login_trend_30_days":     loginTrend,
			"new_users_trend_30_days": newUsersTrend,
		},
		"users": gin.H{
			"recent_users":    recentUsers,
			"never_logged_in": neverLogged,
		},
		"_meta": gin.H{
			"range_start":  start,
			"range_end":    now,
			"generated_at": time.Now(),
		},
	}

	c.JSON(http.StatusOK, response)
}
