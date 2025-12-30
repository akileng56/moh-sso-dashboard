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
import { type AuditLog } from "../lib/api/audit";

export const AuditLogDrawer: React.FC<{
  log: AuditLog | null;
  open: boolean;
  onClose: () => void;
}> = ({ log, open, onClose }) => {
  if (!open || !log) return null;

  const success =
    typeof log.metadata?.success === "boolean" ? log.metadata.success : null;

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
              {new Date(log.createdAt).toLocaleString()}
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
              <strong>User ID:</strong> {log.userId ?? "—"}
            </div>
            <div>
              <strong>Client:</strong> {log.metadata?.client_id ?? "—"}
            </div>
            <div>
              <strong>IP:</strong> {log.metadata?.ip ?? "—"}
            </div>
            <div>
              <strong>User Agent:</strong> {log.metadata?.user_agent ?? "—"}
            </div>
            <div>
              <strong>Country/City:</strong>{" "}
              {(log.metadata?.country ?? "—") +
                " / " +
                (log.metadata?.city ?? "—")}
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
