package announcements

import "strings"

func announcementLevelValue(value string) AnnouncementLevel {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return AnnouncementLevelINFO
	}
	return AnnouncementLevel(value)
}

func announcementStatusValue(value string) AnnouncementStatus {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return AnnouncementStatusDRAFT
	}
	return AnnouncementStatus(value)
}

func announcementAudienceTypeValue(value string) AnnouncementAudienceType {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return AnnouncementAudienceTypeALLUSERS
	}
	return AnnouncementAudienceType(value)
}
