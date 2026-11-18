package admin_units

import (
	"context"
	"errors"
	"strings"
)

// ErrScopeForbidden identifies an invalid or forbidden selected context.
var ErrScopeForbidden = errors.New("health context access denied")

type VisibilityNode struct{ Code string }
type VisibilityAlias struct{ Namespace, ExternalID string }

// VisibilityReader supplies only the health-context values needed for DWH visibility.
// Authentication and context selection remain the host's responsibility.
type VisibilityReader interface {
	GetNode(context.Context, string) (VisibilityNode, error)
	ListAliasesInScope(context.Context, string, bool) ([]VisibilityAlias, error)
}

// ResolveVisibility preserves the national-descendants exception and requires
// DWH aliases for every other selected context. Context IDs are host-validated.
func ResolveVisibility(ctx context.Context, reader VisibilityReader, contextID string, descendants bool) (map[string]struct{}, bool, error) {
	if contextID == "" || reader == nil {
		return nil, true, ErrScopeForbidden
	}
	node, err := reader.GetNode(ctx, contextID)
	if err != nil {
		return nil, true, err
	}
	if node.Code == "UG" && descendants {
		return nil, false, nil
	}
	aliases, err := reader.ListAliasesInScope(ctx, contextID, descendants)
	if err != nil {
		return nil, true, err
	}
	allowed := make(map[string]struct{})
	for _, alias := range aliases {
		if alias.Namespace != "dwh-org-unit" {
			continue
		}
		if id := strings.TrimSpace(alias.ExternalID); id != "" {
			allowed[id] = struct{}{}
		}
	}
	if len(allowed) == 0 {
		return nil, true, errors.New("selected health context has no DWH organisation-unit mapping")
	}
	return allowed, true, nil
}

func orgUnitMatches(allowed map[string]struct{}, ids ...*string) bool {
	for _, id := range ids {
		if id == nil {
			continue
		}
		if _, ok := allowed[strings.TrimSpace(*id)]; ok {
			return true
		}
	}
	return false
}

func FilterOrgUnits(allowed map[string]struct{}, items []OrgUnitFull) []OrgUnitFull {
	result := make([]OrgUnitFull, 0, len(items))
	for _, item := range items {
		if orgUnitMatches(allowed, item.OrgUnitID, item.CountryUID, item.RegionUID, item.DistrictUID, item.SubCountyUID, item.DivisionUID, item.FacilityUID) {
			result = append(result, item)
		}
	}
	return result
}

func FilterFacilities(allowed map[string]struct{}, items []Facility) []Facility {
	result := make([]Facility, 0, len(items))
	for _, item := range items {
		if orgUnitMatches(allowed, item.OrgUnitID, item.CountryUID, item.RegionUID, item.DistrictUID, item.SubCountyUID, item.DivisionUID, item.FacilityUID) {
			result = append(result, item)
		}
	}
	return result
}

func FilterSimpleOrgUnits(allowed map[string]struct{}, items []OrgUnitSimple) []OrgUnitSimple {
	result := make([]OrgUnitSimple, 0, len(items))
	for _, item := range items {
		if orgUnitMatches(allowed, item.OrgUnitID) {
			result = append(result, item)
		}
	}
	return result
}
