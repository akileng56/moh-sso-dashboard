package admin_units

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func textPtr(s string) *string { return &s }

type visibilityFixture struct {
	ctx               context.Context
	id                string
	descendants       bool
	aliasCalls        int
	code              string
	aliases           []VisibilityAlias
	nodeErr, aliasErr error
}

func (f *visibilityFixture) GetNode(ctx context.Context, id string) (VisibilityNode, error) {
	f.ctx, f.id = ctx, id
	return VisibilityNode{Code: f.code}, f.nodeErr
}
func (f *visibilityFixture) ListAliasesInScope(ctx context.Context, id string, descendants bool) ([]VisibilityAlias, error) {
	f.ctx, f.id, f.descendants = ctx, id, descendants
	f.aliasCalls++
	return f.aliases, f.aliasErr
}
func TestVisibilityResolutionPreservesScopeAndMapping(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, "request")
	for _, descendants := range []bool{false, true} {
		f := &visibilityFixture{code: "district", aliases: []VisibilityAlias{
			{Namespace: "dwh-org-unit", ExternalID: " d1 "},
			{Namespace: "dwh-org-unit", ExternalID: "d1"},
			{Namespace: "dwh-org-unit", ExternalID: " "},
			{Namespace: "other", ExternalID: "outside"},
		}}
		allowed, scoped, err := ResolveVisibility(ctx, f, "context-id", descendants)
		if err != nil || !scoped || !reflect.DeepEqual(allowed, map[string]struct{}{"d1": {}}) || f.ctx != ctx || f.id != "context-id" || f.descendants != descendants {
			t.Fatalf("resolution: %v %v %v", allowed, scoped, err)
		}
	}
	for _, code := range []string{"UG", "ug", " UG "} {
		f := &visibilityFixture{code: code, aliases: []VisibilityAlias{{Namespace: "dwh-org-unit", ExternalID: "ug"}}}
		_, scoped, err := ResolveVisibility(ctx, f, "context-id", true)
		if err != nil || scoped != (code != "UG") || (code == "UG" && f.aliasCalls != 0) {
			t.Fatalf("national exception changed: %q", code)
		}
	}
	f := &visibilityFixture{code: "UG", aliases: []VisibilityAlias{{Namespace: "dwh-org-unit", ExternalID: "ug"}}}
	if _, scoped, err := ResolveVisibility(ctx, f, "context-id", false); err != nil || !scoped || f.aliasCalls != 1 {
		t.Fatal("node-only national scope bypassed aliases")
	}
}
func TestVisibilityFailuresRemainRestrictedAndPreserveCauses(t *testing.T) {
	cause := errors.New("lookup failed")
	for _, reader := range []VisibilityReader{nil, &visibilityFixture{}, &visibilityFixture{nodeErr: cause}, &visibilityFixture{aliasErr: cause}} {
		ids, scoped, err := ResolveVisibility(context.Background(), reader, "id", false)
		if err == nil || !scoped || ids != nil {
			t.Fatal("failed scope became unrestricted")
		}
		if f, ok := reader.(*visibilityFixture); ok && (f.nodeErr != nil || f.aliasErr != nil) && !errors.Is(err, cause) {
			t.Fatal("lookup cause lost")
		}
	}
	if _, _, err := ResolveVisibility(context.Background(), &visibilityFixture{}, "", false); !errors.Is(err, ErrScopeForbidden) {
		t.Fatal("missing context accepted")
	}
}
func TestVisibilityFilteringRetainsExistingMatchingAndShape(t *testing.T) {
	allowed := map[string]struct{}{"d1": {}}
	full := []OrgUnitFull{{DimOrgHierarchyKey: 1, DistrictUID: textPtr(" d1 ")}, {DimOrgHierarchyKey: 2, DistrictUID: textPtr("d2")}}
	facilities := []Facility{{DimOrgHierarchyKey: 1, SubCountyUID: textPtr("d1")}, {DimOrgHierarchyKey: 2, FacilityUID: textPtr("other")}}
	simple := []OrgUnitSimple{{DimOrgHierarchyKey: 1, OrgUnitID: textPtr("d1")}, {DimOrgHierarchyKey: 2, OrgUnitID: nil}}
	if got := FilterOrgUnits(allowed, full); len(got) != 1 || got[0].DimOrgHierarchyKey != 1 {
		t.Fatal("ancestor matching changed")
	}
	if got := FilterFacilities(allowed, facilities); len(got) != 1 || got[0].DimOrgHierarchyKey != 1 {
		t.Fatal("facility matching changed")
	}
	if got := FilterSimpleOrgUnits(allowed, simple); len(got) != 1 || got[0].DimOrgHierarchyKey != 1 {
		t.Fatal("simple matching changed")
	}
	if len(full) != 2 || len(facilities) != 2 || len(simple) != 2 {
		t.Fatal("input mutated")
	}
	if got := FilterOrgUnits(nil, full); got == nil || len(got) != 0 {
		t.Fatal("empty allowed set leaked records")
	}
	if got := FilterSimpleOrgUnits(allowed, nil); got == nil || len(got) != 0 {
		t.Fatal("scoped nil shape changed")
	}
}

