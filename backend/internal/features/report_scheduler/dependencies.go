package report_scheduler

import (
	"context"
	"encoding/json"
	"io"

	"github.com/google/uuid"
)

type Address struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Attachment persists report bytes through the host email outbox.
type Attachment struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type,omitempty"`
	DataBase64  string `json:"data_base64,omitempty"`
}

type Message struct {
	To          []Address         `json:"to"`
	Subject     string            `json:"subject"`
	TextBody    string            `json:"text_body,omitempty"`
	HTMLBody    string            `json:"html_body,omitempty"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type User struct {
	FullName string
	Email    string
	Enabled  bool
}

type Notification struct {
	Type       string
	Title      string
	Message    string
	Severity   string
	TargetRole string
	Metadata   json.RawMessage
}

// EmailDelivery supplies the email operations used by scheduled reports.
type EmailDelivery interface {
	Queue(ctx context.Context, message Message) error
	ResolveGroupEmailRecipients(ctx context.Context, groupIDs, groupPaths []string) ([]Address, error)
}

// UserLookup resolves report recipients and schedule owners.
type UserLookup interface {
	GetUserByID(id uuid.UUID) (*User, error)
}

// ArtifactStorage exposes only the operations required for report downloads.
type ArtifactStorage interface {
	Download(context.Context, string) (io.ReadCloser, error)
	GetDownloadURL(context.Context, string, string) (string, error)
}

type Audit interface {
	Log(context.Context, uuid.NullUUID, string, map[string]any) error
}

type Notifications interface {
	Notify(context.Context, Notification) error
}
