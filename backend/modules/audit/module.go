package audit

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

const ID = "audit"
const ReadPermission = "audit:read"
const ExportPolicy = "audit-export"
const ExportLimit = 10
const ExportWindow = time.Minute

var ErrInvalidDependencies = errors.New("invalid Audit dependencies")

type accessFailure int

func (e accessFailure) Error() string   { return http.StatusText(int(e)) }
func (e accessFailure) StatusCode() int { return int(e) }

var ErrUnauthorized error = accessFailure(401)
var ErrForbidden error = accessFailure(403)

// Dependencies contain verified platform capabilities and caller-owned resources.
type Dependencies struct {
	Database     *sql.DB
	Repository   Repository
	Cache        Cache
	RequireAdmin func(*http.Request) error
	AuthorizeAny func(*http.Request, ...string) error
	// LimitExport applies the named per-user policy using verified identity.
	// Returning nil explicitly grants the request. A missing capability is invalid.
	LimitExport func(*http.Request, string, int, time.Duration) (http.Header, error)
	RequestID   func(*http.Request) string
	StreamError func(*http.Request, error)
	BasePath    string
}
type Registrar interface{ Handle(string, http.Handler) }
type Module struct {
	mux      *http.ServeMux
	patterns []string
	base     string
}

func NewModule(deps Dependencies) (*Module, error) {
	if deps.RequireAdmin == nil || deps.AuthorizeAny == nil || deps.LimitExport == nil {
		return nil, fmt.Errorf("%w: admin, permission and export-limit capabilities are required", ErrInvalidDependencies)
	}
	if (deps.Database == nil) == (deps.Repository == nil) {
		return nil, fmt.Errorf("%w: provide exactly one database or repository", ErrInvalidDependencies)
	}
	base := deps.BasePath
	if base == "" {
		base = "/admin/audit-logs"
	}
	if !validBase(base) {
		return nil, fmt.Errorf("%w: invalid base path", ErrInvalidDependencies)
	}
	repo := deps.Repository
	if repo == nil {
		var err error
		repo, err = NewRepository(deps.Database)
		if err != nil {
			return nil, err
		}
	}
	handler := newHTTPHandler(NewService(repo), deps.Cache)
	m := &Module{mux: http.NewServeMux(), base: base}
	endpoints := []struct {
		suffix string
		serve  func(*httpRequest)
	}{
		{"", handler.ListAuditLogs}, {"/actions", handler.ListAuditActions},
		{"/metrics/overview", handler.AuditMetricsOverview},
		{"/metrics/failed-logins-by-day", handler.FailedLoginsByDay},
		{"/metrics/top-failure-ips", handler.TopFailureIPs},
		{"/export", handler.ExportAuditLogs}, {"/{id}", handler.GetAuditLog},
	}
	for _, item := range endpoints {
		pattern := "GET " + base + item.suffix
		m.patterns = append(m.patterns, pattern)
		m.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			c := &httpRequest{Request: r, Writer: w, streamError: deps.StreamError}
			if deps.RequestID != nil {
				c.requestID = deps.RequestID(r)
			}
			if r.Method != http.MethodGet {
				w.Header().Set("Allow", "GET")
				w.WriteHeader(405)
				return
			}
			if err := deps.RequireAdmin(r); err != nil {
				c.reject(err)
				return
			}
			if err := deps.AuthorizeAny(r, ReadPermission); err != nil {
				c.reject(err)
				return
			}
			if item.suffix == "/export" {
				headers, err := deps.LimitExport(r, ExportPolicy, ExportLimit, ExportWindow)
				for name, values := range headers {
					w.Header().Del(name)
					for _, value := range values {
						w.Header().Add(name, value)
					}
				}
				if err != nil {
					c.reject(err)
					return
				}
			}
			item.serve(c)
		})
	}
	return m, nil
}
func Register(registrar Registrar, deps Dependencies) (*Module, error) {
	if registrar == nil {
		return nil, fmt.Errorf("%w: registrar is required", ErrInvalidDependencies)
	}
	m, err := NewModule(deps)
	if err != nil {
		return nil, err
	}
	for _, pattern := range m.patterns {
		registrar.Handle(pattern, m)
	}
	return m, nil
}
func (m *Module) Patterns() []string                               { return append([]string(nil), m.patterns...) }
func (m *Module) ServeHTTP(w http.ResponseWriter, r *http.Request) { m.mux.ServeHTTP(w, r) }

type httpRequest struct {
	Request     *http.Request
	Writer      http.ResponseWriter
	requestID   string
	streamError func(*http.Request, error)
}

func (c *httpRequest) Query(key string) string { return c.Request.URL.Query().Get(key) }
func (c *httpRequest) DefaultQuery(key, fallback string) string {
	if value, ok := c.Request.URL.Query()[key]; ok && len(value) > 0 {
		return value[0]
	}
	return fallback
}
func (c *httpRequest) Param(key string) string  { return c.Request.PathValue(key) }
func (c *httpRequest) Header(key, value string) { c.Writer.Header().Set(key, value) }
func (c *httpRequest) ok(status int, data any) {
	c.write(status, map[string]any{"success": true, "data": data})
}
func (c *httpRequest) fail(status int, code, message string) {
	c.write(status, map[string]any{"success": false, "error": map[string]string{"code": code, "message": message}})
}
func (c *httpRequest) write(status int, body map[string]any) {
	meta := map[string]string{}
	if c.requestID != "" {
		meta["requestId"] = c.requestID
	}
	body["meta"] = meta
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(status)
	_ = json.NewEncoder(c.Writer).Encode(body)
}
func (c *httpRequest) reject(err error) {
	status := 503
	var failure interface{ StatusCode() int }
	if errors.As(err, &failure) {
		switch failure.StatusCode() {
		case 401, 403, 429:
			status = failure.StatusCode()
		}
	}
	code, message := "AUTHORIZATION_UNAVAILABLE", "authorization unavailable"
	switch status {
	case 401:
		code, message = "UNAUTHORIZED", "unauthorized"
	case 403:
		code, message = "FORBIDDEN", "administrator access and audit read permission are required"
	case 429:
		code, message = "RATE_LIMIT_EXCEEDED", "rate limit exceeded"
		var retry interface{ RetryAfter() time.Duration }
		if errors.As(err, &retry) && retry.RetryAfter() > 0 {
			c.Header("Retry-After", strconv.FormatInt(int64(math.Ceil(retry.RetryAfter().Seconds())), 10))
		}
	}
	c.fail(status, code, message)
}
func validBase(base string) bool {
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
