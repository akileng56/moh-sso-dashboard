package audit

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestBuildExportPreservesTypedPayloadAndUTCManifest(t *testing.T) {
	zone := time.FixedZone("fixture", 3*60*60)
	from := time.Date(2026, 11, 18, 8, 0, 0, 0, zone)
	to := from.Add(24 * time.Hour)
	for _, items := range [][]AuditLogResponse{nil, {}, auditFixtureItems()} {
		export, err := BuildExport(from, to, items)
		if err != nil {
			t.Fatal(err)
		}
		payload, _ := json.Marshal(items)
		hash := sha256.Sum256(payload)
		if !reflect.DeepEqual(export.Items, items) || export.Manifest.RecordCount != len(items) || export.Manifest.HashAlgo != "sha256" || export.Manifest.PayloadSHA256 != hex.EncodeToString(hash[:]) {
			t.Fatalf("export payload/hash changed: %+v", export)
		}
		if export.Manifest.Range["from"] != "2026-11-18T05:00:00Z" || export.Manifest.Range["to"] != "2026-11-19T05:00:00Z" {
			t.Fatalf("manifest range not normalized: %+v", export.Manifest)
		}
		if _, err := time.Parse(time.RFC3339, export.Manifest.GeneratedAt); err != nil || !strings.HasSuffix(export.Manifest.GeneratedAt, "Z") {
			t.Fatalf("invalid generated time: %s", export.Manifest.GeneratedAt)
		}
	}
}

func TestExportRejectsUnsupportedMetadata(t *testing.T) {
	items := auditFixtureItems()
	items[0].Metadata["unsupported"] = make(chan int)
	if _, err := BuildExport(time.Now(), time.Now(), items); err == nil {
		t.Fatal("invalid JSON was hashed as an empty payload")
	}
	var dst bytes.Buffer
	if err := WriteCSV(&dst, items); err == nil {
		t.Fatal("CSV metadata serialization error was ignored")
	}
}

type failingExportWriter struct{ err error }

func (w failingExportWriter) Write([]byte) (int, error) { return 0, w.err }

func TestCSVReturnsWriteAndFlushFailures(t *testing.T) {
	failure := errors.New("writer unavailable")
	for _, size := range []int{1, 10000} {
		items := auditFixtureItems()
		items[0].Metadata["payload"] = strings.Repeat("x", size)
		if err := WriteCSV(failingExportWriter{failure}, items); !errors.Is(err, failure) {
			t.Fatalf("write/flush failure lost: %v", err)
		}
	}
}

func TestCSVEmptyRowsAndOptionalSuccess(t *testing.T) {
	const header = "created_at,user_id,username,action,ip,client_id,success,metadata_json\n"
	for _, items := range [][]AuditLogResponse{nil, {}} {
		var dst bytes.Buffer
		if err := WriteCSV(&dst, items); err != nil || dst.String() != header {
			t.Fatalf("empty export changed: %q %v", dst.String(), err)
		}
	}
	items := auditFixtureItems()
	items = append(items, AuditLogResponse{ID: "unknown-success", Metadata: map[string]any{}})
	var dst bytes.Buffer
	if err := WriteCSV(&dst, items); err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(&dst).ReadAll()
	if err != nil || len(rows) != 3 || rows[1][6] != "false" || rows[2][6] != "" {
		t.Fatalf("optional success changed: %v %v", rows, err)
	}
}
