package announcements

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// These feature-owned contract values preserve standard-library nullability.
// They are independent definitions, not aliases to generated persistence types.

type AnnouncementLevel string

const (
	AnnouncementLevelINFO     AnnouncementLevel = "INFO"
	AnnouncementLevelSUCCESS  AnnouncementLevel = "SUCCESS"
	AnnouncementLevelWARNING  AnnouncementLevel = "WARNING"
	AnnouncementLevelCRITICAL AnnouncementLevel = "CRITICAL"
)

type AnnouncementStatus string

const (
	AnnouncementStatusDRAFT     AnnouncementStatus = "DRAFT"
	AnnouncementStatusSCHEDULED AnnouncementStatus = "SCHEDULED"
	AnnouncementStatusPUBLISHED AnnouncementStatus = "PUBLISHED"
	AnnouncementStatusARCHIVED  AnnouncementStatus = "ARCHIVED"
)

type AnnouncementAudienceType string

const (
	AnnouncementAudienceTypeALLUSERS                    AnnouncementAudienceType = "ALL_USERS"
	AnnouncementAudienceTypeADMINSONLY                  AnnouncementAudienceType = "ADMINS_ONLY"
	AnnouncementAudienceTypeSPECIFICCLIENTS             AnnouncementAudienceType = "SPECIFIC_CLIENTS"
	AnnouncementAudienceTypeSPECIFICROLES               AnnouncementAudienceType = "SPECIFIC_ROLES"
	AnnouncementAudienceTypeSPECIFICUSERS               AnnouncementAudienceType = "SPECIFIC_USERS"
	AnnouncementAudienceTypeSPECIFICGROUPS              AnnouncementAudienceType = "SPECIFIC_GROUPS"
	AnnouncementAudienceTypeSPECIFICHEALTHCONTEXTS      AnnouncementAudienceType = "SPECIFIC_HEALTH_CONTEXTS"
	AnnouncementAudienceTypeHEALTHCONTEXTANDDESCENDANTS AnnouncementAudienceType = "HEALTH_CONTEXT_AND_DESCENDANTS"
)

type Announcement struct {
	ID                      uuid.UUID                `json:"id"`
	Title                   string                   `json:"title"`
	Message                 string                   `json:"message"`
	Summary                 sql.NullString           `json:"summary"`
	Level                   AnnouncementLevel        `json:"level"`
	Tag                     sql.NullString           `json:"tag"`
	LinkUrl                 sql.NullString           `json:"link_url"`
	Priority                int32                    `json:"priority"`
	IsPinned                bool                     `json:"is_pinned"`
	Status                  AnnouncementStatus       `json:"status"`
	PublishAt               sql.NullTime             `json:"publish_at"`
	ExpiresAt               sql.NullTime             `json:"expires_at"`
	AudienceType            AnnouncementAudienceType `json:"audience_type"`
	CreatedBy               uuid.UUID                `json:"created_by"`
	UpdatedBy               uuid.NullUUID            `json:"updated_by"`
	PublishedBy             uuid.NullUUID            `json:"published_by"`
	ArchivedBy              uuid.NullUUID            `json:"archived_by"`
	CreatedAt               time.Time                `json:"created_at"`
	UpdatedAt               time.Time                `json:"updated_at"`
	PublishedAt             sql.NullTime             `json:"published_at"`
	ArchivedAt              sql.NullTime             `json:"archived_at"`
	DeletedAt               sql.NullTime             `json:"deleted_at"`
	DeletedBy               uuid.NullUUID            `json:"deleted_by"`
	Version                 int32                    `json:"version"`
	NotifyByEmail           bool                     `json:"notify_by_email"`
	EmailNotificationSentAt sql.NullTime             `json:"email_notification_sent_at"`
	LinkLabel               sql.NullString           `json:"link_label"`
	NotifyBySms             bool                     `json:"notify_by_sms"`
	SmsMessage              sql.NullString           `json:"sms_message"`
	SmsNotificationQueuedAt sql.NullTime             `json:"sms_notification_queued_at"`
}

type AnnouncementAttachment struct {
	ID               uuid.UUID      `json:"id"`
	AnnouncementID   uuid.UUID      `json:"announcement_id"`
	FileName         string         `json:"file_name"`
	OriginalFileName string         `json:"original_file_name"`
	ContentType      sql.NullString `json:"content_type"`
	FileSize         int64          `json:"file_size"`
	StorageProvider  string         `json:"storage_provider"`
	StorageKey       string         `json:"storage_key"`
	Checksum         sql.NullString `json:"checksum"`
	UploadedBy       uuid.NullUUID  `json:"uploaded_by"`
	IncludeInEmail   bool           `json:"include_in_email"`
	Inline           bool           `json:"inline"`
	ContentID        sql.NullString `json:"content_id"`
	SortOrder        int32          `json:"sort_order"`
	CreatedAt        time.Time      `json:"created_at"`
	DeletedAt        sql.NullTime   `json:"deleted_at"`
	DeletedBy        uuid.NullUUID  `json:"deleted_by"`
}

type CreateAnnouncementAttachmentParams struct {
	ID               uuid.UUID      `json:"id"`
	AnnouncementID   uuid.UUID      `json:"announcement_id"`
	FileName         string         `json:"file_name"`
	OriginalFileName string         `json:"original_file_name"`
	ContentType      sql.NullString `json:"content_type"`
	FileSize         int64          `json:"file_size"`
	StorageProvider  string         `json:"storage_provider"`
	StorageKey       string         `json:"storage_key"`
	Checksum         sql.NullString `json:"checksum"`
	UploadedBy       uuid.NullUUID  `json:"uploaded_by"`
	IncludeInEmail   bool           `json:"include_in_email"`
	Inline           bool           `json:"inline"`
	ContentID        sql.NullString `json:"content_id"`
	SortOrder        int32          `json:"sort_order"`
}

