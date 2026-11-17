package reportscheduler

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
)

func TestRecipientOptionsPostgresDirectory(t *testing.T) {
	dsn := os.Getenv("TEST_REPORT_SCHEDULER_DB_URL")
	if dsn == "" {
		t.Skip("set TEST_REPORT_SCHEDULER_DB_URL to a disposable PostgreSQL database")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	ctx := context.Background()
	// Session-local temporary tables shadow real tables and are removed on close.
	_, err = db.ExecContext(ctx, `CREATE TEMP TABLE users(id TEXT,username TEXT,first_name TEXT,last_name TEXT,email TEXT,enabled BOOLEAN);
 CREATE TEMP TABLE ihp_rbac_groups(id TEXT,name TEXT,display_name TEXT,path TEXT,enabled BOOLEAN);
 INSERT INTO users VALUES ('one','testuser','Test','User','test@example.invalid',TRUE),('two','noemail','No','Email','',TRUE),('three','disabled','Hidden','User','hidden@example.invalid',FALSE),('four','service-account-worker','','','service@example.invalid',TRUE);
 INSERT INTO ihp_rbac_groups VALUES ('group-one','Schedulers','Report Schedulers','/MOH/Schedulers',TRUE),('group-two','Disabled','Disabled','/Hidden',FALSE);`)
	if err != nil {
		t.Fatal(err)
	}
	repo := &postgresRepository{db: db}
	options, err := repo.ListRecipientOptions(ctx, "email", "", 100, nil)
	if err != nil || len(options) != 1 || options[0].Value != "one" || !strings.Contains(options[0].Label, "test@example.invalid") {
		t.Fatalf("email directory: %+v %v", options, err)
	}
	options, err = repo.ListRecipientOptions(ctx, "portal", "", 100, nil)
	if err != nil || len(options) != 2 {
		t.Fatalf("portal directory: %+v %v", options, err)
	}
	options, err = repo.ListRecipientOptions(ctx, "group", "Report", 100, nil)
	if err != nil || len(options) != 1 || options[0].Value != "group-one" || !strings.Contains(options[0].Label, "/MOH/Schedulers") {
		t.Fatalf("groups: %+v %v", options, err)
	}
	options, err = repo.ListRecipientOptions(ctx, "email", "not-found", 100, nil)
	if err != nil || len(options) != 0 {
		t.Fatalf("search not applied: %+v %v", options, err)
	}
	options, err = repo.ListRecipientOptions(ctx, "portal", "not-found", 100, []string{"one"})
	if err != nil || len(options) != 1 || options[0].Value != "one" {
		t.Fatalf("saved user selection missing: %+v %v", options, err)
	}
	options, err = repo.ListRecipientOptions(ctx, "group", "not-found", 100, []string{"/MOH/Schedulers"})
	if err != nil || len(options) != 1 || options[0].Value != "/MOH/Schedulers" {
		t.Fatalf("saved group path not preserved: %+v %v", options, err)
	}
	options, err = repo.ListRecipientOptions(ctx, "email", "not-found", 100, []string{"test@example.invalid"})
	if err != nil || len(options) != 1 || options[0].Type != "email" || options[0].Value != "test@example.invalid" {
		t.Fatalf("saved email not preserved: %+v %v", options, err)
	}
	options, err = repo.ListRecipientOptions(ctx, "portal", "", 1, nil)
	if err != nil || len(options) != 1 {
		t.Fatalf("limit not applied: %+v %v", options, err)
	}
}
