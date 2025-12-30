import React, { useMemo, useState } from "react";
import {
  DataTable,
  InlineLoading,
  Pagination,
  Tag,
  Tile,
  Button,
} from "@carbon/react";
import { useAuditLogs } from "../../../hooks/useAuditLogs.ts";
import { type AuditLog } from "../../../lib/api/audit.ts";
import { AuditLogDrawer } from "../../../components/audit/AuditLogDrawer.tsx";
import { AuditLogFilters } from "../../../components/audit/AuditLogFilters.tsx.tsx";
import { AuditMetricsPanel } from "../../../components/audit/AuditMetricsPanel.tsx";
import { EmptyState } from "../../../components/emptystate/EmptyState.tsx";
import { ErrorState } from "../../../components/errorstate/ErrorState.tsx";

function toRFC3339(d: Date) {
  return d.toISOString();
}

type SuccessFilter = "true" | "false";

export default function AuditLogs() {
  const now = new Date();
  const start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);

  const [from, setFrom] = useState(toRFC3339(start));
  const [to, setTo] = useState(toRFC3339(now));
  const [action, setAction] = useState<string>();
  const [clientId, setClientId] = useState<string>();
  const [success, setSuccess] = useState<SuccessFilter>();
  const [selected, setSelected] = useState<AuditLog | null>(null);

  const filters = useMemo(
    () => ({ from, to, action, client_id: clientId, success }),
    [from, to, action, clientId, success]
  );

  const { data, loading, error } = useAuditLogs(filters);

  const rows = useMemo(
    () =>
      (data?.items ?? []).map((log) => ({
        id: log.id,
        time: new Date(log.createdAt).toLocaleString(),
        actor: log.username ?? "System",
        action: log.action,
        client: log.metadata?.client_id ?? "",
        result:
          typeof log.metadata?.success === "boolean"
            ? log.metadata.success
              ? "success"
              : "failure"
            : "",
        raw: log,
      })),
    [data?.items]
  );

  const headers = [
    { key: "time", header: "Time" },
    { key: "actor", header: "Actor" },
    { key: "action", header: "Action" },
    { key: "client", header: "Client" },
    { key: "result", header: "Result" },
  ];

  function clearFilters() {
    setAction(undefined);
    setClientId(undefined);
    setSuccess(undefined);
  }

  function buildExportUrl(format: "csv" | "json") {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([k, v]) => {
      if (v !== undefined) params.set(k, String(v));
    });
    params.set("format", format);
    return `/admin/audit-logs/export?${params.toString()}`;
  }

  return (
    <div style={{ padding: 16, display: "grid", gap: 16 }}>
      {/* Header */}
      <div>
        <h3 style={{ margin: 0 }}>Audit Logs</h3>
        <p style={{ marginTop: 6, opacity: 0.8 }}>
          Security and administrative activity across the platform.
        </p>
      </div>

      {/* Metrics */}
      <AuditMetricsPanel from={from} to={to} />

      {/* Filters */}
      <Tile>
        <AuditLogFilters
          onFromChange={setFrom}
          onToChange={setTo}
          onClientChange={setClientId}
          onActionChange={setAction}
          onSuccessChange={setSuccess}
          onClear={clearFilters}
          onExportCsv={() => window.open(buildExportUrl("csv"))}
          onExportJson={() => window.open(buildExportUrl("json"))}
          from={""}
          to={""}
        />
      </Tile>

      {/* States + Table */}
      <Tile>
        {loading && <InlineLoading description="Loading audit logs…" />}

        {error && (
          <ErrorState
            title="Failed to load audit logs"
            description={error}
            primaryAction={{
              label: "Retry",
              onClick: () => window.location.reload(),
            }}
          />
        )}

        {!loading && !error && rows.length === 0 && (
          <EmptyState
            title="No audit logs found"
            description="There are no audit events matching the selected filters."
            secondaryAction={{
              label: "Clear filters",
              onClick: clearFilters,
            }}
          />
        )}

        {!loading && !error && rows.length > 0 && (
          <DataTable rows={rows} headers={headers}>
            {({ rows, headers, getHeaderProps, getRowProps }) => (
              <table style={{ width: "100%" }}>
                <thead>
                  <tr>
                    {headers.map((h) => (
                      <th key={h.key} {...getHeaderProps({ header: h })}>
                        {h.header}
                      </th>
                    ))}
                    <th>Details</th>
                  </tr>
                </thead>
                <tbody>
                  {rows.map((row) => {
                    const raw = (row as any).raw as AuditLog;
                    return (
                      <tr
                        {...getRowProps({ row })}
                        onClick={() => setSelected(raw)}
                        style={{ cursor: "pointer" }}
                      >
                        {row.cells.map((cell) => (
                          <td key={cell.id}>
                            {cell.info.header === "result" ? (
                              <Tag
                                type={
                                  cell.value === "success" ? "green" : "red"
                                }
                              >
                                {cell.value}
                              </Tag>
                            ) : (
                              cell.value
                            )}
                          </td>
                        ))}
                        <td>
                          <Button
                            size="sm"
                            kind="ghost"
                            onClick={(e) => {
                              e.stopPropagation();
                              setSelected(raw);
                            }}
                          >
                            View
                          </Button>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            )}
          </DataTable>
        )}

        <Pagination
          page={1}
          pageSize={50}
          pageSizes={[50, 100, 200]}
          totalItems={data?.items?.length ?? 0}
          onChange={() => {}}
        />
      </Tile>

      <AuditLogDrawer
        log={selected}
        open={Boolean(selected)}
        onClose={() => setSelected(null)}
      />
    </div>
  );
}
