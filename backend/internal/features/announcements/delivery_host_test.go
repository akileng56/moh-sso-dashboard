package announcements

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/config"
	emailutil "github.com/moh-sso-dashboard/internal/email"
	models "github.com/moh-sso-dashboard/internal/model"
)

type hostSenderFixture struct {
	ctx  context.Context
	item models.Notification
	err  error
}

func (f *hostSenderFixture) Notify(ctx context.Context, item models.Notification) (*models.Notification, error) {
	f.ctx, f.item = ctx, item
	return nil, f.err
}

func TestHostDeliveryConversionPreservesEveryEnvelopeFieldAndAttachment(t *testing.T) {
	when := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	attachment := Attachment{FileName: "file.pdf", ContentType: "application/pdf", Path: "fixture/path", Data: []byte("bytes"), DataBase64: "YQ==", Inline: true, ContentID: "cid"}
	item := Notification{ID: uuid.New(), Type: "event", Title: "title", Message: "message", Severity: "info",
		TargetRole: "manager", ClientID: "client", UserID: "user", Metadata: json.RawMessage(`{"id":1}`), Read: true, CreatedAt: when}
	for _, channel := range []NotificationChannel{NotificationChannelInApp, NotificationChannelEmail, NotificationChannelSMS, NotificationChannelWebhook} {
		item.Deliveries = append(item.Deliveries, NotificationDeliveryRequest{Channel: channel,
			Recipient: map[string]any{"to": "fixture@example.org"}, TemplateName: "announcement",
			TemplateData: map[string]any{"title": "title"}, Payload: map[string]any{"data": true},
			Attachments: []Attachment{attachment}, ScheduledAt: &when, MaxAttempts: 4})
	}
	sender := &hostSenderFixture{err: errors.New("delivery failed")}
	ctx := context.Background()
	if err := NewNotificationPublisher(sender).Publish(ctx, item); !errors.Is(err, sender.err) {
		t.Fatalf("delivery cause lost: %v", err)
	}
	got, want := sender.item, item
	// Compare wire representations as well as byte data excluded from JSON.
	hostJSON, _ := json.Marshal(got)
	portableJSON, _ := json.Marshal(want)
	if string(hostJSON) != string(portableJSON) || sender.ctx != ctx || len(got.Deliveries) != len(item.Deliveries) {
		t.Fatalf("delivery envelope changed: host=%s portable=%s", hostJSON, portableJSON)
	}
	for i := range got.Deliveries {
		if !reflect.DeepEqual(got.Deliveries[i].Attachments[0], models.Attachment(attachment)) {
			t.Fatal("attachment bytes/path/inline fields lost")
		}
	}
	for _, deliveries := range [][]NotificationDeliveryRequest{nil, {}} {
		item.Deliveries = deliveries
		if (toHostNotification(item).Deliveries == nil) != (deliveries == nil) {
			t.Fatal("nil/empty delivery distinction lost")
		}
	}
	if NewNotificationPublisher(nil) != nil {
		t.Fatal("nil sender became an enabled publisher")
	}
}

func TestHostEmailNormalizerMatchesExistingPolicyAndPortableCoreFailsClosed(t *testing.T) {
	cfg := &config.Config{}
	cfg.Email.AllowedAttachmentTypes = "pdf"
	cfg.Email.MaxAttachmentBytes = 20
	items := []Attachment{{FileName: " test.pdf ", DataBase64: "YQ==", Inline: true, ContentID: " cid "}}
	got, err := NewEmailAttachmentNormalizer(cfg).NormalizeAttachments(items)
	original, originalErr := emailutil.NormalizeAttachments(cfg, toHostAttachments(items))
	if (err == nil) != (originalErr == nil) || !reflect.DeepEqual(got, fromHostAttachments(original)) {
		t.Fatalf("normalization changed: %+v %v", got, err)
	}
	invalid := []Attachment{{FileName: "x.pdf", Path: "fixture/path", DataBase64: "YQ=="}}
	if _, err := NewEmailAttachmentNormalizer(cfg).NormalizeAttachments(invalid); err == nil || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("conflicting attachment sources accepted: %v", err)
	}
	cfg.Email.MaxAttachmentBytes = 1
	if _, err := NewEmailAttachmentNormalizer(cfg).NormalizeAttachments([]Attachment{{FileName: "x.pdf", Data: []byte("too large")}}); err == nil {
		t.Fatal("host attachment limits bypassed")
	}
	core := NewServiceWithSettings(nil, ServicePorts{}, nil)
	if _, err := core.normalizeAnnouncementEmailOptions(AnnouncementEmailOptions{Attachments: items, IncludeAttachmentsInEmail: true}); err == nil {
		t.Fatal("missing email policy silently accepted")
	}
	host := NewServiceWithPorts(nil, ServicePorts{}, cfg)
	if host.emailAttachments == nil {
		t.Fatal("host constructor omitted normalization adapter")
	}
}

func TestSettingsExcludeSecretsAndPreserveDecisionValues(t *testing.T) {
	cfg := &config.Config{DBPassword: "DO-NOT-COPY", S3SecretAccessKey: "DO-NOT-COPY", AppBaseURL: "https://api.example.org", StorageProvider: "s3"}
	cfg.SMS.Enabled = true
	cfg.SMS.MaxLength = 140
	cfg.Announcement.MaxAttachments = 3
	cfg.Announcement.MaxAttachmentBytes = 1024
	cfg.Announcement.AllowedAttachmentTypes = "pdf"
	cfg.Email.AllowedAttachmentTypes = "pdf,png"
	cfg.Notification.PlatformName = "Portal"
	cfg.Notification.SystemAdminName = "Operator"
	cfg.Notification.SystemAdminEmail = "operator@example.org"
	cfg.Notification.AdminDashboardURL = "https://portal.example.org/admin"
	cfg.FrontendBaseURL = "https://portal.example.org"
	cfg.FrontendRedirectURI = "https://portal.example.org/portal"
	settings := AnnouncementSettings(cfg)
	data, _ := json.Marshal(settings)
	if strings.Contains(string(data), "DO-NOT-COPY") || settings.SMS.MaxLength != 140 ||
		settings.Announcement.MaxAttachmentBytes != 1024 || settings.Notification.PlatformName != "Portal" ||
		settings.AppBaseURL != cfg.AppBaseURL || settings.StorageProvider != "s3" ||
		settings.Notification.SystemAdminEmail != cfg.Notification.SystemAdminEmail ||
		settings.Email.AllowedAttachmentTypes != cfg.Email.AllowedAttachmentTypes {
		t.Fatal("settings lost decisions or copied credentials")
	}
	if AnnouncementSettings(nil) != nil {
		t.Fatal("nil settings changed default behavior")
	}
	for _, event := range []announcementEvent{AnnouncementCreated, AnnouncementUpdated, AnnouncementDeleted, AnnouncementArchived, AnnouncementScheduled, AnnouncementPublished, AnnouncementDrafted} {
		host := models.NotificationType(event)
		if event.Title() != host.Title() || event.Severity() != host.Severity() {
			t.Fatalf("event labels changed: %s", event)
		}
	}
}
