package admin_units

import (
	"context"

	units "github.com/mohuganda/moh-sso-dashboard/backend/modules/admin-units"
)

// Compatibility names refer to the selected package, with no retained core.
type OrgUnit = units.OrgUnit
type OrgUnitFull = units.OrgUnitFull
type OrgUnitSimple = units.OrgUnitSimple
type Facility = units.Facility
type TreeNode = units.TreeNode
type Repository = units.Repository
type Service = units.Service
type VisibilityReader = units.VisibilityReader
type VisibilityNode = units.VisibilityNode
type VisibilityAlias = units.VisibilityAlias

var ErrScopeForbidden = units.ErrScopeForbidden

func NewService(repository Repository) Service     { return units.NewService(repository) }
func BuildHierarchyTree(rows []OrgUnit) []TreeNode { return units.BuildHierarchyTree(rows) }
func ResolveVisibility(ctx context.Context, reader VisibilityReader, id string, descendants bool) (map[string]struct{}, bool, error) {
	return units.ResolveVisibility(ctx, reader, id, descendants)
}
func FilterOrgUnits(allowed map[string]struct{}, items []OrgUnitFull) []OrgUnitFull {
	return units.FilterOrgUnits(allowed, items)
}
func FilterFacilities(allowed map[string]struct{}, items []Facility) []Facility {
	return units.FilterFacilities(allowed, items)
}
func FilterSimpleOrgUnits(allowed map[string]struct{}, items []OrgUnitSimple) []OrgUnitSimple {
	return units.FilterSimpleOrgUnits(allowed, items)
}
