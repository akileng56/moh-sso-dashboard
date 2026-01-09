import { Tile, Button, Tag, Stack, InlineLoading } from "@carbon/react";
import {
  Launch,
  Add,
  UserFollow,
  Security,
  Notification,
} from "@carbon/icons-react";
import { useMemo } from "react";
import { useSelector } from "react-redux";
import "./home.css";

import { SignalTile } from "../../../components/home/signal-tile/signal-tile.component";
import { EmptyState } from "../../../components/emptystate/EmptyState";
import { ErrorState } from "../../../components/errorstate/ErrorState";

import { selectUser } from "../../../store/auth/auth.selectors";
import { useListClientsQuery } from "../../../store/api/clients.api";
import { useAuditOverviewQuery } from "../../../store/api/metrics.api";
import {
  useGetNotificationsQuery,
  useMarkNotificationAsReadMutation,
} from "../../../store/api/notifications.api";
import { QuickAction } from "../../../components/home/quick-action/quick-action.component";
import { CreateUserPanel } from "../../../components/panels/create-user-panel";
import { ImportUsersPanel } from "../../../components/panels/import-users-panel";
import { useHeaderPanel } from "../../../components/header-panel/header-panel.context";

/* -----------------------------
 * Utils
 * ----------------------------- */
const toRFC3339 = (d: Date) => d.toISOString();

