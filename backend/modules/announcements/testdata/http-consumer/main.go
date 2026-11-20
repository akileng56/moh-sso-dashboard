// Public-API HTTP consumer; runs outside the host source tree against PostgreSQL.
package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	a "github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements"
)

func must[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}
func check(ok bool, msg string) {
	if !ok {
		panic(msg)
	}
}
func ok(e error) {
	if e != nil {
		panic(e)
	}
}

type identityKey struct{}
type identity struct {
	actor       a.Actor
	admin       bool
	permissions []string
}
type capabilities struct {
	mu            sync.Mutex
	blobs         map[string][]byte
	events        []string
	notifications []a.Notification
	actor         uuid.UUID
	contextCalls  int
}

func (c *capabilities) Upload(ctx context.Context, key string, r io.Reader, _ int64, _ string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.blobs[key] = body
	return nil
}
func (c *capabilities) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	body, ok := c.blobs[key]
	if !ok {
		return nil, fmt.Errorf("missing blob")
	}
	return io.NopCloser(bytes.NewReader(body)), nil
}
func (c *capabilities) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.blobs, key)
	return nil
}
func (c *capabilities) Log(_ context.Context, id uuid.NullUUID, event string, _ map[string]any) error {
	check(id.Valid && id.UUID == c.actor, "unverified audit actor")
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, event)
	return nil
}
func (c *capabilities) Publish(_ context.Context, item a.Notification) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.notifications = append(c.notifications, item)
	return nil
}
func (c *capabilities) NormalizeAttachments(items []a.Attachment) ([]a.Attachment, error) {
	return items, nil
}
func (c *capabilities) ListUsers(context.Context) ([]a.AudienceUser, error) {
	return []a.AudienceUser{{ID: c.actor.String(), Enabled: true, Email: "fixture@example.org"}}, nil
}
func (c *capabilities) GetUserByID(context.Context, uuid.UUID) (*a.AudienceUser, error) {
	return &a.AudienceUser{ID: c.actor.String(), Enabled: true, Email: "fixture@example.org"}, nil
}
func (c *capabilities) GetUsersByRealmRole(ctx context.Context, _ string) ([]a.AudienceUser, error) {
	return c.ListUsers(ctx)
}
func (c *capabilities) GetUserClientRoles(context.Context, string) ([]a.ClientAccess, error) {
	return nil, nil
}
func (c *capabilities) ValidateGroup(context.Context, string) error { return nil }
func (c *capabilities) ListGroupMembers(context.Context, string) ([]a.AudienceGroupMember, error) {
	return []a.AudienceGroupMember{{UserID: c.actor.String(), Email: "fixture@example.org"}}, nil
}
func (c *capabilities) ResolveAudienceUserIDs(_ context.Context, actor string, _ []uuid.UUID, _ bool) ([]string, error) {
	check(actor == c.actor.String(), "health preview lost verified actor")
	c.contextCalls++
	return []string{c.actor.String()}, nil
}
func (c *capabilities) ListAudienceUserIDs(context.Context, []uuid.UUID, bool) ([]string, error) {
	return []string{c.actor.String()}, nil
}

type harness struct {
	module  *a.Module
	who     identity
	covered map[string]bool
}

