package announcements

type endpoint struct {
	method, path       string
	admin              bool
	permission, policy string
	serve              func(*httpRequest)
}

func endpoints(h *httpHandler) []endpoint {
	return []endpoint{
		{"GET", "/announcements/public", false, "", "", h.ListPublicAnnouncements},
		{"GET", "/announcements/active", false, "", "", h.ListActivePublishedAnnouncements},
		{"GET", "/announcements/{id}/attachments/{attachmentId}/download", false, ReadPermission, "", h.DownloadAnnouncementAttachment},
		{"GET", "/announcements/me", false, ReadPermission, "", h.ListMyAnnouncements},
		{"GET", "/announcements/user", false, ReadPermission, "", h.ListAnnouncementsForUser},
		{"GET", "/announcements/role/{role_name}", false, ReadPermission, "", h.ListAnnouncementsForRole},
		{"GET", "/announcements/client/{client_id}", false, ReadPermission, "", h.ListAnnouncementsForClient},
		{"GET", "/admin/announcements", true, ReadPermission, "", h.ListAnnouncementsAdmin},
		{"GET", "/admin/announcements/stats", true, ReadPermission, "", h.GetAnnouncementStats},
		{"POST", "/admin/announcements/audience-preview", true, WritePermission, "", h.PreviewAnnouncementAudience},
		{"GET", "/admin/announcements/{id}/attachments", true, ReadPermission, "", h.ListAnnouncementAttachments},
		{"GET", "/admin/announcements/{id}/attachments/{attachmentId}/download", true, ReadPermission, "", h.DownloadAnnouncementAttachment},
		{"POST", "/admin/announcements/{id}/attachments", true, WritePermission, WritePolicy, h.UploadAnnouncementAttachment},
		{"PATCH", "/admin/announcements/{id}/attachments/{attachmentId}", true, WritePermission, WritePolicy, h.UpdateAnnouncementAttachment},
		{"DELETE", "/admin/announcements/{id}/attachments/{attachmentId}", true, WritePermission, WritePolicy, h.DeleteAnnouncementAttachment},
		{"GET", "/admin/announcements/{id}", true, ReadPermission, "", h.GetAnnouncementByID},
		{"POST", "/admin/announcements", true, WritePermission, WritePolicy, h.CreateAnnouncement},
		{"PUT", "/admin/announcements/{id}", true, WritePermission, WritePolicy, h.UpdateAnnouncement},
		{"DELETE", "/admin/announcements/{id}", true, WritePermission, WritePolicy, h.DeleteAnnouncement},
		{"POST", "/admin/announcements/{id}/restore", true, WritePermission, WritePolicy, h.RestoreAnnouncement},
		{"POST", "/admin/announcements/{id}/publish", true, PublishPermission, PublishPolicy, h.PublishAnnouncementNow},
		{"POST", "/admin/announcements/{id}/draft", true, PublishPermission, PublishPolicy, h.MoveAnnouncementToDraft},
		{"POST", "/admin/announcements/{id}/schedule", true, PublishPermission, PublishPolicy, h.ScheduleAnnouncement},
		{"POST", "/admin/announcements/{id}/archive", true, PublishPermission, PublishPolicy, h.ArchiveAnnouncement},
		{"PATCH", "/admin/announcements/{id}/pin", true, WritePermission, WritePolicy, h.SetAnnouncementPinned},
		{"PATCH", "/admin/announcements/{id}/priority", true, WritePermission, WritePolicy, h.SetAnnouncementPriority},
	}
}
