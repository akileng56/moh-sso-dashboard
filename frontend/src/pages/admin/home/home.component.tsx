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

import { QuickAction } from "../../../components/home/QuickAction";
import { SignalTile } from "../../../components/home/SignalTile";
import { EmptyState } from "../../../components/emptystate/EmptyState";
import { ErrorState } from "../../../components/errorstate/ErrorState";

import { selectUser } from "../../../store/auth/auth.selectors";
import { useListClientsQuery } from "../../../store/api/clients.api";
import { useAuditOverviewQuery } from "../../../store/api/metrics.api";
import { useGetNotificationsQuery } from "../../../store/api/notifications.api";

/* -----------------------------
 * Utils
 * ----------------------------- */
const toRFC3339 = (d: Date) => d.toISOString();

export default function HomePage() {
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
   * Notifications
   * ----------------------------- */
  const { data: notifications = [], isLoading: notificationsLoading } =
    useGetNotificationsQuery({
      unread: false,
      limit: 5,
      offset: 0,
    });

  return (
    <div style={{ padding: 16, display: "grid", gap: 16 }}>
      <div>
        {" "}
        <h3 style={{ margin: 0 }}>Admin Overview</h3>{" "}
        <p style={{ marginTop: 6, opacity: 0.8 }}>
          {" "}
          System status, applications, and quick actions.{" "}
        </p>{" "}
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
            <SignalTile label="Active users today" value={0} />
            <SignalTile
              label="Failed logins (24h)"
              value={metrics.failed_logins}
              severity="warning"
            />
            <SignalTile label="Suspicious logins" value={0} severity="danger" />
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
       * QUICK ACTIONS
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
       * NOTIFICATIONS
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
                <span className="notification-text">{n.message}</span>
              </div>
            ))}
          </Stack>
        )}
      </Tile>
    </div>
  );
}
