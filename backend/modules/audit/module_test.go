package audit_test

import (
	"errors"
	audit "github.com/mohuganda/moh-sso-dashboard/backend/modules/audit"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func moduleDeps(repo audit.Repository) audit.Dependencies {
	return audit.Dependencies{Repository: repo, RequireAdmin: func(*http.Request) error { return nil }, AuthorizeAny: func(_ *http.Request, p ...string) error {
		if len(p) != 1 || p[0] != "audit:read" {
			panic(p)
		}
		return nil
	}, LimitExport: func(_ *http.Request, name string, limit int, window time.Duration) (http.Header, error) {
		if name != "audit-export" || limit != 10 || window != time.Minute {
			panic("export policy changed")
		}
		return http.Header{"X-RateLimit-Limit": {"10"}}, nil
	}, RequestID: func(*http.Request) string { return "trace" }}
}

const auditWindow = "?from=2026-01-01T00:00:00Z&to=2027-01-01T00:00:00Z"

func TestModuleGuardsAllReadsBeforeCache(t *testing.T) {
	for _, stage := range []string{"admin", "permission"} {
		for _, failure := range []error{audit.ErrUnauthorized, audit.ErrForbidden, errors.New("private secret")} {
			repo := &consumerRepository{}
			cache := consumerCache{}
			d := moduleDeps(repo)
			d.Cache = cache
			admin, permission := 0, 0
			d.RequireAdmin = func(*http.Request) error {
				admin++
				if stage == "admin" {
					return failure
				}
				return nil
			}
			d.AuthorizeAny = func(*http.Request, ...string) error { permission++; return failure }
			m, err := audit.NewModule(d)
			if err != nil {
				t.Fatal(err)
			}
			for _, pattern := range m.Patterns() {
				route := strings.ReplaceAll(strings.TrimPrefix(pattern, "GET "), "{id}", "11111111-1111-1111-1111-111111111111")
				w := httptest.NewRecorder()
				m.ServeHTTP(w, httptest.NewRequest("GET", route+auditWindow, nil))
				want := 503
				if failure == audit.ErrUnauthorized {
					want = 401
				}
				if failure == audit.ErrForbidden {
					want = 403
				}
				if w.Code != want || repo.reads != 0 || len(cache) != 0 || strings.Contains(w.Body.String(), "private") {
					t.Fatal(pattern, w.Code, w.Body.String())
				}
			}
			if admin != 7 || stage == "admin" && permission != 0 || stage == "permission" && permission != 7 {
				t.Fatal("wrong guard ordering")
			}
		}
	}
}

type limited struct{}

func (limited) Error() string             { return "private limit store" }
func (limited) StatusCode() int           { return 429 }
func (limited) RetryAfter() time.Duration { return 1500 * time.Millisecond }
func TestModuleExportLimitAndSafeStreaming(t *testing.T) {
	repo := &consumerRepository{item: audit.AuditLogResponse{ID: "row", Action: "login", Metadata: map[string]any{"large": strings.Repeat("x", 9000)}}}
	d := moduleDeps(repo)
	d.LimitExport = func(*http.Request, string, int, time.Duration) (http.Header, error) {
		return http.Header{"X-RateLimit-Remaining": {"0"}}, limited{}
	}
	m, err := audit.NewModule(d)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	m.ServeHTTP(w, httptest.NewRequest("GET", "/admin/audit-logs/export"+auditWindow, nil))
	if w.Code != 429 || w.Header().Get("Retry-After") != "2" || w.Header().Get("X-RateLimit-Remaining") != "0" || strings.Contains(w.Body.String(), "private") {
		t.Fatal(w.Code, w.Body.String())
	}
	d = moduleDeps(repo)
	m, err = audit.NewModule(d)
	if err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{"csv", "json"} {
		w = httptest.NewRecorder()
		w.Header().Set("X-RateLimit-Limit", "1000")
		m.ServeHTTP(w, httptest.NewRequest("GET", "/admin/audit-logs/export"+auditWindow+"&format="+format, nil))
		if w.Code != 200 || !strings.Contains(w.Header().Get("Content-Disposition"), "audit_logs_export."+format) || w.Header().Get("X-RateLimit-Limit") != "10" {
			t.Fatal(w.Code, w.Header())
		}
		if format == "csv" && w.Header().Get("X-Audit-Export-Manifest") == "" {
			t.Fatal("large CSV lost manifest headers")
		}
	}
}
func TestModuleRoutesValidationAndCachedQueries(t *testing.T) {
	repo := &consumerRepository{}
	d := moduleDeps(repo)
	d.Cache = consumerCache{}
	mux := http.NewServeMux()
	m, err := audit.Register(mux, d)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Patterns()) != 7 {
		t.Fatal(m.Patterns())
	}
	p := m.Patterns()
	p[0] = "bad"
	if m.Patterns()[0] == "bad" {
		t.Fatal("mutable patterns")
	}
	for _, pattern := range m.Patterns() {
		route := strings.ReplaceAll(strings.TrimPrefix(pattern, "GET "), "{id}", "11111111-1111-1111-1111-111111111111")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", route+auditWindow, nil))
		if w.Code != 200 {
			t.Fatal(pattern, w.Code, w.Body.String())
		}
	}
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		m.ServeHTTP(w, httptest.NewRequest("GET", "/admin/audit-logs"+auditWindow, nil))
	}
	if repo.reads != 1 {
		t.Fatal("raw-query cache identity changed", repo.reads)
	}
	for _, route := range []string{"/admin/audit-logs", "/admin/audit-logs/invalid", "/admin/audit-logs" + auditWindow + "&limit=201", "/admin/audit-logs" + auditWindow + "&user_id=invalid", "/admin/audit-logs" + auditWindow + "&cursor_created_at=bad", "/admin/audit-logs/metrics/top-failure-ips" + auditWindow + "&limit=51"} {
		w := httptest.NewRecorder()
		m.ServeHTTP(w, httptest.NewRequest("GET", route, nil))
		if w.Code != 400 {
			t.Fatal(route, w.Code)
		}
	}
	for _, change := range []func(*audit.Dependencies){func(d *audit.Dependencies) { d.RequireAdmin = nil }, func(d *audit.Dependencies) { d.AuthorizeAny = nil }, func(d *audit.Dependencies) { d.LimitExport = nil }, func(d *audit.Dependencies) { d.Repository = nil }, func(d *audit.Dependencies) { d.BasePath = "/bad/{id}" }} {
		bad := d
		change(&bad)
		if _, err := audit.NewModule(bad); !errors.Is(err, audit.ErrInvalidDependencies) {
			t.Fatal(err)
		}
	}
	if _, err := audit.NewRepository(nil); err == nil {
		t.Fatal("nil database accepted")
	}
}

type brokenResponse struct{ header http.Header }

func (w *brokenResponse) Header() http.Header       { return w.header }
func (w *brokenResponse) WriteHeader(int)           {}
func (w *brokenResponse) Write([]byte) (int, error) { return 0, errors.New("client disconnected") }
func TestCSVStreamingFailureReachesSharedReporter(t *testing.T) {
	repo := &consumerRepository{item: audit.AuditLogResponse{Action: "login", Metadata: map[string]any{"large": strings.Repeat("x", 9000)}}}
	d := moduleDeps(repo)
	reports := 0
	d.StreamError = func(_ *http.Request, err error) {
		if err == nil {
			t.Fatal("missing stream error")
		}
		reports++
	}
	m, err := audit.NewModule(d)
	if err != nil {
		t.Fatal(err)
	}
	writer := &brokenResponse{header: http.Header{}}
	m.ServeHTTP(writer, httptest.NewRequest("GET", "/admin/audit-logs/export"+auditWindow, nil))
	if reports != 1 || writer.header.Get("X-Audit-Export-Manifest") == "" {
		t.Fatal("lost streaming error or manifest", reports)
	}
}
