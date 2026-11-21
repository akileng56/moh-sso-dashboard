package announcements

import (
	"database/sql"
	"github.com/google/uuid"
)

type CreateAnnouncementInput struct {
	Title         string
	Message       string
	Summary       sql.NullString
	Level         string
	Tag           sql.NullString
	LinkURL       sql.NullString
	LinkLabel     sql.NullString
	Priority      int32
	IsPinned      bool
	Status        string
	PublishAt     sql.NullTime
	ExpiresAt     sql.NullTime
	AudienceType  string
	NotifyByEmail bool
	NotifyBySMS   bool
	SMSMessage    sql.NullString
	CreatedBy     uuid.UUID
}

type UpdateAnnouncementInput struct {
	ID            uuid.UUID
	Title         string
	Message       string
	Summary       sql.NullString
	Level         string
	Tag           sql.NullString
	LinkURL       sql.NullString
	LinkLabel     sql.NullString
	Priority      int32
	IsPinned      bool
	PublishAt     sql.NullTime
	ExpiresAt     sql.NullTime
	AudienceType  string
	NotifyByEmail bool
	NotifyBySMS   bool
	SMSMessage    sql.NullString
	UpdatedBy     uuid.UUID
}

type AnnouncementEmailRecipient struct {
	ID       uuid.UUID
	Email    string
	Username string
	FullName string
}

type AnnouncementSMSRecipient struct {
	ID          uuid.UUID
	PhoneNumber string
	Username    string
	FullName    string
}

type AnnouncementHealthContextAudience struct {
	ContextNodeID      uuid.UUID
	IncludeDescendants bool
}
