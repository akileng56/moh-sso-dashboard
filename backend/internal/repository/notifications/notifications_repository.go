package notifications

import (
	"context"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/model"
)

type NotificationsRepository interface {
	Notify(ctx context.Context, notification model.Notification) (*model.Notification, error)

	ListNotifications(context.Context, string, *bool, int32, int32) ([]model.Notification, error)

	GetNotificationByID(ctx context.Context, id uuid.UUID) (*model.Notification, error)

	MarkNotificationAsRead(ctx context.Context, id uuid.UUID) error
	DeleteNotification(ctx context.Context, id uuid.UUID) error
	DeleteOldNotifications(ctx context.Context) error

	CountNotifications(ctx context.Context, targetRole string) (int64, error)
	CountUnreadNotifications(ctx context.Context, targetRole string) (int64, error)
}
