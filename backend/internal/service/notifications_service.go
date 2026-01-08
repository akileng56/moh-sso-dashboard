package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/notifications"
	"github.com/moh-sso-dashboard/internal/utils"
)

type NotificationsService interface {
	Notify(ctx context.Context, notification model.Notification) (*model.Notification, error)

	// 🔔 Auth helpers
	NotifyLoginFailed(ctx context.Context, clientID, ip, userAgent string, err error)
	NotifySuspiciousLogin(ctx context.Context, ip, userAgent string)
	NotifyTokenRefreshFailed(ctx context.Context, ip, userAgent string)
	NotifyAccountLocked(ctx context.Context, userID uuid.UUID)

	// 🔔 System / Ops
	NotifySystemStartup(ctx context.Context, version string)
	NotifySystemShutdown(ctx context.Context, reason string)
	NotifyConfigChanged(ctx context.Context, changedBy string, keys []string)
	NotifyBackupCompleted(ctx context.Context, backupID string, durationSeconds int)
	NotifyBackupFailed(ctx context.Context, backupID string, err error)

	ListNotifications(
		ctx context.Context,
		role string,
		unread *bool,
		limit int32,
		offset int32,
	) ([]model.Notification, error)

	GetNotificationByID(ctx context.Context, notificationID string) (*model.Notification, error)
	MarkNotificationAsRead(ctx context.Context, notificationID string) error
	DeleteNotification(ctx context.Context, notificationID string) error
	DeleteOldNotifications(ctx context.Context) error
	CountNotifications(ctx context.Context, targetRole string) (int64, error)
	CountUnreadNotificationsCount(ctx context.Context, targetRole string) (int64, error)
}

type notificationsService struct {
	notificationsRepo repository.NotificationsRepository
}

func NewNotificationsService(
	notificationsRepo repository.NotificationsRepository,
) NotificationsService {
	return &notificationsService{
		notificationsRepo: notificationsRepo,
	}
}

// ----------------------------------------------------
// CORE NOTIFY
// ----------------------------------------------------
func (s *notificationsService) Notify(
	ctx context.Context,
	notification model.Notification,
) (*model.Notification, error) {

	notification.CreatedAt = time.Now()
	notification.Read = false

	n, err := s.notificationsRepo.Notify(ctx, notification)
	if err != nil {
		log.Printf("error creating notification: %v", err)
		return nil, err
	}

	return n, nil
}

// ----------------------------------------------------
// 🔔 AUTH / SECURITY HELPERS
// ----------------------------------------------------

func (s *notificationsService) NotifyLoginFailed(
	ctx context.Context,
	clientID, ip, userAgent string,
	err error,
) {
	nt := model.LoginFailed

	_, _ = s.Notify(ctx, model.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Authentication failed during login",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]interface{}{
			"client_id":  clientID,
			"ip":         ip,
			"user_agent": userAgent,
			"error":      err.Error(),
		}),
	})
}

func (s *notificationsService) NotifySuspiciousLogin(
	ctx context.Context,
	ip, userAgent string,
) {
	nt := model.SuspiciousLogin

	_, _ = s.Notify(ctx, model.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(), // critical
		Message:    "Suspicious login attempt detected",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]interface{}{
			"ip":         ip,
			"user_agent": userAgent,
		}),
	})
}

func (s *notificationsService) NotifyTokenRefreshFailed(
	ctx context.Context,
	ip, userAgent string,
) {
	nt := model.TokenRefreshFailed

	_, _ = s.Notify(ctx, model.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Token refresh failed",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]interface{}{
			"ip":         ip,
			"user_agent": userAgent,
		}),
	})
}

func (s *notificationsService) NotifyAccountLocked(
	ctx context.Context,
	userID uuid.UUID,
) {
	nt := model.AccountLocked

	_, _ = s.Notify(ctx, model.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(), // critical
		Message:    "User account locked due to repeated login failures",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]interface{}{
			"user_id": userID.String(),
		}),
	})
}

// ----------------------------------------------------
// CRUD / QUERY METHODS (UNCHANGED)
// ----------------------------------------------------

func (s *notificationsService) ListNotifications(
	ctx context.Context,
	role string,
	unread *bool,
	limit int32,
	offset int32,
) ([]model.Notification, error) {

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.notificationsRepo.ListNotifications(ctx, role, unread, offset, limit)
}

func (s *notificationsService) GetNotificationByID(
	ctx context.Context,
	notificationID string,
) (*model.Notification, error) {

	id, err := uuid.Parse(notificationID)
	if err != nil {
		return nil, fmt.Errorf("invalid notification id: %w", err)
	}

	return s.notificationsRepo.GetNotificationByID(ctx, id)
}

func (s *notificationsService) MarkNotificationAsRead(
	ctx context.Context,
	notificationID string,
) error {

	id, err := uuid.Parse(notificationID)
	if err != nil {
		return err
	}

	return s.notificationsRepo.MarkNotificationAsRead(ctx, id)
}

func (s *notificationsService) DeleteNotification(
	ctx context.Context,
	notificationID string,
) error {

	id, err := uuid.Parse(notificationID)
	if err != nil {
		return err
	}

	return s.notificationsRepo.DeleteNotification(ctx, id)
}

func (s *notificationsService) DeleteOldNotifications(ctx context.Context) error {
	return s.notificationsRepo.DeleteOldNotifications(ctx)
}

func (s *notificationsService) CountUnreadNotificationsCount(
	ctx context.Context,
	targetRole string,
) (int64, error) {
	return s.notificationsRepo.CountUnreadNotifications(ctx, targetRole)
}

func (s *notificationsService) CountNotifications(
	ctx context.Context,
	targetRole string,
) (int64, error) {
	return s.notificationsRepo.CountNotifications(ctx, targetRole)
}

// -------------------------------------

// ----------------------------------------------------
// 🔔 SYSTEM / OPERATIONS HELPERS
// ----------------------------------------------------

func (s *notificationsService) NotifySystemStartup(
	ctx context.Context,
	version string,
) {
	nt := model.SystemStartup

	_, _ = s.Notify(ctx, model.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "SSO service started successfully",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]interface{}{
			"version": version,
			"time":    time.Now().UTC(),
		}),
	})
}

func (s *notificationsService) NotifySystemShutdown(
	ctx context.Context,
	reason string,
) {
	nt := model.SystemShutdown

	_, _ = s.Notify(ctx, model.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(), // critical
		Message:    "SSO service shutting down",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]interface{}{
			"reason": reason,
			"time":   time.Now().UTC(),
		}),
	})
}

func (s *notificationsService) NotifyConfigChanged(
	ctx context.Context,
	changedBy string,
	keys []string,
) {
	nt := model.ConfigChanged

	_, _ = s.Notify(ctx, model.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(), // warning
		Message:    "System configuration updated",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]interface{}{
			"changed_by": changedBy,
			"keys":       keys,
		}),
	})
}

func (s *notificationsService) NotifyBackupCompleted(
	ctx context.Context,
	backupID string,
	durationSeconds int,
) {
	nt := model.BackupCompleted

	_, _ = s.Notify(ctx, model.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Database backup completed successfully",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]interface{}{
			"backup_id": backupID,
			"duration":  durationSeconds,
		}),
	})
}

func (s *notificationsService) NotifyBackupFailed(
	ctx context.Context,
	backupID string,
	err error,
) {
	nt := model.BackupFailed

	_, _ = s.Notify(ctx, model.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(), // critical
		Message:    "Database backup failed",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]interface{}{
			"backup_id": backupID,
			"error":     err.Error(),
		}),
	})
}
