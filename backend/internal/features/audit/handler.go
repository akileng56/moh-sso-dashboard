package audit

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/http/response"
)

/* =========================================================
 * Handler
 * ========================================================= */

type Handler struct {
	service Service
	reader  *CachedReader
}

func NewHandler(service Service, cache Cache) *Handler {
	return &Handler{
		service: service,
		reader:  NewCachedReader(service, cache),
	}
}

/* =========================================================
 * Helpers
 * ========================================================= */

func mustParseTimeRFC3339(c *gin.Context, key string) (time.Time, bool) {
	v := c.Query(key)
	if v == "" {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_FAILED", key+" is required (RFC3339)")
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_DATE_FORMAT", "Invalid "+key+" format (RFC3339)")
		return time.Time{}, false
	}
	return t, true
}

func parseUUIDParam(v string) (*uuid.UUID, bool) {
	if v == "" {
		return nil, true
	}
	id, err := uuid.Parse(v)
	if err != nil {
		return nil, false
	}
	return &id, true
}

/* =========================================================
 * List Audit Logs (cursor pagination)
 * ========================================================= */

func (h *Handler) ListAuditLogs(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		return
	}

	// filters
	action := c.Query("action")
	clientID := c.Query("client_id")
	ip := c.Query("ip")
	success := c.Query("success")

	var userID *uuid.UUID
	if v := c.Query("user_id"); v != "" {
		id, ok := parseUUIDParam(v)
		if !ok {
			response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid user_id format")
			return
		}
		userID = id
	}

	// cursor
	var cursorCreatedAt *time.Time
	var cursorID *uuid.UUID

	if v := c.Query("cursor_created_at"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "INVALID_DATE_FORMAT", "Invalid cursor_created_at")
			return
		}
		cursorCreatedAt = &t
	}

	if v := c.Query("cursor_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid cursor_id")
			return
		}
		cursorID = &id
	}

	limit := int32(50)
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 200 {
			response.Fail(c, http.StatusBadRequest, "VALIDATION_FAILED", "limit must be between 1 and 200")
			return
		}
		limit = int32(n)
	}

	resp, err := h.reader.ListAuditLogs(c.Request.Context(), ListAuditLogsInput{
		StartTime:       from,
		EndTime:         to,
		Action:          action,
		UserID:          userID,
		ClientID:        clientID,
		IP:              ip,
		Success:         success,
		CursorCreatedAt: cursorCreatedAt,
		CursorID:        cursorID,
		Limit:           limit,
	}, c.Request.URL.RawQuery)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list audit logs")
		return
	}

	response.OK(c, http.StatusOK, resp)
}

/* =========================================================
 * Get Single Audit Log
 * ========================================================= */

func (h *Handler) GetAuditLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid audit log ID")
		return
	}

	result, err := h.service.GetAuditLog(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "AUDIT_LOG_NOT_FOUND", "Audit log not found")
		return
	}

	response.OK(c, http.StatusOK, result)
}

/* =========================================================
 * Audit Metadata / Metrics
 * ========================================================= */

func (h *Handler) ListAuditActions(c *gin.Context) {
	result, err := h.service.ListAuditActions(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list audit actions")
		return
	}

	response.OK(c, http.StatusOK, result)
}

func (h *Handler) AuditMetricsOverview(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		return
	}

	result, err := h.service.AuditMetricsOverview(c.Request.Context(), AuditWindowInput{StartTime: from, EndTime: to})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to compute audit metrics")
		return
	}

	response.OK(c, http.StatusOK, result)
}

func (h *Handler) FailedLoginsByDay(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		return
	}

	result, err := h.service.FailedLoginsByDay(c.Request.Context(), AuditWindowInput{StartTime: from, EndTime: to})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to compute failed logins series")
		return
	}

	response.OK(c, http.StatusOK, result)
}

func (h *Handler) TopFailureIPs(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		return
	}

	limit := int32(10)
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 50 {
			response.Fail(c, http.StatusBadRequest, "VALIDATION_FAILED", "limit must be between 1 and 50")
			return
		}
		limit = int32(n)
	}

	result, err := h.service.TopFailureIPs(c.Request.Context(), TopFailureIPsInput{
		StartTime: from,
		EndTime:   to,
		Limit:     limit,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to compute top IPs")
		return
	}

	response.OK(c, http.StatusOK, result)
}

/* =========================================================
 * Export (CSV / JSON) – streaming, no envelope
 * ========================================================= */

func (h *Handler) ExportAuditLogs(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		return
	}

	format := c.DefaultQuery("format", "csv")
	action := c.Query("action")
	clientID := c.Query("client_id")
	ip := c.Query("ip")
	success := c.Query("success")

	var userID *uuid.UUID
	if v := c.Query("user_id"); v != "" {
		id, ok := parseUUIDParam(v)
		if !ok {
			response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid user_id format")
			return
		}
		userID = id
	}

	items, err := h.service.ExportAuditLogs(c.Request.Context(), ExportAuditLogsInput{
		StartTime: from,
		EndTime:   to,
		Action:    action,
		UserID:    userID,
		ClientID:  clientID,
		IP:        ip,
		Success:   success,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Export failed")
		return
	}

	export, err := BuildExport(from, to, items)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Export failed")
		return
	}

	if format == "json" {
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=audit_logs_export.json")
		response.OK(c, http.StatusOK, export)
		return
	}

	// Headers must precede streaming, including exports larger than CSV's buffer.
	manifestBytes, _ := json.Marshal(export.Manifest)
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=audit_logs_export.csv")
	c.Header("X-Audit-Export-Manifest", string(manifestBytes))
	if err := WriteCSV(c.Writer, export.Items); err != nil {
		_ = c.Error(err)
		c.Abort()
	}
}
