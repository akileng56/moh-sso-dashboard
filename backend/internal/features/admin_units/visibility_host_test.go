package admin_units

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	healthcontext "github.com/moh-sso-dashboard/internal/features/health_context"
)

type routeService struct{ Service }

func (routeService) ListOrgUnits(context.Context) ([]OrgUnitFull, error) {
	return []OrgUnitFull{
		{DimOrgHierarchyKey: 1, OrgUnitID: textPtr("allowed"), OrgUnitName: textPtr("Allowed"), Level: "1"},
		{DimOrgHierarchyKey: 2, OrgUnitID: textPtr("outside"), OrgUnitName: textPtr("Outside"), Level: "1"},
	}, nil
}
func (routeService) ListFacilities(context.Context) ([]Facility, error) {
	return []Facility{{DimOrgHierarchyKey: 1, RegionUID: textPtr("allowed")}, {DimOrgHierarchyKey: 2, FacilityUID: textPtr("outside")}}, nil
}
func (routeService) ListDistricts(context.Context) ([]OrgUnitSimple, error) { return simpleRows(), nil }
func (routeService) ListRegions(context.Context) ([]OrgUnitSimple, error)   { return simpleRows(), nil }
func (routeService) ListNational(context.Context) ([]OrgUnitSimple, error)  { return simpleRows(), nil }
func (routeService) ListSubCounties(context.Context, *string) ([]OrgUnitSimple, error) {
	return simpleRows(), nil
}
func (routeService) ListLocalGovt(context.Context, *string) ([]OrgUnitSimple, error) {
	return simpleRows(), nil
}
func (routeService) ListDistrictsByRegion(context.Context, *string) ([]OrgUnitSimple, error) {
	return simpleRows(), nil
}
func simpleRows() []OrgUnitSimple {
	return []OrgUnitSimple{{DimOrgHierarchyKey: 1, OrgUnitID: textPtr("allowed")}, {DimOrgHierarchyKey: 2, OrgUnitID: textPtr("outside")}}
}

func TestAllAdminUnitRoutesPreserveVisibilityAndErrorContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, scenario := range []string{"unselected", "scoped", "national-descendants", "invalid-type", "nil-id", "missing-port", "unmapped", "forbidden", "lookup-error"} {
		t.Run(scenario, func(t *testing.T) {
			reader := &visibilityFixture{code: "district", aliases: []VisibilityAlias{{Namespace: "dwh-org-unit", ExternalID: "allowed"}}}
			var port VisibilityReader = reader
			if scenario == "national-descendants" {
				reader.code = "UG"
			}
			if scenario == "unmapped" {
				reader.aliases = nil
			}
			if scenario == "forbidden" {
				reader.nodeErr = ErrScopeForbidden
			}
			if scenario == "lookup-error" {
				reader.nodeErr = errors.New("warehouse mapping unavailable")
			}
			if scenario == "missing-port" {
				port = nil
			}
			engine := gin.New()
			group := engine.Group("/visualiser", func(c *gin.Context) {
				if scenario == "unselected" {
					return
				}
				var id any = uuid.New()
				if scenario == "invalid-type" {
					id = uuid.NewString()
				}
				if scenario == "nil-id" {
					id = uuid.Nil
				}
				c.Set("health_context_id", id)
				c.Set("health_context_scope_mode", healthcontext.ScopeNodeOnly)
				if scenario == "national-descendants" {
					c.Set("health_context_scope_mode", healthcontext.ScopeNodeAndDescendants)
				}
			})
			RegisterVisualiserRoutes(group, NewHandlerWithPorts(routeService{}, port))
			for _, endpoint := range []struct{ method, path string }{
				{"GET", "orgunits"}, {"GET", "facilities"}, {"GET", "district"}, {"POST", "subcounties"},
				{"POST", "localgovt"}, {"POST", "districts"}, {"GET", "region"}, {"GET", "national"}, {"GET", "hierarchy"},
			} {
				res := httptest.NewRecorder()
				req := httptest.NewRequest(endpoint.method, "/visualiser/adminunits/"+endpoint.path, strings.NewReader("{}"))
				req.Header.Set("Content-Type", "application/json")
				engine.ServeHTTP(res, req)
				success := scenario == "unselected" || scenario == "scoped" || scenario == "national-descendants"
				if success {
					if res.Code != 200 {
						t.Fatalf("%s: %d %s", endpoint.path, res.Code, res.Body.String())
					}
					var body struct{ Data []map[string]any }
					if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
						t.Fatal(err)
					}
					want := 2
					if scenario == "scoped" {
						want = 1
					}
					if len(body.Data) != want {
						t.Fatalf("%s: expected %d rows: %s", endpoint.path, want, res.Body.String())
					}
					if scenario == "scoped" {
						key := "dim_org_hierarchy_key"
						if endpoint.path == "hierarchy" {
							key = "id"
						}
						if body.Data[0][key] != float64(1) {
							t.Fatal("outside row leaked")
						}
					}
				} else {
					code := "HEALTH_CONTEXT_MAPPING_REQUIRED"
					if scenario == "invalid-type" || scenario == "nil-id" || scenario == "missing-port" || scenario == "forbidden" {
						code = "HEALTH_CONTEXT_FORBIDDEN"
					}
					if res.Code != 403 || !strings.Contains(res.Body.String(), code) || strings.Contains(res.Body.String(), "\"data\"") {
						t.Fatalf("%s: %d %s", endpoint.path, res.Code, res.Body.String())
					}
				}
			}
			if (scenario == "unselected" || scenario == "invalid-type" || scenario == "nil-id") && reader.ctx != nil {
				t.Fatal("unselected/invalid selection invoked provider")
			}
		})
	}
}

