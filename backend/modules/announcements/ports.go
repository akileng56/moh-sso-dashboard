package announcements

import (
	"context"
	"io"

	"github.com/google/uuid"
)

// AudienceUser is the user information required to resolve announcement recipients.
type AudienceUser struct {
	ID, Username, Email, FirstName, LastName, FullName string
	Enabled                                            bool
}

type AudienceGroupMember struct {
	UserID, Username, Email string
}

// ClientAccess preserves both identifiers used by existing client audiences.
type ClientAccess struct {
	ID, ClientID string
}

type UserDirectory interface {
	ListUsers(context.Context) ([]AudienceUser, error)
	GetUserByID(context.Context, uuid.UUID) (*AudienceUser, error)
	GetUsersByRealmRole(context.Context, string) ([]AudienceUser, error)
	GetUserClientRoles(context.Context, string) ([]ClientAccess, error)
}

type GroupDirectory interface {
	ValidateGroup(context.Context, string) error
	ListGroupMembers(context.Context, string) ([]AudienceGroupMember, error)
}

// HealthAudience retains actor-scoped resolution and stored-audience expansion
// as distinct operations. The host remains authoritative for contextual policy.
type HealthAudience interface {
	ResolveAudienceUserIDs(context.Context, string, []uuid.UUID, bool) ([]string, error)
	ListAudienceUserIDs(context.Context, []uuid.UUID, bool) ([]string, error)
}

type AttachmentStorage interface {
	Upload(context.Context, string, io.Reader, int64, string) error
	Download(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
}
