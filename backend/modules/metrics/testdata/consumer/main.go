package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	metrics "github.com/mohuganda/moh-sso-dashboard/backend/modules/metrics"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"time"
)

type fake struct{ metrics.Repository }

func (fake) CountUsers(context.Context) (int64, error) { return 7, nil }
func check(ok bool, message string) {
	if !ok {
		panic(message)
	}
}
func main() {
	gin.SetMode(gin.TestMode)
	var repo metrics.Repository = fake{}
	var db *sql.DB
	if dsn := os.Getenv("METRICS_TEST_DSN"); dsn != "" {
		var err error
		db, err = sql.Open("postgres", dsn)
		check(err == nil, "open fixture")
		defer db.Close()
		repo, err = metrics.NewRepository(db)
		check(err == nil, "construct repository")
		verifyRepository(repo)
	}
	deps := metrics.Dependencies{Repository: repo, RequireAdmin: func(*http.Request) error { return nil }, AuthorizeAny: func(_ *http.Request, p ...string) error {
		check(len(p) == 1 && p[0] == "metrics:read", "permission changed")
		return nil
	}, RequestID: func(*http.Request) string { return "external-trace" }}
	mux := http.NewServeMux()
	module, err := metrics.Register(mux, deps)
	check(err == nil, "register")
	paths := module.Patterns()
	check(len(paths) == 19, "route count")
	for _, pattern := range paths {
		if db == nil && !strings.HasSuffix(pattern, "/system/count-users") {
			continue
		}
		route := strings.TrimPrefix(pattern, "GET ")
		route = strings.ReplaceAll(route, "{userID}", "00000000-0000-0000-0000-000000000001")
		query := "?start=" + time.Now().Add(-48*time.Hour).UTC().Format(time.RFC3339) + "&end=" + time.Now().Add(time.Hour).UTC().Format(time.RFC3339) + "&client_id=client-one"
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", route+query, nil))
		var body struct {
			Success bool
			Data    json.RawMessage
			Meta    struct {
				RequestID string `json:"requestId"`
			}
		}
		check(w.Code == 200 && json.Unmarshal(w.Body.Bytes(), &body) == nil && body.Success && body.Meta.RequestID == "external-trace", "HTTP route "+pattern+" "+w.Body.String())
	}
	var doc struct{ Paths map[string]any }
	check(json.Unmarshal(module.OpenAPI(), &doc) == nil && len(doc.Paths) == 19, "OpenAPI contract")
	if db != nil {
		check(db.Ping() == nil, "module closed caller database")
		_, err = db.Exec("DELETE FROM users")
		check(err != nil, "fixture is not read-only")
		fmt.Println("PASS: Metrics PostgreSQL consumer, 32 repository reads, 19 HTTP routes, NULLs, bound filters, cancellation and caller-owned pool")
	} else {
		fmt.Println("PASS: standalone Metrics HTTP consumer, verified capability wiring and package-owned contract")
	}
}
func verifyRepository(repo metrics.Repository) {
	ctx := context.Background()
	start := time.Now().Add(-48 * time.Hour)
	end := time.Now().Add(time.Hour)
	methods := reflect.TypeOf((*metrics.Repository)(nil)).Elem()
	for i := 0; i < methods.NumMethod(); i++ {
		method := methods.Method(i)
		args := []reflect.Value{}
		times := 0
		for j := 0; j < method.Type.NumIn(); j++ {
			t := method.Type.In(j)
			switch {
			case t.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()):
				args = append(args, reflect.ValueOf(ctx))
			case t == reflect.TypeOf(time.Time{}):
				v := start
				if times > 0 {
					v = end
				}
				times++
				args = append(args, reflect.ValueOf(v))
			case t == reflect.TypeOf(uuid.UUID{}):
				args = append(args, reflect.ValueOf(uuid.MustParse("00000000-0000-0000-0000-000000000001")))
			case t.Kind() == reflect.String:
				v := "client-one"
				if strings.Contains(method.Name, "Suspicious") {
					v = "UG"
				}
				args = append(args, reflect.ValueOf(v))
			case t.Kind() == reflect.Int32:
				args = append(args, reflect.ValueOf(int32(10)))
			default:
				panic("unexpected repository argument")
			}
		}
		results := reflect.ValueOf(repo).MethodByName(method.Name).Call(args)
		check(results[len(results)-1].IsNil(), "repository query "+method.Name)
	}
	check(methods.NumMethod() == 32, "repository method count")
	users, err := repo.CountUsers(ctx)
	check(err == nil && users == 3, "user count")
	disabled, err := repo.CountDisabledUsers(ctx)
	check(err == nil && disabled == 1, "disabled count")
	rows, err := repo.NewUsersInRange(ctx, start, end)
	check(err == nil && len(rows) == 3, "user rows")
	for _, row := range rows {
		check(!row.FirstName.Valid && len(row.Roles) == 1, "NULL or roles array lost")
	}
	count, err := repo.LoginCountForClientInRange(ctx, "' OR true --", start, end)
	check(err == nil && count == 0, "SQL parameter binding")
	missing, err := repo.LastLoginForUser(ctx, uuid.Nil)
	check(err == sql.ErrNoRows && !missing.Valid, "missing last login semantics")
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = repo.CountUsers(canceled)
	check(err != nil, "cancellation ignored")
}
