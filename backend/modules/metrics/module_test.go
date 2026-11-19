package metrics_test

import (
	"context"
	"encoding/json"
	"errors"
	metrics "github.com/mohuganda/moh-sso-dashboard/backend/modules/metrics"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type moduleRepo struct {
	metrics.Repository
	calls int
	fail  error
}

func (r *moduleRepo) CountUsers(context.Context) (int64, error) { r.calls++; return 7, r.fail }
func capabilities(repo metrics.Repository) metrics.Dependencies {
	return metrics.Dependencies{Repository: repo, RequireAdmin: func(*http.Request) error { return nil }, AuthorizeAny: func(_ *http.Request, p ...string) error {
		if len(p) != 1 || p[0] != metrics.ReadPermission {
			panic("wrong permission")
		}
		return nil
	}, RequestID: func(*http.Request) string { return "trusted-trace" }}
}
func TestModuleGuardsEveryRouteBeforeRepositoryAccess(t *testing.T) {
	for _, stage := range []string{"admin", "permission"} {
		for _, failure := range []error{metrics.ErrUnauthorized, metrics.ErrForbidden, errors.New("private provider credential")} {
			repo := &moduleRepo{}
			d := capabilities(repo)
			adminCalls, permissionCalls := 0, 0
			d.RequireAdmin = func(*http.Request) error {
				adminCalls++
				if stage == "admin" {
					return failure
				}
				return nil
			}
			d.AuthorizeAny = func(*http.Request, ...string) error { permissionCalls++; return failure }
			m, err := metrics.NewModule(d)
			if err != nil {
				t.Fatal(err)
			}
			if len(m.Patterns()) != 19 {
				t.Fatal(m.Patterns())
			}
			for _, pattern := range m.Patterns() {
				route := strings.TrimPrefix(pattern, "GET ")
				route = strings.ReplaceAll(route, "{userID}", "6596c0c5-4446-4caa-9bd0-c7ad991567ba")
				w := httptest.NewRecorder()
				m.ServeHTTP(w, httptest.NewRequest("GET", route, nil))
				want := 503
				if failure == metrics.ErrUnauthorized {
					want = 401
				}
				if failure == metrics.ErrForbidden {
					want = 403
				}
				if w.Code != want || repo.calls != 0 || strings.Contains(w.Body.String(), "private") {
					t.Fatal(pattern, w.Code, w.Body.String())
				}
			}
			if adminCalls != 19 || stage == "admin" && permissionCalls != 0 || stage == "permission" && permissionCalls != 19 {
				t.Fatal("guard order", adminCalls, permissionCalls)
			}
		}
	}
}
func TestModulePublicHTTPConsumerContract(t *testing.T) {
	repo := &moduleRepo{}
	mux := http.NewServeMux()
	m, err := metrics.Register(mux, capabilities(repo))
	if err != nil {
		t.Fatal(err)
	}
	p := m.Patterns()
	p[0] = "bad"
	if m.Patterns()[0] == "bad" {
		t.Fatal("mutable patterns")
	}
	for _, method := range []string{"GET", "HEAD", "POST"} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, "/admin/metrics/system/count-users", nil))
		if method == "GET" {
			if w.Code != 200 || !strings.Contains(w.Body.String(), `"total_users":7`) || !strings.Contains(w.Body.String(), "trusted-trace") {
				t.Fatal(w.Code, w.Body.String())
			}
		} else if w.Code == 200 {
			t.Fatal("expanded method surface", method)
		}
	}
	repo.fail = errors.New("private database password")
	w := httptest.NewRecorder()
	m.ServeHTTP(w, httptest.NewRequest("GET", "/admin/metrics/system/count-users", nil))
	if w.Code != 500 || strings.Contains(w.Body.String(), "private") {
		t.Fatal(w.Code, w.Body.String())
	}
	before := repo.calls
	for _, route := range []string{"/system/login-trend-range", "/users/last-login/invalid", "/clients/login-count", "/users/new-trend?start=invalid&end=invalid"} {
		w = httptest.NewRecorder()
		m.ServeHTTP(w, httptest.NewRequest("GET", "/admin/metrics"+route, nil))
		if w.Code != 400 || repo.calls != before {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	var doc struct {
		Paths      map[string]json.RawMessage
		Components struct{ Schemas map[string]json.RawMessage }
	}
	if err = json.Unmarshal(m.OpenAPI(), &doc); err != nil || len(doc.Paths) != 19 || doc.Components.Schemas["metrics.OverviewResponse"] == nil {
		t.Fatal(err, string(m.OpenAPI()))
	}
}
func TestModuleValidatesDependenciesAndCustomMount(t *testing.T) {
	d := capabilities(&moduleRepo{})
	for _, change := range []func(*metrics.Dependencies){
		func(d *metrics.Dependencies) { d.RequireAdmin = nil }, func(d *metrics.Dependencies) { d.AuthorizeAny = nil }, func(d *metrics.Dependencies) { d.Repository = nil }, func(d *metrics.Dependencies) { d.BasePath = "/bad/{wildcard}" },
	} {
		bad := d
		change(&bad)
		if _, err := metrics.NewModule(bad); !errors.Is(err, metrics.ErrInvalidDependencies) {
			t.Fatal(err)
		}
	}
	if _, err := metrics.NewRepository(nil); err == nil {
		t.Fatal("nil database accepted")
	}
	if _, err := metrics.Register(nil, d); !errors.Is(err, metrics.ErrInvalidDependencies) {
		t.Fatal(err)
	}
	d.BasePath = "/custom/stats"
	m, err := metrics.NewModule(d)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	m.ServeHTTP(w, httptest.NewRequest("GET", "/custom/stats/system/count-users", nil))
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
}
