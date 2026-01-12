package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/service"
)

type ImportHandler struct {
	importService *service.ImportService
	config        *config.Config
}

func NewImportHandler(
	importService *service.ImportService,
	config *config.Config,
) *ImportHandler {
	return &ImportHandler{
		importService: importService,
		config:        config,
	}
}

/* =========================================================
 * Preview CSV
 * ========================================================= */

func (h *ImportHandler) Preview(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"CSV file is required",
		)
		return
	}

	f, err := file.Open()
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_FILE",
			"Unable to open uploaded file",
		)
		return
	}
	defer f.Close()

	createdBy := c.GetString("user_email")
	if createdBy == "" {
		createdBy = "unknown"
	}

	resp, err := h.importService.PreviewCSV(
		c.Request.Context(),
		f,
		file.Filename,
		createdBy,
	)
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			err.Error(),
		)
		return
	}

	response.OK(c, http.StatusOK, resp)
}

/* =========================================================
 * Execute Import Job
 * ========================================================= */

func (h *ImportHandler) Execute(c *gin.Context) {
	var body struct {
		JobID string `json:"jobId"`
	}

	if err := c.ShouldBindJSON(&body); err != nil || body.JobID == "" {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"jobId is required",
		)
		return
	}

	jobID, err := uuid.Parse(body.JobID)
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_UUID",
			"Invalid jobId format",
		)
		return
	}

	resp, err := h.importService.Execute(
		c.Request.Context(),
		jobID,
	)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to execute import job",
		)
		return
	}

	response.OK(c, http.StatusOK, resp)
}

/* =========================================================
 * Get Import Job
 * ========================================================= */

func (h *ImportHandler) GetJob(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("jobId"))
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_UUID",
			"Invalid jobId format",
		)
		return
	}

	resp, err := h.importService.GetJob(
		c.Request.Context(),
		jobID,
	)
	if err != nil {
		response.Fail(
			c,
			http.StatusNotFound,
			"JOB_NOT_FOUND",
			"Import job not found",
		)
		return
	}

	response.OK(c, http.StatusOK, resp)
}

/* =========================================================
 * Download CSV Template (streaming – no envelope)
 * ========================================================= */

func (h *ImportHandler) DownloadTemplateCSV(c *gin.Context) {
	c.Header("Content-Type", "text/csv")
	c.Header(
		"Content-Disposition",
		`attachment; filename="users_import_template.csv"`,
	)

	c.String(
		http.StatusOK,
		"username,email,first_name,last_name,role,enabled,client_ids\n",
	)
}

/* =========================================================
 * Download Error CSV (streaming – no envelope)
 * ========================================================= */

func (h *ImportHandler) DownloadErrorsCSV(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("jobId"))
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_UUID",
			"Invalid jobId format",
		)
		return
	}

	job, err := h.importService.GetJob(
		c.Request.Context(),
		jobID,
	)
	if err != nil {
		response.Fail(
			c,
			http.StatusNotFound,
			"JOB_NOT_FOUND",
			"Import job not found",
		)
		return
	}

	csv := h.importService.BuildErrorCSV(job.Rows)

	c.Header("Content-Type", "text/csv")
	c.Header(
		"Content-Disposition",
		`attachment; filename="users_import_errors.csv"`,
	)

	c.String(http.StatusOK, csv)
}