export default function HomePage() {
  const { openPanel } = useHeaderPanel();
  /* -----------------------------
   * Date range (last 7 days)
   * ----------------------------- */
  const { from, to } = useMemo(() => {
    const now = new Date();
    const start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
    return {
      from: toRFC3339(start),
      to: toRFC3339(now),
    };
  }, []);

  /* -----------------------------
   * Identity
   * ----------------------------- */
  const user = useSelector(selectUser);

  /* -----------------------------
   * Applications
   * ----------------------------- */
  const {
    data: clients = [],
    isLoading: appsLoading,
    isError: appsError,
    refetch: refetchApps,
  } = useListClientsQuery();

  /* -----------------------------
   * Metrics
   * ----------------------------- */
  const {
    data: metrics,
    isLoading: metricsLoading,
    isError: metricsError,
    refetch: refetchMetrics,
  } = useAuditOverviewQuery({ from, to }, { skip: !from || !to });

  /* -----------------------------
   * Security Health Score
   * ----------------------------- */
  const securityScore = useMemo(() => {
    if (!metrics) return 100;

    let score = 100;
    score -= Math.min(metrics.failed_logins * 2, 40);
    score -= Math.min((metrics.suspicious_logins ?? 0) * 5, 40);

    return Math.max(score, 0);
  }, [metrics]);

  /* -----------------------------
   * Notifications
   * ----------------------------- */
  const { data: notifications = [], isLoading: notificationsLoading } =
    useGetNotificationsQuery({
      unread: true,
      limit: 5,
      offset: 0,
    });

  const [markNotificationAsRead] = useMarkNotificationAsReadMutation();

  return (
    <div style={{ padding: 16, display: "grid", gap: 16 }}>
      {/* ==================================================
       * HEADER
       * ================================================== */}
      <div>
        <h3 style={{ margin: 0 }}>Admin Overview</h3>
        <p style={{ marginTop: 6, opacity: 0.8 }}>
          System status, applications, and quick actions.
        </p>
      </div>
      {/* ==================================================
       * HERO / WELCOME
       * ================================================== */}
      <Tile className="home-hero">
        {!user ? (
          <InlineLoading description="Loading user…" />
        ) : (
          <Stack gap={3}>
            <h3 style={{ margin: 0 }}>Welcome back, {user.username}</h3>

            <Stack orientation="horizontal" gap={3}>
              <Tag type="blue">{user.realm_roles.join(", ")}</Tag>
              <span>{user.email}</span>
              <span className="muted">
                Last login:{" "}
                {user.last_login_at
                  ? new Date(user.last_login_at).toLocaleString()
                  : "—"}
              </span>
            </Stack>
          </Stack>
        )}
      </Tile>
      {/* ==================================================
       * SYSTEM SIGNALS
       * ================================================== */}
      <Tile>
        <h4>System signals</h4>

        {metricsLoading && <InlineLoading description="Loading metrics…" />}

        {metricsError && (
          <ErrorState
            title="Failed to load metrics"
            description="System metrics are currently unavailable."
            primaryAction={{ label: "Retry", onClick: refetchMetrics }}
          />
        )}

        {metrics && (
          <div className="home-grid">
            <SignalTile
              label="Security health score"
              value={`${securityScore}%`}
              severity={
                securityScore > 80
                  ? "success"
                  : securityScore > 50
                  ? "warning"
                  : "danger"
              }
            />
            <SignalTile
              label="Failed logins (24h)"
              value={metrics.failed_logins}
              severity="warning"
            />
            <SignalTile
              label="Suspicious logins"
              value={metrics.suspicious_logins ?? 0}
              severity="danger"
            />
          </div>
        )}
      </Tile>
      {/* ==================================================
       * APPLICATIONS
       * ================================================== */}
      <Tile>
        <h4>
          Your applications{" "}
          <Tag size="sm" type="cool-gray">
            {clients.length}
          </Tag>
        </h4>

        {appsLoading && <InlineLoading description="Loading applications…" />}

        {appsError && (
          <ErrorState
            title="Failed to load applications"
            description="Unable to fetch assigned applications."
            primaryAction={{ label: "Retry", onClick: refetchApps }}
          />
        )}

        {!appsLoading && !appsError && clients.length === 0 && (
          <EmptyState
            title="No applications assigned"
            description="You do not have access to any applications yet."
          />
        )}

        {!appsLoading && !appsError && clients.length > 0 && (
          <div className="home-grid">
            {clients.map((client) => (
              <Tile key={client.clientId} className="app-tile">
                <Stack gap={4}>
                  <div className="app-header">
                    <strong>{client.name}</strong>
                    {!client.enabled && <Tag type="gray">Disabled</Tag>}
                  </div>

                  <p className="app-description">{client.description ?? "—"}</p>

                  <Button
                    size="sm"
                    kind={client.enabled ? "primary" : "secondary"}
                    renderIcon={Launch}
                    disabled={!client.enabled}
                  >
                    {client.enabled ? "Launch" : "Disabled"}
                  </Button>
                </Stack>
              </Tile>
            ))}
          </div>
        )}
      </Tile>
      {/* ==================================================
       * CLIENT USAGE METRICS
       * ================================================== */}
      {/* <Tile>
        <h4>Client usage (last 7 days)</h4>

        {!metrics?.top_clients?.length && (
          <EmptyState
            title="No usage data"
            description="No client activity recorded for this period."
          />
        )}

        {metrics?.top_clients?.length > 0 && (
          <Stack gap={3}>
            {metrics.top_clients.map((c) => (
              <div
                key={c.client_id}
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <strong>{c.name}</strong>
                <Tag type="cool-gray">{c.logins} logins</Tag>
              </div>
            ))}
          </Stack>
        )}
      </Tile> */}
      {/* ==================================================
       * QUICK ACTIONS
       * ================================================== */}
      <Tile>
        <h4>Quick actions</h4>

        <div className="home-grid">
          <QuickAction
            icon={<Add />}
            label="Create user"
            description="Add a new user to the system"
            onClick={() =>
              openPanel({
                title: "Create user",
                content: <CreateUserPanel />,
                size: "md",
              })
            }
          />

          <QuickAction
            icon={<UserFollow />}
            label="Import users"
            description="Bulk upload users via CSV"
            onClick={() =>
              openPanel({
                title: "Import users",
                content: <ImportUsersPanel />,
                size: "lg",
              })
            }
          />

          <QuickAction
            icon={<Security />}
            label="Security alerts"
            description="Review suspicious activity"
            href="/admin/metrics/security"
            tone="warning"
          />
        </div>
      </Tile>

      {/* ==================================================
       * NOTIFICATIONS (ACTIONABLE)
       * ================================================== */}
      <Tile>
        <h4>Notifications</h4>

        {notificationsLoading && (
          <InlineLoading description="Loading notifications…" />
        )}

        {!notificationsLoading && notifications.length === 0 && (
          <EmptyState
            title="No notifications"
            description="You're all caught up."
          />
        )}

        {!notificationsLoading && notifications.length > 0 && (
          <Stack gap={3}>
            {notifications.map((n) => (
              <div
                key={n.id}
                className={`notification-item ${
                  !n.read ? "notification-unread" : ""
                }`}
              >
                <Notification size={16} />

                <div className="notification-content">
                  <strong className="notification-title">{n.title}</strong>
                  <span className="notification-message">{n.message}</span>

                  <span className="notification-meta">
                    {new Date(n.created_at).toLocaleString()}
                  </span>
                </div>

                <Tag
                  size="sm"
                  type={
                    n.severity === "critical"
                      ? "red"
                      : n.severity === "warning"
                      ? "yellow"
                      : "blue"
                  }
                >
                  {n.severity}
                </Tag>

                <Button
                  size="sm"
                  kind="ghost"
                  onClick={() => markNotificationAsRead(n.id)}
                >
                  Dismiss
                </Button>
              </div>
            ))}
          </Stack>
        )}
      </Tile>
    </div>
  );
}
