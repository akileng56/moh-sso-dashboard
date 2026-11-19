package announcements

import (
	"context"
	"github.com/moh-sso-dashboard/internal/config"
	emailutil "github.com/moh-sso-dashboard/internal/email"
	models "github.com/moh-sso-dashboard/internal/model"
)

type HostNotificationSender interface {
	Notify(context.Context, models.Notification) (*models.Notification, error)
}
type hostNotificationPublisher struct{ sender HostNotificationSender }

func NewNotificationPublisher(sender HostNotificationSender) NotificationPublisher {
	if sender == nil {
		return nil
	}
	return &hostNotificationPublisher{sender: sender}
}
func (p *hostNotificationPublisher) Publish(ctx context.Context, item Notification) error {
	_, err := p.sender.Notify(ctx, toHostNotification(item))
	return err
}

type hostAttachmentNormalizer struct{ cfg *config.Config }

func NewEmailAttachmentNormalizer(cfg *config.Config) EmailAttachmentNormalizer {
	return &hostAttachmentNormalizer{cfg: cfg}
}
func (a *hostAttachmentNormalizer) NormalizeAttachments(items []Attachment) ([]Attachment, error) {
	result, err := emailutil.NormalizeAttachments(a.cfg, toHostAttachments(items))
	return fromHostAttachments(result), err
}

func toHostAttachments(items []Attachment) []models.Attachment {
	if items == nil {
		return nil
	}
	result := make([]models.Attachment, len(items))
	for i, item := range items {
		result[i] = models.Attachment(item)
	}
	return result
}
func fromHostAttachments(items []models.Attachment) []Attachment {
	if items == nil {
		return nil
	}
	result := make([]Attachment, len(items))
	for i, item := range items {
		result[i] = Attachment(item)
	}
	return result
}

func toHostNotification(item Notification) models.Notification {
	out := models.Notification{
		ID: item.ID, Type: item.Type, Title: item.Title, Message: item.Message,
		Severity: item.Severity, TargetRole: item.TargetRole, ClientID: item.ClientID,
		UserID: item.UserID, Metadata: item.Metadata, Read: item.Read, CreatedAt: item.CreatedAt,
	}
	if item.Deliveries != nil {
		out.Deliveries = make([]models.NotificationDeliveryRequest, len(item.Deliveries))
		for i, delivery := range item.Deliveries {
			out.Deliveries[i] = models.NotificationDeliveryRequest{
				Channel: models.NotificationChannel(delivery.Channel), Recipient: delivery.Recipient,
				TemplateName: delivery.TemplateName, TemplateData: delivery.TemplateData,
				Payload: delivery.Payload, Attachments: toHostAttachments(delivery.Attachments),
				ScheduledAt: delivery.ScheduledAt, MaxAttempts: delivery.MaxAttempts,
			}
		}
	}
	return out
}
