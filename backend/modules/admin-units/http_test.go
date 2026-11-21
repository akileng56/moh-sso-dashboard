package admin_units_test

import (
	"context"
	"encoding/json"
	"errors"
	units "github.com/mohuganda/moh-sso-dashboard/backend/modules/admin-units"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type httpRepo struct {
	units.Repository
	calls  int
	filter *string
	fail   error
	empty  bool
}

func sp(s string) *string { return &s }
func (r *httpRepo) simple(f *string) ([]units.OrgUnitSimple, error) {
	r.calls++
	r.filter = f
	if r.empty {
		return nil, r.fail
	}
	return []units.OrgUnitSimple{{OrgUnitID: sp("allowed")}, {OrgUnitID: sp("hidden")}}, r.fail
}
func (r *httpRepo) ListOrgUnits(context.Context) ([]units.OrgUnitFull, error) {
	r.calls++
	if r.empty {
		return nil, r.fail
	}
	return []units.OrgUnitFull{{DimOrgHierarchyKey: 1, OrgUnitID: sp("allowed"), OrgUnitName: sp("visible"), Level: "1", IsCurrent: true}, {DimOrgHierarchyKey: 2, OrgUnitID: sp("hidden"), OrgUnitName: sp("secret"), Level: "1"}}, r.fail
}
func (r *httpRepo) ListFacilities(context.Context) ([]units.Facility, error) {
	r.calls++
	if r.empty {
		return nil, r.fail
	}
	return []units.Facility{{OrgUnitID: sp("allowed")}, {OrgUnitID: sp("hidden")}}, r.fail
}
func (r *httpRepo) ListDistricts(context.Context) ([]units.OrgUnitSimple, error) {
	return r.simple(nil)
}
func (r *httpRepo) ListSubCounties(_ context.Context, f *string) ([]units.OrgUnitSimple, error) {
	return r.simple(f)
}
func (r *httpRepo) ListLocalGovt(_ context.Context, f *string) ([]units.OrgUnitSimple, error) {
	return r.simple(f)
}
func (r *httpRepo) ListDistrictsByRegion(_ context.Context, f *string) ([]units.OrgUnitSimple, error) {
	return r.simple(f)
}
func (r *httpRepo) ListRegions(context.Context) ([]units.OrgUnitSimple, error)  { return r.simple(nil) }
func (r *httpRepo) ListNational(context.Context) ([]units.OrgUnitSimple, error) { return r.simple(nil) }
func dependencies(repo *httpRepo) units.Dependencies {
	return units.Dependencies{
		Repository: repo,
		RequireSystem: func(_ *http.Request, system string) error {
			if system != units.RequiredSystem {
				panic(system)
			}
			return nil
		},
		AuthorizeAny: func(_ *http.Request, p ...string) error {
			if !reflect.DeepEqual(p, units.RequiredPermissions()) {
				panic(p)
			}
			return nil
		},
		SelectContext:   func(*http.Request) (string, bool, error) { return "verified", true, nil },
		ContextNodeCode: func(context.Context, string) (string, error) { return "district", nil },
		ContextAliases: func(context.Context, string, bool) (map[string][]string, error) {
			return map[string][]string{"dwh-org-unit": {"allowed"}, "other": {"hidden"}}, nil
		},
		RequestID: func(*http.Request) string { return "trace" },
	}
}
func request(t *testing.T, d units.Dependencies, method, suffix, body string) *httptest.ResponseRecorder {
	t.Helper()
	m, err := units.NewModule(d)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	m.ServeHTTP(w, httptest.NewRequest(method, "/visualizer/adminunits/"+suffix, strings.NewReader(body)))
	return w
}
func TestHTTPRoutesScopeAndShape(t *testing.T) {
	for _, suffix := range []string{"orgunits", "facilities", "district", "subcounties", "localgovt", "districts", "region", "national", "hierarchy"} {
		t.Run(suffix, func(t *testing.T) {
			repo := &httpRepo{}
			method := "GET"
			body := ""
			if suffix == "subcounties" || suffix == "localgovt" || suffix == "districts" {
				method = "POST"
				body = "{}"
			}
			w := request(t, dependencies(repo), method, suffix, body)
			if w.Code != 200 || repo.calls != 1 || strings.Contains(w.Body.String(), "hidden") || strings.Contains(w.Body.String(), "secret") {
				t.Fatalf("%d %s calls=%d", w.Code, w.Body.String(), repo.calls)
			}
			var b struct {
				Data []map[string]any
				Meta map[string]string
			}
			if err := json.Unmarshal(w.Body.Bytes(), &b); err != nil {
				t.Fatal(err)
			}
			if len(b.Data) != 1 || b.Meta["requestId"] != "trace" {
				t.Fatal(w.Body.String())
			}
			repo.empty = true
			w = request(t, dependencies(repo), method, suffix, body)
			if w.Code != 200 || !strings.Contains(w.Body.String(), `"data":[]`) {
				t.Fatal(w.Body.String())
			}
		})
	}
}
func TestHTTPGuardsFailClosedInOrder(t *testing.T) {
	for stage := 0; stage < 3; stage++ {
		for _, failure := range []error{units.ErrUnauthorized, units.ErrForbidden, errors.New("private credential")} {
			repo := &httpRepo{}
			d := dependencies(repo)
			var order []int
			gate := func(n int) error {
				order = append(order, n)
				if n == stage {
					return failure
				}
				return nil
			}
			d.RequireSystem = func(*http.Request, string) error { return gate(0) }
			d.SelectContext = func(*http.Request) (string, bool, error) { return "", false, gate(1) }
			d.AuthorizeAny = func(*http.Request, ...string) error { return gate(2) }
			w := request(t, d, "GET", "orgunits", "")
			want := 503
			if failure == units.ErrUnauthorized {
				want = 401
			}
			if failure == units.ErrForbidden {
				want = 403
			}
			if w.Code != want || repo.calls != 0 || len(order) != stage+1 || strings.Contains(w.Body.String(), "private") {
				t.Fatalf("%d %v %s", w.Code, order, w.Body.String())
			}
		}
	}
}
func TestHTTPBodyAndNullableFilters(t *testing.T) {
	for _, body := range []string{"", "{", "[]", `{"region":1}`, "{} {}"} {
		repo := &httpRepo{}
		w := request(t, dependencies(repo), "POST", "districts", body)
		if w.Code != 400 || repo.calls != 0 {
			t.Fatalf("%q: %d %s", body, w.Code, w.Body.String())
		}
	}
	repo := &httpRepo{}
	d := dependencies(repo)
	d.MaxBodyBytes = 8
	for _, body := range []string{`{"region":"too long"}`, "{}" + strings.Repeat(" ", 40)} {
		w := request(t, d, "POST", "districts", body)
		if w.Code != 413 || repo.calls != 0 {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
	}
	for _, body := range []string{"{}", `{"region":null}`, `{"region":""}`, `{"region":"Central"}`} {
		repo = &httpRepo{}
		w := request(t, dependencies(repo), "POST", "districts", body)
		if w.Code != 200 {
			t.Fatal(w.Body.String())
		}
		if strings.Contains(body, "Central") && (repo.filter == nil || *repo.filter != "Central") {
			t.Fatal("filter lost")
		}
		if body == `{"region":null}` && repo.filter != nil {
			t.Fatal("NULL changed")
		}
	}
}
func TestHTTPContextModesAndSafeErrors(t *testing.T) {
	for _, mode := range []string{"legacy", "national", "national-node", "unmapped", "denied", "lookup-error", "warehouse-error"} {
		t.Run(mode, func(t *testing.T) {
			repo := &httpRepo{}
			d := dependencies(repo)
			want := 200
			count := 1
			switch mode {
			case "legacy":
				d.SelectContext = func(*http.Request) (string, bool, error) { return "", false, nil }
				count = 2
			case "national":
				d.ContextNodeCode = func(context.Context, string) (string, error) { return "UG", nil }
				count = 2
			case "national-node":
				d.ContextNodeCode = func(context.Context, string) (string, error) { return "UG", nil }
				d.SelectContext = func(*http.Request) (string, bool, error) { return "verified", false, nil }
			case "unmapped":
				d.ContextAliases = func(context.Context, string, bool) (map[string][]string, error) {
					return map[string][]string{"other": {"allowed"}}, nil
				}
				want = 403
			case "denied":
				d.ContextNodeCode = func(context.Context, string) (string, error) { return "", units.ErrForbidden }
				want = 403
			case "lookup-error":
				d.ContextAliases = func(context.Context, string, bool) (map[string][]string, error) {
					return nil, errors.New("private lookup")
				}
				want = 403
			case "warehouse-error":
				repo.fail = errors.New("private database")
				want = 500
			}
			w := request(t, d, "GET", "orgunits", "")
			if w.Code != want || strings.Contains(w.Body.String(), "private") {
				t.Fatal(w.Code, w.Body.String())
			}
			if want == 200 {
				var b struct{ Data []any }
				json.Unmarshal(w.Body.Bytes(), &b)
				if len(b.Data) != count {
					t.Fatal(w.Body.String())
				}
			}
		})
	}
}
func TestHTTPRegistrationAndOwnedOpenAPI(t *testing.T) {
	repo := &httpRepo{}
	d := dependencies(repo)
	m, err := units.Register(http.NewServeMux(), d)
	if err != nil {
		t.Fatal(err)
	}
	p := m.Patterns()
	if len(p) != 9 {
		t.Fatal(p)
	}
	p[0] = "bad"
	if m.Patterns()[0] == "bad" {
		t.Fatal("mutable patterns")
	}
	var doc struct {
		Paths      map[string]json.RawMessage
		Components struct{ Schemas map[string]json.RawMessage }
	}
	if err := json.Unmarshal(m.OpenAPI(), &doc); err != nil || len(doc.Paths) != 9 || len(doc.Components.Schemas) != 4 {
		t.Fatal(err, string(m.OpenAPI()))
	}
	for _, change := range []func(*units.Dependencies){
		func(d *units.Dependencies) { d.RequireSystem = nil }, func(d *units.Dependencies) { d.AuthorizeAny = nil },
		func(d *units.Dependencies) { d.SelectContext = nil }, func(d *units.Dependencies) { d.ContextNodeCode = nil },
		func(d *units.Dependencies) { d.ContextAliases = nil }, func(d *units.Dependencies) { d.Repository = nil },
		func(d *units.Dependencies) { d.BasePath = "/bad/{wildcard}" }, func(d *units.Dependencies) { d.MaxBodyBytes = -1 },
	} {
		bad := d
		change(&bad)
		if _, err := units.NewModule(bad); !errors.Is(err, units.ErrInvalidDependencies) {
			t.Fatal("invalid dependencies accepted", err)
		}
	}
	w := request(t, d, "HEAD", "district", "")
	if w.Code != 200 || w.Body.Len() != 0 {
		t.Fatal(w.Code, w.Body.String())
	}
	if _, err := units.NewRepository(nil).ListOrgUnits(context.Background()); err == nil {
		t.Fatal("nil warehouse accepted")
	}
}
