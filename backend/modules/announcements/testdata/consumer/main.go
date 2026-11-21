// This program is copied into an independent Go module by the package verifier.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	announcements "github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements"
)

type repository struct {
	announcements.Repository
	item       announcements.Announcement
	attachment announcements.AnnouncementAttachment
	created    announcements.CreateAnnouncementParams
	ctx        context.Context
	failure    error
}

func (r *repository) Create(ctx context.Context, p announcements.CreateAnnouncementParams) (announcements.Announcement, error) {
	r.ctx, r.created = ctx, p
	return r.item, r.failure
}
func (r *repository) PublishNow(ctx context.Context, id, actor uuid.UUID) (announcements.Announcement, error) {
	if id != r.item.ID || actor != r.item.CreatedBy {
		panic("publication identifiers changed")
	}
	r.ctx = ctx
	r.item.Status = announcements.AnnouncementStatusPUBLISHED
	return r.item, r.failure
}
func (r *repository) GetByID(context.Context, uuid.UUID) (announcements.Announcement, error) {
	return r.item, r.failure
}
func (*repository) CountAttachments(context.Context, uuid.UUID) (int64, error) { return 0, nil }
func (r *repository) CreateAttachment(_ context.Context, p announcements.CreateAnnouncementAttachmentParams) (announcements.AnnouncementAttachment, error) {
	r.attachment = announcements.AnnouncementAttachment{ID: p.ID, AnnouncementID: p.AnnouncementID, StorageKey: p.StorageKey, OriginalFileName: p.OriginalFileName, Checksum: p.Checksum}
	return r.attachment, r.failure
}
func (r *repository) GetAttachmentByID(_ context.Context, p announcements.GetAnnouncementAttachmentByIDParams) (announcements.AnnouncementAttachment, error) {
	if p.ID != r.attachment.ID || p.AnnouncementID != r.item.ID {
		panic("attachment identifiers changed")
	}
	return r.attachment, r.failure
}

type storage struct {
	body, key string
	deleted   bool
}

func (s *storage) Upload(_ context.Context, key string, body io.Reader, size int64, contentType string) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	if size != int64(len(data)) || contentType != "text/csv" {
		return errors.New("storage metadata changed")
	}
	s.key, s.body = key, string(data)
	return nil
}
func (s *storage) Download(_ context.Context, key string) (io.ReadCloser, error) {
	if key != s.key {
		return nil, errors.New("wrong storage key")
	}
	return io.NopCloser(strings.NewReader(s.body)), nil
}
func (s *storage) Delete(_ context.Context, key string) error { s.deleted = key == s.key; return nil }

type publisher struct{ items []announcements.Notification }

func (p *publisher) Publish(_ context.Context, item announcements.Notification) error {
	p.items = append(p.items, item)
	return errors.New("best-effort downstream failure")
}

func main() {
	type requestKey struct{}
	ctx := context.WithValue(context.Background(), requestKey{}, "consumer-request")
	actor, id := uuid.New(), uuid.New()
	repo := &repository{item: announcements.Announcement{ID: id, CreatedBy: actor, Title: "External consumer", Status: announcements.AnnouncementStatusDRAFT}}
	blobs := &storage{}
	events := &publisher{}
	core := announcements.NewServiceWithSettings(repo, announcements.ServicePorts{Storage: blobs, Notifications: events}, &announcements.Settings{})
	item, err := core.CreateAnnouncementFromInput(ctx, announcements.CreateAnnouncementInput{
		Title: "External consumer", Message: "Message", Level: "info", Status: "draft", AudienceType: "ALL_USERS", CreatedBy: actor,
		Summary: sql.NullString{String: "retained nullable value", Valid: false},
	})
	if err != nil || item.ID != id || repo.ctx != ctx || repo.created.Level != announcements.AnnouncementLevelINFO || repo.created.Summary.Valid || repo.created.Summary.String != "retained nullable value" {
		panic("create contract changed")
	}
	item, err = core.PublishAnnouncementNow(ctx, id, actor)
	if err != nil || item.Status != announcements.AnnouncementStatusPUBLISHED || len(events.items) != 2 || events.items[1].Type != string(announcements.AnnouncementPublished) || events.items[1].TargetRole != "admin" {
		panic("best-effort publication changed")
	}
	const body = "name,value\nregistry,fixture\n"
	attachment, err := core.UploadAttachment(ctx, id, announcements.UploadAnnouncementAttachmentInput{
		FileName: "fixture.csv", ContentType: "text/csv", Reader: strings.NewReader(body), Size: int64(len(body)), UploadedBy: actor,
	})
	if err != nil || !attachment.Checksum.Valid || len(attachment.Checksum.String) != 64 {
		panic(fmt.Sprintf("upload failed: %v", err))
	}
	download, err := core.OpenAttachmentDownload(ctx, id, attachment.ID)
	if err != nil {
		panic(err)
	}
	data, err := io.ReadAll(download.Reader)
	if err != nil {
		panic(err)
	}
	if err := download.Reader.Close(); err != nil {
		panic(err)
	}
	if string(data) != body || blobs.deleted {
		panic("download bytes changed")
	}
	cause := errors.New("persistence unavailable")
	repo.failure = cause
	if _, err := core.CreateAnnouncement(ctx, announcements.CreateAnnouncementParams{}); !errors.Is(err, cause) {
		panic("repository cause lost")
	}
	if announcements.ModulePath != "github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements" || announcements.Version() != "dev" {
		panic("local module identity changed")
	}
	fmt.Println("PASS: separate Announcements consumer (dev), lifecycle, notifications, nullable models, attachments and errors")
}
