package announcements

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// UserDirectoryFuncs permits composition without a host feature adapter type.
// All required operations must be supplied; missing functions fail explicitly.
type UserDirectoryFuncs struct {
	List         func(context.Context) ([]AudienceUser, error)
	Get          func(context.Context, uuid.UUID) (*AudienceUser, error)
	ForRealmRole func(context.Context, string) ([]AudienceUser, error)
	ClientRoles  func(context.Context, string) ([]ClientAccess, error)
}

func (f UserDirectoryFuncs) ListUsers(ctx context.Context) ([]AudienceUser, error) {
	if f.List == nil {
		return nil, errors.New("user listing unavailable")
	}
	return f.List(ctx)
}
func (f UserDirectoryFuncs) GetUserByID(ctx context.Context, id uuid.UUID) (*AudienceUser, error) {
	if f.Get == nil {
		return nil, errors.New("user lookup unavailable")
	}
	return f.Get(ctx, id)
}
func (f UserDirectoryFuncs) GetUsersByRealmRole(ctx context.Context, role string) ([]AudienceUser, error) {
	if f.ForRealmRole == nil {
		return nil, errors.New("realm role lookup unavailable")
	}
	return f.ForRealmRole(ctx, role)
}
func (f UserDirectoryFuncs) GetUserClientRoles(ctx context.Context, id string) ([]ClientAccess, error) {
	if f.ClientRoles == nil {
		return nil, errors.New("client role lookup unavailable")
	}
	return f.ClientRoles(ctx, id)
}

type GroupDirectoryFuncs struct {
	Validate func(context.Context, string) error
	Members  func(context.Context, string) ([]AudienceGroupMember, error)
}

func (f GroupDirectoryFuncs) ValidateGroup(ctx context.Context, id string) error {
	if f.Validate == nil {
		return errors.New("group lookup unavailable")
	}
	return f.Validate(ctx, id)
}
func (f GroupDirectoryFuncs) ListGroupMembers(ctx context.Context, id string) ([]AudienceGroupMember, error) {
	if f.Members == nil {
		return nil, errors.New("group membership unavailable")
	}
	return f.Members(ctx, id)
}

type NotificationPublisherFunc func(context.Context, Notification) error

func (f NotificationPublisherFunc) Publish(ctx context.Context, item Notification) error {
	if f == nil {
		return errors.New("notification publisher unavailable")
	}
	return f(ctx, item)
}

type EmailAttachmentNormalizerFunc func([]Attachment) ([]Attachment, error)

func (f EmailAttachmentNormalizerFunc) NormalizeAttachments(items []Attachment) ([]Attachment, error) {
	if f == nil {
		return nil, errors.New("email attachment normalization unavailable")
	}
	return f(items)
}
