package announcements

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/authz"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	rbacfeature "github.com/moh-sso-dashboard/internal/features/rbac"
	userfeature "github.com/moh-sso-dashboard/internal/features/users"
	"github.com/moh-sso-dashboard/internal/keycloak"
	"github.com/moh-sso-dashboard/internal/middleware"
	models "github.com/moh-sso-dashboard/internal/model"
)

type audienceUsersFixture struct {
	UserDirectory
	users  []AudienceUser
	byID   map[uuid.UUID]*AudienceUser
	access []ClientAccess
	ctx    context.Context
	role   string
	err    error
}

func (f *audienceUsersFixture) ListUsers(ctx context.Context) ([]AudienceUser, error) {
	f.ctx = ctx
	return f.users, f.err
}
func (f *audienceUsersFixture) GetUserByID(ctx context.Context, id uuid.UUID) (*AudienceUser, error) {
	f.ctx = ctx
	return f.byID[id], f.err
}
func (f *audienceUsersFixture) GetUsersByRealmRole(ctx context.Context, role string) ([]AudienceUser, error) {
	f.ctx = ctx
	f.role = role
	return f.users, f.err
}
func (f *audienceUsersFixture) GetUserClientRoles(ctx context.Context, _ string) ([]ClientAccess, error) {
	f.ctx = ctx
	return f.access, f.err
}

type audienceGroupsFixture struct {
	ids     []string
	members []AudienceGroupMember
	ctx     context.Context
	err     error
}

func (f *audienceGroupsFixture) ValidateGroup(ctx context.Context, id string) error {
	f.ctx = ctx
	f.ids = append(f.ids, id)
	return f.err
}
func (f *audienceGroupsFixture) ListGroupMembers(ctx context.Context, _ string) ([]AudienceGroupMember, error) {
	f.ctx = ctx
	return f.members, f.err
}

type healthAudienceFixture struct {
	ctx         context.Context
	actor       string
	ids         []uuid.UUID
	descendants bool
	stored      bool
	users       []string
	err         error
}

func (f *healthAudienceFixture) ResolveAudienceUserIDs(ctx context.Context, actor string, ids []uuid.UUID, descendants bool) ([]string, error) {
	f.ctx, f.actor, f.ids, f.descendants = ctx, actor, ids, descendants
	return f.users, f.err
}
func (f *healthAudienceFixture) ListAudienceUserIDs(ctx context.Context, ids []uuid.UUID, descendants bool) ([]string, error) {
	f.ctx, f.ids, f.descendants, f.stored = ctx, ids, descendants, true
	return f.users, f.err
}

func TestAudiencePortsPreserveUserFilteringAndDeduplication(t *testing.T) {
	ctx := context.Background()
	users := &audienceUsersFixture{users: []AudienceUser{
		{ID: uuid.NewString(), Enabled: true, Email: " Same@example.org ", FirstName: "First", LastName: "Last"},
		{ID: uuid.NewString(), Enabled: true, Email: "same@example.org", FullName: "Other"},
		{ID: uuid.NewString(), Enabled: false, Email: "disabled@example.org"},
		{ID: uuid.NewString(), Enabled: true, Email: " "},
	}}
	svc := NewServiceWithPorts(nil, ServicePorts{Users: users})
	recipients, err := svc.resolveAllUsersEmailRecipients(ctx)
	if err != nil || len(recipients) != 1 || !strings.EqualFold(recipients[0].Email, "same@example.org") {
		t.Fatalf("recipients=%+v error=%v", recipients, err)
	}
	if users.ctx != ctx {
		t.Fatal("request context not forwarded")
	}
	recipients, err = svc.resolveUsersByRealmRoles(ctx, []string{" ", " manager "})
	if err != nil || len(recipients) != 1 || users.role != "manager" {
		t.Fatalf("realm audience changed: %+v %v", recipients, err)
	}
	for _, access := range []ClientAccess{{ID: " CLIENT-UUID "}, {ClientID: "client-uuid"}} {
		users.access = []ClientAccess{access}
		if !userHasAnyAnnouncementClientAccess(users.access, map[string]struct{}{"client-uuid": {}}) {
			t.Fatal("client identifier compatibility changed")
		}
	}
	cause := errors.New("directory unavailable")
	users.err = cause
	if _, err := svc.resolveAllUsersEmailRecipients(ctx); !errors.Is(err, cause) {
		t.Fatalf("directory error lost: %v", err)
	}
}

