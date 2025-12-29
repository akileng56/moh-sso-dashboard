package handler

import (
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/sqlc-dev/pqtype"
)

type AuditHandler struct {
	store db.Store
}

func NewAuditHandler(store db.Store) *AuditHandler {
	return &AuditHandler{store: store}
}

func mustParseTimeRFC3339(c *gin.Context, key string) (time.Time, bool) {
	v := c.Query(key)
	if v == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
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

func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from is required (RFC3339)"})
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to is required (RFC3339)"})
		return
	}

	action := c.Query("action")
	if action == "" {
		action = ""
	}

	userIDStr := c.Query("user_id")
	var userID *uuid.UUID
	if userIDStr != "" {
		uid, ok := parseUUIDParam(userIDStr)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id uuid"})
			return
		}
		userID = uid
	}

	clientID := c.Query("client_id")
	ip := c.Query("ip")
	success := c.Query("success") // "true"/"false"

	// cursor
	var cursorCreatedAt *time.Time
	var cursorID *uuid.UUID

	if v := c.Query("cursor_created_at"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor_created_at"})
			return
		}
		cursorCreatedAt = &t
	}
	if v := c.Query("cursor_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor_id"})
			return
		}
		cursorID = &id
	}

	limit := int32(50)
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be 1..200"})
			return
		}
		limit = int32(n)
	}

	// SQLC expects NULLable params; adapt based on your generated types if needed.
	rows, err := h.store.ListAuditLogs(c.Request.Context(), db.ListAuditLogsParams{
		StartTime: toNullTime(&from),
		EndTime:   toNullTime(&to),

		Action:   toNullString(action),
		UserID:   toNullUUID(userID),
		ClientID: toNullString(clientID),
		Ip:       toNullString(ip),
		Success:  toNullString(success),

		CursorCreatedAt: toNullTime(cursorCreatedAt),
		CursorID:        toNullUUID(cursorID),

		RowLimit: limit + 1,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list audit logs"})
		return
	}

	hasMore := len(rows) > int(limit)
	if hasMore {
		rows = rows[:limit]
	}

	var nextCursorCreatedAt *time.Time
	var nextCursorID *int64

	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]

		// created_at: sql.NullTime → *time.Time
		if last.CreatedAt.Valid {
			t := last.CreatedAt.Time
			nextCursorCreatedAt = &t
		}

		// id: BIGSERIAL → int64
		id := last.ID
		nextCursorID = &id
	}

	c.JSON(http.StatusOK, gin.H{
		"items": rows,
		"next_cursor": gin.H{
			"cursor_created_at": nextCursorCreatedAt,
			"cursor_id":         nextCursorID,
		},
		"has_more": hasMore,
	})
}

func (h *AuditHandler) GetAuditLog(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid audit log id",
		})
		return
	}

	row, err := h.store.GetAuditLog(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "audit log not found",
		})
		return
	}

	c.JSON(http.StatusOK, row)
}

func (h *AuditHandler) ListAuditActions(c *gin.Context) {
	rows, err := h.store.ListAuditActions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list actions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"actions": rows})
}

func (h *AuditHandler) AuditMetricsOverview(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from is required (RFC3339)"})
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to is required (RFC3339)"})
		return
	}

	row, err := h.store.AuditMetricsOverview(c.Request.Context(), db.AuditMetricsOverviewParams{
		StartTime: toNullTime(&from),
		EndTime:   toNullTime(&to),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute metrics"})
		return
	}

	c.JSON(http.StatusOK, row)
}

func (h *AuditHandler) FailedLoginsByDay(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from is required (RFC3339)"})
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to is required (RFC3339)"})
		return
	}

	rows, err := h.store.FailedLoginsByDay(c.Request.Context(), db.FailedLoginsByDayParams{
		StartTime: toNullTime(&from),
		EndTime:   toNullTime(&to),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute series"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"series": rows})
}

func (h *AuditHandler) TopFailureIPs(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from is required (RFC3339)"})
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to is required (RFC3339)"})
		return
	}

	limit := int32(10)
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 50 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be 1..50"})
			return
		}
		limit = int32(n)
	}

	rows, err := h.store.TopFailureIPs(c.Request.Context(), db.TopFailureIPsParams{
		StartTime: toNullTime(&from),
		EndTime:   toNullTime(&to),
		RowLimit:  limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute top ips"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": rows})
}

