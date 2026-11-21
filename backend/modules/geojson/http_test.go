package geojson_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	geojson "github.com/mohuganda/moh-sso-dashboard/backend/modules/geojson"
)

type httpAssetReader struct {
	calls   int
	streams []*consumerStream
	err     error
}

func (r *httpAssetReader) Open(_ context.Context, _ string) (geojson.Asset, error) {
	r.calls++
	stream := &consumerStream{Reader: bytes.NewReader([]byte("abcdef"))}
	r.streams = append(r.streams, stream)
	return geojson.Asset{Content: stream, ModifiedAt: time.Unix(100, 0).UTC()}, r.err
}
func allowed(*http.Request, ...string) error { return nil }
func serveHTTP(h http.Handler, url string, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest("GET", url, nil)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, request)
	return recorder
}

type statusFailure struct{ status int }

func (e statusFailure) Error() string   { return "private authorization diagnostic" }
func (e statusFailure) StatusCode() int { return e.status }

type countingRegistrar struct{ calls int }

func (r *countingRegistrar) Handle(string, http.Handler) { r.calls++ }

func TestModuleRejectsMissingSecurityBeforeRegistration(t *testing.T) {
	registrar := &countingRegistrar{}
	if _, err := geojson.Register(registrar, geojson.Dependencies{}); !errors.Is(err, geojson.ErrInvalidDependencies) || registrar.calls != 0 {
		t.Fatal("unsecured endpoint exposed", err, registrar.calls)
	}
	if _, err := geojson.Register(nil, geojson.Dependencies{AuthorizeAny: allowed}); !errors.Is(err, geojson.ErrInvalidDependencies) {
		t.Fatal(err)
	}
	if _, err := geojson.NewModule(geojson.Dependencies{AuthorizeAny: allowed, Assets: &httpAssetReader{}, Directory: "maps"}); !errors.Is(err, geojson.ErrInvalidDependencies) {
		t.Fatal("ambiguous storage accepted", err)
	}
	for _, prefix := range []string{"/", "relative", "//maps", "/a/../maps", "/maps/", "/maps/{id}", "/bad%zz", "/bad path", "/bad?query"} {
		if _, err := geojson.NewModule(geojson.Dependencies{AuthorizeAny: allowed, BasePath: prefix}); !errors.Is(err, geojson.ErrInvalidDependencies) {
			t.Fatalf("invalid route prefix %q: %v", prefix, err)
		}
	}
}
func TestModuleAuthorizationFailsClosedBeforeStorage(t *testing.T) {
	for _, test := range []struct {
		err    error
		status int
		code   string
	}{
		{geojson.ErrUnauthorized, 401, "UNAUTHORIZED"},
		{fmt.Errorf("wrapped: %w", statusFailure{403}), 403, "FORBIDDEN"},
		{statusFailure{500}, 503, "AUTHORIZATION_UNAVAILABLE"},
		{errors.New("private provider credentials"), 503, "AUTHORIZATION_UNAVAILABLE"},
	} {
		t.Run(test.code+fmt.Sprint(test.status), func(t *testing.T) {
			assets := &httpAssetReader{}
			module, err := geojson.NewModule(geojson.Dependencies{
				Assets: assets,
				AuthorizeAny: func(_ *http.Request, permissions ...string) error {
					if !reflect.DeepEqual(permissions, []string{"surveillance:read", "data_quality:read", "report_browser:read"}) {
						t.Fatal("policy changed", permissions)
					}
					return test.err
				},
				RequestID: func(*http.Request) string { return "trusted-trace" },
			})
			if err != nil {
				t.Fatal(err)
			}
			response := serveHTTP(module, "/geojson/geo_water", map[string]string{"Range": "bytes=0-1", "If-None-Match": "*"})
			if response.Code != test.status || assets.calls != 0 || strings.Contains(response.Body.String(), "private") {
				t.Fatal(response.Code, response.Body.String(), assets.calls)
			}
			var payload struct {
				Success bool
				Error   struct{ Code string }
				Meta    struct{ RequestID string }
			}
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil || payload.Success || payload.Error.Code != test.code || payload.Meta.RequestID != "trusted-trace" {
				t.Fatal(payload, err)
			}
		})
	}
}
func TestModuleOwnsStreamingRangesConditionalReadsAndClosure(t *testing.T) {
	assets := &httpAssetReader{}
	authorizations := 0
	module, err := geojson.NewModule(geojson.Dependencies{Assets: assets, AuthorizeAny: func(r *http.Request, permissions ...string) error {
		authorizations++
		// A provider cannot change future request policy by mutating this slice.
		if permissions[0] != "surveillance:read" {
			t.Fatal("shared mutable policy")
		}
		permissions[0] = "mutated"
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	full := serveHTTP(module, "/geojson/geo_water", nil)
	if full.Code != 200 || full.Body.String() != "abcdef" || full.Header().Get("Content-Type") != "application/geo+json" || full.Header().Get("Cache-Control") != "public, max-age=86400" || full.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal(full.Code, full.Header(), full.Body.String())
	}
	partial := serveHTTP(module, "/geojson/geo_water", map[string]string{"Range": "bytes=1-3"})
	if partial.Code != 206 || partial.Body.String() != "bcd" || partial.Header().Get("Content-Range") != "bytes 1-3/6" {
		t.Fatal(partial.Code, partial.Header(), partial.Body.String())
	}
	unchanged := serveHTTP(module, "/geojson/geo_water", map[string]string{"If-Modified-Since": full.Header().Get("Last-Modified")})
	if unchanged.Code != 304 || unchanged.Body.Len() != 0 || authorizations != 3 {
		t.Fatal("conditional request bypassed access or body contract", unchanged.Code, authorizations)
	}
	for _, stream := range assets.streams {
		if stream.closed != 1 {
			t.Fatal("request leaked or double-closed stream", stream.closed)
		}
	}
	unknown := serveHTTP(module, "/geojson/not-a-layer", nil)
	if unknown.Code != 404 || assets.calls != 3 {
		t.Fatal("unknown layer reached storage", unknown.Code, assets.calls)
	}
	assets.err = geojson.ErrUnknownLayer
	failed := serveHTTP(module, "/geojson/geo_water", nil)
	if failed.Code != 500 || !strings.Contains(failed.Body.String(), "GEOJSON_LOAD_FAILED") || assets.streams[len(assets.streams)-1].closed != 1 {
		t.Fatal("reader failure misclassified or leaked", failed.Code, failed.Body.String())
	}
}
func TestModuleHEADClosesStreamWithoutWritingBody(t *testing.T) {
	assets := &httpAssetReader{}
	authorized := false
	module, err := geojson.NewModule(geojson.Dependencies{Assets: assets, AuthorizeAny: func(*http.Request, ...string) error { authorized = true; return nil }})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	module.ServeHTTP(response, httptest.NewRequest("HEAD", "/geojson/geo_water", nil))
	if response.Code != 200 || response.Body.Len() != 0 || response.Header().Get("Content-Length") != "6" || !authorized || len(assets.streams) != 1 || assets.streams[0].closed != 1 {
		t.Fatal("HEAD lost authorization, headers or stream ownership", response.Code, response.Header(), response.Body.String())
	}
}

func TestRegisteredModuleServesVersionedAssetsWithoutHostFiles(t *testing.T) {
	mux := http.NewServeMux()
	module, err := geojson.Register(mux, geojson.Dependencies{AuthorizeAny: allowed, BasePath: "/api/v1/geojson"})
	if err != nil {
		t.Fatal(err)
	}
	if module.Pattern() != "GET /api/v1/geojson/{name}" {
		t.Fatal(module.Pattern())
	}
	for _, layer := range []string{"geo_water", "geo_districts", "geo_subcounties"} {
		response := serveHTTP(mux, "/api/v1/geojson/"+layer, nil)
		if response.Code != 200 {
			t.Fatal(layer, response.Code, response.Body.String())
		}
		var collection struct {
			Type     string
			Features []json.RawMessage
		}
		if err := json.Unmarshal(response.Body.Bytes(), &collection); err != nil || collection.Type != "FeatureCollection" || len(collection.Features) == 0 {
			t.Fatal("bundled map invalid", layer, err)
		}
		sum := sha256.Sum256(response.Body.Bytes())
		expected := string(rune(34)) + hex.EncodeToString(sum[:]) + string(rune(34))
		if response.Header().Get("ETag") != expected {
			t.Fatal("embedded asset identity drift", layer, response.Header().Get("ETag"), expected)
		}
		unchanged := serveHTTP(mux, "/api/v1/geojson/"+layer, map[string]string{"If-None-Match": expected})
		if unchanged.Code != 304 || unchanged.Body.Len() != 0 {
			t.Fatal("embedded conditional read failed", unchanged.Code)
		}
	}
	denied, err := geojson.NewModule(geojson.Dependencies{AuthorizeAny: func(*http.Request, ...string) error { return statusFailure{403} }})
	if err != nil {
		t.Fatal(err)
	}
	if got := serveHTTP(denied, "/geojson/geo_water", map[string]string{"If-None-Match": "*"}); got.Code != 403 {
		t.Fatal("embedded conditional request bypassed authorization", got.Code)
	}
}
func TestPackageDirectoryOverrideAndPathBoundary(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "uga_water.geojson"), []byte("override"), 0600); err != nil {
		t.Fatal(err)
	}
	reader := geojson.NewDirectoryReader(directory)
	for _, filename := range []string{"", "..", "../secret", "/secret", "nested/map.geojson"} {
		if _, err := reader.Open(context.Background(), filename); err == nil {
			t.Fatal("unsafe filename accepted", filename)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := reader.Open(ctx, "uga_water.geojson"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	module, err := geojson.NewModule(geojson.Dependencies{AuthorizeAny: allowed, Directory: directory})
	if err != nil {
		t.Fatal(err)
	}
	response := serveHTTP(module, "/geojson/geo_water", nil)
	if response.Code != 200 || response.Body.String() != "override" {
		t.Fatal("directory override unused", response.Code, response.Body.String())
	}
	asset, err := geojson.BundledAssets().Open(context.Background(), "uga_water.geojson")
	if err != nil {
		t.Fatal(err)
	}
	defer asset.Content.Close()
	if _, err := io.Copy(io.Discard, asset.Content); err != nil {
		t.Fatal(err)
	}
}

func TestModuleOwnsDetachedAPIContract(t *testing.T) {
	module, err := geojson.NewModule(geojson.Dependencies{AuthorizeAny: allowed})
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Paths map[string]map[string]struct {
			Responses   map[string]any
			Permissions []string `json:"x-route-any-permissions"`
		}
	}
	if err := json.Unmarshal(module.OpenAPI(), &doc); err != nil {
		t.Fatal(err)
	}
	operation := doc.Paths["/geojson/{name}"]["get"]
	if !reflect.DeepEqual(operation.Permissions, geojson.RequiredPermissions()) {
		t.Fatal("documentation and policy differ")
	}
	for _, status := range []string{"200", "206", "304", "401", "403", "404", "416", "429", "500", "503"} {
		if operation.Responses[status] == nil {
			t.Fatal("response undocumented", status)
		}
	}
	first := module.OpenAPI()
	first[0] = '!'
	if !json.Valid(module.OpenAPI()) {
		t.Fatal("mutable contract escaped")
	}
}