func TestGroupPortPreservesValidationLookupAndFallback(t *testing.T) {
	ctx := context.Background()
	groupID := uuid.New()
	disabledID := uuid.New()
	groups := &audienceGroupsFixture{members: []AudienceGroupMember{
		{UserID: disabledID.String(), Email: "disabled@example.org"},
		{UserID: "external-id", Username: "External", Email: "external@example.org"},
		{UserID: "external-id", Username: "External", Email: "EXTERNAL@example.org"},
	}}
	users := &audienceUsersFixture{byID: map[uuid.UUID]*AudienceUser{
		disabledID: {ID: disabledID.String(), Enabled: false, Email: "disabled@example.org"},
	}}
	svc := NewServiceWithPorts(nil, ServicePorts{Users: users, Groups: groups})
	if err := svc.ValidateGroupAudience(ctx, []uuid.UUID{groupID, groupID}); err != nil || len(groups.ids) != 1 {
		t.Fatalf("group validation changed: %v", err)
	}
	if err := svc.ValidateGroupAudience(ctx, []uuid.UUID{uuid.Nil}); err == nil {
		t.Fatal("empty group ID accepted")
	}
	recipients, err := svc.resolveUsersByRBACGroups(ctx, []uuid.UUID{uuid.Nil, groupID})
	if err != nil || len(recipients) != 1 || recipients[0].ID != uuid.Nil || !strings.EqualFold(recipients[0].Email, "external@example.org") {
		t.Fatalf("group recipients=%+v error=%v", recipients, err)
	}
	if users.ctx != ctx || groups.ctx != ctx {
		t.Fatal("group lookup lost request context")
	}
	cause := errors.New("group lookup failed")
	groups.err = cause
	if err := svc.ValidateGroupAudience(ctx, []uuid.UUID{groupID}); !errors.Is(err, cause) {
		t.Fatalf("group error lost: %v", err)
	}
}

func TestHealthAudiencePortPreservesActorScopeAndStoredExpansion(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()
	node := uuid.New()
	health := &healthAudienceFixture{users: []string{userID}}
	users := &audienceUsersFixture{users: []AudienceUser{
		{ID: userID, Enabled: true, Email: "allowed@example.org"},
		{ID: uuid.NewString(), Enabled: true, Email: "outside@example.org"},
	}}
	svc := NewServiceWithPorts(nil, ServicePorts{Users: users, HealthContexts: health})
	recipients, err := svc.ResolveHealthContextRecipients(ctx, "actor-id", []uuid.UUID{node}, true)
	if err != nil || len(recipients) != 1 || recipients[0].Email != "allowed@example.org" {
		t.Fatalf("scoped audience changed: %+v %v", recipients, err)
	}
	if health.actor != "actor-id" || !health.descendants || health.ctx != ctx || !reflect.DeepEqual(health.ids, []uuid.UUID{node}) || health.stored {
		t.Fatal("actor/context scope not forwarded")
	}
	if _, err := svc.resolveStoredHealthContextRecipients(ctx, []AnnouncementHealthContextAudience{{ContextNodeID: node, IncludeDescendants: true}}); err != nil || !health.stored {
		t.Fatalf("stored audience path changed: %v", err)
	}
	cause := errors.New("scope denied")
	health.err = cause
	if _, err := svc.ResolveHealthContextRecipients(ctx, "actor-id", []uuid.UUID{node}, false); !errors.Is(err, cause) {
		t.Fatalf("scope error lost: %v", err)
	}
	svc.SetHealthContextService(nil)
	if svc.healthContexts != nil {
		t.Fatal("nil health context became a typed-nil dependency")
	}
}

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
	if svc.users != nil || svc.groups != nil || svc.healthContexts != nil || svc.notifications != nil || svc.storage != nil {
		t.Fatal("optional dependencies changed")
	}
}