type coreRepository struct {
	Repository
	ctx    context.Context
	filter *string
	rows   []OrgUnit
	err    error
}

func (r *coreRepository) ListHierarchy(ctx context.Context) ([]OrgUnit, error) {
	r.ctx = ctx
	return r.rows, r.err
}
func (r *coreRepository) ListSubCounties(ctx context.Context, filter *string) ([]OrgUnitSimple, error) {
	r.ctx, r.filter = ctx, filter
	return nil, r.err
}
func TestPortableServicePreservesContextFiltersAndRepositoryErrors(t *testing.T) {
	ctx := context.Background()
	repo := &coreRepository{}
	service := NewService(repo)
	blank := ""
	for _, filter := range []*string{nil, &blank, textPtr("district")} {
		items, err := service.ListSubCounties(ctx, filter)
		if err != nil || items != nil || repo.ctx != ctx || repo.filter != filter {
			t.Fatal("repository delegation changed")
		}
	}
	cause := errors.New("warehouse unavailable")
	repo.err = cause
	if _, err := service.GetHierarchy(ctx); !errors.Is(err, cause) {
		t.Fatal("hierarchy error lost")
	}
}
func TestHierarchyPreservesFallbackDeduplicationAndLeafShape(t *testing.T) {
	country, region, district, sub, facility := "ug", "r1", "d1", "s1", "f1"
	row := func(id int64, level string) OrgUnit {
		return OrgUnit{DimOrgHierarchyKey: id, Level: level, CountryUID: &country, RegionUID: &region, Region: textPtr("Central"), DistrictUID: &district, District: textPtr("District"), SubCountyUID: &sub, SubCounty: textPtr("Subcounty"), FacilityUID: &facility, FacilityName: textPtr("Hospital")}
	}
	national := row(1, "1") // nil name retains the legacy National fallback.
	regionRow, districtRow, subRow, facRow := row(2, "2"), row(3, "3"), row(5, "5"), row(6, "6")
	localDuplicate := row(4, "4")
	localDuplicate.OrgUnitID = &sub
	localDuplicate.OrgUnitName = textPtr("Subcounty")
	repo := &coreRepository{rows: []OrgUnit{national, regionRow, regionRow, districtRow, localDuplicate, subRow, facRow, facRow}}
	tree, err := NewService(repo).GetHierarchy(context.Background())
	if err != nil || len(tree) != 1 || tree[0].Name == nil || *tree[0].Name != "National" {
		t.Fatalf("tree: %+v %v", tree, err)
	}
	n := tree[0]
	for _, kind := range []string{"region", "district", "subcounty", "facility"} {
		if len(n.Children) != 1 || n.Children[0].Type != kind {
			t.Fatalf("hierarchy/dedup changed at %s: %+v", kind, n)
		}
		n = n.Children[0]
	}
	if n.Children != nil {
		t.Fatal("leaf null children changed")
	}
	if got := buildHierarchyTree([]OrgUnit{facRow}); got == nil || len(got) != 0 {
		t.Fatal("missing national behavior changed")
	}
}
