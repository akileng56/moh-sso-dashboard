package announcements

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"
)

// OpenAPI declares the same 26 operations registered by the package.
func (m *Module) OpenAPI() []byte {
	schemas := map[string]any{}
	paths := map[string]any{}
	for _, route := range endpoints(&httpHandler{}) {
		var model any = AnnouncementResponse{}
		if route.method == "DELETE" {
			model = MessageResponse{}
		}
		if route.path == "/admin/announcements/audience-preview" {
			model = AnnouncementAudiencePreviewResponse{}
		}
		if route.path == "/admin/announcements/stats" {
			model = AnnouncementStatsResponse{}
		}
		download := strings.HasSuffix(route.path, "/download")
		if strings.HasSuffix(route.path, "/attachments") {
			if route.method == "GET" {
				model = []AnnouncementAttachmentResponse{}
			} else {
				model = AnnouncementAttachmentResponse{}
			}
		} else if strings.HasSuffix(route.path, "/attachments/{attachmentId}") && route.method == "PATCH" {
			model = AnnouncementAttachmentResponse{}
		}
		if route.method == "GET" && !download && !strings.HasSuffix(route.path, "/{id}") && !strings.HasSuffix(route.path, "/stats") && !strings.HasSuffix(route.path, "/attachments") {
			model = []AnnouncementResponse{}
		}
		data := modelSchema(reflect.TypeOf(model), schemas)
		if reflect.TypeOf(model).Kind() == reflect.Slice {
			delete(data, "nullable")
		}
		success := map[string]any{"type": "object", "properties": map[string]any{
			"success": map[string]any{"type": "boolean"}, "data": data,
			"meta": map[string]any{"type": "object", "properties": map[string]any{"requestId": map[string]any{"type": "string"}}},
		}}
		status := "200"
		if route.method == "POST" && (route.path == "/admin/announcements" || strings.HasSuffix(route.path, "/attachments")) {
			status = "201"
		}
		responses := map[string]any{status: jsonResponse("Announcements result", success)}
		if download {
			responses["200"] = map[string]any{"description": "Attachment stream", "content": map[string]any{"application/octet-stream": map[string]any{"schema": map[string]any{"type": "string", "format": "binary"}}}}
		}
		for _, status := range []string{"400", "401", "403", "404", "429", "500", "503"} {
			responses[status] = jsonResponse("Validation, access, rate or provider failure", map[string]any{"type": "object", "properties": map[string]any{
				"success": map[string]any{"type": "boolean"},
				"error":   map[string]any{"type": "object", "properties": map[string]any{"code": map[string]any{"type": "string"}, "message": map[string]any{"type": "string"}}},
				"meta":    map[string]any{"type": "object"},
			}})
		}
		params := []any{}
		for _, name := range []string{"id", "attachmentId", "client_id", "role_name"} {
			if strings.Contains(route.path, "{"+name+"}") {
				schema := map[string]any{"type": "string"}
				if name != "role_name" {
					schema["format"] = "uuid"
				}
				params = append(params, map[string]any{"name": name, "in": "path", "required": true, "schema": schema})
			}
		}
		if reflect.TypeOf(model).Kind() == reflect.Slice && route.method == "GET" && !strings.HasSuffix(route.path, "/attachments") {
			for _, name := range []string{"limit", "offset"} {
				params = append(params, map[string]any{"name": name, "in": "query", "schema": map[string]any{"type": "integer", "format": "int32"}})
			}
		}
		if route.path == "/announcements/me" {
			for _, name := range []string{"role", "client_id"} {
				params = append(params, map[string]any{"name": name, "in": "query", "schema": map[string]any{"type": "string"}})
			}
		}
		operation := map[string]any{
			"operationId": "announcements_" + strings.ToLower(route.method) + strings.NewReplacer("/", "_", "{", "", "}", "").Replace(route.path),
			"tags":        []string{ID}, "parameters": params, "responses": responses,
			"x-admin-required": route.admin, "x-go-handler": ModulePath + ".(*Module).ServeHTTP",
		}
		if route.permission != "" {
			operation["x-route-permissions"] = []string{route.permission}
		} else {
			operation["security"] = []any{}
		}
		if route.policy != "" {
			operation["x-rate-policy"] = route.policy
		}
		if input := requestModel(route.method, route.path); input != nil {
			content := map[string]any{"application/json": map[string]any{"schema": modelSchema(reflect.TypeOf(input), schemas)}}
			if strings.HasSuffix(route.path, "/attachments") && route.method == "POST" {
				content["multipart/form-data"] = map[string]any{"schema": map[string]any{"type": "object", "required": []string{"file"}, "properties": map[string]any{
					"file": map[string]any{"type": "string", "format": "binary"}, "include_in_email": map[string]any{"type": "boolean"}, "inline": map[string]any{"type": "boolean"}, "content_id": map[string]any{"type": "string"}, "sort_order": map[string]any{"type": "integer"},
				}}}
			}
			operation["requestBody"] = map[string]any{"required": !strings.HasSuffix(route.path, "/publish"), "content": content}
		}
		methods, ok := paths[route.path].(map[string]any)
		if !ok {
			methods = map[string]any{}
			paths[route.path] = methods
		}
		methods[strings.ToLower(route.method)] = operation
	}
	document := map[string]any{"openapi": "3.0.3", "info": map[string]any{"title": "Announcements module", "version": Version()}, "paths": paths,
		"security": []any{map[string]any{"bearerAuth": []string{}}, map[string]any{"sessionCookie": []string{}}, map[string]any{"legacyAccessCookie": []string{}}},
		"components": map[string]any{"schemas": schemas, "securitySchemes": map[string]any{
			"bearerAuth":         map[string]any{"type": "http", "scheme": "bearer", "bearerFormat": "JWT"},
			"sessionCookie":      map[string]any{"type": "apiKey", "in": "cookie", "name": "sso_session"},
			"legacyAccessCookie": map[string]any{"type": "apiKey", "in": "cookie", "name": "access_token"},
		}},
	}
	result, err := json.Marshal(document)
	if err != nil {
		panic(err)
	}
	return result
}
func requestModel(method, path string) any {
	if path == "/admin/announcements" && method == "POST" {
		return createAnnouncementRequest{}
	}
	if path == "/admin/announcements/{id}" && method == "PUT" {
		return updateAnnouncementRequest{}
	}
	if path == "/admin/announcements/audience-preview" {
		return announcementAudiencePreviewRequest{}
	}
	if method == "POST" {
		switch {
		case strings.HasSuffix(path, "/publish"):
			return publishAnnouncementRequest{}
		case strings.HasSuffix(path, "/schedule"):
			return scheduleAnnouncementRequest{}
		case strings.HasSuffix(path, "/attachments"):
			return createAnnouncementAttachmentRequest{}
		}
	}
	if method == "PATCH" {
		switch {
		case strings.HasSuffix(path, "/pin"):
			return setPinnedRequest{}
		case strings.HasSuffix(path, "/priority"):
			return setPriorityRequest{}
		case strings.HasSuffix(path, "/attachments/{attachmentId}"):
			return updateAnnouncementAttachmentRequest{}
		}
	}
	return nil
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
		name := "announcements." + t.Name()
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
				if (t.Name()[0] >= 'a' && t.Name()[0] <= 'z' && field.Tag.Get("binding") == "required") || (!(t.Name()[0] >= 'a' && t.Name()[0] <= 'z') && !optional) {
					required = append(required, key)
				}
			}
			definitions[name] = map[string]any{"type": "object", "properties": properties, "required": required}
		}
		return map[string]any{"$ref": "#/components/schemas/" + name}
	}
	return map[string]any{}
}