type attachmentRepoFixture struct {
	Repository
	persistErr error
	creates    int
	params     db.CreateAnnouncementAttachmentParams
}

func (*attachmentRepoFixture) GetByID(context.Context, uuid.UUID) (db.Announcement, error) {
	return db.Announcement{}, nil
}
func (*attachmentRepoFixture) CountAttachments(context.Context, uuid.UUID) (int64, error) {
	return 0, nil
}
func (f *attachmentRepoFixture) CreateAttachment(_ context.Context, params db.CreateAnnouncementAttachmentParams) (db.AnnouncementAttachment, error) {
	f.creates++
	f.params = params
	return db.AnnouncementAttachment{ID: params.ID}, f.persistErr
}

type attachmentStorageFixture struct {
	ctx          context.Context
	key, deleted string
	data         []byte
	uploadErr    error
}

func (f *attachmentStorageFixture) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	f.ctx, f.key = ctx, key
	f.data, _ = io.ReadAll(r)
	if size != int64(len(f.data)) || contentType != "application/pdf" {
		return errors.New("storage metadata mismatch")
	}
	return f.uploadErr
}
func (*attachmentStorageFixture) Download(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(nil)), nil
}
func (f *attachmentStorageFixture) Delete(_ context.Context, key string) error {
	f.deleted = key
	return nil
}

func TestAttachmentPortPreservesChecksumAndFailureCleanup(t *testing.T) {
	for _, persistFails := range []bool{false, true} {
		repo := &attachmentRepoFixture{}
		cause := errors.New("persistence failed")
		if persistFails {
			repo.persistErr = cause
		}
		files := &attachmentStorageFixture{}
		svc := NewServiceWithPorts(repo, ServicePorts{Storage: files})
		ctx := context.Background()
		announcement := uuid.New()
		actor := uuid.New()
		_, err := svc.UploadAttachment(ctx, announcement, UploadAnnouncementAttachmentInput{
			FileName: "report.pdf", ContentType: "application/pdf", Reader: strings.NewReader("data"), Size: 4, UploadedBy: actor,
		})
		sum := sha256.Sum256([]byte("data"))
		if files.ctx != ctx || !strings.HasPrefix(files.key, "announcements/"+announcement.String()+"/") ||
			string(files.data) != "data" || repo.params.Checksum.String != hex.EncodeToString(sum[:]) ||
			repo.params.UploadedBy.UUID != actor || !repo.params.UploadedBy.Valid {
			t.Fatal("attachment payload, key or checksum changed")
		}
		if persistFails {
			if !errors.Is(err, cause) || files.deleted != files.key {
				t.Fatalf("orphan cleanup lost: %v", err)
			}
		} else if err != nil || files.deleted != "" {
			t.Fatalf("successful upload removed blob: %v", err)
		}
	}
}

type publisherFixture struct {
	ctx  context.Context
	item models.Notification
}

func (f *publisherFixture) Notify(ctx context.Context, item models.Notification) (*models.Notification, error) {
	f.ctx, f.item = ctx, item
	return nil, errors.New("optional delivery failure")
}

func TestNotificationPortPreservesBestEffortPayloadAndDefaultAudience(t *testing.T) {
	publisher := &publisherFixture{}
	ctx := context.Background()
	svc := NewServiceWithPorts(nil, ServicePorts{Notifications: publisher})
	notification := models.Notification{Type: "fixture", Metadata: []byte(`{"id":"fixture"}`), Deliveries: []models.NotificationDeliveryRequest{{Channel: models.NotificationChannelEmail, TemplateName: "announcement"}}}
	svc.notify(ctx, notification)
	notification.TargetRole = "admin"
	if publisher.ctx != ctx || !reflect.DeepEqual(publisher.item, notification) {
		t.Fatal("notification envelope or delivery policy changed")
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
