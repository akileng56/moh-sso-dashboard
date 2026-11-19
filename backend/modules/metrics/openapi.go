package metrics

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"
)

func (m *Module) OpenAPI() []byte {
	schemas := map[string]any{}
	paths := map[string]any{}
	models := map[string]any{
		"/overview":                     OverviewResponse{},
		"/system/count-users":           TotalUsersResponse{},
		"/system/count-disabled-users":  DisabledUsersResponse{},
		"/system/active-today":          ActiveUsersTodayResponse{},
		"/system/active-this-week":      ActiveUsersThisWeekResponse{},
		"/system/login-trend":           []LoginTrendPoint{},
		"/system/login-trend-range":     []LoginTrendByDayPoint{},
		"/security/failed-logins":       FailedLoginsResponse{},
		"/security/failed-logins-range": FailedLoginsResponse{},
		"/security/suspicious-logins":   []SuspiciousLogin{},
		"/clients/count":                TotalClientsResponse{},
		"/clients/most-accessed":        []MostAccessedClient{},
		"/clients/login-count":          LoginCountResponse{},
		"/clients/active-today":         []ActiveUsersPerClient{},
		"/users/new-range":              []UserSummary{},
		"/users/new-trend":              []NewUsersTrendPoint{},
		"/users/never-logged-in":        []UserSummary{},
		"/users/last-login/{userID}":    LastLoginResponse{},
		"/users/client-usage/{userID}":  []UserClientUsagePoint{},
	}
	ranged := map[string]bool{"/system/login-trend-range": true, "/security/failed-logins-range": true, "/security/suspicious-logins": true, "/clients/most-accessed": true, "/clients/login-count": true, "/users/new-range": true, "/users/new-trend": true, "/users/client-usage/{userID}": true}
	for _, pattern := range m.Patterns() {
		route := strings.TrimPrefix(pattern, "GET ")
		suffix := strings.TrimPrefix(route, m.base)
		model, ok := models[suffix]
		if !ok {
			panic("missing Metrics operation metadata")
		}
		data := modelSchema(reflect.TypeOf(model), schemas)
		if reflect.TypeOf(model).Kind() == reflect.Slice {
			delete(data, "nullable")
		}
		envelope := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"success": map[string]any{"type": "boolean"},
				"data":    data,
				"meta":    map[string]any{"type": "object", "properties": map[string]any{"requestId": map[string]any{"type": "string"}}},
			},
		}
		responses := map[string]any{"200": jsonResponse("Metrics result", envelope)}
		for _, status := range []string{"400", "401", "403", "500", "503"} {
			failure := map[string]any{
				"type": "object",
				"properties": map[string]any{
					"success": map[string]any{"type": "boolean"},
					"error":   map[string]any{"type": "object", "properties": map[string]any{"code": map[string]any{"type": "string"}, "message": map[string]any{"type": "string"}}},
					"meta":    map[string]any{"type": "object"},
				},
			}
			responses[status] = jsonResponse("Validation, access or service failure", failure)
		}
		parameters := []any{}
		param := func(name, location string, required bool, schema map[string]any) {
			parameters = append(parameters, map[string]any{"name": name, "in": location, "required": required, "schema": schema})
		}
		if ranged[suffix] {
			for _, name := range []string{"start", "end"} {
				param(name, "query", true, map[string]any{"type": "string", "format": "date-time"})
			}
		}
		if strings.Contains(suffix, "{userID}") {
			param("userID", "path", true, map[string]any{"type": "string", "format": "uuid"})
		}
		if suffix == "/clients/login-count" {
			param("client_id", "query", true, map[string]any{"type": "string"})
		}
		if suffix == "/clients/most-accessed" {
			param("limit", "query", false, map[string]any{"type": "integer", "format": "int32"})
		}
		if suffix == "/security/suspicious-logins" {
			param("home", "query", false, map[string]any{"type": "string", "default": "UG"})
		}
		paths[route] = map[string]any{"get": map[string]any{"operationId": "metrics_" + strings.NewReplacer("/", "_", "{", "", "}", "").Replace(suffix), "tags": []string{"metrics"}, "parameters": parameters, "responses": responses, "x-admin-required": true, "x-route-permissions": []string{ReadPermission}, "x-go-handler": ModulePath + ".(*Module).ServeHTTP"}}
	}
	document := map[string]any{"openapi": "3.0.3", "info": map[string]any{"title": "Metrics module", "version": Version()}, "paths": paths,
		"security":   []any{map[string]any{"bearerAuth": []string{}}, map[string]any{"sessionCookie": []string{}}, map[string]any{"legacyAccessCookie": []string{}}},
		"components": map[string]any{"schemas": schemas, "securitySchemes": map[string]any{"bearerAuth": map[string]any{"type": "http", "scheme": "bearer", "bearerFormat": "JWT"}, "sessionCookie": map[string]any{"type": "apiKey", "in": "cookie", "name": "sso_session"}, "legacyAccessCookie": map[string]any{"type": "apiKey", "in": "cookie", "name": "access_token"}}}}
	bytes, err := json.Marshal(document)
	if err != nil {
		panic(err)
	}
	return bytes
}
func jsonResponse(description string, schema map[string]any) map[string]any {
	return map[string]any{"description": description, "content": map[string]any{"application/json": map[string]any{"schema": schema}}}
}
func modelSchema(t reflect.Type, definitions map[string]any) map[string]any {
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
		name := "metrics." + t.Name()
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
