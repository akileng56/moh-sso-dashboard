// The verifier copies this program into an independent consumer module.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	units "github.com/mohuganda/moh-sso-dashboard/backend/modules/admin-units"
)

func ptr(s string) *string { return &s }

type repository struct {
	units.Repository
	ctx    context.Context
	filter *string
	rows   []units.OrgUnit
	err    error
}

func (r *repository) ListOrgUnits(ctx context.Context) ([]units.OrgUnitFull, error) {
	r.ctx = ctx
	return []units.OrgUnitFull{{OrgUnitID: ptr("unit"), RowVersion: 7, IsCurrent: true}}, r.err
}
func (r *repository) ListFacilities(ctx context.Context) ([]units.Facility, error) {
	r.ctx = ctx
	return []units.Facility{{FacilityUID: ptr("f1"), DistrictUID: ptr(" d1 ")}}, r.err
}
func (r *repository) simple(ctx context.Context, filter *string) ([]units.OrgUnitSimple, error) {
	r.ctx, r.filter = ctx, filter
	return []units.OrgUnitSimple{{OrgUnitID: ptr("d1")}, {OrgUnitID: nil}, {OrgUnitID: ptr("other")}}, r.err
}
func (r *repository) ListDistricts(ctx context.Context) ([]units.OrgUnitSimple, error) {
	return r.simple(ctx, nil)
}
func (r *repository) ListRegions(ctx context.Context) ([]units.OrgUnitSimple, error) {
	return r.simple(ctx, nil)
}
func (r *repository) ListNational(ctx context.Context) ([]units.OrgUnitSimple, error) {
	return r.simple(ctx, nil)
}
func (r *repository) ListSubCounties(ctx context.Context, f *string) ([]units.OrgUnitSimple, error) {
	return r.simple(ctx, f)
}
func (r *repository) ListLocalGovt(ctx context.Context, f *string) ([]units.OrgUnitSimple, error) {
	return r.simple(ctx, f)
}
func (r *repository) ListDistrictsByRegion(ctx context.Context, f *string) ([]units.OrgUnitSimple, error) {
	return r.simple(ctx, f)
}
func (r *repository) ListHierarchy(ctx context.Context) ([]units.OrgUnit, error) {
	r.ctx = ctx
	return r.rows, r.err
}

type visibility struct {
	ctx         context.Context
	id          string
	descendants bool
	code        string
	calls       int
	err         error
}

func (v *visibility) GetNode(ctx context.Context, id string) (units.VisibilityNode, error) {
	v.ctx, v.id = ctx, id
	return units.VisibilityNode{Code: v.code}, v.err
}
func (v *visibility) ListAliasesInScope(ctx context.Context, id string, descendants bool) ([]units.VisibilityAlias, error) {
	v.ctx, v.id, v.descendants = ctx, id, descendants
	v.calls++
	return []units.VisibilityAlias{{Namespace: "dwh-org-unit", ExternalID: " d1 "}, {Namespace: "dwh-org-unit", ExternalID: "d1"}, {Namespace: "other", ExternalID: "other"}}, v.err
}

func main() {
	type requestKey struct{}
	ctx := context.WithValue(context.Background(), requestKey{}, "external-request")
	country, region, district, sub, facility := "ug", "r1", "d1", "s1", "f1"
	row := func(id int64, level string) units.OrgUnit {
		return units.OrgUnit{DimOrgHierarchyKey: id, Level: level, CountryUID: &country, RegionUID: &region, Region: ptr("Central"), DistrictUID: &district, District: ptr("District"), SubCountyUID: &sub, SubCounty: ptr("Subcounty"), FacilityUID: &facility, FacilityName: ptr("Hospital")}
	}
	rows := []units.OrgUnit{row(1, "1"), row(2, "2"), row(3, "3"), row(5, "5"), row(6, "6"), row(6, "6")}
	repo := &repository{rows: rows}
	service := units.NewService(repo)
	full, err := service.ListOrgUnits(ctx)
	if err != nil || len(full) != 1 || full[0].RowVersion != 7 || !full[0].IsCurrent {
		panic("full model contract changed")
	}
	facilities, err := service.ListFacilities(ctx)
	if err != nil {
		panic(err)
	}
	simple, err := service.ListDistricts(ctx)
	if err != nil {
		panic(err)
	}
	for _, list := range []func(context.Context) ([]units.OrgUnitSimple, error){service.ListRegions, service.ListNational} {
		if _, err := list(ctx); err != nil {
			panic(err)
		}
	}
	blank := ""
	for _, list := range []func(context.Context, *string) ([]units.OrgUnitSimple, error){service.ListSubCounties, service.ListLocalGovt, service.ListDistrictsByRegion} {
		for _, filter := range []*string{nil, &blank, ptr("optional-filter")} {
			if _, err := list(ctx, filter); err != nil || repo.ctx != ctx || repo.filter != filter {
				panic("nullable filter/context contract changed")
			}
		}
	}
	tree, err := service.GetHierarchy(ctx)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(tree, units.BuildHierarchyTree(rows)) || len(tree) != 1 || tree[0].Name == nil || *tree[0].Name != "National" {
		panic("public hierarchy contract changed")
	}
	node := tree[0]
	for _, kind := range []string{"region", "district", "subcounty", "facility"} {
		if len(node.Children) != 1 || node.Children[0].Type != kind {
			panic("hierarchy or deduplication changed")
		}
		node = node.Children[0]
	}
	encoded, err := json.Marshal(node)
	if err != nil {
		panic(err)
	}
	var leaf map[string]any
	if err := json.Unmarshal(encoded, &leaf); err != nil {
		panic(err)
	}
	if children, exists := leaf["children"]; !exists || children != nil {
		panic("leaf nullability changed")
	}

	reader := &visibility{code: "district"}
	allowed, scoped, err := units.ResolveVisibility(ctx, reader, "selected-context", true)
	if err != nil || !scoped || len(allowed) != 1 || reader.ctx != ctx || reader.id != "selected-context" || !reader.descendants {
		panic("visibility contract changed")
	}
	if len(units.FilterFacilities(allowed, facilities)) != 1 || len(units.FilterSimpleOrgUnits(allowed, simple)) != 1 {
		panic("visibility filtering changed")
	}
	if len(units.FilterOrgUnits(allowed, []units.OrgUnitFull{{DistrictUID: &district}, {DistrictUID: ptr("outside")}})) != 1 {
		panic("ancestor matching changed")
	}
	reader.code = "UG"
	before := reader.calls
	if _, restricted, err := units.ResolveVisibility(ctx, reader, "national-context", true); err != nil || restricted || reader.calls != before {
		panic("national exception changed")
	}
	if _, restricted, err := units.ResolveVisibility(ctx, nil, "selected-context", false); !errors.Is(err, units.ErrScopeForbidden) || !restricted {
		panic("missing visibility provider became unrestricted")
	}
	cause := errors.New("warehouse unavailable")
	repo.err = cause
	if _, err := service.GetHierarchy(ctx); !errors.Is(err, cause) {
		panic("repository error cause lost")
	}
	reader.err = cause
	if _, restricted, err := units.ResolveVisibility(ctx, reader, "selected-context", false); !errors.Is(err, cause) || !restricted {
		panic("scope error became unrestricted")
	}
	if units.ModulePath != "github.com/mohuganda/moh-sso-dashboard/backend/modules/admin-units" || units.Version() != "dev" {
		panic("module identity changed")
	}
	fmt.Println("PASS: separate Administrative Units consumer (dev), lookup contracts, hierarchy, visibility, nullable values and errors")
}