type healthRepository struct {
	healthcontext.Repository
	ctx         context.Context
	id          uuid.UUID
	descendants bool
	node        healthcontext.Node
	aliases     []healthcontext.Alias
	err         error
}

func (r *healthRepository) GetNode(ctx context.Context, id uuid.UUID) (healthcontext.Node, error) {
	r.ctx, r.id = ctx, id
	return r.node, r.err
}
func (r *healthRepository) ListAliasesInScope(ctx context.Context, id uuid.UUID, descendants bool) ([]healthcontext.Alias, error) {
	r.ctx, r.id, r.descendants = ctx, id, descendants
	return r.aliases, r.err
}
func TestHealthVisibilityAdapterMapsOnlyRequiredValuesAndPreservesCauses(t *testing.T) {
	if NewHealthContextVisibilityReader(nil) != nil {
		t.Fatal("typed nil adapter")
	}
	repo := &healthRepository{node: healthcontext.Node{Code: "UG"}, aliases: []healthcontext.Alias{{Namespace: "dwh-org-unit", ExternalID: " raw-id "}}}
	reader := NewHealthContextVisibilityReader(healthcontext.NewService(repo))
	ctx := context.WithValue(context.Background(), struct{}{}, "request")
	id := uuid.New()
	node, err := reader.GetNode(ctx, id.String())
	if err != nil || node.Code != "UG" || repo.id != id || repo.ctx != ctx {
		t.Fatal("node translation/context changed")
	}
	aliases, err := reader.ListAliasesInScope(ctx, id.String(), true)
	if err != nil || len(aliases) != 1 || aliases[0].ExternalID != " raw-id " || aliases[0].Namespace != "dwh-org-unit" || !repo.descendants || repo.ctx != ctx {
		t.Fatal("alias translation changed")
	}
	repo.err = healthcontext.ErrContextForbidden
	if _, err := reader.GetNode(ctx, id.String()); !errors.Is(err, ErrScopeForbidden) || !errors.Is(err, healthcontext.ErrContextForbidden) {
		t.Fatalf("forbidden classification/cause lost: %v", err)
	}
	repo.err = errors.New("database unavailable")
	if _, err := reader.ListAliasesInScope(ctx, id.String(), false); !errors.Is(err, repo.err) {
		t.Fatal("provider cause lost")
	}
	if _, err := reader.GetNode(ctx, "invalid"); !errors.Is(err, ErrScopeForbidden) {
		t.Fatal("invalid adapter id accepted")
	}
}
