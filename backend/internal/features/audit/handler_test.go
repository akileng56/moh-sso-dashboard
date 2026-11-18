package audit

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/ratelimit"
)

func TestAuditRoutesRegisterStaticRoutesBeforeIDRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterAdminRoutes(router.Group("/admin"), &Handler{}, ratelimit.New(nil))

	order := map[string]int{}
	for index, route := range router.Routes() {
		if route.Method == "GET" {
			order[route.Path] = index
		}
	}

	idIndex, ok := order["/admin/audit-logs/:id"]
	if !ok {
		t.Fatal("expected audit detail route to be registered")
	}

	for _, path := range []string{
		"/admin/audit-logs/metrics/overview",
		"/admin/audit-logs/metrics/failed-logins-by-day",
		"/admin/audit-logs/metrics/top-failure-ips",
		"/admin/audit-logs/export",
	} {
		staticIndex, ok := order[path]
		if !ok {
			t.Fatalf("expected %s to be registered", path)
		}
		if staticIndex > idIndex {
			t.Fatalf("expected %s to be registered before /:id", path)
		}
	}
}

func TestBuildAuditManifestUsesRecordCount(t *testing.T) {
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)

	manifest := buildAuditManifest(from, to, []AuditLogResponse{{ID: "one"}, {ID: "two"}}, 2, nil)

	if got := manifest.RecordCount; got != 2 {
		t.Fatalf("expected record_count 2, got %v", got)
	}
	if manifest.PayloadSHA256 == "" {
		t.Fatal("expected payload hash to be set")
	}
}
