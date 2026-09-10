package data_quality

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/http/response"
)

// ── tables ───────────────────────────────────────────────────────────────

func (h *Handler) ListDQATables(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	tables, err := h.dqaStore.ListTables(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DQA_TABLES_LIST_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, tables)
}

func (h *Handler) UpsertDQATable(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	var req dqaTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if err := h.dqaStore.UpsertTable(c.Request.Context(), req.toMapping()); err != nil {
		response.Fail(c, http.StatusInternalServerError, "DQA_TABLE_SAVE_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, req.toMapping())
}

func (h *Handler) DeleteDQATable(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	tableID := c.Param("tableId")
	removed, err := h.dqaStore.DeleteTable(c.Request.Context(), tableID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DQA_TABLE_DELETE_FAILED", err.Error())
		return
	}
	if !removed {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "table mapping not found")
		return
	}
	c.Status(http.StatusNoContent)
}

// ── rules ────────────────────────────────────────────────────────────────

func (h *Handler) ListDQARules(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	tableID := c.Query("table_id")
	records, err := h.dqaStore.ListRules(c.Request.Context(), tableID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DQA_RULES_LIST_FAILED", err.Error())
		return
	}
	out := make([]dqaRuleResponse, 0, len(records))
	for _, r := range records {
		out = append(out, ruleRecordToResponse(r))
	}
	response.OK(c, http.StatusOK, out)
}

func (h *Handler) UpsertDQARule(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	var req dqaRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	rule := req.toRule()

	// Compile before saving: a rule that can't be compiled is rejected
	// outright rather than silently stored broken.
	compiled, err := h.dqaStore.CompilePreview(c.Request.Context(), rule)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "RULE_COMPILE_FAILED", err.Error())
		return
	}

	rec, err := h.dqaStore.UpsertRule(c.Request.Context(), rule, compiled.MeasureSQL, getUserEmailOrUsername(c))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DQA_RULE_SAVE_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, ruleRecordToResponse(rec))
}

func (h *Handler) DeleteDQARule(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	tableID, code := c.Param("tableId"), c.Param("code")
	removed, err := h.dqaStore.DeleteRule(c.Request.Context(), tableID, code)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DQA_RULE_DELETE_FAILED", err.Error())
		return
	}
	if !removed {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "rule not found")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) CompileDQARule(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	var req dqaRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	compiled, err := h.dqaStore.CompilePreview(c.Request.Context(), req.toRule())
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "RULE_COMPILE_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, dqaCompilePreviewResponse{
		Kind: compiled.Kind, MeasureSQL: compiled.MeasureSQL, SampleSQL: compiled.SampleSQL,
	})
}

// ── seed ─────────────────────────────────────────────────────────────────

func (h *Handler) SeedDQABuiltinRules(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	var req dqaSeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if req.TableID == "" {
		req.TableID = "vht_monthly"
	}
	if req.PhysicalTable == "" {
		response.Fail(c, http.StatusBadRequest, "INVALID_BODY", "physical_table is required")
		return
	}

	ctx := c.Request.Context()
	if err := h.dqaStore.UpsertTable(ctx, TableMapping{
		TableID: req.TableID, PhysicalTable: req.PhysicalTable,
		Description: "eCHIS VHT monthly report extract", IsActive: true,
	}); err != nil {
		response.Fail(c, http.StatusInternalServerError, "DQA_TABLE_SAVE_FAILED", err.Error())
		return
	}

	rules := BuiltinVHTMonthlyRules(req.TableID, req.PhysicalTable)
	saved := make([]dqaRuleResponse, 0, len(rules))
	var errs []RuleCompileErr
	for _, rule := range rules {
		compiled, err := h.dqaStore.CompilePreview(ctx, rule)
		if err != nil {
			errs = append(errs, RuleCompileErr{Code: rule.Code, Error: err.Error()})
			continue
		}
		rec, err := h.dqaStore.UpsertRule(ctx, rule, compiled.MeasureSQL, getUserEmailOrUsername(c))
		if err != nil {
			errs = append(errs, RuleCompileErr{Code: rule.Code, Error: err.Error()})
			continue
		}
		saved = append(saved, ruleRecordToResponse(rec))
	}

	response.OK(c, http.StatusOK, gin.H{
		"table_id": req.TableID, "seeded": len(saved), "failed": len(errs),
		"rules": saved, "errors": errs,
	})
}

type dqaSeedRequest struct {
	TableID       string `json:"table_id"`
	PhysicalTable string `json:"physical_table" binding:"required"`
}

// ── runs + flags ─────────────────────────────────────────────────────────

func (h *Handler) RunDQATable(c *gin.Context) {
	if h.dqaStore == nil || h.dwhDB == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	var req dqaRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	result, err := h.dqaStore.RunTable(c.Request.Context(), h.dwhDB, req.TableID, getUserEmailOrUsername(c))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DQA_RUN_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusCreated, result)
}

func (h *Handler) ListDQARuns(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	tableID := c.Query("table_id")
	limit := parseListLimit(c, 25, 200)
	offset := parseListOffset(c)
	runs, err := h.dqaStore.ListRuns(c.Request.Context(), tableID, limit, offset)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DQA_RUNS_LIST_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, runs)
}

func (h *Handler) ListDQAFlags(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	runID, err := strconv.ParseInt(c.Param("runId"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_RUN_ID", "runId must be an integer")
		return
	}
	filter := FlagFilter{
		Severity:  c.Query("severity"),
		District:  c.Query("district"),
		Facility:  c.Query("facility"),
		VHT:       c.Query("vht"),
		Region:    c.Query("region"),
		Subcounty: c.Query("subcounty"),
		Year:      parseOptionalInt(c.Query("year")),
		Month:     parseOptionalInt(c.Query("month")),
	}
	limit := parseListLimit(c, 100, 1000)
	offset := parseListOffset(c)
	flags, err := h.dqaStore.ListFlags(c.Request.Context(), runID, filter, limit, offset)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DQA_FLAGS_LIST_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, flags)
}

func parseOptionalInt(raw string) *int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &v
}

// ListDQAFlagFilters returns the distinct filter values present in a run, so the
// dashboard's Year/Month/District/Facility/VHT dropdowns reflect real data.
func (h *Handler) ListDQAFlagFilters(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	runID, err := strconv.ParseInt(c.Param("runId"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_RUN_ID", "runId must be an integer")
		return
	}
	opts, err := h.dqaStore.FlagFilterOptions(c.Request.Context(), runID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DQA_FLAG_FILTERS_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, opts)
}
