package geojson

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/http/response"
)

type Handler struct {
	assets AssetReader
}

// NewHandler preserves the host's directory-based construction API.
func NewHandler(basePath string) *Handler {
	return NewHandlerWithReader(NewDirectoryReader(basePath))
}

// NewHandlerWithReader accepts the asset port without owning storage policy.
func NewHandlerWithReader(assets AssetReader) *Handler {
	return &Handler{assets: assets}
}

func (h *Handler) GetGeoJSON(c *gin.Context) {
	name := c.Param("name")

	files := map[string]string{
		"geo_water":       "uga_water.geojson",
		"geo_districts":   "uganda_districts.geojson",
		"geo_subcounties": "uganda_subcounties.geojson",
	}

	filename, exists := files[name]
	if !exists {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "requested geojson does not exist")
		return
	}

	if h == nil || h.assets == nil {
		response.Fail(c, http.StatusInternalServerError, "GEOJSON_LOAD_FAILED", "failed to load geojson file")
		return
	}
	asset, err := h.assets.Open(c.Request.Context(), filename)
	if err != nil {
		// The caller closes any returned stream, including on failure.
		if asset.Content != nil {
			_ = asset.Content.Close()
		}
		response.Fail(c, http.StatusInternalServerError, "GEOJSON_LOAD_FAILED", "failed to load geojson file")
		return
	}
	if asset.Content == nil {
		response.Fail(c, http.StatusInternalServerError, "GEOJSON_LOAD_FAILED", "failed to load geojson file")
		return
	}
	defer asset.Content.Close()

	c.Header("Content-Type", "application/geo+json")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("X-Content-Type-Options", "nosniff")
	http.ServeContent(c.Writer, c.Request, filename, asset.ModifiedAt, asset.Content)
}
