package announcements

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"testing"
)

func TestFunctionalProvidersForwardContextValuesAndErrors(t *testing.T) {
	cause := errors.New("provider failed")
	ctx := context.Background()
	id := uuid.New()
	users := UserDirectoryFuncs{Get: func(c context.Context, got uuid.UUID) (*AudienceUser, error) {
		if c != ctx || got != id {
			t.Fatal("lookup arguments lost")
		}
		return nil, cause
	}}
	if _, err := users.GetUserByID(ctx, id); !errors.Is(err, cause) {
		t.Fatal(err)
	}
	if _, err := users.ListUsers(ctx); err == nil {
		t.Fatal("missing provider accepted")
	}
	if _, err := users.GetUsersByRealmRole(ctx, "admin"); err == nil {
		t.Fatal("missing role provider")
	}
	if _, err := users.GetUserClientRoles(ctx, id.String()); err == nil {
		t.Fatal("missing client provider")
	}
	groups := GroupDirectoryFuncs{}
	if err := groups.ValidateGroup(ctx, id.String()); err == nil {
		t.Fatal("missing group provider")
	}
	if _, err := groups.ListGroupMembers(ctx, id.String()); err == nil {
		t.Fatal("missing members provider")
	}
	var publish NotificationPublisherFunc
	if err := publish.Publish(ctx, Notification{}); err == nil {
		t.Fatal("missing publisher")
	}
	var normalize EmailAttachmentNormalizerFunc
	if _, err := normalize.NormalizeAttachments(nil); err == nil {
		t.Fatal("missing normalizer")
	}
	publish = func(c context.Context, n Notification) error {
		if c != ctx || n.ID != id {
			t.Fatal("notification arguments")
		}
		return cause
	}
	if err := publish.Publish(ctx, Notification{ID: id}); !errors.Is(err, cause) {
		t.Fatal(err)
	}
}
