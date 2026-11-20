package admin_units

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
)

const ID = "admin-units"
const RequiredSystem = "data-statistics"

var ErrInvalidDependencies = errors.New("invalid Administrative Units module dependencies")

type accessFailure int

func (e accessFailure) Error() string   { return http.StatusText(int(e)) }
func (e accessFailure) StatusCode() int { return int(e) }

var ErrUnauthorized error = accessFailure(401)
var ErrForbidden error = accessFailure(403)

func RequiredPermissions() []string {
	return []string{"report_browser:read", "documents:read", "data_quality:read", "surveillance:read"}
}

type Dependencies struct {
	Database   *sql.DB
	Repository Repository
	// The platform authenticates the caller before granting system/permission
	// access. Each error can implement StatusCode, ErrorCode and PublicMessage.
	RequireSystem func(*http.Request, string) error
	AuthorizeAny  func(*http.Request, ...string) error
	// SelectContext must verify membership, expiry and selection requirements.
	// Empty ID with no error is an explicit platform-approved legacy unscoped request.
	SelectContext func(*http.Request) (id string, descendants bool, err error)
	// Primitive shared health-context ports avoid module-specific host mapping.
	ContextNodeCode func(context.Context, string) (string, error)
	ContextAliases  func(context.Context, string, bool) (map[string][]string, error)
	RequestID       func(*http.Request) string
	BasePath        string
	MaxBodyBytes    int64
}
type Registrar interface{ Handle(string, http.Handler) }
type Module struct {
	mux      *http.ServeMux
	service  Service
	deps     Dependencies
	patterns []string
}
type endpoint struct{ method, suffix, code, message string }

var endpoints = []endpoint{
	{"GET", "orgunits", "LIST_ORG_UNITS_FAILED", "failed to list organizational units"},
	{"GET", "facilities", "LIST_FACILITIES_FAILED", "failed to list facilities"},
	{"GET", "district", "LIST_DISTRICTS_FAILED", "failed to list districts"},
	{"POST", "subcounties", "LIST_SUBCOUNTIES_FAILED", "failed to list subcounties"},
	{"POST", "localgovt", "LIST_LOCAL_GOVT_FAILED", "failed to list local governments"},
	{"POST", "districts", "LIST_REGIONS_FAILED", "failed to list regions"},
	{"GET", "region", "LIST_NATIONAL_FAILED", "failed to list national hierarchy"},
	{"GET", "national", "LIST_HIERARCHY_FAILED", "failed to list hierarchy"},
	{"GET", "hierarchy", "LIST_HIERARCHY_FAILED", "failed to list hierarchy"},
}

func NewModule(deps Dependencies) (*Module, error) {
	if deps.RequireSystem == nil || deps.AuthorizeAny == nil || deps.SelectContext == nil || deps.ContextNodeCode == nil || deps.ContextAliases == nil {
		return nil, fmt.Errorf("%w: platform security and context capabilities are required", ErrInvalidDependencies)
	}
	if (deps.Database == nil) == (deps.Repository == nil) {
		return nil, fmt.Errorf("%w: provide exactly one warehouse database or repository", ErrInvalidDependencies)
	}
	base := deps.BasePath
	if base == "" {
		base = "/visualizer/adminunits"
	}
	if !validBasePath(base) {
		return nil, fmt.Errorf("%w: invalid base path", ErrInvalidDependencies)
	}
	if deps.MaxBodyBytes < 0 {
		return nil, fmt.Errorf("%w: body limit must be positive", ErrInvalidDependencies)
	}
	if deps.MaxBodyBytes == 0 {
		deps.MaxBodyBytes = 1 << 20
	}
	repository := deps.Repository
	if repository == nil {
		repository = NewRepository(deps.Database)
	}
	m := &Module{mux: http.NewServeMux(), service: NewService(repository), deps: deps}
	for _, item := range endpoints {
		item := item
		pattern := item.method + " " + base + "/" + item.suffix
		m.patterns = append(m.patterns, pattern)
		m.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) { m.serve(w, r, item) })
	}
	return m, nil
}
func Register(routes Registrar, deps Dependencies) (*Module, error) {
	if routes == nil {
		return nil, fmt.Errorf("%w: registrar is required", ErrInvalidDependencies)
	}
	module, err := NewModule(deps)
	if err != nil {
		return nil, err
	}
	for _, pattern := range module.patterns {
		routes.Handle(pattern, module)
	}
	return module, nil
}
func (m *Module) Patterns() []string                               { return append([]string(nil), m.patterns...) }
func (m *Module) ServeHTTP(w http.ResponseWriter, r *http.Request) { m.mux.ServeHTTP(w, r) }

