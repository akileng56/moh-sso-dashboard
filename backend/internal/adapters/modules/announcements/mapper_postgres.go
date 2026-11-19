package announcements

import db "github.com/moh-sso-dashboard/internal/db/sqlc"

func toSQLAnnouncement(value Announcement) db.Announcement {
	return db.Announcement{
		ID:                      value.ID,
		Title:                   value.Title,
		Message:                 value.Message,
		Summary:                 value.Summary,
		Level:                   db.AnnouncementLevel(value.Level),
		Tag:                     value.Tag,
		LinkUrl:                 value.LinkUrl,
		Priority:                value.Priority,
		IsPinned:                value.IsPinned,
		Status:                  db.AnnouncementStatus(value.Status),
		PublishAt:               value.PublishAt,
		ExpiresAt:               value.ExpiresAt,
		AudienceType:            db.AnnouncementAudienceType(value.AudienceType),
		CreatedBy:               value.CreatedBy,
		UpdatedBy:               value.UpdatedBy,
		PublishedBy:             value.PublishedBy,
		ArchivedBy:              value.ArchivedBy,
		CreatedAt:               value.CreatedAt,
		UpdatedAt:               value.UpdatedAt,
		PublishedAt:             value.PublishedAt,
		ArchivedAt:              value.ArchivedAt,
		DeletedAt:               value.DeletedAt,
		DeletedBy:               value.DeletedBy,
		Version:                 value.Version,
		NotifyByEmail:           value.NotifyByEmail,
		EmailNotificationSentAt: value.EmailNotificationSentAt,
		LinkLabel:               value.LinkLabel,
		NotifyBySms:             value.NotifyBySms,
		SmsMessage:              value.SmsMessage,
		SmsNotificationQueuedAt: value.SmsNotificationQueuedAt,
	}
}

func fromSQLAnnouncement(value db.Announcement) Announcement {
	return Announcement{
		ID:                      value.ID,
		Title:                   value.Title,
		Message:                 value.Message,
		Summary:                 value.Summary,
		Level:                   AnnouncementLevel(value.Level),
		Tag:                     value.Tag,
		LinkUrl:                 value.LinkUrl,
		Priority:                value.Priority,
		IsPinned:                value.IsPinned,
		Status:                  AnnouncementStatus(value.Status),
		PublishAt:               value.PublishAt,
		ExpiresAt:               value.ExpiresAt,
		AudienceType:            AnnouncementAudienceType(value.AudienceType),
		CreatedBy:               value.CreatedBy,
		UpdatedBy:               value.UpdatedBy,
		PublishedBy:             value.PublishedBy,
		ArchivedBy:              value.ArchivedBy,
		CreatedAt:               value.CreatedAt,
		UpdatedAt:               value.UpdatedAt,
		PublishedAt:             value.PublishedAt,
		ArchivedAt:              value.ArchivedAt,
		DeletedAt:               value.DeletedAt,
		DeletedBy:               value.DeletedBy,
		Version:                 value.Version,
		NotifyByEmail:           value.NotifyByEmail,
		EmailNotificationSentAt: value.EmailNotificationSentAt,
		LinkLabel:               value.LinkLabel,
		NotifyBySms:             value.NotifyBySms,
		SmsMessage:              value.SmsMessage,
		SmsNotificationQueuedAt: value.SmsNotificationQueuedAt,
	}
}

func toSQLAnnouncementAttachment(value AnnouncementAttachment) db.AnnouncementAttachment {
	return db.AnnouncementAttachment{
		ID:               value.ID,
		AnnouncementID:   value.AnnouncementID,
		FileName:         value.FileName,
		OriginalFileName: value.OriginalFileName,
		ContentType:      value.ContentType,
		FileSize:         value.FileSize,
		StorageProvider:  value.StorageProvider,
		StorageKey:       value.StorageKey,
		Checksum:         value.Checksum,
		UploadedBy:       value.UploadedBy,
		IncludeInEmail:   value.IncludeInEmail,
		Inline:           value.Inline,
		ContentID:        value.ContentID,
		SortOrder:        value.SortOrder,
		CreatedAt:        value.CreatedAt,
		DeletedAt:        value.DeletedAt,
		DeletedBy:        value.DeletedBy,
	}
}

