package announcements

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func newTestService(repo Repository, ports ServicePorts) *Service {
	return NewServiceWithSettings(repo, ports, nil)
}

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
	svc := newTestService(nil, ServicePorts{Users: users})
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
	svc := newTestService(nil, ServicePorts{Users: users, Groups: groups})
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
	svc := newTestService(nil, ServicePorts{Users: users, HealthContexts: health})
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
}

type attachmentRepoFixture struct {
	Repository
	persistErr error
	count      int64
	creates    int
	params     CreateAnnouncementAttachmentParams
}

func (*attachmentRepoFixture) GetByID(context.Context, uuid.UUID) (Announcement, error) {
	return Announcement{}, nil
}
func (f *attachmentRepoFixture) CountAttachments(context.Context, uuid.UUID) (int64, error) {
	return f.count, nil
}
func (f *attachmentRepoFixture) CreateAttachment(_ context.Context, params CreateAnnouncementAttachmentParams) (AnnouncementAttachment, error) {
	f.creates++
	f.params = params
	return AnnouncementAttachment{ID: params.ID}, f.persistErr
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
		svc := newTestService(repo, ServicePorts{Storage: files})
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
	item Notification
}

func (f *publisherFixture) Publish(ctx context.Context, item Notification) error {
	f.ctx, f.item = ctx, item
	return errors.New("optional delivery failure")
}

func TestNotificationPortPreservesBestEffortPayloadAndDefaultAudience(t *testing.T) {
	publisher := &publisherFixture{}
	ctx := context.Background()
	svc := newTestService(nil, ServicePorts{Notifications: publisher})
	notification := Notification{Type: "fixture", Metadata: []byte(`{"id":"fixture"}`), Deliveries: []NotificationDeliveryRequest{{Channel: NotificationChannelEmail, TemplateName: "announcement"}}}
	svc.notify(ctx, notification)
	notification.TargetRole = "admin"
	if publisher.ctx != ctx || !reflect.DeepEqual(publisher.item, notification) {
		t.Fatal("notification envelope or delivery policy changed")
	}
}

func TestAttachmentPolicyRejectsBeforeStorageAndPersistence(t *testing.T) {
	for _, scenario := range []string{"count", "extension", "bytes", "inline", "name", "id"} {
		t.Run(scenario, func(t *testing.T) {
			settings := &Settings{}
			settings.Announcement.MaxAttachments = 1
			settings.Announcement.MaxAttachmentBytes = 3
			settings.Announcement.AllowedAttachmentTypes = "pdf"
			repo := &attachmentRepoFixture{}
			files := &attachmentStorageFixture{}
			input := UploadAnnouncementAttachmentInput{
				FileName: "report.pdf", ContentType: "application/pdf",
				Reader: strings.NewReader("ok"), Size: 2,
			}
			id := uuid.New()
			switch scenario {
			case "count":
				repo.count = 1
			case "extension":
				input.FileName = "report.exe"
			case "bytes":
				input.Reader = strings.NewReader("over limit")
				input.Size = 10
			case "inline":
				input.Inline = true
			case "name":
				input.FileName = ""
			case "id":
				id = uuid.Nil
			}
			service := NewServiceWithSettings(repo, ServicePorts{Storage: files}, settings)
			if _, err := service.UploadAttachment(context.Background(), id, input); err == nil {
				t.Fatal("invalid attachment accepted")
			}
			if repo.creates != 0 || files.key != "" || files.deleted != "" {
				t.Fatal("rejected upload touched persistence or blob storage")
			}
		})
	}
}

func TestMissingEmailNormalizerRejectsRequiredAttachments(t *testing.T) {
	service := NewServiceWithSettings(nil, ServicePorts{}, nil)
	attachment := Attachment{FileName: "notice.pdf", Data: []byte("data")}
	_, err := service.normalizeAnnouncementEmailOptions(
		AnnouncementEmailOptions{Attachments: []Attachment{attachment}, IncludeAttachmentsInEmail: true})
	if err == nil || !strings.Contains(err.Error(), "normalizer") {
		t.Fatalf("missing normalizer must fail closed when required: %v", err)
	}
	for _, options := range []AnnouncementEmailOptions{
		{},
		{Attachments: []Attachment{attachment}, IncludeAttachmentsInEmail: false},
	} {
		if _, err := service.normalizeAnnouncementEmailOptions(options); err != nil {
			t.Fatalf("normalization is optional without included email attachments: %v", err)
		}
	}
}
