import React, { useEffect, useState } from "react";
import { Tile, InlineLoading } from "@carbon/react";

type Overview = {
  totalEvents: number;
  totalFailures: number;
  failedLogins: number;
  successfulLogins: number;
};

export const AuditMetricsPanel: React.FC<{ from: string; to: string }> = ({
  from,
  to,
}) => {
  const [data, setData] = useState<Overview | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);

    fetch(
      `/admin/audit-logs/metrics/overview?from=${encodeURIComponent(
        from
      )}&to=${encodeURIComponent(to)}`,
      {
        credentials: "include",
      }
    )
      .then((r) => r.json())
      .then((json) => {
        if (!cancelled) setData(json);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [from, to]);

  if (loading && !data)
    return <InlineLoading description="Loading metrics..." />;

  return (
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "repeat(4, minmax(0, 1fr))",
        gap: 12,
      }}
    >
      <Tile>
        <strong>Total Events</strong>
        <div style={{ fontSize: 24 }}>{data?.totalEvents ?? "—"}</div>
      </Tile>
      <Tile>
        <strong>Total Failures</strong>
        <div style={{ fontSize: 24 }}>{data?.totalFailures ?? "—"}</div>
      </Tile>
      <Tile>
        <strong>Failed Logins</strong>
        <div style={{ fontSize: 24 }}>{data?.failedLogins ?? "—"}</div>
      </Tile>
      <Tile>
        <strong>Successful Logins</strong>
        <div style={{ fontSize: 24 }}>{data?.successfulLogins ?? "—"}</div>
      </Tile>
    </div>
  );
};
