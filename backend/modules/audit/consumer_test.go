package audit_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	audit "github.com/mohuganda/moh-sso-dashboard/backend/modules/audit"
)

type consumerRepository struct {
	item  audit.AuditLogResponse
	input audit.ListAuditLogsInput
	ctx   context.Context
	reads int
}

var _ audit.Repository = (*consumerRepository)(nil)

func (r *consumerRepository) ListAuditLogs(ctx context.Context, input audit.ListAuditLogsInput) (audit.AuditLogListResponse, error) {
	r.ctx = ctx
	r.input = input
	r.reads++
	return audit.AuditLogListResponse{Items: []audit.AuditLogResponse{r.item}}, nil
}
func (r *consumerRepository) GetAuditLog(context.Context, uuid.UUID) (audit.AuditLogResponse, error) {
	return r.item, nil
}
func (*consumerRepository) ListAuditActions(context.Context) (audit.AuditActionsResponse, error) {
	return audit.AuditActionsResponse{Actions: []string{"login"}}, nil
}
func (*consumerRepository) AuditMetricsOverview(context.Context, audit.AuditWindowInput) (audit.AuditMetricsOverviewResponse, error) {
	return audit.AuditMetricsOverviewResponse{TotalEvents: 1}, nil
}
func (*consumerRepository) FailedLoginsByDay(context.Context, audit.AuditWindowInput) (audit.AuditFailedLoginsByDayResponse, error) {
	return audit.AuditFailedLoginsByDayResponse{Series: []audit.AuditFailedLoginsByDayPoint{}}, nil
}
func (*consumerRepository) TopFailureIPs(context.Context, audit.TopFailureIPsInput) (audit.AuditTopFailureIPsResponse, error) {
	return audit.AuditTopFailureIPsResponse{Items: []audit.AuditTopFailureIPPoint{}}, nil
}
func (r *consumerRepository) ExportAuditLogs(context.Context, audit.ExportAuditLogsInput) ([]audit.AuditLogResponse, error) {
	return []audit.AuditLogResponse{r.item}, nil
}

type consumerCache map[string][]byte

func (c consumerCache) Get(_ context.Context, key string, dest any) (bool, error) {
	data, found := c[key]
	if !found {
		return false, nil
	}
	err := json.Unmarshal(data, dest)
	return err == nil, err
}
func (c consumerCache) Set(_ context.Context, key string, value any, _ time.Duration) error {
	data, err := json.Marshal(value)
	if err == nil {
		c[key] = data
	}
	return err
}

func TestExternalConsumerReadCacheAndExport(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo := &consumerRepository{item: audit.AuditLogResponse{ID: id.String(), Action: "login", Metadata: map[string]any{"client_id": "fixture"}}}
	service := audit.NewService(repo)
	reader := audit.NewCachedReader(service, consumerCache{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := 0; i < 2; i++ {
		page, err := reader.ListAuditLogs(ctx, audit.ListAuditLogsInput{Action: "login", UserID: &id, Limit: 50}, "action=login&user=fixture")
		if err != nil || len(page.Items) != 1 || page.Items[0].ID != id.String() {
			t.Fatalf("public read failed: %+v %v", page, err)
		}
	}
	if repo.reads != 1 || repo.ctx != ctx || repo.input.UserID == nil || *repo.input.UserID != id {
		t.Fatal("public cache/query boundary lost data")
	}
	if item, err := service.GetAuditLog(ctx, id); err != nil || item.ID != id.String() {
		t.Fatal("public detail query failed")
	}
	if actions, err := service.ListAuditActions(ctx); err != nil || len(actions.Actions) != 1 {
		t.Fatal("public action query failed")
	}
	window := audit.AuditWindowInput{StartTime: time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC)}
	if metrics, err := service.AuditMetricsOverview(ctx, window); err != nil || metrics.TotalEvents != 1 {
		t.Fatal("public overview query failed")
	}
	if _, err := service.FailedLoginsByDay(ctx, window); err != nil {
		t.Fatal(err)
	}
	if _, err := service.TopFailureIPs(ctx, audit.TopFailureIPsInput{Limit: 10}); err != nil {
		t.Fatal(err)
	}
	items, err := service.ExportAuditLogs(ctx, audit.ExportAuditLogsInput{StartTime: window.StartTime, EndTime: window.StartTime.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	export, err := audit.BuildExport(window.StartTime, window.StartTime.Add(time.Hour), items)
	if err != nil || export.Manifest.RecordCount != 1 {
		t.Fatalf("public export failed: %+v %v", export, err)
	}
	var output bytes.Buffer
	if err := audit.WriteCSV(&output, export.Items); err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(&output).ReadAll()
	if err != nil || len(rows) != 2 || rows[1][3] != "login" {
		t.Fatalf("public CSV failed: %v %v", rows, err)
	}
}

func ExampleWriteCSV() {
	var output bytes.Buffer
	if err := audit.WriteCSV(&output, nil); err != nil {
		panic(err)
	}
	fmt.Print(output.String())
	// Output: created_at,user_id,username,action,ip,client_id,success,metadata_json
}