type CreateAnnouncementParams struct {
	Title         string                   `json:"title"`
	Message       string                   `json:"message"`
	Summary       sql.NullString           `json:"summary"`
	Level         AnnouncementLevel        `json:"level"`
	Tag           sql.NullString           `json:"tag"`
	LinkUrl       sql.NullString           `json:"link_url"`
	LinkLabel     sql.NullString           `json:"link_label"`
	Priority      int32                    `json:"priority"`
	IsPinned      bool                     `json:"is_pinned"`
	Status        AnnouncementStatus       `json:"status"`
	PublishAt     sql.NullTime             `json:"publish_at"`
	ExpiresAt     sql.NullTime             `json:"expires_at"`
	AudienceType  AnnouncementAudienceType `json:"audience_type"`
	NotifyByEmail bool                     `json:"notify_by_email"`
	NotifyBySms   bool                     `json:"notify_by_sms"`
	SmsMessage    sql.NullString           `json:"sms_message"`
	CreatedBy     uuid.UUID                `json:"created_by"`
}

type DraftAnnouncementParams struct {
	ID        uuid.UUID     `json:"id"`
	UpdatedBy uuid.NullUUID `json:"updated_by"`
}

type GetAnnouncementAttachmentByIDParams struct {
	ID             uuid.UUID `json:"id"`
	AnnouncementID uuid.UUID `json:"announcement_id"`
}

type GetAnnouncementStatsRow struct {
	Total          int64 `json:"total"`
	DraftCount     int64 `json:"draft_count"`
	ScheduledCount int64 `json:"scheduled_count"`
	PublishedCount int64 `json:"published_count"`
	ArchivedCount  int64 `json:"archived_count"`
	ActiveCount    int64 `json:"active_count"`
}

type ListActivePublishedAnnouncementsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListAnnouncementsAdminParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListAnnouncementsByStatusParams struct {
	Status AnnouncementStatus `json:"status"`
	Limit  int32              `json:"limit"`
	Offset int32              `json:"offset"`
}

type ListAnnouncementsCreatedByUserParams struct {
	CreatedBy uuid.UUID `json:"created_by"`
	Limit     int32     `json:"limit"`
	Offset    int32     `json:"offset"`
}

type ListAnnouncementsForClientParams struct {
	ClientID uuid.UUID `json:"client_id"`
	Limit    int32     `json:"limit"`
	Offset   int32     `json:"offset"`
}

type ListAnnouncementsForRoleParams struct {
	RoleName   string `json:"role_name"`
	PageOffset int32  `json:"page_offset"`
	PageLimit  int32  `json:"page_limit"`
}

type ListAnnouncementsForUserParams struct {
	UserID uuid.UUID `json:"user_id"`
	Limit  int32     `json:"limit"`
	Offset int32     `json:"offset"`
}

type ScheduleAnnouncementParams struct {
	ID        uuid.UUID     `json:"id"`
	PublishAt sql.NullTime  `json:"publish_at"`
	UpdatedBy uuid.NullUUID `json:"updated_by"`
}

type SearchAnnouncementsAdminParams struct {
	SearchText sql.NullString `json:"search_text"`
	PageOffset int32          `json:"page_offset"`
	PageLimit  int32          `json:"page_limit"`
}

type SoftDeleteAnnouncementAttachmentParams struct {
	ID             uuid.UUID     `json:"id"`
	AnnouncementID uuid.UUID     `json:"announcement_id"`
	DeletedBy      uuid.NullUUID `json:"deleted_by"`
}

type UpdateAnnouncementAttachmentParams struct {
	ID             uuid.UUID      `json:"id"`
	AnnouncementID uuid.UUID      `json:"announcement_id"`
	IncludeInEmail bool           `json:"include_in_email"`
	Inline         bool           `json:"inline"`
	ContentID      sql.NullString `json:"content_id"`
	SortOrder      int32          `json:"sort_order"`
}

type UpdateAnnouncementParams struct {
	ID            uuid.UUID                `json:"id"`
	Title         string                   `json:"title"`
	Message       string                   `json:"message"`
	Summary       sql.NullString           `json:"summary"`
	Level         AnnouncementLevel        `json:"level"`
	Tag           sql.NullString           `json:"tag"`
	LinkUrl       sql.NullString           `json:"link_url"`
	LinkLabel     sql.NullString           `json:"link_label"`
	Priority      int32                    `json:"priority"`
	IsPinned      bool                     `json:"is_pinned"`
	PublishAt     sql.NullTime             `json:"publish_at"`
	ExpiresAt     sql.NullTime             `json:"expires_at"`
	AudienceType  AnnouncementAudienceType `json:"audience_type"`
	NotifyByEmail bool                     `json:"notify_by_email"`
	NotifyBySms   bool                     `json:"notify_by_sms"`
	SmsMessage    sql.NullString           `json:"sms_message"`
	UpdatedBy     uuid.NullUUID            `json:"updated_by"`
}

type UpdateAnnouncementStatusParams struct {
	ID        uuid.UUID          `json:"id"`
	Status    AnnouncementStatus `json:"status"`
	UpdatedBy uuid.NullUUID      `json:"updated_by"`
}
