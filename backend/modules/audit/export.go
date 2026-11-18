package audit

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"io"
	"strconv"
	"time"
)

// BuildExport prepares the public export payload and its integrity manifest.
// It validates JSON serialization before an HTTP adapter starts streaming.
func BuildExport(from, to time.Time, items []AuditLogResponse) (AuditExportResponse, error) {
	payload, err := json.Marshal(items)
	if err != nil {
		return AuditExportResponse{}, err
	}
	sum := sha256.Sum256(payload)
	return AuditExportResponse{
		Manifest: AuditExportManifest{
			GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
			Range:         map[string]string{"from": from.UTC().Format(time.RFC3339), "to": to.UTC().Format(time.RFC3339)},
			RecordCount:   len(items),
			HashAlgo:      "sha256",
			PayloadSHA256: hex.EncodeToString(sum[:]),
		},
		Items: items,
	}, nil
}

// WriteCSV writes the existing audit columns to any writer. HTTP headers and
// handling of partial response failures belong to the caller. Encoding and
// write/flush errors are returned without appending a different response format.
func WriteCSV(dst io.Writer, rows []AuditLogResponse) error {
	w := csv.NewWriter(dst)
	if err := w.Write([]string{"created_at", "user_id", "username", "action", "ip", "client_id", "success", "metadata_json"}); err != nil {
		return err
	}
	for _, row := range rows {
		metadata, err := json.Marshal(row.Metadata)
		if err != nil {
			return err
		}
		success := ""
		if row.Success != nil {
			success = strconv.FormatBool(*row.Success)
		}
		if err := w.Write([]string{row.CreatedAt, row.UserID, row.Username, row.Action, row.IP, row.ClientID, success, string(metadata)}); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}
