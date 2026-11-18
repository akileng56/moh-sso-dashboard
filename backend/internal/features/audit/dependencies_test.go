package audit

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/authz"
)

type auditReadFixture struct {
	Repository
	inputs []ListAuditLogsInput
	items  []AuditLogResponse
	err    error
}

func (r *auditReadFixture) ListAuditLogs(_ context.Context, input ListAuditLogsInput) (AuditLogListResponse, error) {
	r.inputs = append(r.inputs, input)
	return AuditLogListResponse{Items: r.items, HasMore: true, NextCursor: AuditCursor{CursorID: input.CursorID, CursorCreatedAt: input.CursorCreatedAt}}, r.err
}
func (r *auditReadFixture) ExportAuditLogs(context.Context, ExportAuditLogsInput) ([]AuditLogResponse, error) {
	return r.items, r.err
}

type auditCacheFixture struct {
	payload        []byte
	getErr, setErr error
	gets, sets     int
	key            string
	ttl            time.Duration
}

func (c *auditCacheFixture) Get(_ context.Context, key string, dest any) (bool, error) {
	c.gets++
	c.key = key
	if c.getErr != nil {
		return true, c.getErr
	} // A failed adapter may leave partial data.
	if c.payload == nil {
		return false, nil
	}
	if err := json.Unmarshal(c.payload, dest); err != nil {
		return false, err
	}
	return true, nil
}
func (c *auditCacheFixture) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	c.sets++
	c.key = key
	c.ttl = ttl
	if c.setErr != nil {
		return c.setErr
	}
	var err error
	c.payload, err = json.Marshal(value)
	return err
}

func auditFixtureItems() []AuditLogResponse {
	success := false
	return []AuditLogResponse{{ID: "11111111-1111-1111-1111-111111111111", CreatedAt: "2026-11-18T06:00:00Z", Username: "fixture", Action: "login_failed", IP: "127.0.0.1", ClientID: "dashboard-web", Success: &success, Metadata: map[string]any{"success": false, "message": "fixture, quoted \"value\""}}}
}

func TestAuditReadCachePortPreservesQueriesAndResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	failure := errors.New("cache unavailable")
	for _, tc := range []struct {
		name                        string
		cache                       *auditCacheFixture
		repoErr                     error
		wantCalls, wantSets, status int
	}{
		{"disabled", nil, nil, 2, 0, 200},
		{"miss then hit", &auditCacheFixture{}, nil, 1, 1, 200},
		{"lookup error", &auditCacheFixture{getErr: failure}, nil, 2, 2, 200},
		{"corrupt cache", &auditCacheFixture{payload: []byte("broken JSON"), setErr: failure}, nil, 2, 2, 200},
		{"write error", &auditCacheFixture{setErr: failure}, nil, 2, 2, 200},
		{"repository failure", &auditCacheFixture{}, errors.New("database unavailable"), 2, 0, 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &auditReadFixture{items: auditFixtureItems(), err: tc.repoErr}
			var cache Cache
			if tc.cache != nil {
				cache = tc.cache
			}
			router := gin.New()
			router.GET("/logs", NewHandler(NewService(repo), cache).ListAuditLogs)
			id := uuid.MustParse("22222222-2222-2222-2222-222222222222")
			query := url.Values{"from": {"2026-11-18T00:00:00Z"}, "to": {"2026-11-19T00:00:00Z"}, "action": {"login_failed"}, "user_id": {id.String()}, "client_id": {"dashboard-web"}, "ip": {"127.0.0.1"}, "success": {"false"}, "limit": {"1"}, "cursor_id": {id.String()}, "cursor_created_at": {"2026-11-18T06:00:00Z"}}.Encode()
			var first AuditLogListResponse
			for i := 0; i < 2; i++ {
				res := httptest.NewRecorder()
				router.ServeHTTP(res, httptest.NewRequest("GET", "/logs?"+query, nil))
				if res.Code != tc.status {
					t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
				}
				if tc.status == 200 {
					var body struct {
						Success bool                 `json:"success"`
						Data    AuditLogListResponse `json:"data"`
					}
					if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
						t.Fatal(err)
					}
					if !body.Success || !reflect.DeepEqual(body.Data.Items, repo.items) || !body.Data.HasMore || body.Data.NextCursor.CursorID == nil || *body.Data.NextCursor.CursorID != id {
						t.Fatalf("response contract lost: %+v", body)
					}
					if i == 0 {
						first = body.Data
					} else if !reflect.DeepEqual(first, body.Data) {
						t.Fatal("cache changed response data")
					}
				}
			}
			if len(repo.inputs) != tc.wantCalls {
				t.Fatalf("repository calls=%d want=%d", len(repo.inputs), tc.wantCalls)
			}
			input := repo.inputs[0]
			if input.Limit != 1 || input.Action != "login_failed" || input.UserID == nil || *input.UserID != id || input.ClientID != "dashboard-web" || input.IP != "127.0.0.1" || input.Success != "false" || input.CursorID == nil || *input.CursorID != id || input.CursorCreatedAt == nil || input.StartTime.Format(time.RFC3339) != "2026-11-18T00:00:00Z" || input.EndTime.Format(time.RFC3339) != "2026-11-19T00:00:00Z" {
				t.Fatalf("query input lost: %+v", input)
			}
			if tc.cache != nil && (tc.cache.sets != tc.wantSets || tc.cache.key != "audit_logs:"+query || (tc.wantSets > 0 && tc.cache.ttl != 20*time.Second)) {
				t.Fatalf("cache contract lost: %+v", tc.cache)
			}
		})
	}
}

