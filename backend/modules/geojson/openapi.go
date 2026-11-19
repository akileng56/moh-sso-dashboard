package geojson

import (
	"encoding/json"
	"strings"
)

// OpenAPI supplies the module-owned operation contract. Framework hosts merge
// it against their actual registered routes; no dashboard handler source is
// needed to document the self-contained module.
func (m *Module) OpenAPI() []byte {
	pattern := m.pattern
	if pattern == "" {
		pattern = "GET /geojson/{name}"
	}
	route := strings.TrimPrefix(pattern, "GET ")
	errorSchema := map[string]any{"type": "object", "properties": map[string]any{
		"success": map[string]any{"type": "boolean"},
		"error":   map[string]any{"type": "object", "properties": map[string]any{"code": map[string]any{"type": "string"}, "message": map[string]any{"type": "string"}}},
		"meta":    map[string]any{"type": "object", "properties": map[string]any{"requestId": map[string]any{"type": "string"}}},
	}}
	responses := map[string]any{}
	for code, description := range map[string]string{
		"401": "Unauthorized", "403": "Forbidden", "404": "Not Found", "429": "Too Many Requests",
		"500": "Internal Server Error", "503": "Authorization service unavailable",
	} {
		responses[code] = map[string]any{"description": description, "content": map[string]any{"application/json": map[string]any{"schema": errorSchema}}}
	}
	for code, description := range map[string]string{"200": "GeoJSON layer", "206": "Partial GeoJSON layer"} {
		responses[code] = map[string]any{"description": description, "content": map[string]any{"application/geo+json": map[string]any{"schema": map[string]any{"type": "string", "format": "binary"}}}}
	}
	responses["304"] = map[string]any{"description": "Not Modified"}
	responses["416"] = map[string]any{"description": "Requested range not satisfiable"}
	op := map[string]any{
		"operationId": "get_api_v1_geojson_name", "summary": "Get Geo JSON", "tags": []string{"geojson"},
		"parameters": []any{map[string]any{"in": "path", "name": "name", "required": true, "schema": map[string]any{"type": "string"}}},
		"responses":  responses, "x-go-handler": ModulePath + ".(*Module).ServeHTTP",
		"x-route-any-permissions": RequiredPermissions(),
	}
	document := map[string]any{
		"security": []any{map[string]any{"bearerAuth": []string{}}, map[string]any{"sessionCookie": []string{}}, map[string]any{"legacyAccessCookie": []string{}}},
		"components": map[string]any{"securitySchemes": map[string]any{
			"bearerAuth":         map[string]any{"type": "http", "scheme": "bearer", "bearerFormat": "JWT"},
			"sessionCookie":      map[string]any{"type": "apiKey", "in": "cookie", "name": "sso_session"},
			"legacyAccessCookie": map[string]any{"type": "apiKey", "in": "cookie", "name": "access_token"},
		}},

		"openapi": "3.0.3", "info": map[string]any{"title": "GeoJSON module", "version": Version()},
		"paths": map[string]any{route: map[string]any{"get": op}},
	}
	data, err := json.Marshal(document)
	if err != nil {
		panic(err)
	} // Only static JSON-compatible values are encoded.
	return data
}
