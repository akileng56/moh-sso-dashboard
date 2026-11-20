package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	db "github.com/mohuganda/moh-sso-dashboard/backend/modules/audit/internal/postgres"
)

func toAuditLogResponse(row db.ListAuditLogsRow) AuditLogResponse {
	return auditLogResponse(row.ID.String(), row.CreatedAt.Valid, row.CreatedAt.Time, row.UserID.Valid, row.UserID.UUID.String(), row.Username, row.Action, row.Metadata)
}

func toAuditLogResponseFromGet(row db.GetAuditLogRow) AuditLogResponse {
	return auditLogResponse(row.ID.String(), row.CreatedAt.Valid, row.CreatedAt.Time, row.UserID.Valid, row.UserID.UUID.String(), row.Username, row.Action, row.Metadata)
}

func toAuditLogResponseFromExport(row db.ExportAuditLogsRow) AuditLogResponse {
	return auditLogResponse(row.ID.String(), row.CreatedAt.Valid, row.CreatedAt.Time, row.UserID.Valid, row.UserID.UUID.String(), row.Username, row.Action, row.Metadata)
}

func toAuditLogResponses(rows []db.ListAuditLogsRow) []AuditLogResponse {
	out := make([]AuditLogResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toAuditLogResponse(row))
	}
	return out
}

func toAuditMetricsOverviewResponse(row db.AuditMetricsOverviewRow) AuditMetricsOverviewResponse {
	return AuditMetricsOverviewResponse{
		TotalEvents:      row.TotalEvents,
		TotalFailures:    row.TotalFailures,
		FailedLogins:     row.FailedLogins,
		SuccessfulLogins: row.SuccessfulLogins,
	}
}

func toAuditFailedLoginsByDayResponse(rows []db.FailedLoginsByDayRow) AuditFailedLoginsByDayResponse {
	out := make([]AuditFailedLoginsByDayPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, AuditFailedLoginsByDayPoint{Day: row.Day.UnixMilli(), Count: row.Count})
	}
	return AuditFailedLoginsByDayResponse{Series: out}
}

func toAuditTopFailureIPsResponse(rows []db.TopFailureIPsRow) AuditTopFailureIPsResponse {
	out := make([]AuditTopFailureIPPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, AuditTopFailureIPPoint{
			IP:    toAuditString(row.Ip),
			Count: row.Count,
		})
	}
	return AuditTopFailureIPsResponse{Items: out}
}

func toExportAuditLogResponses(rows []db.ExportAuditLogsRow) []AuditLogResponse {
	out := make([]AuditLogResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toAuditLogResponseFromExport(row))
	}
	return out
}

func auditLogResponse(
	id string,
	createdAtValid bool,
	createdAt time.Time,
	userIDValid bool,
	userID string,
	username string,
	action string,
	rawMetadata db.NullJSON,
) AuditLogResponse {
	metadata := extractAuditMetadata(rawMetadata)
	if metadata == nil {
		metadata = map[string]any{}
	}

	response := AuditLogResponse{
		ID:       id,
		Username: username,
		Action:   action,
		Metadata: metadata,
		IP:       stringValue(metadata["ip"]),
		ClientID: stringValue(metadata["client_id"]),
		Success:  boolPtr(metadata["success"]),
	}
	if createdAtValid {
		response.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	}
	if userIDValid {
		response.UserID = userID
	}

	return response
}

func extractAuditMetadata(r db.NullJSON) map[string]any {
	if !r.Valid || len(r.RawMessage) == 0 {
		return nil
	}

	var meta map[string]any
	if err := json.Unmarshal(r.RawMessage, &meta); err != nil {
		return nil
	}
	return meta
}

func stringValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func boolPtr(value any) *bool {
	switch v := value.(type) {
	case bool:
		return &v
	case string:
		if v == "true" {
			out := true
			return &out
		}
		if v == "false" {
			out := false
			return &out
		}
	}
	return nil
}

func toAuditString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case uuid.UUID:
		return v.String()
	default:
		return ""
	}
}
