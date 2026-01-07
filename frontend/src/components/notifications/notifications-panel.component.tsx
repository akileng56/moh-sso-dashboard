import {
  StructuredListWrapper,
  StructuredListBody,
  StructuredListRow,
  StructuredListCell,
  Tag,
  Button,
} from "@carbon/react";
import { Close } from "@carbon/icons-react";
import type { Notification } from "../../store/types/notifications.types";

type Props = {
  notifications: Notification[];
  onClose: () => void;
};

export function NotificationsPanel({ notifications, onClose }: Props) {
  return (
    <div style={{ width: 360, padding: "1rem" }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          marginBottom: "1rem",
        }}
      >
        <strong>Notifications</strong>
        <Button
          size="sm"
          kind="ghost"
          hasIconOnly
          iconDescription="Close"
          renderIcon={Close}
          onClick={onClose}
        />
      </div>

      {notifications.length === 0 ? (
        <p style={{ opacity: 0.7 }}>No notifications</p>
      ) : (
        <StructuredListWrapper>
          <StructuredListBody>
            {notifications.map((n) => (
              <StructuredListRow
                key={n.id}
                style={{
                  cursor: "pointer",
                  backgroundColor: n.read ? "inherit" : "#f4f4f4",
                }}
              >
                <StructuredListCell>
                  <div style={{ fontWeight: 600 }}>{n.title}</div>
                  <div style={{ fontSize: "0.8rem", opacity: 0.7 }}>
                    {new Date(n.created_at).toLocaleString()}
                  </div>
                  <div style={{ marginTop: 4 }}>{n.message}</div>
                </StructuredListCell>

                <StructuredListCell>
                  <Tag size="sm" type={mapSeverity(n.severity)}>
                    {n.severity}
                  </Tag>
                </StructuredListCell>
              </StructuredListRow>
            ))}
          </StructuredListBody>
        </StructuredListWrapper>
      )}
    </div>
  );
}

function mapSeverity(sev: string) {
  switch (sev) {
    case "critical":
      return "red";
    case "warning":
      return "yellow";
    default:
      return "gray";
  }
}
