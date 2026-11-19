package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	audit "github.com/mohuganda/moh-sso-dashboard/backend/modules/audit"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

type fake struct{ audit.Repository }

func (fake) ListAuditActions(context.Context) (audit.AuditActionsResponse, error) {
	return audit.AuditActionsResponse{Actions: []string{"login"}}, nil
}
func check(ok bool, what string) {
	if !ok {
		panic(what)
	}
}
func main() {
	var repo audit.Repository = fake{}
	var db *sql.DB
	if dsn := os.Getenv("AUDIT_TEST_DSN"); dsn != "" {
		var err error
		db, err = sql.Open("postgres", dsn)
		check(err == nil, "open PostgreSQL")
		defer db.Close()
		repo, err = audit.NewRepository(db)
		check(err == nil, "construct read repository")
		verifyRepository(repo)
	}
	m, err := audit.Register(http.NewServeMux(), audit.Dependencies{
		Repository: repo, RequireAdmin: func(*http.Request) error { return nil },
		AuthorizeAny: func(_ *http.Request, permissions ...string) error {
			check(len(permissions) == 1 && permissions[0] == "audit:read", "permissions")
			return nil
		},
		LimitExport: func(_ *http.Request, policy string, limit int, window time.Duration) (http.Header, error) {
			check(policy == "audit-export" && limit == 10 && window == time.Minute, "export limit policy")
			return http.Header{"X-RateLimit-Limit": {"10"}}, nil
		},
		RequestID: func(*http.Request) string { return "external-trace" },
	})
	check(err == nil && len(m.Patterns()) == 7, "module construction")
	for _, pattern := range m.Patterns() {
		if db == nil && !strings.HasSuffix(pattern, "/actions") {
			continue
		}
		path := strings.TrimPrefix(pattern, "GET ")
		path = strings.ReplaceAll(path, "{id}", "10000000-0000-0000-0000-000000000003")
		query := "?from=2020-01-01T00:00:00Z&to=2100-01-01T00:00:00Z"
		w := httptest.NewRecorder()
		m.ServeHTTP(w, httptest.NewRequest("GET", path+query, nil))
		check(w.Code == 200, "HTTP route "+pattern+" "+w.Body.String())
		if strings.HasSuffix(path, "/export") {
			var manifest audit.AuditExportManifest
			check(w.Header().Get("Content-Type") == "text/csv" && json.Unmarshal([]byte(w.Header().Get("X-Audit-Export-Manifest")), &manifest) == nil && manifest.RecordCount == 3, "CSV headers/manifest")
			w = httptest.NewRecorder()
			m.ServeHTTP(w, httptest.NewRequest("GET", path+query+"&format=json", nil))
			var body struct{ Data audit.AuditExportResponse }
			check(w.Code == 200 && json.Unmarshal(w.Body.Bytes(), &body) == nil, "JSON export")
			bytes, _ := json.Marshal(body.Data.Items)
			sum := sha256.Sum256(bytes)
			check(body.Data.Manifest.PayloadSHA256 == hex.EncodeToString(sum[:]) && body.Data.Manifest.PayloadSHA256 == manifest.PayloadSHA256, "export integrity")
		} else {
			check(strings.Contains(w.Body.String(), "external-trace"), "trace metadata")
		}
	}
	var document struct {
		Paths      map[string]any
		Components struct{ Schemas map[string]any }
	}
	check(json.Unmarshal(m.OpenAPI(), &document) == nil && len(document.Paths) == 7 && document.Components.Schemas["audit.AuditLogResponse"] != nil, "package-owned API")
	if db != nil {
		check(db.Ping() == nil, "module closed caller pool")
		_, err = db.Exec("DELETE FROM audit_logs")
		check(err != nil, "fixture must be SELECT-only")
		fmt.Println("PASS: Audit PostgreSQL consumer, seven reads/routes, keyset pagination, NULL metadata, bound filters, cancellation and CSV/JSON integrity")
	} else {
		fmt.Println("PASS: standalone Audit HTTP consumer and module-owned contract")
	}
}
func verifyRepository(repo audit.Repository) {
	ctx := context.Background()
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(80, 0, 0)
	input := audit.ListAuditLogsInput{StartTime: from, EndTime: to, Limit: 1}
	var ids []string
	for i := 0; i < 3; i++ {
		page, err := repo.ListAuditLogs(ctx, input)
		check(err == nil && len(page.Items) == 1, "keyset page")
		ids = append(ids, page.Items[0].ID)
		check(page.HasMore == (i < 2), "lookahead pagination")
		if i == 0 {
			check(page.Items[0].UserID == "" && page.Items[0].Success == nil && len(page.Items[0].Metadata) == 0, "SQL NULL metadata/user")
		}
		input.CursorCreatedAt = page.NextCursor.CursorCreatedAt
		input.CursorID = page.NextCursor.CursorID
	}
	check(ids[0] != ids[1] && ids[1] != ids[2] && ids[0] != ids[2], "cursor repeated tied timestamp")
	detail, err := repo.GetAuditLog(ctx, uuid.MustParse(ids[0]))
	check(err == nil && detail.ID == ids[0], "detail")
	_, err = repo.GetAuditLog(ctx, uuid.Nil)
	check(err == sql.ErrNoRows, "detail missing-row semantics")
	actions, err := repo.ListAuditActions(ctx)
	check(err == nil && len(actions.Actions) == 2, "actions")
	overview, err := repo.AuditMetricsOverview(ctx, audit.AuditWindowInput{StartTime: from, EndTime: to})
	check(err == nil && overview.TotalEvents == 3 && overview.FailedLogins == 1, "overview")
	days, err := repo.FailedLoginsByDay(ctx, audit.AuditWindowInput{StartTime: from, EndTime: to})
	check(err == nil && len(days.Series) == 1 && days.Series[0].Day == time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC).UnixMilli(), "failed login days")
	ips, err := repo.TopFailureIPs(ctx, audit.TopFailureIPsInput{StartTime: from, EndTime: to, Limit: 10})
	check(err == nil && len(ips.Items) == 1 && ips.Items[0].IP == "127.0.0.1", "failure IPs")
	items, err := repo.ExportAuditLogs(ctx, audit.ExportAuditLogsInput{StartTime: from, EndTime: to})
	check(err == nil && len(items) == 3, "export reads")
	items, err = repo.ExportAuditLogs(ctx, audit.ExportAuditLogsInput{StartTime: from, EndTime: to, Action: "' OR true --"})
	check(err == nil && len(items) == 0, "bound filter")
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = repo.ListAuditActions(canceled)
	check(err != nil, "cancellation")
}
