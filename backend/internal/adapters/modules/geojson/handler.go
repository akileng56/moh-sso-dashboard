package geojson

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/http/response"
)

type Handler struct {
	service *Service
}

// NewHandler preserves the host's directory-based construction API.
func NewHandler(basePath string) *Handler {
	return NewHandlerWithReader(NewDirectoryReader(basePath))
}

// NewHandlerWithReader accepts the asset port without owning storage policy.
func NewHandlerWithReader(assets AssetReader) *Handler {
	return &Handler{service: NewService(assets)}
}

func (h *Handler) GetGeoJSON(c *gin.Context) {
	var service *Service
	if h != nil {
		service = h.service
	}
	layer, err := service.Open(c.Request.Context(), c.Param("name"))
	if err != nil {
		// Only resolution can produce a 404; a reader failure remains a load error.
		if errors.Is(err, ErrUnknownLayer) && !errors.Is(err, ErrLoadFailed) {
			response.Fail(c, http.StatusNotFound, "NOT_FOUND", "requested geojson does not exist")
		} else {
			response.Fail(c, http.StatusInternalServerError, "GEOJSON_LOAD_FAILED", "failed to load geojson file")
		}
		return
	}
	defer layer.Content.Close()

	c.Header("Content-Type", "application/geo+json")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("X-Content-Type-Options", "nosniff")
	http.ServeContent(c.Writer, c.Request, layer.Filename, layer.ModifiedAt, layer.Content)
}
