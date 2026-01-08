import {
  StructuredListWrapper,
  StructuredListBody,
  StructuredListRow,
  StructuredListCell,
  Tag,
  Button,
  Stack,
} from "@carbon/react";
import { Close } from "@carbon/icons-react";
import type { Notification } from "../../store/types/notifications.types";
import "./notifications-panel.css";

type Props = {
  notifications: Notification[];
  onClose: () => void;
  onMarkRead?: (id: string) => void;
  onView?: (notification: Notification) => void;
};

export function NotificationsPanel({
  notifications,
  onClose,
  onMarkRead,
  onView,
}: Props) {
  return (
    <div className="notifications-panel">
      {/* Header */}
      <div className="notifications-panel__header">
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
        <p className="notifications-panel__empty">No notifications</p>
      ) : (
        <StructuredListWrapper>
          <StructuredListBody>
            {notifications.map((n) => (
              <StructuredListRow
                key={n.id}
                className={`notification-row ${
                  !n.read ? "notification-row--unread" : ""
                }`}
                tabIndex={0}
                onClick={() => onView?.(n)}
              >
                <StructuredListCell>
                  <Stack gap={1}>
                    <span className="notification-title">{n.title}</span>

                    <span className="notification-message">{n.message}</span>

                    <span className="notification-meta">
                      {new Date(n.created_at).toLocaleString()}
                    </span>
                  </Stack>
                </StructuredListCell>

                <StructuredListCell className="notification-actions">
                  <Stack gap={2}>
                    <Tag size="sm" type={mapSeverity(n.severity)}>
                      {n.severity}
                    </Tag>

                    {!n.read && onMarkRead && (
                      <Button
                        size="sm"
                        kind="ghost"
                        onClick={(e) => {
                          e.stopPropagation();
                          onMarkRead(n.id);
                        }}
                      >
                        Mark read
                      </Button>
                    )}
                  </Stack>
                </StructuredListCell>
              </StructuredListRow>
            ))}
          </StructuredListBody>
        </StructuredListWrapper>
      )}
    </div>
  );
}

/* -----------------------------
 * Severity mapping (Carbon-safe)
 * ----------------------------- */
function mapSeverity(
  severity: "info" | "warning" | "critical"
): "red" | "yellow" | "gray" {
  switch (severity) {
    case "critical":
      return "red";
    case "warning":
      return "yellow";
    default:
      return "gray";
  }
}
