package audit

import (
	"encoding/json"
	"github.com/google/uuid"
	"reflect"
	"strings"
	"time"
)

func (m *Module) OpenAPI() []byte {
	schemas := map[string]any{}
	paths := map[string]any{}
	models := map[string]any{"": AuditLogListResponse{}, "/actions": AuditActionsResponse{}, "/metrics/overview": AuditMetricsOverviewResponse{}, "/metrics/failed-logins-by-day": AuditFailedLoginsByDayResponse{}, "/metrics/top-failure-ips": AuditTopFailureIPsResponse{}, "/export": AuditExportResponse{}, "/{id}": AuditLogResponse{}}
	for _, pattern := range m.Patterns() {
		route := strings.TrimPrefix(pattern, "GET ")
		suffix := strings.TrimPrefix(route, m.base)
		data := modelSchema(reflect.TypeOf(models[suffix]), schemas)
		success := map[string]any{"type": "object", "properties": map[string]any{"success": map[string]any{"type": "boolean"}, "data": data, "meta": map[string]any{"type": "object"}}}
		responses := map[string]any{"200": jsonResponse("Audit result", success)}
		for _, status := range []string{"400", "401", "403", "404", "429", "500", "503"} {
			responses[status] = jsonResponse("Safe validation, access or service error", map[string]any{"type": "object", "properties": map[string]any{"success": map[string]any{"type": "boolean"}, "error": map[string]any{"type": "object", "properties": map[string]any{"code": map[string]any{"type": "string"}, "message": map[string]any{"type": "string"}}}, "meta": map[string]any{"type": "object"}}})
		}
		parameters := []any{}
		param := func(name, where string, required bool, schema map[string]any) {
			parameters = append(parameters, map[string]any{"name": name, "in": where, "required": required, "schema": schema})
		}
		if suffix != "/actions" && suffix != "/{id}" {
			for _, name := range []string{"from", "to"} {
				param(name, "query", true, map[string]any{"type": "string", "format": "date-time"})
			}
		}
		if suffix == "/{id}" {
			param("id", "path", true, map[string]any{"type": "string", "format": "uuid"})
		}
		if suffix == "" || suffix == "/export" {
			for _, name := range []string{"action", "client_id", "ip", "success", "user_id"} {
				schema := map[string]any{"type": "string"}
				if name == "user_id" {
					schema["format"] = "uuid"
				}
				param(name, "query", false, schema)
			}
		}
		if suffix == "" {
			param("cursor_created_at", "query", false, map[string]any{"type": "string", "format": "date-time"})
			param("cursor_id", "query", false, map[string]any{"type": "string", "format": "uuid"})
			param("limit", "query", false, map[string]any{"type": "integer", "minimum": 1, "maximum": 200, "default": 50})
		}
		if suffix == "/metrics/top-failure-ips" {
			param("limit", "query", false, map[string]any{"type": "integer", "minimum": 1, "maximum": 50, "default": 10})
		}
		operation := map[string]any{"operationId": "audit" + strings.NewReplacer("/", "_", "{", "", "}", "").Replace(suffix), "tags": []string{"audit"}, "parameters": parameters, "responses": responses, "x-admin-required": true, "x-route-permissions": []string{ReadPermission}, "x-go-handler": ModulePath + ".(*Module).ServeHTTP"}
		if suffix == "/export" {
			param("format", "query", false, map[string]any{"type": "string", "default": "csv", "description": "json returns an envelope; other values retain CSV fallback"})
			operation["parameters"] = parameters
			response := responses["200"].(map[string]any)
			response["content"].(map[string]any)["text/csv"] = map[string]any{"schema": map[string]any{"type": "string", "format": "binary"}}
			response["headers"] = map[string]any{"Content-Disposition": map[string]any{"schema": map[string]any{"type": "string"}}, "X-Audit-Export-Manifest": map[string]any{"description": "JSON manifest for CSV, including record count and JSON-item SHA-256", "schema": map[string]any{"type": "string"}}}
			operation["x-rate-limit-policy"] = map[string]any{"name": ExportPolicy, "limit": ExportLimit, "windowSeconds": int(ExportWindow.Seconds()), "key": "verified-user"}
		}
		paths[route] = map[string]any{"get": operation}
	}
	if definition, ok := schemas["audit.AuditFailedLoginsByDayPoint"].(map[string]any); ok {
		definition["properties"].(map[string]any)["day"].(map[string]any)["description"] = "Database day-bucket start, Unix milliseconds"
	}
	data, err := json.Marshal(map[string]any{"openapi": "3.0.3", "info": map[string]any{"title": "Audit read module", "version": Version()}, "paths": paths, "components": map[string]any{"schemas": schemas, "securitySchemes": map[string]any{
		"bearerAuth":         map[string]any{"type": "http", "scheme": "bearer"},
		"sessionCookie":      map[string]any{"type": "apiKey", "in": "cookie", "name": "sso_session"},
		"legacyAccessCookie": map[string]any{"type": "apiKey", "in": "cookie", "name": "access_token"},
	}},
		"security": []any{map[string]any{"bearerAuth": []string{}}, map[string]any{"sessionCookie": []string{}}, map[string]any{"legacyAccessCookie": []string{}}}})
	if err != nil {
		panic(err)
	}
	return data
}
func jsonResponse(description string, schema map[string]any) map[string]any {
	return map[string]any{"description": description, "content": map[string]any{"application/json": map[string]any{"schema": schema}}}
}
func modelSchema(t reflect.Type, definitions map[string]any) map[string]any {
	if t == reflect.TypeOf(uuid.UUID{}) {
		return map[string]any{"type": "string", "format": "uuid"}
	}
	if t == reflect.TypeOf(time.Time{}) {
		return map[string]any{"type": "string", "format": "date-time"}
	}
	switch t.Kind() {
	case reflect.Pointer:
		schema := modelSchema(t.Elem(), definitions)
		if ref, ok := schema["$ref"]; ok {
			return map[string]any{"nullable": true, "allOf": []any{map[string]any{"$ref": ref}}}
		}
		schema["nullable"] = true
		return schema
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Int, reflect.Int32, reflect.Int64:
		return map[string]any{"type": "integer", "format": "int64"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Slice:
		return map[string]any{"type": "array", "nullable": true, "items": modelSchema(t.Elem(), definitions)}
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": modelSchema(t.Elem(), definitions)}
	case reflect.Struct:
		name := "audit." + t.Name()
		if _, seen := definitions[name]; !seen {
			properties := map[string]any{}
			required := []string{}
			definitions[name] = map[string]any{} // Break recursive model references.
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				tag := strings.Split(field.Tag.Get("json"), ",")
				if tag[0] == "-" {
					continue
				}
				key := tag[0]
				if key == "" {
					key = field.Name
				}
				properties[key] = modelSchema(field.Type, definitions)
				optional := false
				for _, option := range tag[1:] {
					if option == "omitempty" {
						optional = true
					}
				}
				if !optional {
					required = append(required, key)
				}
			}
			definitions[name] = map[string]any{"type": "object", "properties": properties, "required": required}
		}
		return map[string]any{"$ref": "#/components/schemas/" + name}
	}
	return map[string]any{}
}