func (h *harness) request(method, path string, body io.Reader, contentType string, want int) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	req = req.WithContext(context.WithValue(req.Context(), identityKey{}, h.who))
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	h.module.ServeHTTP(w, req)
	if w.Code != want {
		panic(fmt.Sprintf("%s %s: got %d, want %d: %s", method, path, w.Code, want, w.Body.String()))
	}
	if want >= 200 && want < 300 {
		for _, pattern := range h.module.Patterns() {
			parts := strings.SplitN(pattern, " ", 2)
			if parts[0] != method {
				continue
			}
			template := strings.Split(parts[1], "/")
			actual := strings.Split(strings.SplitN(path, "?", 2)[0], "/")
			if len(template) != len(actual) {
				continue
			}
			match := true
			for i, part := range template {
				if !strings.HasPrefix(part, "{") && part != actual[i] {
					match = false
				}
			}
			if match {
				h.covered[pattern] = true
			}
		}
	}
	return w
}
func (h *harness) json(method, path string, body any, want int) map[string]any {
	var data io.Reader
	if body != nil {
		data = bytes.NewReader(must(json.Marshal(body)))
	}
	w := h.request(method, path, data, "application/json", want)
	var envelope map[string]any
	ok(json.Unmarshal(w.Body.Bytes(), &envelope))
	check(envelope["meta"].(map[string]any)["requestId"] == "http-fixture", "request ID lost")
	return envelope
}
func (h *harness) create(title, audience string, extra map[string]any) string {
	body := map[string]any{"title": title, "message": "HTTP body", "level": "INFO", "audience_type": audience}
	for k, v := range extra {
		body[k] = v
	}
	return h.json("POST", "/admin/announcements", body, 201)["data"].(map[string]any)["id"].(string)
}
func main() {
	pool := must(sql.Open("postgres", os.Getenv("ANNOUNCEMENTS_TEST_DSN")))
	defer pool.Close()
	actor := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	client := uuid.New()
	caps := &capabilities{actor: actor, blobs: map[string][]byte{}}
	settings := &a.Settings{}
	settings.SMS.Enabled = true
	settings.SMS.MaxLength = 160
	settings.Announcement.AllowedAttachmentTypes = "txt"
	deps := a.Dependencies{
		Database: pool, Settings: settings, Audit: caps,
		Ports: a.ServicePorts{Users: caps, Groups: caps, HealthContexts: caps, Notifications: caps, Storage: caps, EmailAttachments: caps},
		Actor: func(r *http.Request) (a.Actor, error) {
			who, _ := r.Context().Value(identityKey{}).(identity)
			return who.actor, nil
		},
		RequireAdmin: func(r *http.Request) error {
			who, _ := r.Context().Value(identityKey{}).(identity)
			if !who.admin {
				return a.ErrForbidden
			}
			return nil
		},
		AuthorizeAny: func(r *http.Request, permissions ...string) error {
			who, _ := r.Context().Value(identityKey{}).(identity)
			for _, need := range permissions {
				for _, have := range who.permissions {
					if have == need {
						return nil
					}
				}
			}
			return a.ErrForbidden
		},
		Limit: func(_ *http.Request, policy string, limit int, window time.Duration) (http.Header, error) {
			check(window == time.Minute, "rate window")
			check((policy == a.WritePolicy && limit == 60) || (policy == a.PublishPolicy && limit == 30), "rate policy")
			return http.Header{"X-RateLimit-Policy": []string{policy}}, nil
		},
		RequestID: func(*http.Request) string { return "http-fixture" },
	}
	full := identity{actor: a.Actor{UserID: actor, RealmRoles: []string{"admin", "manager"}, ClientIDs: []string{client.String()}}, admin: true, permissions: []string{a.ReadPermission, a.WritePermission, a.PublishPermission}}
	h := &harness{module: must(a.NewModule(deps)), who: full, covered: map[string]bool{}}
	id := h.create("HTTP fixture", "ALL_USERS", map[string]any{"notify_by_email": true, "notify_by_sms": true, "sms_message": "Fixture SMS"})
	base := "/admin/announcements/" + id
	h.json("GET", "/admin/announcements", nil, 200)
	h.json("GET", "/admin/announcements/stats", nil, 200)
	h.json("GET", base, nil, 200)
	h.json("PUT", base, map[string]any{"title": "Updated HTTP", "message": "Body", "level": "WARNING", "audience_type": "ALL_USERS", "notify_by_email": true, "notify_by_sms": true, "sms_message": "SMS"}, 200)
	h.json("PATCH", base+"/pin", map[string]any{"is_pinned": true}, 200)
	h.json("PATCH", base+"/priority", map[string]any{"priority": 5}, 200)
	h.json("POST", "/admin/announcements/audience-preview", map[string]any{"audience_type": "SPECIFIC_HEALTH_CONTEXTS", "health_context_ids": []string{"30000000-0000-0000-0000-000000000001"}}, 200)
	check(caps.contextCalls == 1, "health capability unused")
	attachment := h.json("POST", base+"/attachments", map[string]any{"file_name": "note.txt", "content_type": "text/plain", "data_base64": "aGVsbG8="}, 201)["data"].(map[string]any)
	attachmentID := attachment["id"].(string)
	attachmentBase := base + "/attachments/" + attachmentID
	h.json("GET", base+"/attachments", nil, 200)
	h.request("GET", attachmentBase+"/download", nil, "", 200)
	h.json("PATCH", attachmentBase, map[string]any{"include_in_email": false}, 200)
	h.json("POST", base+"/schedule", map[string]any{"publish_at": time.Now().Add(time.Hour).Format(time.RFC3339)}, 200)
	h.json("POST", base+"/draft", nil, 200)
	h.json("PUT", base, map[string]any{"title": "Updated HTTP", "message": "Body", "level": "INFO", "audience_type": "ALL_USERS", "publish_at": time.Now().Add(-time.Minute).Format(time.RFC3339), "notify_by_email": true, "notify_by_sms": true, "sms_message": "SMS"}, 200)
	h.json("POST", base+"/publish", nil, 200)
	h.request("GET", "/announcements/"+id+"/attachments/"+attachmentID+"/download", nil, "", 200)
	for _, path := range []string{"/announcements/public", "/announcements/active", "/announcements/me", "/announcements/user", "/announcements/role/manager", "/announcements/client/" + client.String()} {
		h.json("GET", path, nil, 200)
	}
	h.json("POST", base+"/archive", nil, 200)
	h.json("DELETE", base, nil, 200)
	h.json("POST", base+"/restore", nil, 200)

	// Multipart and JSON use the same package-owned upload/cleanup policy.
	var multipartBody bytes.Buffer
	writer := multipart.NewWriter(&multipartBody)
	file := must(writer.CreateFormFile("file", "multipart.txt"))
	_, err := file.Write([]byte("multipart bytes"))
	ok(err)
	ok(writer.Close())
	multi := h.request("POST", base+"/attachments", &multipartBody, writer.FormDataContentType(), 201)
	check(strings.Contains(multi.Body.String(), "multipart.txt"), "multipart upload")
	h.json("DELETE", attachmentBase, nil, 200)
	h.request("GET", attachmentBase+"/download", nil, "", 404)

	// Private audiences must never appear in public feeds or other users' downloads.
	private := h.create("PRIVATE-MARKER", "SPECIFIC_USERS", map[string]any{"status": "PUBLISHED", "user_ids": []string{uuid.NewString()}})
	privateBase := "/admin/announcements/" + private
	privateAttachment := h.json("POST", privateBase+"/attachments", map[string]any{"file_name": "secret.txt", "content_type": "text/plain", "data_base64": "c2VjcmV0"}, 201)["data"].(map[string]any)["id"].(string)
	h.who = identity{actor: a.Actor{UserID: actor}, permissions: []string{a.ReadPermission}}
	for _, path := range []string{"/announcements/public", "/announcements/active", "/announcements/me", "/announcements/user"} {
		result := h.json("GET", path, nil, 200)
		check(!strings.Contains(string(must(json.Marshal(result))), "PRIVATE-MARKER"), "audience leak: "+path)
	}
	h.request("GET", "/announcements/"+private+"/attachments/"+privateAttachment+"/download", nil, "", 404)
	h.json("GET", "/announcements/me?role=admin", nil, 403)
	h.json("GET", "/announcements/client/"+client.String(), nil, 403)
	h.who = full
	own := h.create("OWN-MARKER", "SPECIFIC_USERS", map[string]any{"status": "PUBLISHED", "user_ids": []string{actor.String()}})
	_ = own
	h.create("ROLE-MARKER", "SPECIFIC_ROLES", map[string]any{"status": "PUBLISHED", "role_names": []string{"manager"}})
	h.create("CLIENT-MARKER", "SPECIFIC_CLIENTS", map[string]any{"status": "PUBLISHED", "client_ids": []string{client.String()}})
	h.create("GROUP-MARKER", "SPECIFIC_GROUPS", map[string]any{"status": "PUBLISHED", "group_ids": []string{"20000000-0000-0000-0000-000000000001"}})
	h.create("HEALTH-MARKER", "SPECIFIC_HEALTH_CONTEXTS", map[string]any{"status": "PUBLISHED", "health_context_ids": []string{"30000000-0000-0000-0000-000000000001"}})
	mine := string(must(json.Marshal(h.json("GET", "/announcements/me", nil, 200))))
	for _, marker := range []string{"OWN-MARKER", "ROLE-MARKER", "CLIENT-MARKER", "GROUP-MARKER", "HEALTH-MARKER"} {
		check(strings.Contains(mine, marker), "missing eligible audience: "+marker)
	}
	check(!strings.Contains(mine, "PRIVATE-MARKER"), "unrelated user audience leak")
	h.who = identity{actor: a.Actor{UserID: uuid.New()}, permissions: []string{a.ReadPermission}}
	other := string(must(json.Marshal(h.json("GET", "/announcements/me", nil, 200))))
	for _, marker := range []string{"OWN-MARKER", "ROLE-MARKER", "CLIENT-MARKER", "GROUP-MARKER", "HEALTH-MARKER", "PRIVATE-MARKER"} {
		check(!strings.Contains(other, marker), "other actor leak: "+marker)
	}
	h.who = identity{}
	forged := httptest.NewRequest("GET", "/announcements/me", nil)
	forged.Header.Set("X-User-ID", actor.String())
	forged.Header.Set("X-Roles", "admin")
	denied := httptest.NewRecorder()
	h.module.ServeHTTP(denied, forged)
	check(denied.Code == 401, "forged identity trusted")

	h.who = full
	// Provider payloads and audit events still flow through the package service.
	check(len(caps.events) >= 12, "audit events missing")
	email, sms := false, false
	for _, n := range caps.notifications {
		for _, d := range n.Deliveries {
			if d.Channel == a.NotificationChannelEmail {
				email = true
			}
			if d.Channel == a.NotificationChannelSMS {
				sms = true
			}
		}
	}
	check(email && sms, "delivery capabilities not exercised")
	check(len(h.covered) == 26, fmt.Sprintf("only %d/26 HTTP operations exercised", len(h.covered)))
	ok(pool.Ping())
	fmt.Println("PASS: standalone Announcements HTTP/PostgreSQL consumer; 26 routes, verified access, audience isolation, attachment streams, audit and email/SMS capabilities")
}
