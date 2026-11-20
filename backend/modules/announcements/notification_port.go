package announcements

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Attachment struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type,omitempty"`

	// Use one of these:
	Path string `json:"path,omitempty"`
	Data []byte `json:"-"`

	// DataBase64 is persisted through queues/outbox and decoded before send.
	DataBase64 string `json:"data_base64,omitempty"`

	// Optional inline image support
	Inline    bool   `json:"inline,omitempty"`
	ContentID string `json:"content_id,omitempty"`
}

type Notification struct {
	ID         uuid.UUID       `json:"id"`
	Type       string          `json:"type"`
	Title      string          `json:"title"`
	Message    string          `json:"message"`
	Severity   string          `json:"severity"`    // info | warning | critical
	TargetRole string          `json:"target_role"` // admin | super_admin | etc
	ClientID   string          `json:"client_id,omitempty"`
	UserID     string          `json:"user_id,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"` // JSONB from Postgres
	Read       bool            `json:"read"`
	CreatedAt  time.Time       `json:"created_at"`

	// Optional delivery channels for this notification.
	// If empty, the service will create a default in-app delivery.
	Deliveries []NotificationDeliveryRequest `json:"deliveries,omitempty"`
}

type NotificationDeliveryRequest struct {
	Channel      NotificationChannel `json:"channel"`
	Recipient    map[string]any      `json:"recipient,omitempty"`
	TemplateName string              `json:"template_name,omitempty"`
	TemplateData map[string]any      `json:"template_data,omitempty"`
	Payload      map[string]any      `json:"payload,omitempty"`
	Attachments  []Attachment        `json:"attachments,omitempty"`
	ScheduledAt  *time.Time          `json:"scheduled_at,omitempty"`
	MaxAttempts  int32               `json:"max_attempts,omitempty"`
}

type NotificationChannel string

const (
	NotificationChannelInApp   NotificationChannel = "in_app"
	NotificationChannelEmail   NotificationChannel = "email"
	NotificationChannelSMS     NotificationChannel = "sms"
	NotificationChannelWebhook NotificationChannel = "webhook"
)

type NotificationPublisher interface {
	Publish(context.Context, Notification) error
}

type EmailAttachmentNormalizer interface {
	NormalizeAttachments([]Attachment) ([]Attachment, error)
}

type ServicePorts struct {
	Users            UserDirectory
	Groups           GroupDirectory
	HealthContexts   HealthAudience
	Notifications    NotificationPublisher
	Storage          AttachmentStorage
	EmailAttachments EmailAttachmentNormalizer
}

type announcementEvent string

const (
	AnnouncementCreated   announcementEvent = "ANNOUNCEMENT_CREATED"
	AnnouncementUpdated   announcementEvent = "ANNOUNCEMENT_UPDATED"
	AnnouncementDeleted   announcementEvent = "ANNOUNCEMENT_DELETED"
	AnnouncementArchived  announcementEvent = "ANNOUNCEMENT_ARCHIVED"
	AnnouncementScheduled announcementEvent = "ANNOUNCEMENT_SCHEDULED"
	AnnouncementPublished announcementEvent = "ANNOUNCEMENT_PUBLISHED"
	AnnouncementDrafted   announcementEvent = "ANNOUNCEMENT_DRAFTED"
)

// Preserve the existing host announcement event labels.
func (announcementEvent) Title() string    { return "System notification" }
func (announcementEvent) Severity() string { return "info" }
