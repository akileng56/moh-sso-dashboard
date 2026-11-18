// Package audit supplies host HTTP, SQLC and cache adapters for the published
// Audit read module. Event recording and migrations remain in the host.
package audit

import (
	"io"
	"time"

	auditmodule "github.com/mohuganda/moh-sso-dashboard/backend/modules/audit"
)

type ListAuditLogsInput = auditmodule.ListAuditLogsInput
type AuditWindowInput = auditmodule.AuditWindowInput
type TopFailureIPsInput = auditmodule.TopFailureIPsInput
type ExportAuditLogsInput = auditmodule.ExportAuditLogsInput
type AuditLogResponse = auditmodule.AuditLogResponse
type AuditActionsResponse = auditmodule.AuditActionsResponse
type AuditCursor = auditmodule.AuditCursor
type AuditLogListResponse = auditmodule.AuditLogListResponse
type AuditMetricsOverviewResponse = auditmodule.AuditMetricsOverviewResponse
type AuditFailedLoginsByDayPoint = auditmodule.AuditFailedLoginsByDayPoint
type AuditFailedLoginsByDayResponse = auditmodule.AuditFailedLoginsByDayResponse
type AuditTopFailureIPPoint = auditmodule.AuditTopFailureIPPoint
type AuditTopFailureIPsResponse = auditmodule.AuditTopFailureIPsResponse
type AuditExportManifest = auditmodule.AuditExportManifest
type AuditExportResponse = auditmodule.AuditExportResponse
type AuditMetricsSummary = auditmodule.AuditMetricsSummary
type AuditCountResponse = auditmodule.AuditCountResponse
type AuditWindowResponse = auditmodule.AuditWindowResponse
type Repository = auditmodule.Repository
type Service = auditmodule.Service
type Cache = auditmodule.Cache
type CachedReader = auditmodule.CachedReader

func NewService(repository Repository) Service { return auditmodule.NewService(repository) }
func NewCachedReader(service Service, cache Cache) *CachedReader {
	return auditmodule.NewCachedReader(service, cache)
}
func BuildExport(from, to time.Time, items []AuditLogResponse) (AuditExportResponse, error) {
	return auditmodule.BuildExport(from, to, items)
}
func WriteCSV(dst io.Writer, items []AuditLogResponse) error { return auditmodule.WriteCSV(dst, items) }
