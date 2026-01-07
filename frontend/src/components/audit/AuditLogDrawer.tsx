import React from "react";
import {
  ComposedModal,
  ModalBody,
  ModalHeader,
  ModalFooter,
  Button,
  CodeSnippet,
  Tag,
} from "@carbon/react";
import { type AuditLog } from "../../store/types/audit.types";

export const AuditLogDrawer: React.FC<{
  log: AuditLog | null;
  open: boolean;
  onClose: () => void;
}> = ({ log, open, onClose }) => {
  if (!open || !log) return null;

  const success =
    typeof log.metadata?.RawMessage.success === "boolean"
      ? log.metadata?.RawMessage.success
      : null;

  return (
    <ComposedModal open={open} onClose={onClose} size="lg">
      <ModalHeader title="Audit Log Details" />
      <ModalBody>
        <div style={{ display: "grid", gap: 12 }}>
          <div>
            <div
              style={{
                display: "flex",
                gap: 8,
                alignItems: "center",
                flexWrap: "wrap",
              }}
            >
              <strong>{log.action}</strong>
              {success !== null && (
                <Tag type={success ? "green" : "red"}>
                  {success ? "success" : "failure"}
                </Tag>
              )}
            </div>
            <div style={{ opacity: 0.8 }}>
              {new Date(log.created_at.Time).toLocaleString()}
            </div>
          </div>

          <div style={{ display: "grid", gap: 6 }}>
            <div>
              <strong>ID:</strong> {log.id}
            </div>
            <div>
              <strong>Actor:</strong> {log.username ?? "System"}
            </div>
            <div>
              <strong>User ID:</strong> {log.user_id ?? "—"}
            </div>
            <div>
              <strong>Client:</strong>{" "}
              {log.metadata?.RawMessage.client_id ?? "—"}
            </div>
            <div>
              <strong>IP:</strong> {log.metadata?.RawMessage.ip ?? "—"}
            </div>
            <div>
              <strong>User Agent:</strong>{" "}
              {log.metadata?.RawMessage.user_agent ?? "—"}
            </div>
            <div>
              <strong>Country/City:</strong>{" "}
              {(log.metadata?.RawMessage.country ?? "—") +
                " / " +
                (log.metadata?.RawMessage.city ?? "—")}
            </div>
          </div>

          <div>
            <strong>Metadata</strong>
            <div style={{ marginTop: 8 }}>
              <CodeSnippet type="multi" hideCopyButton={false}>
                {JSON.stringify(log.metadata ?? {}, null, 2)}
              </CodeSnippet>
            </div>
          </div>
        </div>
      </ModalBody>

      <ModalFooter>
        <Button kind="secondary" onClick={onClose}>
          Close
        </Button>
      </ModalFooter>
    </ComposedModal>
  );
};
