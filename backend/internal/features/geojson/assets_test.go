package geojson

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

type assetReaderFunc func(context.Context, string) (Asset, error)

func (f assetReaderFunc) Open(ctx context.Context, filename string) (Asset, error) {
	return f(ctx, filename)
}

type trackedAsset struct {
	*bytes.Reader
	closed int
}

func (a *trackedAsset) Close() error { a.closed++; return nil }

func TestInjectedAssetReaderPreservesHTTPContentSemantics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	modified := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name, header, value, body string
		status                    int
	}{
		{"full", "", "", "0123456789", http.StatusOK},
		{"range", "Range", "bytes=2-5", "2345", http.StatusPartialContent},
		{"not-modified", "If-Modified-Since", modified.Format(http.TimeFormat), "", http.StatusNotModified},
		{"invalid-range", "Range", "bytes=20-30", "", http.StatusRequestedRangeNotSatisfiable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stream := &trackedAsset{Reader: bytes.NewReader([]byte("0123456789"))}
			calls := 0
			reader := assetReaderFunc(func(ctx context.Context, filename string) (Asset, error) {
				calls++
				if filename != "uga_water.geojson" {
					t.Fatalf("unexpected asset %q", filename)
				}
				if ctx.Err() != nil {
					t.Fatal(ctx.Err())
				}
				return Asset{Content: stream, ModifiedAt: modified}, nil
			})
			router := gin.New()
			router.GET("/geojson/:name", NewHandlerWithReader(reader).GetGeoJSON)
			req := httptest.NewRequest(http.MethodGet, "/geojson/geo_water", nil)
			if tc.header != "" {
				req.Header.Set(tc.header, tc.value)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tc.status {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
			if tc.status != http.StatusRequestedRangeNotSatisfiable && rec.Body.String() != tc.body {
				t.Fatalf("unexpected body %q", rec.Body.String())
			}
			if calls != 1 || stream.closed != 1 {
				t.Fatalf("opens=%d closes=%d", calls, stream.closed)
			}
			if tc.status == http.StatusOK || tc.status == http.StatusPartialContent {
				if rec.Header().Get("Content-Type") != "application/geo+json" ||
					rec.Header().Get("Cache-Control") != "public, max-age=86400" ||
					rec.Header().Get("X-Content-Type-Options") != "nosniff" {
					t.Fatalf("serving headers changed: %v", rec.Header())
				}
			}
			if tc.status == http.StatusPartialContent && rec.Header().Get("Content-Range") != "bytes 2-5/10" {
				t.Fatalf("unexpected range: %s", rec.Header().Get("Content-Range"))
			}
		})
	}
}

func TestAssetReaderIsNotCalledForNamesOutsideAllowlist(t *testing.T) {
	for _, name := range []string{"../secret", "/secret", "geo_water.geojson", "geo_water/..", "unknown"} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			handler := NewHandlerWithReader(assetReaderFunc(func(context.Context, string) (Asset, error) {
				calls++
				return Asset{}, errors.New("must not open")
			}))
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Params = gin.Params{{Key: "name", Value: name}}
			handler.GetGeoJSON(c)
			if rec.Code != http.StatusNotFound || calls != 0 {
				t.Fatalf("status=%d opens=%d", rec.Code, calls)
			}
		})
	}
}

func TestAssetReaderFailuresStaySafeAndCloseStreams(t *testing.T) {
	for _, tc := range []struct {
		name              string
		failed, hasStream bool
	}{
		{"reader-error", true, false}, {"partial-reader-error", true, true}, {"nil-content", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stream := &trackedAsset{Reader: bytes.NewReader(nil)}
			handler := NewHandlerWithReader(assetReaderFunc(func(context.Context, string) (Asset, error) {
				asset := Asset{}
				if tc.hasStream {
					asset.Content = stream
				}
				if tc.failed {
					return asset, errors.New("/private/maps/internal-file.geojson")
				}
				return asset, nil
			}))
			router := gin.New()
			router.GET("/geojson/:name", handler.GetGeoJSON)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/geojson/geo_water", nil))
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("status=%d", rec.Code)
			}
			assertAPIErrorCode(t, rec.Body.Bytes(), "GEOJSON_LOAD_FAILED")
			if strings.Contains(rec.Body.String(), "/private") {
				t.Fatal("storage details leaked")
			}
			want := 0
			if tc.hasStream {
				want = 1
			}
			if stream.closed != want {
				t.Fatalf("closes=%d want=%d", stream.closed, want)
			}
		})
	}
}

func TestProtectedGeoJSONReadsRequireAnyExistingMapPermission(t *testing.T) {
	for _, tc := range []struct {
		name       string
		permission authz.Permission
		status     int
	}{
		{"anonymous", "", http.StatusUnauthorized},
		{"authenticated-without-permission", "", http.StatusForbidden},
		{"surveillance", authz.PermissionSurveillanceRead, http.StatusOK},
		{"data-quality", authz.PermissionDataQualityRead, http.StatusOK},
		{"report-browser", authz.PermissionReportBrowserRead, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			reader := assetReaderFunc(func(context.Context, string) (Asset, error) {
				calls++
				return Asset{Content: &trackedAsset{Reader: bytes.NewReader([]byte("{}"))}}, nil
			})
			router := gin.New()
			group := router.Group("/api/v1", func(c *gin.Context) {
				if tc.name == "anonymous" {
					return
				}
				actor := authz.Context{UserID: "fixture"}
				if tc.permission != "" {
					actor.Permissions = []authz.Permission{tc.permission}
				}
				c.Set("user_id", actor.UserID)
				c.Set(authz.ContextKey, actor)
			}, middleware.RequireAuth())
			RegisterProtectedRoutes(group, NewHandlerWithReader(reader))
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/geojson/geo_water", nil))
			wantCalls := 0
			if tc.status == http.StatusOK {
				wantCalls = 1
			}
			if rec.Code != tc.status || calls != wantCalls {
				t.Fatalf("status=%d opens=%d", rec.Code, calls)
			}
		})
	}
}

func TestDirectoryReaderHonorsContextAndCanonicalFilenames(t *testing.T) {
	dir := t.TempDir()
	body := []byte("{}")
	if err := os.WriteFile(filepath.Join(dir, "uga_water.geojson"), body, 0600); err != nil {
		t.Fatal(err)
	}
	reader := NewDirectoryReader(dir)
	asset, err := reader.Open(context.Background(), "uga_water.geojson")
	if err != nil {
		t.Fatal(err)
	}
	if asset.Content == nil || asset.ModifiedAt.IsZero() {
		t.Fatalf("incomplete asset: %+v", asset)
	}
	_ = asset.Content.Close()
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := reader.Open(canceled, "uga_water.geojson"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
	for _, name := range []string{"", ".", "..", "../uga_water.geojson", "/tmp/map.geojson", "nested/map.geojson", "nested\\map.geojson", "missing.geojson"} {
		if asset, err := reader.Open(context.Background(), name); err == nil || asset.Content != nil {
			t.Fatalf("expected failure without stream for %q", name)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "directory.geojson"), 0700); err != nil {
		t.Fatal(err)
	}
	if asset, err := reader.Open(context.Background(), "directory.geojson"); err == nil || asset.Content != nil {
		t.Fatal("directory must not be served")
	}
}
