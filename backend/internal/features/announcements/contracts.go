package announcements

import announcementmodule "github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements"

// Compatibility aliases expose only the selected package contracts.
type AnnouncementLevel = announcementmodule.AnnouncementLevel
type AnnouncementStatus = announcementmodule.AnnouncementStatus
type AnnouncementAudienceType = announcementmodule.AnnouncementAudienceType
type Announcement = announcementmodule.Announcement
type AnnouncementAttachment = announcementmodule.AnnouncementAttachment
type CreateAnnouncementAttachmentParams = announcementmodule.CreateAnnouncementAttachmentParams
type CreateAnnouncementParams = announcementmodule.CreateAnnouncementParams
type DraftAnnouncementParams = announcementmodule.DraftAnnouncementParams
type GetAnnouncementAttachmentByIDParams = announcementmodule.GetAnnouncementAttachmentByIDParams
type GetAnnouncementStatsRow = announcementmodule.GetAnnouncementStatsRow
type ListActivePublishedAnnouncementsParams = announcementmodule.ListActivePublishedAnnouncementsParams
type ListAnnouncementsAdminParams = announcementmodule.ListAnnouncementsAdminParams
type ListAnnouncementsByStatusParams = announcementmodule.ListAnnouncementsByStatusParams
type ListAnnouncementsCreatedByUserParams = announcementmodule.ListAnnouncementsCreatedByUserParams
type ListAnnouncementsForClientParams = announcementmodule.ListAnnouncementsForClientParams
type ListAnnouncementsForRoleParams = announcementmodule.ListAnnouncementsForRoleParams
type ListAnnouncementsForUserParams = announcementmodule.ListAnnouncementsForUserParams
type ScheduleAnnouncementParams = announcementmodule.ScheduleAnnouncementParams
type SearchAnnouncementsAdminParams = announcementmodule.SearchAnnouncementsAdminParams
type SoftDeleteAnnouncementAttachmentParams = announcementmodule.SoftDeleteAnnouncementAttachmentParams
type UpdateAnnouncementAttachmentParams = announcementmodule.UpdateAnnouncementAttachmentParams
type UpdateAnnouncementParams = announcementmodule.UpdateAnnouncementParams
type UpdateAnnouncementStatusParams = announcementmodule.UpdateAnnouncementStatusParams
type AnnouncementRepository = announcementmodule.AnnouncementRepository
type Repository = announcementmodule.Repository
type Attachment = announcementmodule.Attachment
type Notification = announcementmodule.Notification
type NotificationDeliveryRequest = announcementmodule.NotificationDeliveryRequest
type NotificationChannel = announcementmodule.NotificationChannel
type NotificationPublisher = announcementmodule.NotificationPublisher
type EmailAttachmentNormalizer = announcementmodule.EmailAttachmentNormalizer
type ServicePorts = announcementmodule.ServicePorts
type Settings = announcementmodule.Settings
type AudienceUser = announcementmodule.AudienceUser
type AudienceGroupMember = announcementmodule.AudienceGroupMember
type ClientAccess = announcementmodule.ClientAccess
type UserDirectory = announcementmodule.UserDirectory
type GroupDirectory = announcementmodule.GroupDirectory
type HealthAudience = announcementmodule.HealthAudience
type AttachmentStorage = announcementmodule.AttachmentStorage
type CreateAnnouncementInput = announcementmodule.CreateAnnouncementInput
type UpdateAnnouncementInput = announcementmodule.UpdateAnnouncementInput
type AnnouncementEmailRecipient = announcementmodule.AnnouncementEmailRecipient
type AnnouncementSMSRecipient = announcementmodule.AnnouncementSMSRecipient
type AnnouncementHealthContextAudience = announcementmodule.AnnouncementHealthContextAudience
type AnnouncementEmailOptions = announcementmodule.AnnouncementEmailOptions
type AnnouncementEmailAttachmentLink = announcementmodule.AnnouncementEmailAttachmentLink
type UploadAnnouncementAttachmentInput = announcementmodule.UploadAnnouncementAttachmentInput
type UpdateAnnouncementAttachmentInput = announcementmodule.UpdateAnnouncementAttachmentInput
type AnnouncementAttachmentDownload = announcementmodule.AnnouncementAttachmentDownload

const (
	AnnouncementLevelINFO                               = announcementmodule.AnnouncementLevelINFO
	AnnouncementLevelSUCCESS                            = announcementmodule.AnnouncementLevelSUCCESS
	AnnouncementLevelWARNING                            = announcementmodule.AnnouncementLevelWARNING
	AnnouncementLevelCRITICAL                           = announcementmodule.AnnouncementLevelCRITICAL
	AnnouncementStatusDRAFT                             = announcementmodule.AnnouncementStatusDRAFT
	AnnouncementStatusSCHEDULED                         = announcementmodule.AnnouncementStatusSCHEDULED
	AnnouncementStatusPUBLISHED                         = announcementmodule.AnnouncementStatusPUBLISHED
	AnnouncementStatusARCHIVED                          = announcementmodule.AnnouncementStatusARCHIVED
	AnnouncementAudienceTypeALLUSERS                    = announcementmodule.AnnouncementAudienceTypeALLUSERS
	AnnouncementAudienceTypeADMINSONLY                  = announcementmodule.AnnouncementAudienceTypeADMINSONLY
	AnnouncementAudienceTypeSPECIFICCLIENTS             = announcementmodule.AnnouncementAudienceTypeSPECIFICCLIENTS
	AnnouncementAudienceTypeSPECIFICROLES               = announcementmodule.AnnouncementAudienceTypeSPECIFICROLES
	AnnouncementAudienceTypeSPECIFICUSERS               = announcementmodule.AnnouncementAudienceTypeSPECIFICUSERS
	AnnouncementAudienceTypeSPECIFICGROUPS              = announcementmodule.AnnouncementAudienceTypeSPECIFICGROUPS
	AnnouncementAudienceTypeSPECIFICHEALTHCONTEXTS      = announcementmodule.AnnouncementAudienceTypeSPECIFICHEALTHCONTEXTS
	AnnouncementAudienceTypeHEALTHCONTEXTANDDESCENDANTS = announcementmodule.AnnouncementAudienceTypeHEALTHCONTEXTANDDESCENDANTS
	NotificationChannelInApp                            = announcementmodule.NotificationChannelInApp
	NotificationChannelEmail                            = announcementmodule.NotificationChannelEmail
	NotificationChannelSMS                              = announcementmodule.NotificationChannelSMS
	NotificationChannelWebhook                          = announcementmodule.NotificationChannelWebhook
	AnnouncementCreated                                 = announcementmodule.AnnouncementCreated
	AnnouncementUpdated                                 = announcementmodule.AnnouncementUpdated
	AnnouncementDeleted                                 = announcementmodule.AnnouncementDeleted
	AnnouncementArchived                                = announcementmodule.AnnouncementArchived
	AnnouncementScheduled                               = announcementmodule.AnnouncementScheduled
	AnnouncementPublished                               = announcementmodule.AnnouncementPublished
	AnnouncementDrafted                                 = announcementmodule.AnnouncementDrafted
)