func fromSQLAnnouncementAttachment(value db.AnnouncementAttachment) AnnouncementAttachment {
	return AnnouncementAttachment{
		ID:               value.ID,
		AnnouncementID:   value.AnnouncementID,
		FileName:         value.FileName,
		OriginalFileName: value.OriginalFileName,
		ContentType:      value.ContentType,
		FileSize:         value.FileSize,
		StorageProvider:  value.StorageProvider,
		StorageKey:       value.StorageKey,
		Checksum:         value.Checksum,
		UploadedBy:       value.UploadedBy,
		IncludeInEmail:   value.IncludeInEmail,
		Inline:           value.Inline,
		ContentID:        value.ContentID,
		SortOrder:        value.SortOrder,
		CreatedAt:        value.CreatedAt,
		DeletedAt:        value.DeletedAt,
		DeletedBy:        value.DeletedBy,
	}
}

func toSQLCreateAnnouncementAttachmentParams(value CreateAnnouncementAttachmentParams) db.CreateAnnouncementAttachmentParams {
	return db.CreateAnnouncementAttachmentParams{
		ID:               value.ID,
		AnnouncementID:   value.AnnouncementID,
		FileName:         value.FileName,
		OriginalFileName: value.OriginalFileName,
		ContentType:      value.ContentType,
		FileSize:         value.FileSize,
		StorageProvider:  value.StorageProvider,
		StorageKey:       value.StorageKey,
		Checksum:         value.Checksum,
		UploadedBy:       value.UploadedBy,
		IncludeInEmail:   value.IncludeInEmail,
		Inline:           value.Inline,
		ContentID:        value.ContentID,
		SortOrder:        value.SortOrder,
	}
}

func fromSQLCreateAnnouncementAttachmentParams(value db.CreateAnnouncementAttachmentParams) CreateAnnouncementAttachmentParams {
	return CreateAnnouncementAttachmentParams{
		ID:               value.ID,
		AnnouncementID:   value.AnnouncementID,
		FileName:         value.FileName,
		OriginalFileName: value.OriginalFileName,
		ContentType:      value.ContentType,
		FileSize:         value.FileSize,
		StorageProvider:  value.StorageProvider,
		StorageKey:       value.StorageKey,
		Checksum:         value.Checksum,
		UploadedBy:       value.UploadedBy,
		IncludeInEmail:   value.IncludeInEmail,
		Inline:           value.Inline,
		ContentID:        value.ContentID,
		SortOrder:        value.SortOrder,
	}
}

func toSQLCreateAnnouncementParams(value CreateAnnouncementParams) db.CreateAnnouncementParams {
	return db.CreateAnnouncementParams{
		Title:         value.Title,
		Message:       value.Message,
		Summary:       value.Summary,
		Level:         db.AnnouncementLevel(value.Level),
		Tag:           value.Tag,
		LinkUrl:       value.LinkUrl,
		LinkLabel:     value.LinkLabel,
		Priority:      value.Priority,
		IsPinned:      value.IsPinned,
		Status:        db.AnnouncementStatus(value.Status),
		PublishAt:     value.PublishAt,
		ExpiresAt:     value.ExpiresAt,
		AudienceType:  db.AnnouncementAudienceType(value.AudienceType),
		NotifyByEmail: value.NotifyByEmail,
		NotifyBySms:   value.NotifyBySms,
		SmsMessage:    value.SmsMessage,
		CreatedBy:     value.CreatedBy,
	}
}

func fromSQLCreateAnnouncementParams(value db.CreateAnnouncementParams) CreateAnnouncementParams {
	return CreateAnnouncementParams{
		Title:         value.Title,
		Message:       value.Message,
		Summary:       value.Summary,
		Level:         AnnouncementLevel(value.Level),
		Tag:           value.Tag,
		LinkUrl:       value.LinkUrl,
		LinkLabel:     value.LinkLabel,
		Priority:      value.Priority,
		IsPinned:      value.IsPinned,
		Status:        AnnouncementStatus(value.Status),
		PublishAt:     value.PublishAt,
		ExpiresAt:     value.ExpiresAt,
		AudienceType:  AnnouncementAudienceType(value.AudienceType),
		NotifyByEmail: value.NotifyByEmail,
		NotifyBySms:   value.NotifyBySms,
		SmsMessage:    value.SmsMessage,
		CreatedBy:     value.CreatedBy,
	}
}

