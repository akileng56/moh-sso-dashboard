package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	units "github.com/mohuganda/moh-sso-dashboard/backend/modules/admin-units"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
)

func check(ok bool, message string) {
	if !ok {
		panic(message)
	}
}
func main() {
	db, err := sql.Open("postgres", os.Getenv("ADMIN_UNITS_TEST_DSN"))
	if err != nil {
		panic("open warehouse")
	}
	defer db.Close()
	ctx := context.Background()
	repo := units.NewRepository(db)
	rows, err := repo.ListOrgUnits(ctx)
	check(err == nil && len(rows) == 6, "current/sentinel filtering")
	for _, row := range rows {
		check(row.OrgUnitName == nil, "SQL NULL was changed")
	}
	facilities, err := repo.ListFacilities(ctx)
	check(err == nil && len(facilities) == 1 && facilities[0].Level == "6", "facilities scan")
	hierarchy, err := repo.ListHierarchy(ctx)
	check(err == nil && len(hierarchy) == 6, "hierarchy scan")
	for name, fn := range map[string]func(context.Context) ([]units.OrgUnitSimple, error){"district": repo.ListDistricts, "region": repo.ListRegions, "national": repo.ListNational} {
		rows, err := fn(ctx)
		check(err == nil && len(rows) == 1, name+" level filter")
	}
	for name, fn := range map[string]func(context.Context, *string) ([]units.OrgUnitSimple, error){"subcounty": repo.ListSubCounties, "localgovt": repo.ListLocalGovt, "districts": repo.ListDistrictsByRegion} {
		for _, filter := range []*string{nil, ptr(""), ptr("fixture")} {
			rows, err := fn(ctx, filter)
			check(err == nil && len(rows) == 1, name+" optional bound filter")
		}
		rows, err := fn(ctx, ptr("' OR true --"))
		check(err == nil && len(rows) == 0, name+" SQL parameter binding")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = repo.ListOrgUnits(canceled)
	check(err != nil, "context cancellation ignored")
	m, err := units.NewModule(units.Dependencies{
		Database:        db,
		RequireSystem:   func(*http.Request, string) error { return nil },
		AuthorizeAny:    func(*http.Request, ...string) error { return nil },
		SelectContext:   func(*http.Request) (string, bool, error) { return "selected", true, nil },
		ContextNodeCode: func(context.Context, string) (string, error) { return "UG", nil },
		ContextAliases: func(context.Context, string, bool) (map[string][]string, error) {
			panic("national should not read aliases")
		},
	})
	check(err == nil, "construct SQL-backed HTTP module")
	for _, pattern := range m.Patterns() {
		fields := strings.SplitN(pattern, " ", 2)
		w := httptest.NewRecorder()
		m.ServeHTTP(w, httptest.NewRequest(fields[0], fields[1], strings.NewReader("{}")))
		var body struct {
			Success bool
			Data    []any
		}
		err = json.Unmarshal(w.Body.Bytes(), &body)
		check(w.Code == 200 && err == nil && body.Success && len(body.Data) > 0, "SQL-backed endpoint "+pattern)
	}
	check(db.PingContext(ctx) == nil, "module closed caller database")
	_, err = db.ExecContext(ctx, "DELETE FROM dwh.dim_org_hierarchy")
	check(err != nil, "fixture reader is not read-only")
	fmt.Println("PASS: PostgreSQL read-only warehouse, nine HTTP routes, all repository queries, SQL NULLs, bound filters, cancellation and pool ownership")
}
func ptr(s string) *string { return &s }
