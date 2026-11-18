package audit

import (
	"context"

	"github.com/google/uuid"
)

// Repository is the read-only audit query boundary. Core audit event recording
// and schema migrations remain host responsibilities.
type Repository interface {
	ListAuditLogs(ctx context.Context, input ListAuditLogsInput) (AuditLogListResponse, error)
	GetAuditLog(ctx context.Context, id uuid.UUID) (AuditLogResponse, error)
	ListAuditActions(ctx context.Context) (AuditActionsResponse, error)
	AuditMetricsOverview(ctx context.Context, input AuditWindowInput) (AuditMetricsOverviewResponse, error)
	FailedLoginsByDay(ctx context.Context, input AuditWindowInput) (AuditFailedLoginsByDayResponse, error)
	TopFailureIPs(ctx context.Context, input TopFailureIPsInput) (AuditTopFailureIPsResponse, error)
	ExportAuditLogs(ctx context.Context, input ExportAuditLogsInput) ([]AuditLogResponse, error)
}
