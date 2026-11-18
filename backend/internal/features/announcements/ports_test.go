package announcements

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/authz"
	rbacfeature "github.com/moh-sso-dashboard/internal/features/rbac"
	userfeature "github.com/moh-sso-dashboard/internal/features/users"
	"github.com/moh-sso-dashboard/internal/keycloak"
	"github.com/moh-sso-dashboard/internal/middleware"
	models "github.com/moh-sso-dashboard/internal/model"
)

type legacyUsersFixture struct {
	userfeature.UserRepository
	user  models.User
	ctx   context.Context
	calls int
}

func (f *legacyUsersFixture) ListUsers() ([]models.User, error) {
	f.calls++
	return []models.User{f.user}, nil
}
func (f *legacyUsersFixture) GetUserByID(uuid.UUID) (*models.User, error) {
	f.calls++
	return &f.user, nil
}
func (f *legacyUsersFixture) GetUsersByRealmRole(ctx context.Context, _ string) ([]models.User, error) {
	f.ctx = ctx
	return []models.User{f.user}, nil
}
func (f *legacyUsersFixture) GetUserClientRoles(ctx context.Context, _ string) ([]keycloak.UserClientRoleAssignment, error) {
	f.ctx = ctx
	return []keycloak.UserClientRoleAssignment{{Id: "internal-id", ClientID: "client-id"}}, nil
}

type legacyGroupsFixture struct {
	rbacfeature.Repository
	ctx context.Context
}

func (f *legacyGroupsFixture) GetGroup(ctx context.Context, _ string) (rbacfeature.Group, error) {
	f.ctx = ctx
	return rbacfeature.Group{}, nil
}
func (f *legacyGroupsFixture) ListGroupMembers(ctx context.Context, _ string) ([]rbacfeature.GroupMember, error) {
	f.ctx = ctx
	return []rbacfeature.GroupMember{{UserID: "id", Username: "name", Email: "mail@example.org"}}, nil
}

func TestHostAudienceAdaptersMapDataAndOptionalDependencies(t *testing.T) {
	ctx := context.Background()
	legacy := &legacyUsersFixture{user: models.User{ID: "id", Username: "user", Email: "mail", FirstName: "First", LastName: "Last", FullName: "Full", Enabled: true}}
	directory := NewUserDirectory(legacy)
	users, err := directory.ListUsers(ctx)
	expected := AudienceUser{ID: "id", Username: "user", Email: "mail", FirstName: "First", LastName: "Last", FullName: "Full", Enabled: true}
	if err != nil || len(users) != 1 || users[0] != expected {
		t.Fatalf("host user mapping changed: %+v %v", users, err)
	}
	detail, err := directory.GetUserByID(ctx, uuid.New())
	if err != nil || detail == nil || *detail != expected {
		t.Fatal("host detail mapping changed")
	}
	if _, err := directory.GetUsersByRealmRole(ctx, "user"); err != nil || legacy.ctx != ctx {
		t.Fatal("host role context lost")
	}
	access, err := directory.GetUserClientRoles(ctx, "id")
	if err != nil || legacy.ctx != ctx || !reflect.DeepEqual(access, []ClientAccess{{ID: "internal-id", ClientID: "client-id"}}) {
		t.Fatalf("client identifiers lost: %+v", access)
	}
	before := legacy.calls
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := directory.ListUsers(canceled); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled legacy list started")
	}
	if _, err := directory.GetUserByID(canceled, uuid.New()); !errors.Is(err, context.Canceled) || legacy.calls != before {
		t.Fatal("canceled legacy detail started")
	}
	groupRepo := &legacyGroupsFixture{}
	group := NewGroupDirectory(groupRepo)
	if err := group.ValidateGroup(ctx, "group"); err != nil || groupRepo.ctx != ctx {
		t.Fatal("host group validation failed")
	}
	members, err := group.ListGroupMembers(ctx, "group")
	if err != nil || len(members) != 1 || members[0].UserID != "id" || members[0].Username != "name" || members[0].Email != "mail@example.org" {
		t.Fatal("host group member mapping changed")
	}
	svc := NewService(nil, nil, nil, nil)
	svc.SetRBACRepository(nil)
	svc.SetHealthContextService(nil)
	if svc.ports.Users != nil || svc.ports.Groups != nil || svc.ports.HealthContexts != nil || svc.ports.Notifications != nil || svc.ports.Storage != nil {
		t.Fatal("optional dependencies changed")
	}
}

func TestAnnouncementPermissionsDenyBeforeDependencyAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	protected := router.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "fixture")
		c.Set(authz.ContextKey, authz.Context{UserID: "fixture"})
	}, middleware.RequireAuth())
	handler := NewHandler(nil, nil)
	RegisterProtectedRoutes(protected, handler)
	RegisterAdminRoutes(protected.Group("/admin"), handler, nil)
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/announcements/me"},
		{http.MethodPost, "/api/v1/admin/announcements/audience-preview"},
		{http.MethodPost, "/api/v1/admin/announcements/" + uuid.NewString() + "/publish"},
	} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s: expected denial before nil dependencies, got %d", tc.path, rec.Code)
		}
	}
}
