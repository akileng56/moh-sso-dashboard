import { Tile, Button, Tag, Stack, InlineLoading } from "@carbon/react";
import {
  Launch,
  Add,
  UserFollow,
  Security,
  Notification,
} from "@carbon/icons-react";
import { useSelector } from "react-redux";
import "./home.css";

import { QuickAction } from "../../../components/home/QuickAction";
import { SignalTile } from "../../../components/home/SignalTile";
import { EmptyState } from "../../../components/emptystate/EmptyState";
import { ErrorState } from "../../../components/errorstate/ErrorState";

import { selectUser } from "../../../store/auth/auth.selectors";
import { useListClientsQuery } from "../../../store/api/clients.api";
import { useAuditOverviewQuery } from "../../../store/api/metrics.api";
import { useGetNotificationsQuery } from "../../../store/api/notifications.api";

export default function HomePage() {
  // filters
  function toRFC3339(d: Date) {
    return d.toISOString();
  }

  const now = new Date();
  const start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);

  const from = toRFC3339(start);
  const to = toRFC3339(now);

  /* ==================================================
   * Identity
   * ================================================== */
  const user = useSelector(selectUser);

  /* ==================================================
   * Applications
   * ================================================== */
  const {
    data: clients = [],
    isLoading: appsLoading,
    isError: appsError,
    refetch: refetchApps,
  } = useListClientsQuery();

  /* ==================================================
   * Metrics (System Signals)
   * ================================================== */
  const {
    data: metrics,
    isLoading: metricsLoading,
    isError: metricsError,
  } = useAuditOverviewQuery({ from, to }, { skip: !from || !to });

  /* ==================================================
   * Notifications
   * ================================================== */
  const { data: notifications = [], isLoading: notificationsLoading } =
    useGetNotificationsQuery({
      unread: false,
      limit: 5,
      offset: 0,
    });

  return (
    <div style={{ padding: 16, display: "grid", gap: 16 }}>
      {/* ==================================================
       * Header
       * ================================================== */}
      <div>
        <h3 style={{ margin: 0 }}>Admin Overview</h3>
        <p style={{ marginTop: 6, opacity: 0.8 }}>
          System status, applications, and quick actions.
        </p>
      </div>

      {/* ==================================================
       * Identity
       * ================================================== */}
      <Tile>
        {!user ? (
          <InlineLoading description="Loading user…" />
        ) : (
          <Stack gap={3}>
            <h4>Welcome, {user.username}</h4>

            <Stack orientation="horizontal" gap={4}>
              <span>{user.email}</span>
              <Tag key={user.id} type="blue">
                {user.realm_roles.join(", ")}
              </Tag>
              <span>
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
       * Applications
       * ================================================== */}
      <Tile>
        <h4>Your applications</h4>

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
                    kind="primary"
                    renderIcon={Launch}
                    disabled={!client.enabled}
                    onClick={() => {}}
                  >
                    Launch
                  </Button>
                </Stack>
              </Tile>
            ))}
          </div>
        )}
      </Tile>

      {/* ==================================================
       * Quick Actions
       * ================================================== */}
      <Tile>
        <h4>Quick actions</h4>

        <div className="home-grid">
          <QuickAction
            icon={<Add />}
            label="Create user"
            href="/admin/users/new"
          />
          <QuickAction
            icon={<UserFollow />}
            label="Import users"
            href="/admin/users/import"
          />
          <QuickAction
            icon={<Security />}
            label="Security alerts"
            href="/admin/metrics/security"
          />
        </div>
      </Tile>

      {/* ==================================================
       * System Signals
       * ================================================== */}
      <Tile>
        <h4>System signals</h4>

        {metricsLoading && <InlineLoading description="Loading metrics…" />}

        {metricsError && (
          <ErrorState
            title="Failed to load metrics"
            description="System metrics are currently unavailable."
          />
        )}

        {metrics && (
          <div className="home-grid">
            <SignalTile label="Active users today" value={0} />
            <SignalTile
              label="Failed logins (24h)"
              value={metrics?.failed_logins}
              severity="warning"
            />
            <SignalTile label="Suspicious logins" value={0} severity="danger" />
          </div>
        )}
      </Tile>

      {/* ==================================================
       * Notifications
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
              <div key={n.id} className="notification-item">
                <Notification size={16} />
                <span>{n.message}</span>
              </div>
            ))}
          </Stack>
        )}
      </Tile>
    </div>
  );
}
