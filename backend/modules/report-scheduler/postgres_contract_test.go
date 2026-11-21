package reportscheduler_test

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	scheduler "github.com/mohuganda/moh-sso-dashboard/backend/modules/report-scheduler"
)

// Each contender uses a real connection. A single persisted job may be claimed
// once even when multiple worker instances poll concurrently.
func assertSingleClaim(t *testing.T, claim func() (bool, error)) {
	t.Helper()
	type result struct {
		found bool
		err   error
	}
	results := make(chan result, 8)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < cap(results); i++ {
		wg.Add(1)
		go func() { defer wg.Done(); <-start; found, err := claim(); results <- result{found, err} }()
	}
	close(start)
	wg.Wait()
	close(results)
	count := 0
	for result := range results {
		if result.err != nil {
			t.Fatal(result.err)
		}
		if result.found {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("concurrent claims=%d, want 1", count)
	}
}

func TestPostgresHostSchemaClaimAndRecoveryContract(t *testing.T) {
	dsn := os.Getenv("TEST_REPORT_SCHEDULER_SCHEMA_DB_URL")
	if dsn == "" {
		t.Skip("run infra/scripts/test-report-scheduler-package.py with isolated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(8)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	repo := scheduler.NewRepository(db)
	item, err := repo.CreateSchedule(ctx, scheduler.CreateScheduleRequest{
		HealthBIReportID: "fixture/report", ReportName: "Standalone schema contract", Frequency: "daily",
		Timezone: "Africa/Kampala", PeriodStrategy: "previous_day", OutputFormat: "pdf", Parameters: map[string]any{},
	}, "standalone-fixture")
	if err != nil {
		t.Fatal(err)
	}
	// A disposable fixture owns this schema. Delete only this schedule's records.
	defer func() {
		_, _ = db.Exec("DELETE FROM report_executions WHERE schedule_id=$1::uuid", item.ID)
		_, _ = db.Exec("DELETE FROM report_schedules WHERE id=$1::uuid", item.ID)
	}()
	now := time.Now().UTC()
	execSQL := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	execSQL("UPDATE report_schedules SET next_run_at=$2 WHERE id=$1::uuid", item.ID, now.Add(-time.Minute))
	assertSingleClaim(t, func() (bool, error) { _, _, found, err := repo.ClaimDueExecution(ctx, now); return found, err })
	executions, err := repo.ListExecutions(ctx, "standalone-fixture", false, scheduler.HealthContext{}, scheduler.ListOptions{Limit: 10})
	if err != nil || len(executions) != 1 {
		t.Fatalf("persisted execution count=%d error=%v", len(executions), err)
	}
	execution := executions[0]
	if execution.ScheduledFor == nil || execution.ResolvedPeriod == nil {
		t.Fatal("claim lost schedule/period snapshot")
	}
	if err := repo.UpdateExecutionJob(ctx, execution.ID, "job-fixture", "generating", ""); err != nil {
		t.Fatal(err)
	}
	poll := func() (bool, error) { items, err := repo.ClaimGeneratingExecutions(ctx, 1); return len(items) > 0, err }
	assertSingleClaim(t, poll)
	if found, err := poll(); err != nil || found {
		t.Fatalf("fresh poll claim reclaimed: %v %v", found, err)
	}
	execSQL("UPDATE report_executions SET poll_claimed_at=now()-interval '3 minutes' WHERE id=$1::uuid", execution.ID)
	assertSingleClaim(t, poll)
	if err := repo.ScheduleExecutionRetry(ctx, execution.ID, "temporary failure", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, found, err := repo.ClaimRetryExecution(ctx, now); err != nil || found {
		t.Fatalf("future retry claimed: %v %v", found, err)
	}
	if err := repo.ScheduleExecutionRetry(ctx, execution.ID, "temporary failure", now.Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	assertSingleClaim(t, func() (bool, error) { _, found, err := repo.ClaimRetryExecution(ctx, now); return found, err })
	if err := repo.UpdateExecutionJob(ctx, execution.ID, "job-fixture", "delivering", ""); err != nil {
		t.Fatal(err)
	}
	artifact, err := repo.CreateArtifact(ctx, scheduler.Artifact{ExecutionID: execution.ID, FileName: "fixture.pdf", ContentType: "application/pdf", ObjectKey: "fixture.pdf"})
	if err != nil {
		t.Fatal(err)
	}
	delivery, err := repo.CreatePendingDelivery(ctx, execution.ID, artifact.ID, "email", "fixture@example.invalid", "email")
	if err != nil {
		t.Fatal(err)
	}
	duplicate, err := repo.CreatePendingDelivery(ctx, execution.ID, artifact.ID, "email", "fixture@example.invalid", "email")
	if err != nil || duplicate.ID != delivery.ID {
		t.Fatalf("duplicate delivery created: %v", err)
	}
	due := time.Now().UTC()
	claimDelivery := func() (bool, error) { _, found, err := repo.ClaimDueDelivery(ctx, due); return found, err }
	assertSingleClaim(t, claimDelivery)
	execSQL("UPDATE report_deliveries SET last_attempt_at=now()-interval '3 minutes' WHERE id=$1::uuid", delivery.ID)
	assertSingleClaim(t, claimDelivery)
	if err := repo.MarkDeliverySent(ctx, delivery.ID); err != nil {
		t.Fatal(err)
	}
	if found, err := claimDelivery(); err != nil || found {
		t.Fatalf("sent delivery reclaimed: %v %v", found, err)
	}
	if err := repo.UpdateWorkerHeartbeat(ctx, "standalone-fixture", time.Second, now, nil, nil, ""); err != nil {
		t.Fatal(err)
	}
}
