// Package geojson supplies host storage and HTTP adapters for the published
// GeoJSON module. Asset packaging and authorization remain host-owned.
package geojson

import geojsonmodule "github.com/mohuganda/moh-sso-dashboard/backend/modules/geojson"

type Asset = geojsonmodule.Asset
type AssetReader = geojsonmodule.AssetReader
type Layer = geojsonmodule.Layer
type Service = geojsonmodule.Service

var (
	ErrUnknownLayer = geojsonmodule.ErrUnknownLayer
	ErrLoadFailed   = geojsonmodule.ErrLoadFailed
)

func NewService(assets AssetReader) *Service   { return geojsonmodule.NewService(assets) }
func ResolveLayer(name string) (string, error) { return geojsonmodule.ResolveLayer(name) }
