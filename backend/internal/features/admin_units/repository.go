package admin_units

import (
	"context"
)

type Repository interface {
	ListOrgUnits(ctx context.Context) ([]OrgUnitFull, error)
	ListFacilities(ctx context.Context) ([]Facility, error)
	ListDistricts(ctx context.Context) ([]OrgUnitSimple, error)
	ListSubCounties(ctx context.Context, subCounty *string) ([]OrgUnitSimple, error)
	ListLocalGovt(ctx context.Context, district *string) ([]OrgUnitSimple, error)
	ListDistrictsByRegion(ctx context.Context, region *string) ([]OrgUnitSimple, error)
	ListRegions(ctx context.Context) ([]OrgUnitSimple, error)
	ListNational(ctx context.Context) ([]OrgUnitSimple, error)
	ListHierarchy(ctx context.Context) ([]OrgUnit, error)
}