func (m *Module) serve(w http.ResponseWriter, r *http.Request, item endpoint) {
	// Preserve the existing group ordering: system, selected context, functional
	// permission. No warehouse request can run before these guards succeed.
	if err := m.deps.RequireSystem(r, RequiredSystem); err != nil {
		m.guardFailure(w, r, err, "FORBIDDEN", "system access is required")
		return
	}
	id, descendants, err := m.deps.SelectContext(r)
	if err != nil {
		m.guardFailure(w, r, err, "HEALTH_CONTEXT_UNAVAILABLE", "health context access could not be verified")
		return
	}
	if err := m.deps.AuthorizeAny(r, RequiredPermissions()...); err != nil {
		m.guardFailure(w, r, err, "FORBIDDEN", "one of these permissions is required: "+strings.Join(RequiredPermissions(), ", "))
		return
	}
	var filter struct {
		SubCounty *string `json:"sub_county"`
		District  *string `json:"district"`
		Region    *string `json:"region"`
	}
	if item.method == "POST" {
		if r.Body == nil {
			m.fail(w, r, 400, "INVALID_REQUEST", "request body must be a JSON object")
			return
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, m.deps.MaxBodyBytes))
		err := decoder.Decode(&filter)
		if err == nil {
			var extra any
			if next := decoder.Decode(&extra); next != io.EOF {
				if next == nil {
					err = errors.New("multiple JSON values")
				} else {
					err = next
				}
			}
		}
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				m.fail(w, r, 413, "REQUEST_TOO_LARGE", "request body is too large")
			} else {
				m.fail(w, r, 400, "INVALID_REQUEST", "request body must be a JSON object")
			}
			return
		}
	}
	var result any
	switch item.suffix {
	case "orgunits", "hierarchy":
		var rows []OrgUnitFull
		rows, err = m.service.ListOrgUnits(r.Context())
		if err == nil {
			var allowed map[string]struct{}
			var scoped bool
			allowed, scoped, err = m.visibility(r.Context(), id, descendants)
			if err != nil {
				m.scopeFailure(w, r, err)
				return
			}
			if scoped {
				rows = FilterOrgUnits(allowed, rows)
			}
			if item.suffix == "hierarchy" {
				units := make([]OrgUnit, 0, len(rows))
				for _, row := range rows {
					units = append(units, OrgUnit{
						DimOrgHierarchyKey: row.DimOrgHierarchyKey, OrgUnitID: row.OrgUnitID, OrgUnitName: row.OrgUnitName, Level: row.Level,
						CountryUID: row.CountryUID, RegionUID: row.RegionUID, Region: row.Region, DistrictUID: row.DistrictUID, District: row.District,
						SubCountyUID: row.SubCountyUID, SubCounty: row.SubCounty, DivisionUID: row.DivisionUID, Division: row.Division,
						FacilityUID: row.FacilityUID, FacilityName: row.FacilityName,
					})
				}
				result = nonNil(BuildHierarchyTree(units))
			} else {
				result = nonNil(rows)
			}
		}
	case "facilities":
		var rows []Facility
		rows, err = m.service.ListFacilities(r.Context())
		if err == nil {
			var allowed map[string]struct{}
			var scoped bool
			allowed, scoped, err = m.visibility(r.Context(), id, descendants)
			if err != nil {
				m.scopeFailure(w, r, err)
				return
			}
			if scoped {
				rows = FilterFacilities(allowed, rows)
			}
			result = nonNil(rows)
		}
	default:
		var rows []OrgUnitSimple
		switch item.suffix {
		case "district":
			rows, err = m.service.ListDistricts(r.Context())
		case "subcounties":
			rows, err = m.service.ListSubCounties(r.Context(), filter.SubCounty)
		case "localgovt":
			rows, err = m.service.ListLocalGovt(r.Context(), filter.District)
		case "districts":
			rows, err = m.service.ListDistrictsByRegion(r.Context(), filter.Region)
		case "region":
			rows, err = m.service.ListRegions(r.Context())
		case "national":
			rows, err = m.service.ListNational(r.Context())
		}
		if err == nil {
			var allowed map[string]struct{}
			var scoped bool
			allowed, scoped, err = m.visibility(r.Context(), id, descendants)
			if err != nil {
				m.scopeFailure(w, r, err)
				return
			}
			if scoped {
				rows = FilterSimpleOrgUnits(allowed, rows)
			}
			result = nonNil(rows)
		}
	}
	if err != nil {
		m.fail(w, r, 500, item.code, item.message)
		return
	}
	m.write(w, r, 200, map[string]any{"success": true, "data": result})
}
func nonNil[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}
func (m *Module) guardFailure(w http.ResponseWriter, r *http.Request, err error, code, message string) {
	status := 503
	var failure interface{ StatusCode() int }
	if errors.As(err, &failure) {
		if candidate := failure.StatusCode(); candidate >= 400 && candidate <= 599 {
			status = candidate
		}
	}
	if status == 401 {
		code, message = "UNAUTHORIZED", "unauthorized"
	}
	if status == 503 {
		code, message = "AUTHORIZATION_UNAVAILABLE", "authorization unavailable"
	}
	var public interface {
		ErrorCode() string
		PublicMessage() string
	}
	if errors.As(err, &public) {
		code, message = public.ErrorCode(), public.PublicMessage()
	}
	m.fail(w, r, status, code, message)
}
func (m *Module) scopeFailure(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrScopeForbidden) {
		m.fail(w, r, 403, "HEALTH_CONTEXT_FORBIDDEN", "health context access denied")
		return
	}
	m.fail(w, r, 403, "HEALTH_CONTEXT_MAPPING_REQUIRED", "selected health context is not mapped to the data hierarchy")
}
func (m *Module) fail(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	m.write(w, r, status, map[string]any{"success": false, "error": map[string]string{"code": code, "message": message}})
}
func (m *Module) write(w http.ResponseWriter, r *http.Request, status int, body map[string]any) {
	meta := map[string]string{}
	if m.deps.RequestID != nil {
		if id := m.deps.RequestID(r); id != "" {
			meta["requestId"] = id
		}
	}
	body["meta"] = meta
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if r.Method == http.MethodHead {
		return
	}
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

type contextReader struct{ deps Dependencies }

func (r contextReader) GetNode(ctx context.Context, id string) (VisibilityNode, error) {
	code, err := r.deps.ContextNodeCode(ctx, id)
	return VisibilityNode{Code: code}, scopeError(err)
}
func (r contextReader) ListAliasesInScope(ctx context.Context, id string, descendants bool) ([]VisibilityAlias, error) {
	aliases, err := r.deps.ContextAliases(ctx, id, descendants)
	if err != nil {
		return nil, scopeError(err)
	}
	var result []VisibilityAlias
	for namespace, ids := range aliases {
		for _, external := range ids {
			result = append(result, VisibilityAlias{Namespace: namespace, ExternalID: external})
		}
	}
	return result, nil
}
func scopeError(err error) error {
	if err == nil {
		return nil
	}
	var status interface{ StatusCode() int }
	if errors.As(err, &status) && (status.StatusCode() == 401 || status.StatusCode() == 403) {
		return errors.Join(ErrScopeForbidden, err)
	}
	return err
}
func (m *Module) visibility(ctx context.Context, id string, descendants bool) (map[string]struct{}, bool, error) {
	if id == "" {
		return nil, false, nil
	}
	return ResolveVisibility(ctx, contextReader{m.deps}, id, descendants)
}
