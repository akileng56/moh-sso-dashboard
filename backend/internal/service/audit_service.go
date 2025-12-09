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

func (a *AuditService) Log(ctx context.Context, userID uuid.NullUUID, action string, metadata interface{}) error {
	return a.store.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID: userID,
		Action: action,
		Metadata: pqtype.NullRawMessage{
			RawMessage: utils.Encode(metadata),
			Valid:      metadata != nil,
		},
	})
}

func (a *AuditService) LogLogin(
	ctx context.Context,
	userID uuid.NullUUID,
	success bool,
	clientID string,
	tenantID string,
	ip string,
	userAgent string,
	country string,
	city string,
) error {

	metadata := map[string]interface{}{
		"success":    success,
		"client_id":  clientID,
		"ip":         ip,
		"country":    country,
		"city":       city,
		"user_agent": userAgent,
	}

	return a.store.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID: userID,
		Action: "login",
		Metadata: pqtype.NullRawMessage{
			RawMessage: utils.Encode(metadata),
			Valid:      metadata != nil,
		}})
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

	return a.store.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID: userID,
		Action: "password_reset",
		Metadata: pqtype.NullRawMessage{
			RawMessage: utils.Encode(metadata),
			Valid:      metadata != nil,
		}})
}

func (a *AuditService) LogAdminAction(
	ctx context.Context,
	adminID uuid.NullUUID,
	action string, // e.g. "CREATE_USER", "DELETE_CLIENT", etc.
	metadata map[string]interface{},
) error {

	return a.store.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID: adminID,
		Action: action,
		Metadata: pqtype.NullRawMessage{
			RawMessage: utils.Encode(metadata),
			Valid:      metadata != nil,
		}})
}
