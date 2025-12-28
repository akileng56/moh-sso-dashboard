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

// ----------------------------------------------------
// internal helper
// Ensures FK safety for audit logs
// ----------------------------------------------------
func (a *AuditService) safeUserID(
	ctx context.Context,
	userID uuid.NullUUID,
) uuid.NullUUID {

	// Anonymous / system action
	if !userID.Valid {
		return userID
	}

	exists, err := a.store.UserExists(ctx, userID.UUID)
	if err != nil {
		// Fail-safe: never break audit logging
		return uuid.NullUUID{}
	}

	if !exists {
		// User not yet synced locally
		return uuid.NullUUID{}
	}

	return userID
}

// ----------------------------------------------------
// Core audit logger (ALL logs go through here)
// ----------------------------------------------------
func (a *AuditService) Log(
	ctx context.Context,
	userID uuid.NullUUID,
	action string,
	metadata interface{},
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

// ----------------------------------------------------
// Auth / security logs
// ----------------------------------------------------
func (a *AuditService) LogLogin(
	ctx context.Context,
	userID uuid.NullUUID,
	success bool,
	clientID string,
	ip string,
	userAgent string,
	country string,
	city string,
) error {

	metadata := map[string]interface{}{
		"success":    success,
		"client_id":  clientID,
		"ip":         ip,
		"user_agent": userAgent,
		"country":    country,
		"city":       city,
	}

	return a.Log(ctx, userID, "login", metadata)
}

func (a *AuditService) LogPasswordReset(
	ctx context.Context,
	userID uuid.NullUUID,
	success bool,
	ip string,
) error {

	metadata := map[string]interface{}{
		"success": success,
		"ip":      ip,
	}

	return a.Log(ctx, userID, "password_reset", metadata)
}

// ----------------------------------------------------
// Admin actions
// ----------------------------------------------------
func (a *AuditService) LogAdminAction(
	ctx context.Context,
	adminID uuid.NullUUID,
	action string, // e.g. CREATE_USER, DELETE_CLIENT
	metadata map[string]interface{},
) error {

	return a.Log(ctx, adminID, action, metadata)
}
