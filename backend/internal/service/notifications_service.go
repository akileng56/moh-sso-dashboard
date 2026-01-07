package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/notifications"
)

type NotificationsService interface {
	Notify(ctx context.Context, notification model.Notification) (*model.Notification, error)
	ListNotifications(
		ctx context.Context,
		role string,
		unread *bool,
		limit int32,
		offset int32,
	) ([]model.Notification, error)

	GetNotificationByID(
		ctx context.Context,
		notificationID string,
	) (notification *model.Notification, err error)
	MarkNotificationAsRead(ctx context.Context, notificationID string) error
	DeleteNotification(ctx context.Context, notificationID string) error
	DeleteOldNotifications(ctx context.Context) error
	CountNotifications(ctx context.Context, targetRole string) (int64, error)
	CountUnreadNotificationsCount(ctx context.Context, targetRole string) (int64, error)
}

type notificationsService struct {
	notificationsRepo repository.NotificationsRepository
}

func NewNotificationsService(notificationsRepo repository.NotificationsRepository) NotificationsService {
	return &notificationsService{
		notificationsRepo: notificationsRepo,
	}
}

func (s *notificationsService) Notify(
	ctx context.Context,
	notification model.Notification,
) (*model.Notification, error) {

	notification.CreatedAt = time.Now()
	notification.Read = false

	n, err := s.notificationsRepo.Notify(ctx, notification)
	if err != nil {
		log.Printf("error creating notification", "%s", err)
		return nil, err
	}

	return n, nil
}

func (s *notificationsService) ListNotifications(
	ctx context.Context,
	role string,
	unread *bool,
	limit int32,
	offset int32,
) ([]model.Notification, error) {

	// Defensive defaults (in case handler passes bad values)
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	notifications, err := s.notificationsRepo.ListNotifications(
		ctx,
		role,
		unread,
		offset,
		limit,
	)
	if err != nil {
		fmt.Printf(
			"failed to list notifications",
			"error", err,
			"role", role,
		)
		return nil, err
	}

	return notifications, nil
}

func (s *notificationsService) GetNotificationByID(
	ctx context.Context,
	notificationID string,
) (*model.Notification, error) {

	id, err := uuid.Parse(notificationID)
	if err != nil {
		return nil, fmt.Errorf("invalid notification id: %w", err)
	}

	notification, err := s.notificationsRepo.GetNotificationByID(ctx, id)
	if err != nil {
		log.Printf("error getting notification", "%id", id, "error", err)
		return nil, err
	}

	return notification, nil
}

func (s *notificationsService) MarkNotificationAsRead(ctx context.Context, notificationID string) error {
	id, _ := uuid.Parse(notificationID)
	if err := s.notificationsRepo.MarkNotificationAsRead(ctx, id); err != nil {
		log.Printf("Error marking notification %s as read: %v", id, err)
		return err
	}
	return nil
}

func (s *notificationsService) DeleteNotification(ctx context.Context, notificationID string) error {
	id, _ := uuid.Parse(notificationID)
	if err := s.notificationsRepo.DeleteNotification(ctx, id); err != nil {
		log.Printf("Error deleting notification %s: %v", id, err)
		return err
	}
	return nil
}

func (s *notificationsService) DeleteOldNotifications(ctx context.Context) error {
	if err := s.notificationsRepo.DeleteOldNotifications(ctx); err != nil {
		log.Printf("Error deleting old notifications: %v", err)
		return err
	}
	return nil
}

func (s *notificationsService) CountUnreadNotificationsCount(ctx context.Context, targetRole string) (int64, error) {
	return s.notificationsRepo.CountUnreadNotifications(ctx, targetRole)

}

func (s *notificationsService) CountNotifications(ctx context.Context, targetRole string) (int64, error) {
	return s.notificationsRepo.CountNotifications(ctx, targetRole)
}