func toSQLDraftAnnouncementParams(value DraftAnnouncementParams) db.DraftAnnouncementParams {
	return db.DraftAnnouncementParams{
		ID:        value.ID,
		UpdatedBy: value.UpdatedBy,
	}
}

func fromSQLDraftAnnouncementParams(value db.DraftAnnouncementParams) DraftAnnouncementParams {
	return DraftAnnouncementParams{
		ID:        value.ID,
		UpdatedBy: value.UpdatedBy,
	}
}

func toSQLGetAnnouncementAttachmentByIDParams(value GetAnnouncementAttachmentByIDParams) db.GetAnnouncementAttachmentByIDParams {
	return db.GetAnnouncementAttachmentByIDParams{
		ID:             value.ID,
		AnnouncementID: value.AnnouncementID,
	}
}

func fromSQLGetAnnouncementAttachmentByIDParams(value db.GetAnnouncementAttachmentByIDParams) GetAnnouncementAttachmentByIDParams {
	return GetAnnouncementAttachmentByIDParams{
		ID:             value.ID,
		AnnouncementID: value.AnnouncementID,
	}
}

func toSQLGetAnnouncementStatsRow(value GetAnnouncementStatsRow) db.GetAnnouncementStatsRow {
	return db.GetAnnouncementStatsRow{
		Total:          value.Total,
		DraftCount:     value.DraftCount,
		ScheduledCount: value.ScheduledCount,
		PublishedCount: value.PublishedCount,
		ArchivedCount:  value.ArchivedCount,
		ActiveCount:    value.ActiveCount,
	}
}

func fromSQLGetAnnouncementStatsRow(value db.GetAnnouncementStatsRow) GetAnnouncementStatsRow {
	return GetAnnouncementStatsRow{
		Total:          value.Total,
		DraftCount:     value.DraftCount,
		ScheduledCount: value.ScheduledCount,
		PublishedCount: value.PublishedCount,
		ArchivedCount:  value.ArchivedCount,
		ActiveCount:    value.ActiveCount,
	}
}

func toSQLListActivePublishedAnnouncementsParams(value ListActivePublishedAnnouncementsParams) db.ListActivePublishedAnnouncementsParams {
	return db.ListActivePublishedAnnouncementsParams{
		Limit:  value.Limit,
		Offset: value.Offset,
	}
}

func fromSQLListActivePublishedAnnouncementsParams(value db.ListActivePublishedAnnouncementsParams) ListActivePublishedAnnouncementsParams {
	return ListActivePublishedAnnouncementsParams{
		Limit:  value.Limit,
		Offset: value.Offset,
	}
}

func toSQLListAnnouncementsAdminParams(value ListAnnouncementsAdminParams) db.ListAnnouncementsAdminParams {
	return db.ListAnnouncementsAdminParams{
		Limit:  value.Limit,
		Offset: value.Offset,
	}
}

func fromSQLListAnnouncementsAdminParams(value db.ListAnnouncementsAdminParams) ListAnnouncementsAdminParams {
	return ListAnnouncementsAdminParams{
		Limit:  value.Limit,
		Offset: value.Offset,
	}
}

func toSQLListAnnouncementsByStatusParams(value ListAnnouncementsByStatusParams) db.ListAnnouncementsByStatusParams {
	return db.ListAnnouncementsByStatusParams{
		Status: db.AnnouncementStatus(value.Status),
		Limit:  value.Limit,
		Offset: value.Offset,
	}
}

func fromSQLListAnnouncementsByStatusParams(value db.ListAnnouncementsByStatusParams) ListAnnouncementsByStatusParams {
	return ListAnnouncementsByStatusParams{
		Status: AnnouncementStatus(value.Status),
		Limit:  value.Limit,
		Offset: value.Offset,
	}
}

