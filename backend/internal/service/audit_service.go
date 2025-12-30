package service

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/utils"
	"github.com/sqlc-dev/pqtype"
)

type AuditService struct {
	store db.Store
}

func NewAuditService(store db.Store) *AuditService {
	return &AuditService{store: store}
}

//
// ----------------------------------------------------
// Internal helpers
// ----------------------------------------------------
//

// Ensures we never violate FK constraints.
// If the user does not exist locally, we downgrade to NULL (system).
func (a *AuditService) safeUserID(
	ctx context.Context,
	userID uuid.NullUUID,
) uuid.NullUUID {

	// Anonymous / system event
	if !userID.Valid {
		return userID
	}

	exists, err := a.store.UserExists(ctx, userID.UUID)
	if err != nil || !exists {
		return uuid.NullUUID{}
	}

	return userID
}

func (a *AuditService) write(
	ctx context.Context,
	userID uuid.NullUUID,
	action string,
	metadata map[string]interface{},
) error {

	safeID := a.safeUserID(ctx, userID)

	return a.store.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID: safeID,
		Action: action,
		Metadata: pqtype.NullRawMessage{
			RawMessage: utils.Encode(metadata),
			Valid:      metadata != nil,
		},
	})
}

//
// ----------------------------------------------------
// Public API (handlers use ONLY these)
// ----------------------------------------------------
//

// Generic logger (fallback)
func (a *AuditService) Log(
	ctx context.Context,
	userID uuid.NullUUID,
	action string,
	metadata map[string]interface{},
) error {
	return a.write(ctx, userID, action, metadata)
}

// --------------------
// AUTH EVENTS
// --------------------

func (a *AuditService) LoginInitiated(
	ctx context.Context,
	ip string,
	userAgent string,
	clientID string,
) error {

	return a.write(ctx, uuid.NullUUID{}, "auth.login_initiated", map[string]interface{}{
		"ip":         ip,
		"user_agent": userAgent,
		"client_id":  clientID,
	})
}

func (a *AuditService) LoginResult(
	ctx context.Context,
	userID uuid.NullUUID,
	success bool,
	clientID string,
	ip string,
	userAgent string,
	country string,
	city string,
) error {

	return a.write(ctx, userID, "auth.login", map[string]interface{}{
		"success":    success,
		"client_id":  clientID,
		"ip":         ip,
		"user_agent": userAgent,
		"country":    country,
		"city":       city,
	})
}

func (a *AuditService) TokenRefresh(
	ctx context.Context,
	userID uuid.NullUUID,
	success bool,
	ip string,
	userAgent string,
) error {

	action := "auth.refresh_success"
	if !success {
		action = "auth.refresh_failed"
	}

	return a.write(ctx, userID, action, map[string]interface{}{
		"success":    success,
		"ip":         ip,
		"user_agent": userAgent,
	})
}

func (a *AuditService) Logout(
	ctx context.Context,
	userID uuid.NullUUID,
	ip string,
	userAgent string,
) error {

	return a.write(ctx, userID, "auth.logout", map[string]interface{}{
		"ip":         ip,
		"user_agent": userAgent,
	})
}

func (a *AuditService) PasswordReset(
	ctx context.Context,
	userID uuid.NullUUID,
	success bool,
	ip string,
) error {

	return a.write(ctx, userID, "auth.password_reset", map[string]interface{}{
		"success": success,
		"ip":      ip,
	})
}

// --------------------
// CLIENT / ADMIN EVENTS
// --------------------

func (a *AuditService) ClientCreated(
	ctx context.Context,
	adminID uuid.NullUUID,
	clientID string,
	ip string,
	userAgent string,
) error {

	return a.write(ctx, adminID, "client.create", map[string]interface{}{
		"client_id":  clientID,
		"ip":         ip,
		"user_agent": userAgent,
	})
}

func (a *AuditService) ClientDeleted(
	ctx context.Context,
	adminID uuid.NullUUID,
	clientID string,
	ip string,
	userAgent string,
) error {

	return a.write(ctx, adminID, "client.delete", map[string]interface{}{
		"client_id":  clientID,
		"ip":         ip,
		"user_agent": userAgent,
	})
}

func (a *AuditService) AdminAction(
	ctx context.Context,
	adminID uuid.NullUUID,
	action string, // e.g. user.disable, role.assign
	metadata map[string]interface{},
) error {

	return a.write(ctx, adminID, action, metadata)
}
