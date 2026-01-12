package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/service"
)

type ImportHandler struct {
	importSevice *service.ImportService
	config       *config.Config
}

func NewImportHandler(importSevice *service.ImportService, config *config.Config) *ImportHandler {
	return &ImportHandler{
		importSevice: importSevice,
		config:       config,
	}
}

func (h *ImportHandler) Preview(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to open file"})
		return
	}
	defer f.Close()

	createdBy := c.GetString("user_email") // adapt to your auth middleware
	if createdBy == "" {
		createdBy = "unknown"
	}

	resp, err := h.importSevice.PreviewCSV(c.Request.Context(), f, file.Filename, createdBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ImportHandler) Execute(c *gin.Context) {
	var body struct {
		JobID string `json:"jobId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.JobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "jobId is required"})
		return
	}

	id, err := uuid.Parse(body.JobID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid jobId"})
		return
	}

	resp, err := h.importSevice.Execute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ImportHandler) GetJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("jobId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid jobId"})
		return
	}

	resp, err := h.importSevice.GetJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ImportHandler) DownloadTemplateCSV(c *gin.Context) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="users_import_template.csv"`)
	c.String(http.StatusOK, "username,email,first_name,last_name,role,enabled,client_ids\n")
}

func (h *ImportHandler) DownloadErrorsCSV(c *gin.Context) {
	id, err := uuid.Parse(c.Param("jobId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid jobId"})
		return
	}

	job, err := h.importSevice.GetJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	csv := h.importSevice.BuildErrorCSV(job.Rows)

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="users_import_errors.csv"`)
	c.String(http.StatusOK, csv)
}
