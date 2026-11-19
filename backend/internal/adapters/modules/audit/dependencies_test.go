package audit

import (
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

func TestAuditCSVManifestHeaderPrecedesLargeBody(t *testing.T) {
	repo := &auditReadFixture{items: auditFixtureItems()}
	repo.items[0].Metadata["large"] = strings.Repeat("payload", 2000)
	router := gin.New()
	router.GET("/export", NewHandler(NewService(repo), nil).ExportAuditLogs)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest("GET", "/export?from=2026-11-18T00:00:00Z&to=2026-11-19T00:00:00Z", nil))
	if res.Code != 200 || res.Body.Len() <= 4096 {
		t.Fatalf("large export failed: %d bytes=%d", res.Code, res.Body.Len())
	}
	// Result uses the headers actually sent at the first write, not the mutable map.
	sent := res.Result()
	defer sent.Body.Close()
	var manifest AuditExportManifest
	if err := json.Unmarshal([]byte(sent.Header.Get("X-Audit-Export-Manifest")), &manifest); err != nil {
		t.Fatalf("streaming manifest header missing: %v", err)
	}
	payload, _ := json.Marshal(repo.items)
	hash := sha256.Sum256(payload)
	if manifest.RecordCount != 1 || manifest.PayloadSHA256 != hex.EncodeToString(hash[:]) {
		t.Fatalf("streaming manifest changed: %+v", manifest)
	}
	if sent.Header.Get("Content-Type") != "text/csv" || sent.Header.Get("Content-Disposition") != "attachment; filename=audit_logs_export.csv" {
		t.Fatalf("streaming headers changed: %v", sent.Header)
	}
}

func TestAuditExportEncodingFailurePrecedesStreaming(t *testing.T) {
	for _, format := range []string{"json", "csv"} {
		repo := &auditReadFixture{items: auditFixtureItems()}
		repo.items[0].Metadata["unsupported"] = make(chan int)
		router := gin.New()
		router.GET("/export", NewHandler(NewService(repo), nil).ExportAuditLogs)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, httptest.NewRequest("GET", "/export?from=2026-11-18T00:00:00Z&to=2026-11-19T00:00:00Z&format="+format, nil))
		if res.Code != 500 || !strings.Contains(res.Body.String(), "Export failed") || res.Result().Header.Get("Content-Disposition") != "" {
			t.Fatalf("failed export started streaming: %d %s", res.Code, res.Body.String())
		}
	}
}

func TestAuditValidatesFiltersBeforeCacheLookup(t *testing.T) {
	for _, invalid := range []string{"limit=0", "limit=201", "limit=bad", "user_id=bad", "cursor_id=bad", "cursor_created_at=bad"} {
		repo := &auditReadFixture{items: auditFixtureItems()}
		cache := &auditCacheFixture{payload: []byte("{}")}
		router := gin.New()
		router.GET("/logs", NewHandler(NewService(repo), cache).ListAuditLogs)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, httptest.NewRequest("GET", "/logs?from=2026-11-18T00:00:00Z&to=2026-11-19T00:00:00Z&"+invalid, nil))
		if res.Code != 400 || cache.gets != 0 || len(repo.inputs) != 0 {
			t.Fatalf("invalid query reached reader: %s status=%d gets=%d", invalid, res.Code, cache.gets)
		}
	}
}
