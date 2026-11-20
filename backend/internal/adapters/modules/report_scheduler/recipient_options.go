package report_scheduler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/http/response"
)

func (h *Handler) RecipientOptions(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	kind := c.DefaultQuery("kind", "email")
	if kind != "email" && kind != "group" && kind != "portal" {
		response.Fail(c, 400, "INVALID_KIND", "kind must be email, group, or portal")
		return
	}
	limit := 100
	if raw := c.Query("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 100 {
			response.Fail(c, 400, "INVALID_LIMIT", "limit must be between 1 and 100")
			return
		}
		limit = value
	}
	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 200 {
		response.Fail(c, 400, "INVALID_SEARCH", "search is too long")
		return
	}
	if h.service == nil || !h.service.Module().SchedulingEnabled {
		response.Fail(c, 503, "DIRECTORY_UNAVAILABLE", "Recipient directory is unavailable")
		return
	}
	var selected []string
	if raw := c.Query("selected"); raw != "" {
		if len(raw) > 32768 || json.Unmarshal([]byte(raw), &selected) != nil || len(selected) > 100 {
			response.Fail(c, 400, "INVALID_SELECTION", "selected must contain at most 100 recipient values")
			return
		}
	}
	options, err := h.service.ListRecipientOptions(c.Request.Context(), kind, search, limit, selected)
	if err != nil {
		response.Fail(c, 503, "DIRECTORY_UNAVAILABLE", "Unable to load recipient options")
		return
	}
	response.OK(c, http.StatusOK, options)
}
