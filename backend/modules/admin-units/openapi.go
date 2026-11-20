package admin_units

import (
	"encoding/json"
	"reflect"
	"strings"
)

// OpenAPI is the package-owned contract for the nine registered endpoints.
func (m *Module) OpenAPI() []byte {
	paths := map[string]any{}
	schemas := map[string]any{}
	patterns := m.Patterns()
	for i, pattern := range patterns {
		fields := strings.SplitN(pattern, " ", 2)
		model := reflect.TypeOf(OrgUnitSimple{})
		switch endpoints[i].suffix {
		case "orgunits":
			model = reflect.TypeOf(OrgUnitFull{})
		case "facilities":
			model = reflect.TypeOf(Facility{})
		case "hierarchy":
			model = reflect.TypeOf(TreeNode{})
		}
		response := map[string]any{
			"description": "Organizational units visible in the selected health context",
			"content": map[string]any{"application/json": map[string]any{"schema": map[string]any{
				"type": "object", "properties": map[string]any{"success": map[string]any{"type": "boolean"}, "data": map[string]any{"type": "array", "items": modelSchema(model, schemas)}, "meta": map[string]any{"type": "object"}},
			}}},
		}
		responses := map[string]any{"200": response}
		for _, status := range []string{"400", "401", "403", "413", "429", "500", "503"} {
			responses[status] = map[string]any{"description": "Request rejected or service unavailable", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "properties": map[string]any{
				"success": map[string]any{"type": "boolean"}, "error": map[string]any{"type": "object", "properties": map[string]any{"code": map[string]any{"type": "string"}, "message": map[string]any{"type": "string"}}}, "meta": map[string]any{"type": "object"},
			}}}}}
		}
		operation := map[string]any{
			"operationId": "admin_units_" + endpoints[i].suffix, "tags": []string{"admin-units"}, "responses": responses,
			"x-route-any-permissions": RequiredPermissions(), "x-required-system": RequiredSystem,
			"x-health-context-required": true, "x-go-handler": ModulePath + ".(*Module).ServeHTTP",
			"parameters": []any{map[string]any{"in": "header", "name": "X-Health-Context-ID", "description": "Platform-verified selection; omission is accepted only when the platform explicitly allows legacy unscoped access.", "required": false, "schema": map[string]any{"type": "string"}}},
		}
		if fields[0] == "POST" {
			filter := map[string]string{"subcounties": "sub_county", "localgovt": "district", "districts": "region"}[endpoints[i].suffix]
			operation["requestBody"] = map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "properties": map[string]any{filter: map[string]any{"type": "string", "nullable": true}}}}}}
		}
		paths[fields[1]] = map[string]any{strings.ToLower(fields[0]): operation}
	}
	data, err := json.Marshal(map[string]any{
		"openapi": "3.0.3", "info": map[string]any{"title": "Administrative Units module", "version": Version()}, "paths": paths,
		"security": []any{map[string]any{"bearerAuth": []string{}}, map[string]any{"sessionCookie": []string{}}, map[string]any{"legacyAccessCookie": []string{}}},
		"components": map[string]any{"schemas": schemas, "securitySchemes": map[string]any{
			"bearerAuth":         map[string]any{"type": "http", "scheme": "bearer", "bearerFormat": "JWT"},
			"legacyAccessCookie": map[string]any{"type": "apiKey", "in": "cookie", "name": "access_token"},
			"sessionCookie":      map[string]any{"type": "apiKey", "in": "cookie", "name": "sso_session"},
		}},
	})
	if err != nil {
		panic(err)
	}
	return data
}

func modelSchema(t reflect.Type, definitions map[string]any) map[string]any {
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
	case reflect.Int64:
		return map[string]any{"type": "integer", "format": "int64"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Slice:
		return map[string]any{"type": "array", "nullable": true, "items": modelSchema(t.Elem(), definitions)}
	case reflect.Struct:
		name := "admin_units." + t.Name()
		if _, seen := definitions[name]; !seen {
			properties := map[string]any{}
			required := []string{}
			definitions[name] = map[string]any{} // Break recursive TreeNode.Children.
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
