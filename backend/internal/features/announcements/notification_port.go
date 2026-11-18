package announcements

import (
	"context"

	models "github.com/moh-sso-dashboard/internal/model"
)

// NotificationPublisher is the only delivery operation this feature needs.
// The payload remains host-owned until the portable delivery-model phase.
type NotificationPublisher interface {
	Notify(context.Context, models.Notification) (*models.Notification, error)
}

type ServicePorts struct {
	Users          UserDirectory
	Groups         GroupDirectory
	HealthContexts HealthAudience
	Notifications  NotificationPublisher
	Storage        AttachmentStorage
}
