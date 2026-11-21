package audit

import (
	"time"

	"github.com/google/uuid"
)

func toAuditActionsResponse(rows []string) AuditActionsResponse {
	return AuditActionsResponse{Actions: rows}
}

func toAuditCursorResponse(nextCreatedAt *time.Time, nextID *uuid.UUID) AuditCursor {
	return AuditCursor{
		CursorCreatedAt: nextCreatedAt,
		CursorID:        nextID,
	}
}

func toAuditLogListResponse(items []AuditLogResponse, nextCursor AuditCursor, hasMore bool) AuditLogListResponse {
	return AuditLogListResponse{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}
