package router

import (
	"context"
	"database/sql"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/config"
	announcement "github.com/moh-sso-dashboard/internal/features/announcements"
	"github.com/moh-sso-dashboard/internal/middleware"
	"github.com/moh-sso-dashboard/internal/version"
	"github.com/moh-sso-dashboard/module"
)

const otherAnnouncementModules = "metrics,audit,geojson,report-scheduler"

type announcementRouteRepository struct {
	announcement.Repository
	downloads int
}

func (*announcementRouteRepository) ListPublicAnnouncements(context.Context, int32, int32) ([]announcement.Announcement, error) {
	return nil, nil
}
func (*announcementRouteRepository) ListActivePublished(context.Context, announcement.ListActivePublishedAnnouncementsParams) ([]announcement.Announcement, error) {
	return nil, nil
}
func (*announcementRouteRepository) ListForUser(context.Context, announcement.ListAnnouncementsForUserParams) ([]announcement.Announcement, error) {
	return nil, nil
}
func (*announcementRouteRepository) ListAdmin(context.Context, announcement.ListAnnouncementsAdminParams) ([]announcement.Announcement, error) {
	return nil, nil
}
func (r *announcementRouteRepository) GetAttachmentByID(_ context.Context, p announcement.GetAnnouncementAttachmentByIDParams) (announcement.AnnouncementAttachment, error) {
	r.downloads++
	return announcement.AnnouncementAttachment{ID: p.ID, AnnouncementID: p.AnnouncementID, StorageKey: "fixture", OriginalFileName: "notice.txt", ContentType: sql.NullString{String: "text/plain", Valid: true}}, nil
}

type announcementRouteStorage struct {
	announcement.AttachmentStorage
	downloads int
}

func (s *announcementRouteStorage) Download(context.Context, string) (io.ReadCloser, error) {
	s.downloads++
	return io.NopCloser(strings.NewReader("fixture attachment")), nil
}