func toSQLListAnnouncementsCreatedByUserParams(value ListAnnouncementsCreatedByUserParams) db.ListAnnouncementsCreatedByUserParams {
	return db.ListAnnouncementsCreatedByUserParams{
		CreatedBy: value.CreatedBy,
		Limit:     value.Limit,
		Offset:    value.Offset,
	}
}

func fromSQLListAnnouncementsCreatedByUserParams(value db.ListAnnouncementsCreatedByUserParams) ListAnnouncementsCreatedByUserParams {
	return ListAnnouncementsCreatedByUserParams{
		CreatedBy: value.CreatedBy,
		Limit:     value.Limit,
		Offset:    value.Offset,
	}
}

func toSQLListAnnouncementsForClientParams(value ListAnnouncementsForClientParams) db.ListAnnouncementsForClientParams {
	return db.ListAnnouncementsForClientParams{
		ClientID: value.ClientID,
		Limit:    value.Limit,
		Offset:   value.Offset,
	}
}

func fromSQLListAnnouncementsForClientParams(value db.ListAnnouncementsForClientParams) ListAnnouncementsForClientParams {
	return ListAnnouncementsForClientParams{
		ClientID: value.ClientID,
		Limit:    value.Limit,
		Offset:   value.Offset,
	}
}

func toSQLListAnnouncementsForRoleParams(value ListAnnouncementsForRoleParams) db.ListAnnouncementsForRoleParams {
	return db.ListAnnouncementsForRoleParams{
		RoleName:   value.RoleName,
		PageOffset: value.PageOffset,
		PageLimit:  value.PageLimit,
	}
}

func fromSQLListAnnouncementsForRoleParams(value db.ListAnnouncementsForRoleParams) ListAnnouncementsForRoleParams {
	return ListAnnouncementsForRoleParams{
		RoleName:   value.RoleName,
		PageOffset: value.PageOffset,
		PageLimit:  value.PageLimit,
	}
}

func toSQLListAnnouncementsForUserParams(value ListAnnouncementsForUserParams) db.ListAnnouncementsForUserParams {
	return db.ListAnnouncementsForUserParams{
		UserID: value.UserID,
		Limit:  value.Limit,
		Offset: value.Offset,
	}
}

func fromSQLListAnnouncementsForUserParams(value db.ListAnnouncementsForUserParams) ListAnnouncementsForUserParams {
	return ListAnnouncementsForUserParams{
		UserID: value.UserID,
		Limit:  value.Limit,
		Offset: value.Offset,
	}
}

func toSQLScheduleAnnouncementParams(value ScheduleAnnouncementParams) db.ScheduleAnnouncementParams {
	return db.ScheduleAnnouncementParams{
		ID:        value.ID,
		PublishAt: value.PublishAt,
		UpdatedBy: value.UpdatedBy,
	}
}

func fromSQLScheduleAnnouncementParams(value db.ScheduleAnnouncementParams) ScheduleAnnouncementParams {
	return ScheduleAnnouncementParams{
		ID:        value.ID,
		PublishAt: value.PublishAt,
		UpdatedBy: value.UpdatedBy,
	}
}

func toSQLSearchAnnouncementsAdminParams(value SearchAnnouncementsAdminParams) db.SearchAnnouncementsAdminParams {
	return db.SearchAnnouncementsAdminParams{
		SearchText: value.SearchText,
		PageOffset: value.PageOffset,
		PageLimit:  value.PageLimit,
	}
}

func fromSQLSearchAnnouncementsAdminParams(value db.SearchAnnouncementsAdminParams) SearchAnnouncementsAdminParams {
	return SearchAnnouncementsAdminParams{
		SearchText: value.SearchText,
		PageOffset: value.PageOffset,
		PageLimit:  value.PageLimit,
	}
}

