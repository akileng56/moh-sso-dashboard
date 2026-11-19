package main

import (
	"context"
	"fmt"
	units "github.com/mohuganda/moh-sso-dashboard/backend/modules/admin-units"
	"net/http"
	"net/http/httptest"
	"strings"
)

type warehouse struct{ units.Repository }

func (warehouse) ListDistricts(context.Context) ([]units.OrgUnitSimple, error) {
	id := "district"
	return []units.OrgUnitSimple{{OrgUnitID: &id}}, nil
}
func main() {
	mux := http.NewServeMux()
	m, err := units.Register(mux, units.Dependencies{
		Repository:      warehouse{},
		RequireSystem:   func(*http.Request, string) error { return nil },
		AuthorizeAny:    func(*http.Request, ...string) error { return nil },
		SelectContext:   func(*http.Request) (string, bool, error) { return "verified", true, nil },
		ContextNodeCode: func(context.Context, string) (string, error) { return "district", nil },
		ContextAliases: func(context.Context, string, bool) (map[string][]string, error) {
			return map[string][]string{"dwh-org-unit": {"district"}}, nil
		},
	})
	if err != nil {
		panic(err)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/visualizer/adminunits/district", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "district") || len(m.Patterns()) != 9 || len(m.OpenAPI()) == 0 {
		panic(w.Body.String())
	}
	fmt.Println("PASS: standalone Administrative Units HTTP consumer, scoped lookup and package-owned routes")
}
