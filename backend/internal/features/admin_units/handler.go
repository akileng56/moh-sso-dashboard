package admin_units

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/http/response"
)

type Handler struct {
	service    Service
	visibility VisibilityReader
}

// NewHandlerWithPorts composes HTTP handlers without owning infrastructure.
func NewHandlerWithPorts(service Service, visibility VisibilityReader) *Handler {
	return &Handler{service: service, visibility: visibility}
}

func (h *Handler) GetOrgUnits(c *gin.Context) {
	results, err := h.service.ListOrgUnits(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_ORG_UNITS_FAILED", "failed to list organizational units")
		return
	}
	results, err = h.scopeOrgUnits(c, results)
	if err != nil {
		h.handleScopeError(c, err)
		return
	}

	response.OK(c, http.StatusOK, toOrgUnitFullResponses(results))
}

// GetFacilities gets all facilities (level 6)
func (h *Handler) GetFacilities(c *gin.Context) {
	results, err := h.service.ListFacilities(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_FACILITIES_FAILED", "failed to list facilities")
		return
	}
	results, err = h.scopeFacilities(c, results)
	if err != nil {
		h.handleScopeError(c, err)
		return
	}

	response.OK(c, http.StatusOK, toFacilityResponses(results))
}

// GetDistricts gets all districts (level 3)
func (h *Handler) GetDistricts(c *gin.Context) {
	results, err := h.service.ListDistricts(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_DISTRICTS_FAILED", "failed to list districts")
		return
	}
	results, err = h.scopeSimpleOrgUnits(c, results)
	if err != nil {
		h.handleScopeError(c, err)
		return
	}

	response.OK(c, http.StatusOK, toOrgUnitSimpleResponses(results))
}

// GetSubCounties gets subcounties (level 5), optionally filtered by sub_county
func (h *Handler) GetSubCounties(c *gin.Context) {
	type Request struct {
		SubCounty *string `json:"sub_county"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		// No JSON body or invalid JSON: continue without filter
		req = Request{}
	}

	results, err := h.service.ListSubCounties(c.Request.Context(), req.SubCounty)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_SUBCOUNTIES_FAILED", "failed to list subcounties")
		return
	}
	results, err = h.scopeSimpleOrgUnits(c, results)
	if err != nil {
		h.handleScopeError(c, err)
		return
	}

	response.OK(c, http.StatusOK, toOrgUnitSimpleResponses(results))
}

// GetLocalGovt gets local government units (level 4), optionally filtered by district
func (h *Handler) GetLocalGovt(c *gin.Context) {
	type Request struct {
		District *string `json:"district"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		req = Request{}
	}

	results, err := h.service.ListLocalGovt(c.Request.Context(), req.District)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_LOCAL_GOVT_FAILED", "failed to list local governments")
		return
	}
	results, err = h.scopeSimpleOrgUnits(c, results)
	if err != nil {
		h.handleScopeError(c, err)
		return
	}

	response.OK(c, http.StatusOK, toOrgUnitSimpleResponses(results))
}

// GetDistrictsByRegion gets districts (level 3) filtered by region
func (h *Handler) GetDistrictsByRegion(c *gin.Context) {
	type Request struct {
		Region *string `json:"region"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		req = Request{}
	}

	results, err := h.service.ListDistrictsByRegion(c.Request.Context(), req.Region)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_REGIONS_FAILED", "failed to list regions")
		return
	}
	results, err = h.scopeSimpleOrgUnits(c, results)
	if err != nil {
		h.handleScopeError(c, err)
		return
	}

	response.OK(c, http.StatusOK, toOrgUnitSimpleResponses(results))
}

// GetRegions gets all regions (level 2)
func (h *Handler) GetRegions(c *gin.Context) {
	results, err := h.service.ListRegions(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_NATIONAL_FAILED", "failed to list national hierarchy")
		return
	}
	results, err = h.scopeSimpleOrgUnits(c, results)
	if err != nil {
		h.handleScopeError(c, err)
		return
	}

	response.OK(c, http.StatusOK, toOrgUnitSimpleResponses(results))
}

// GetNational gets national level (level 1)
func (h *Handler) GetNational(c *gin.Context) {
	results, err := h.service.ListNational(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_HIERARCHY_FAILED", "failed to list hierarchy")
		return
	}
	results, err = h.scopeSimpleOrgUnits(c, results)
	if err != nil {
		h.handleScopeError(c, err)
		return
	}

	response.OK(c, http.StatusOK, toOrgUnitSimpleResponses(results))
}

// GetHierarchy gets the full organizational hierarchy as a tree structure
func (h *Handler) GetHierarchy(c *gin.Context) {
	data, err := h.service.ListOrgUnits(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_HIERARCHY_FAILED", "failed to list hierarchy")
		return
	}
	data, err = h.scopeOrgUnits(c, data)
	if err != nil {
		h.handleScopeError(c, err)
		return
	}
	orgUnits := make([]OrgUnit, 0, len(data))
	for _, item := range data {
		orgUnits = append(orgUnits, OrgUnit{
			DimOrgHierarchyKey: item.DimOrgHierarchyKey,
			OrgUnitID:          item.OrgUnitID, OrgUnitName: item.OrgUnitName, Level: item.Level,
			CountryUID: item.CountryUID, RegionUID: item.RegionUID, Region: item.Region,
			DistrictUID: item.DistrictUID, District: item.District,
			SubCountyUID: item.SubCountyUID, SubCounty: item.SubCounty,
			DivisionUID: item.DivisionUID, Division: item.Division,
			FacilityUID: item.FacilityUID, FacilityName: item.FacilityName,
		})
	}
	tree := BuildHierarchyTree(orgUnits)

	response.OK(c, http.StatusOK, toTreeNodeResponses(tree))
}

func (h *Handler) scopeOrgUnits(c *gin.Context, items []OrgUnitFull) ([]OrgUnitFull, error) {
	allowed, scoped, err := h.allowedOrgUnitIDs(c)
	if err != nil || !scoped {
		return items, err
	}
	return FilterOrgUnits(allowed, items), nil
}

func (h *Handler) scopeFacilities(c *gin.Context, items []Facility) ([]Facility, error) {
	allowed, scoped, err := h.allowedOrgUnitIDs(c)
	if err != nil || !scoped {
		return items, err
	}
	return FilterFacilities(allowed, items), nil
}

func (h *Handler) scopeSimpleOrgUnits(c *gin.Context, items []OrgUnitSimple) ([]OrgUnitSimple, error) {
	allowed, scoped, err := h.allowedOrgUnitIDs(c)
	if err != nil || !scoped {
		return items, err
	}
	return FilterSimpleOrgUnits(allowed, items), nil
}

func (h *Handler) handleScopeError(c *gin.Context, err error) {
	if errors.Is(err, ErrScopeForbidden) {
		response.Fail(c, http.StatusForbidden, "HEALTH_CONTEXT_FORBIDDEN", "health context access denied")
		return
	}
	response.Fail(c, http.StatusForbidden, "HEALTH_CONTEXT_MAPPING_REQUIRED", "selected health context is not mapped to the data hierarchy")
}