func TestAnnouncementsInitializationAndVersion(t *testing.T) {
	for _, disabled := range []bool{false, true} {
		ids := otherAnnouncementModules
		if disabled {
			ids += ", announcements "
		}
		registry, err := NewModuleRegistry(HandlerSet{}, ids, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		stop, err := registry.Initialize(context.Background())
		if !disabled {
			if err == nil || !strings.Contains(err.Error(), "announcements handler is required") {
				t.Fatalf("missing handler: %v", err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		defer stop(context.Background())
		for _, item := range registry.Inventory() {
			if item.ID == "announcements" && (item.Enabled || item.Version != version.Get().Version) {
				t.Fatalf("inventory: %+v", item)
			}
		}
	}
}

// Exercise the production router as well as registry registration: a leftover
// legacy registrar must neither leak disabled routes nor duplicate enabled ones.
func TestAnnouncementsProductionRouterDisablement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var baseline map[string]bool
	for _, mode := range []string{"legacy", "enabled", "disabled"} {
		handlers := HandlerSet{Announcements: &announcement.Handler{}}
		var registry *module.Registry
		if mode != "legacy" {
			ids := otherAnnouncementModules
			if mode == "disabled" {
				ids += ",announcements"
				handlers.Announcements = nil
			}
			var err error
			registry, err = NewModuleRegistry(handlers, ids, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			stop, err := registry.Initialize(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			defer stop(context.Background())
			found := false
			for _, item := range registry.Inventory() {
				if item.ID == "announcements" {
					found = true
					if item.Enabled != (mode == "enabled") || item.Version != version.Get().Version {
						t.Fatalf("inventory: %+v", item)
					}
				}
			}
			if !found {
				t.Fatal("announcements missing from inventory")
			}
		}
		engine := SetupRouter(RouterDependencies{Config: &config.Config{}, Handlers: handlers, Modules: registry})
		actual := map[string]bool{}
		for _, route := range engine.Routes() {
			if strings.Contains(route.Path, "/announcements") {
				actual[route.Method+" "+route.Path] = true
			}
		}
		if mode == "legacy" {
			baseline = actual
			if len(baseline) != 26 {
				t.Fatalf("announcement route contract count: %d", len(baseline))
			}
			continue
		}
		if mode == "enabled" {
			if len(actual) != len(baseline) {
				t.Fatalf("registry route count: %d vs %d", len(actual), len(baseline))
			}
			for key := range baseline {
				if !actual[key] {
					t.Errorf("lost route %s", key)
				}
			}
			continue
		}
		if len(actual) != 0 {
			t.Fatalf("disabled routes leaked: %v", actual)
		}
		for key := range baseline {
			method, path, _ := strings.Cut(key, " ")
			path = strings.NewReplacer(":attachmentId", uuid.NewString(), ":id", uuid.NewString(), ":client_id", uuid.NewString(), ":role_name", "user").Replace(path)
			res := httptest.NewRecorder()
			engine.ServeHTTP(res, httptest.NewRequest(method, path, nil))
			if res.Code != 404 {
				t.Fatalf("disabled %s: %d", key, res.Code)
			}
		}
	}
}

func TestAnnouncementsRegistryPermissionsAndAttachments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, disabled := range []bool{false, true} {
		repo := &announcementRouteRepository{}
		storage := &announcementRouteStorage{}
		handler := announcement.NewHandler(announcement.NewServiceWithSettings(repo, announcement.ServicePorts{Storage: storage}, nil), nil)
		ids := otherAnnouncementModules
		if disabled {
			ids += ",announcements"
		}
		registry, err := NewModuleRegistry(HandlerSet{Announcements: handler}, ids, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		stop, err := registry.Initialize(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		defer stop(context.Background())
		for _, identity := range []string{"anonymous", "denied", "reader", "writer", "publisher"} {
			engine := gin.New()
			public := engine.Group("/api/v1")
			protected := engine.Group("/api/v1", func(c *gin.Context) {
				if identity == "anonymous" {
					return
				}
				actor := authz.Context{UserID: uuid.NewString()}
				switch identity {
				case "reader":
					actor.Permissions = []authz.Permission{authz.PermissionAnnouncementsRead}
				case "writer":
					actor.Permissions = []authz.Permission{authz.PermissionAnnouncementsRead, authz.PermissionAnnouncementsWrite}
				case "publisher":
					actor.Permissions = []authz.Permission{authz.PermissionAnnouncementsRead, authz.PermissionAnnouncementsPublish}
				}
				c.Set("user_id", actor.UserID)
				c.Set(authz.ContextKey, actor)
			}, middleware.RequireAuth())
			registry.RegisterRoutes(module.Routes{Public: public, Protected: protected, Admin: protected.Group("/admin")})
			id := uuid.NewString()
			attachment := uuid.NewString()
			for _, endpoint := range []struct{ method, path, permission string }{
				{"GET", "/announcements/public", "public"}, {"GET", "/announcements/active", "public"},
				{"GET", "/announcements/me", "read"}, {"GET", "/admin/announcements", "read"},
				{"GET", "/announcements/" + id + "/attachments/" + attachment + "/download", "read"},
				{"GET", "/admin/announcements/" + id + "/attachments/" + attachment + "/download", "read"},
				{"POST", "/admin/announcements", "write"},
				{"POST", "/admin/announcements/" + id + "/publish", "publish"},
			} {
				want := 200
				if endpoint.permission != "public" {
					want = 403
					if identity == "anonymous" {
						want = 401
					} else if endpoint.permission == "read" && identity != "denied" {
						want = 200
					} else if endpoint.permission == "write" && identity == "writer" {
						want = 400
					} else if endpoint.permission == "publish" && identity == "publisher" {
						want = 400
					}
				}
				if disabled {
					want = 404
				}
				beforeRepo, beforeStorage := repo.downloads, storage.downloads
				res := httptest.NewRecorder()
				engine.ServeHTTP(res, httptest.NewRequest(endpoint.method, "/api/v1"+endpoint.path, strings.NewReader("{")))
				if res.Code != want {
					t.Fatalf("disabled=%v identity=%s %s: got %d want %d: %s", disabled, identity, endpoint.path, res.Code, want, res.Body.String())
				}
				if strings.HasSuffix(endpoint.path, "/download") {
					if want == 200 {
						if res.Body.String() != "fixture attachment" || res.Header().Get("Content-Disposition") != `attachment; filename="notice.txt"` {
							t.Fatalf("download: %s", res.Body.String())
						}
					} else if repo.downloads != beforeRepo || storage.downloads != beforeStorage {
						t.Fatal("denied/disabled download touched dependencies")
					}
				}
			}
		}
	}
}
