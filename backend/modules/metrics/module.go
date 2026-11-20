package metrics

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mohuganda/moh-sso-dashboard/backend/modules/metrics/internal/http/response"
)

const ID = "metrics"
const ReadPermission = "metrics:read"

var ErrInvalidDependencies = errors.New("invalid Metrics dependencies")

type accessFailure int

func (e accessFailure) Error() string   { return http.StatusText(int(e)) }
func (e accessFailure) StatusCode() int { return int(e) }

var ErrUnauthorized error = accessFailure(401)
var ErrForbidden error = accessFailure(403)

// Dependencies are shared platform capabilities, never host-specific module DTOs.
// Both authorization callbacks must inspect verified platform identity.
type Dependencies struct {
	Database     *sql.DB
	Repository   Repository
	RequireAdmin func(*http.Request) error
	AuthorizeAny func(*http.Request, ...string) error
	RequestID    func(*http.Request) string
	BasePath     string
}
type Registrar interface{ Handle(string, http.Handler) }
type Module struct {
	engine   *gin.Engine
	patterns []string
	base     string
}

func NewModule(deps Dependencies) (*Module, error) {
	if deps.RequireAdmin == nil || deps.AuthorizeAny == nil {
		return nil, fmt.Errorf("%w: admin and permission capabilities are required", ErrInvalidDependencies)
	}
	if (deps.Database == nil) == (deps.Repository == nil) {
		return nil, fmt.Errorf("%w: provide exactly one database or repository", ErrInvalidDependencies)
	}
	base := deps.BasePath
	if base == "" {
		base = "/admin/metrics"
	}
	if !validBase(base) {
		return nil, fmt.Errorf("%w: invalid base path", ErrInvalidDependencies)
	}
	repository := deps.Repository
	if repository == nil {
		var err error
		repository, err = NewRepository(deps.Database)
		if err != nil {
			return nil, err
		}
	}
	engine := gin.New()
	engine.RedirectTrailingSlash = false
	engine.RedirectFixedPath = false
	group := engine.Group(base, func(c *gin.Context) {
		if deps.RequestID != nil {
			c.Set("request_id", deps.RequestID(c.Request))
		}
		if err := deps.RequireAdmin(c.Request); err != nil {
			reject(c, err)
			return
		}
		if err := deps.AuthorizeAny(c.Request, ReadPermission); err != nil {
			reject(c, err)
			return
		}
		c.Next()
	})
	registerRoutes(group, NewHandler(NewService(repository)))
	m := &Module{engine: engine, base: base}
	parameter := regexp.MustCompile(":([A-Za-z_][A-Za-z0-9_]*)")
	for _, route := range engine.Routes() {
		m.patterns = append(m.patterns, route.Method+" "+parameter.ReplaceAllString(route.Path, "{$1}"))
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
func (m *Module) ServeHTTP(w http.ResponseWriter, r *http.Request) { m.engine.ServeHTTP(w, r) }
func reject(c *gin.Context, err error) {
	status := 503
	var coded interface{ StatusCode() int }
	if errors.As(err, &coded) && (coded.StatusCode() == 401 || coded.StatusCode() == 403) {
		status = coded.StatusCode()
	}
	code, message := "AUTHORIZATION_UNAVAILABLE", "authorization unavailable"
	if status == 401 {
		code, message = "UNAUTHORIZED", "unauthorized"
	}
	if status == 403 {
		code, message = "FORBIDDEN", "administrator access and metrics read permission are required"
	}
	response.Fail(c, status, code, message)
	c.Abort()
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