func toSQLSoftDeleteAnnouncementAttachmentParams(value SoftDeleteAnnouncementAttachmentParams) db.SoftDeleteAnnouncementAttachmentParams {
	return db.SoftDeleteAnnouncementAttachmentParams{
		ID:             value.ID,
		AnnouncementID: value.AnnouncementID,
		DeletedBy:      value.DeletedBy,
	}
}

func fromSQLSoftDeleteAnnouncementAttachmentParams(value db.SoftDeleteAnnouncementAttachmentParams) SoftDeleteAnnouncementAttachmentParams {
	return SoftDeleteAnnouncementAttachmentParams{
		ID:             value.ID,
		AnnouncementID: value.AnnouncementID,
		DeletedBy:      value.DeletedBy,
	}
}

func toSQLUpdateAnnouncementAttachmentParams(value UpdateAnnouncementAttachmentParams) db.UpdateAnnouncementAttachmentParams {
	return db.UpdateAnnouncementAttachmentParams{
		ID:             value.ID,
		AnnouncementID: value.AnnouncementID,
		IncludeInEmail: value.IncludeInEmail,
		Inline:         value.Inline,
		ContentID:      value.ContentID,
		SortOrder:      value.SortOrder,
	}
}

func fromSQLUpdateAnnouncementAttachmentParams(value db.UpdateAnnouncementAttachmentParams) UpdateAnnouncementAttachmentParams {
	return UpdateAnnouncementAttachmentParams{
		ID:             value.ID,
		AnnouncementID: value.AnnouncementID,
		IncludeInEmail: value.IncludeInEmail,
		Inline:         value.Inline,
		ContentID:      value.ContentID,
		SortOrder:      value.SortOrder,
	}
}

func toSQLUpdateAnnouncementParams(value UpdateAnnouncementParams) db.UpdateAnnouncementParams {
	return db.UpdateAnnouncementParams{
		ID:            value.ID,
		Title:         value.Title,
		Message:       value.Message,
		Summary:       value.Summary,
		Level:         db.AnnouncementLevel(value.Level),
		Tag:           value.Tag,
		LinkUrl:       value.LinkUrl,
		LinkLabel:     value.LinkLabel,
		Priority:      value.Priority,
		IsPinned:      value.IsPinned,
		PublishAt:     value.PublishAt,
		ExpiresAt:     value.ExpiresAt,
		AudienceType:  db.AnnouncementAudienceType(value.AudienceType),
		NotifyByEmail: value.NotifyByEmail,
		NotifyBySms:   value.NotifyBySms,
		SmsMessage:    value.SmsMessage,
		UpdatedBy:     value.UpdatedBy,
	}
}

func fromSQLUpdateAnnouncementParams(value db.UpdateAnnouncementParams) UpdateAnnouncementParams {
	return UpdateAnnouncementParams{
		ID:            value.ID,
		Title:         value.Title,
		Message:       value.Message,
		Summary:       value.Summary,
		Level:         AnnouncementLevel(value.Level),
		Tag:           value.Tag,
		LinkUrl:       value.LinkUrl,
		LinkLabel:     value.LinkLabel,
		Priority:      value.Priority,
		IsPinned:      value.IsPinned,
		PublishAt:     value.PublishAt,
		ExpiresAt:     value.ExpiresAt,
		AudienceType:  AnnouncementAudienceType(value.AudienceType),
		NotifyByEmail: value.NotifyByEmail,
		NotifyBySms:   value.NotifyBySms,
		SmsMessage:    value.SmsMessage,
		UpdatedBy:     value.UpdatedBy,
	}
}

func toSQLUpdateAnnouncementStatusParams(value UpdateAnnouncementStatusParams) db.UpdateAnnouncementStatusParams {
	return db.UpdateAnnouncementStatusParams{
		ID:        value.ID,
		Status:    db.AnnouncementStatus(value.Status),
		UpdatedBy: value.UpdatedBy,
	}
}

func fromSQLUpdateAnnouncementStatusParams(value db.UpdateAnnouncementStatusParams) UpdateAnnouncementStatusParams {
	return UpdateAnnouncementStatusParams{
		ID:        value.ID,
		Status:    AnnouncementStatus(value.Status),
		UpdatedBy: value.UpdatedBy,
	}
}