func TestAuditCacheDoesNotBypassReadPermission(t *testing.T) {
	repo := &auditReadFixture{items: auditFixtureItems()}
	cache := &auditCacheFixture{}
	router := gin.New()
	group := router.Group("/admin", func(c *gin.Context) {
		if c.GetHeader("X-Fixture-Reader") == "true" {
			c.Set(authz.ContextKey, authz.Context{Permissions: []authz.Permission{authz.PermissionAuditRead}})
		}
	})
	RegisterAdminRoutes(group, NewHandler(NewService(repo), cache), nil)
	for _, reader := range []bool{true, false} {
		req := httptest.NewRequest("GET", "/admin/audit-logs?from=2026-11-18T00:00:00Z&to=2026-11-19T00:00:00Z", nil)
		if reader {
			req.Header.Set("X-Fixture-Reader", "true")
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		want := 401
		if reader {
			want = 200
		}
		if res.Code != want {
			t.Fatalf("reader=%v status=%d body=%s", reader, res.Code, res.Body.String())
		}
	}
	if cache.gets != 1 || len(repo.inputs) != 1 {
		t.Fatal("unauthorized request reached cache or repository")
	}
}

func TestAuditExportPortPreservesJSONAndCSVContracts(t *testing.T) {
	repo := &auditReadFixture{items: auditFixtureItems()}
	router := gin.New()
	router.GET("/export", NewHandler(NewService(repo), nil).ExportAuditLogs)
	for _, format := range []string{"json", "csv"} {
		res := httptest.NewRecorder()
		router.ServeHTTP(res, httptest.NewRequest("GET", "/export?from=2026-11-18T00:00:00Z&to=2026-11-19T00:00:00Z&format="+format, nil))
		if res.Code != 200 {
			t.Fatal(res.Body.String())
		}
		if format == "json" {
			var body struct {
				Data AuditExportResponse `json:"data"`
			}
			if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			payload, _ := json.Marshal(body.Data.Items)
			hash := sha256.Sum256(payload)
			if !reflect.DeepEqual(body.Data.Items, repo.items) || body.Data.Manifest.RecordCount != 1 || body.Data.Manifest.HashAlgo != "sha256" || body.Data.Manifest.PayloadSHA256 != hex.EncodeToString(hash[:]) {
				t.Fatalf("export contract lost: %+v", body)
			}
		} else {
			rows, err := csv.NewReader(strings.NewReader(res.Body.String())).ReadAll()
			if err != nil || len(rows) != 2 || len(rows[0]) != 8 || rows[0][0] != "created_at" || rows[1][2] != "fixture" || rows[1][6] != "false" {
				t.Fatalf("CSV contract lost: %v %v", rows, err)
			}
			var metadata map[string]any
			if err := json.Unmarshal([]byte(rows[1][7]), &metadata); err != nil || !reflect.DeepEqual(metadata, repo.items[0].Metadata) {
				t.Fatal("CSV metadata quoting changed")
			}
		}
	}
}