// ----------------------------------------------------
// EXPORT (CSV or JSON) + manifest checksum
// ----------------------------------------------------
func (h *AuditHandler) ExportAuditLogs(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from is required (RFC3339)"})
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to is required (RFC3339)"})
		return
	}

	format := c.DefaultQuery("format", "csv") // csv | json
	action := c.Query("action")
	clientID := c.Query("client_id")
	ip := c.Query("ip")
	success := c.Query("success")

	var userID *uuid.UUID
	if v := c.Query("user_id"); v != "" {
		id, ok := parseUUIDParam(v)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id uuid"})
			return
		}
		userID = id
	}

	rows, err := h.store.ExportAuditLogs(c.Request.Context(), db.ExportAuditLogsParams{
		StartTime: toNullTime(&from),
		EndTime:   toNullTime(&to),
		Action:    toNullString(action),
		UserID:    toNullUUID(userID),
		ClientID:  toNullString(clientID),
		Ip:        toNullString(ip),
		Success:   toNullString(success),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "export failed"})
		return
	}

	// Build a compliance manifest (checksum + filter info)
	manifest := map[string]any{
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"range": map[string]string{
			"from": from.UTC().Format(time.RFC3339),
			"to":   to.UTC().Format(time.RFC3339),
		},
		"filters": map[string]any{
			"action":    action,
			"user_id":   c.Query("user_id"),
			"client_id": clientID,
			"ip":        ip,
			"success":   success,
		},
		"record_count": len(rows),
		"hash_algo":    "sha256",
	}

	// Serialize for checksum
	payloadBytes, _ := json.Marshal(rows)
	sum := sha256.Sum256(payloadBytes)
	manifest["payload_sha256"] = hex.EncodeToString(sum[:])

	if format == "json" {
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=audit_logs_export.json")
		c.JSON(http.StatusOK, gin.H{
			"manifest": manifest,
			"items":    rows,
		})
		return
	}

	// CSV
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=audit_logs_export.csv")

	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	// header
	_ = w.Write([]string{
		"id", "created_at", "user_id", "username", "action",
		"ip", "client_id", "success", "metadata_json",
	})

	for _, r := range rows {
		meta := extractAuditMetadata(r.Metadata)

		metaJSON, _ := json.Marshal(meta)

		ipv := ""
		clientv := ""
		successv := ""

		if meta != nil {
			if v, ok := meta["ip"].(string); ok {
				ipv = v
			}
			if v, ok := meta["client_id"].(string); ok {
				clientv = v
			}
			if v, ok := meta["success"]; ok {
				successv = stringify(v)
			}
		}

		userIDVal := ""
		if r.UserID.Valid {
			userIDVal = r.UserID.UUID.String()
		}

		idStr := strconv.FormatInt(r.ID, 10)

		createdAtStr := ""
		if r.CreatedAt.Valid {
			createdAtStr = r.CreatedAt.Time.UTC().Format(time.RFC3339)
		}

		_ = w.Write([]string{
			idStr,
			createdAtStr,
			userIDVal,
			r.Username,
			r.Action,
			ipv,
			clientv,
			successv,
			string(metaJSON),
		})

	}

	// Include manifest in headers for compliance
	manifestBytes, _ := json.Marshal(manifest)
	c.Header("X-Audit-Export-Manifest", string(manifestBytes))
}

// ---- helpers for nullable parameters ----
// Adapt these to your SQLC null types if needed.

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func toNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func stringify(v any) string {
	switch x := v.(type) {
	case bool:
		if x {
			return "true"
		}
		return "false"
	case string:
		return x
	default:
		b, _ := json.Marshal(x)
		return string(b)
	}
}

func extractAuditMetadata(r pqtype.NullRawMessage) map[string]any {
	if !r.Valid || len(r.RawMessage) == 0 {
		return nil
	}

	var meta map[string]any
	if err := json.Unmarshal(r.RawMessage, &meta); err != nil {
		return nil
	}
	return meta
}
