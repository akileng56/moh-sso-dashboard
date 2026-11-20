package announcements

// Settings contains only non-secret values used by announcement decisions.
type Settings struct {
	SMS struct {
		Enabled   bool
		MaxLength int
	}
	Announcement struct {
		MaxAttachments         int
		MaxAttachmentBytes     int64
		AllowedAttachmentTypes string
	}
	Email                                                             struct{ AllowedAttachmentTypes string }
	Notification                                                      struct{ PlatformName, SystemAdminName, SystemAdminEmail, AdminDashboardURL string }
	StorageProvider, AppBaseURL, FrontendBaseURL, FrontendRedirectURI string
}
