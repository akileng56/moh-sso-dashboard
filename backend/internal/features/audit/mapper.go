package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

func toAuditExportResponse(manifest AuditExportManifest, items []AuditLogResponse) AuditExportResponse {
	return AuditExportResponse{Manifest: manifest, Items: items}
}

func toAuditLogListResponse(items []AuditLogResponse, nextCursor AuditCursor, hasMore bool) AuditLogListResponse {
	return AuditLogListResponse{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}

func toAuditExportManifest(from, to time.Time, rows any, recordCount int) AuditExportManifest {
	payload, _ := json.Marshal(rows)
	sum := sha256.Sum256(payload)
	return AuditExportManifest{
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Range:         map[string]string{"from": from.UTC().Format(time.RFC3339), "to": to.UTC().Format(time.RFC3339)},
		RecordCount:   recordCount,
		HashAlgo:      "sha256",
		PayloadSHA256: hex.EncodeToString(sum[:]),
	}
}
