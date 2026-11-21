package geojson

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
)

const ID = "geojson"

// RequiredPermissions is detached and case-sensitive. A trusted platform
// authorizer must grant at least one after verifying the caller's credentials.
func RequiredPermissions() []string {
	return []string{"surveillance:read", "data_quality:read", "report_browser:read"}
}

// Authorization failures from any platform can implement StatusCode() int;
// no module-specific error conversion or host dependency is required.
type accessFailure struct{ status int }

func (e accessFailure) Error() string   { return http.StatusText(e.status) }
func (e accessFailure) StatusCode() int { return e.status }

var ErrUnauthorized error = accessFailure{http.StatusUnauthorized}
var ErrForbidden error = accessFailure{http.StatusForbidden}
var ErrInvalidDependencies = errors.New("invalid GeoJSON module dependencies")

type Dependencies struct {
	// AuthorizeAny is mandatory and must authenticate before checking any of the
	// supplied permissions. A nil function is rejected; unknown errors fail closed.
	AuthorizeAny func(*http.Request, ...string) error
	// RequestID supplies the trusted platform trace ID for the existing envelope.
	RequestID func(*http.Request) string
	// Assets and Directory are mutually exclusive overrides; neither is required.
	// With both absent, the module serves its embedded, versioned map files.
	Assets    AssetReader
	Directory string
	// BasePath defaults to /geojson. A host may select /api/v1/geojson directly.
	BasePath string
}

// Registrar is a shared structural HTTP contract, satisfied by http.ServeMux.
// A framework host can implement it once for all modules.
type Registrar interface{ Handle(string, http.Handler) }

type Module struct {
	mux       *http.ServeMux
	pattern   string
	service   *Service
	authorize func(*http.Request, ...string) error
	requestID func(*http.Request) string
	bundled   bool
}

func NewModule(deps Dependencies) (*Module, error) {
	if deps.AuthorizeAny == nil {
		return nil, fmt.Errorf("%w: authorizer is required", ErrInvalidDependencies)
	}
	if deps.Assets != nil && deps.Directory != "" {
		return nil, fmt.Errorf("%w: provide assets or directory, not both", ErrInvalidDependencies)
	}
	base := deps.BasePath
	if base == "" {
		base = "/geojson"
	}
	if !validBasePath(base) {
		return nil, fmt.Errorf("%w: base path must be a clean absolute route prefix", ErrInvalidDependencies)
	}
	bundled := deps.Assets == nil && deps.Directory == ""

	assets := deps.Assets
	if assets == nil {
		if deps.Directory != "" {
			assets = NewDirectoryReader(deps.Directory)
		} else {
			assets = BundledAssets()
		}
	}
	m := &Module{
		mux: http.NewServeMux(), pattern: "GET " + base + "/{name}", service: NewService(assets),
		authorize: deps.AuthorizeAny, requestID: deps.RequestID, bundled: bundled,
	}
	m.mux.HandleFunc(m.pattern, m.serveLayer)
	return m, nil
}

// Register prepares the entire module before exposing its endpoint. It owns
// no database schema or workers. Registry enablement belongs to composition.
func Register(routes Registrar, deps Dependencies) (*Module, error) {
	if routes == nil {
		return nil, fmt.Errorf("%w: route registrar is required", ErrInvalidDependencies)
	}
	module, err := NewModule(deps)
	if err != nil {
		return nil, err
	}
	routes.Handle(module.pattern, module)
	return module, nil
}
func (m *Module) Pattern() string                                  { return m.pattern }
func (m *Module) ServeHTTP(w http.ResponseWriter, r *http.Request) { m.mux.ServeHTTP(w, r) }
func (m *Module) serveLayer(w http.ResponseWriter, r *http.Request) {
	if err := m.authorize(r, RequiredPermissions()...); err != nil {
		var denied interface{ StatusCode() int }
		if errors.As(err, &denied) {
			switch denied.StatusCode() {
			case http.StatusUnauthorized:
				m.fail(w, r, 401, "UNAUTHORIZED", "unauthorized")
				return
			case http.StatusForbidden:
				m.fail(w, r, 403, "FORBIDDEN", "one of these permissions is required: "+strings.Join(RequiredPermissions(), ", "))
				return
			}
		}
		m.fail(w, r, 503, "AUTHORIZATION_UNAVAILABLE", "authorization unavailable")
		return
	}
	layer, err := m.service.Open(r.Context(), r.PathValue("name"))
	if err != nil {
		if errors.Is(err, ErrUnknownLayer) && !errors.Is(err, ErrLoadFailed) {
			m.fail(w, r, 404, "NOT_FOUND", "requested geojson does not exist")
		} else {
			m.fail(w, r, 500, "GEOJSON_LOAD_FAILED", "failed to load geojson file")
		}
		return
	}
	defer layer.Content.Close()
	if m.bundled {
		w.Header().Set("ETag", bundledETags[layer.Filename])
	}
	w.Header().Set("Content-Type", "application/geo+json")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, layer.Filename, layer.ModifiedAt, layer.Content)
}
func (m *Module) fail(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	requestID := ""
	if m.requestID != nil {
		requestID = m.requestID(r)
	}
	body := struct {
		Success bool `json:"success"`
		Error   struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		Meta struct {
			RequestID string `json:"requestId,omitempty"`
		} `json:"meta"`
	}{}
	body.Error.Code, body.Error.Message, body.Meta.RequestID = code, message, requestID
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func validBasePath(base string) bool {
	if !strings.HasPrefix(base, "/") || base == "/" || path.Clean(base) != base || strings.Contains(base, "//") {
		return false
	}
	for _, r := range base {
		if r == '/' || r == '_' || r == '-' || r == '.' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}
