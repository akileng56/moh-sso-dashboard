package audit

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	db "github.com/mohuganda/moh-sso-dashboard/backend/modules/audit/internal/postgres"
)

func TestAuditLogResponseNormalizesSQLCInternals(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	createdAt := time.Date(2026, 6, 1, 12, 30, 0, 0, time.UTC)

	response := toAuditLogResponse(db.ListAuditLogsRow{
		ID:        id,
		CreatedAt: sql.NullTime{Time: createdAt, Valid: true},
		UserID:    uuid.NullUUID{UUID: userID, Valid: true},
		Username:  "testuser",
		Action:    "login",
		Metadata: db.NullJSON{
			RawMessage: []byte(`{"ip":"127.0.0.1","client_id":"dashboard-web","success":true}`),
			Valid:      true,
		},
	})

	if response.ID != id.String() {
		t.Fatalf("expected id %s, got %s", id, response.ID)
	}
	if response.CreatedAt != createdAt.Format(time.RFC3339) {
		t.Fatalf("expected createdAt %s, got %s", createdAt.Format(time.RFC3339), response.CreatedAt)
	}
	if response.UserID != userID.String() {
		t.Fatalf("expected userId %s, got %s", userID, response.UserID)
	}
	if response.IP != "127.0.0.1" || response.ClientID != "dashboard-web" {
		t.Fatalf("expected extracted metadata fields, got ip=%s client=%s", response.IP, response.ClientID)
	}
	if response.Success == nil || !*response.Success {
		t.Fatalf("expected success true, got %v", response.Success)
	}
}
