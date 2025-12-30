import React, { useMemo, useState } from "react";
import {
  DataTable,
  InlineLoading,
  Pagination,
  Tag,
  Tile,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow,
} from "@carbon/react";

import { useAuditLogs } from "../../../hooks/useAuditLogs";
import { type AuditLog } from "../../../lib/api/audit";
import { AuditLogDrawer } from "../../../components/audit/AuditLogDrawer";
import { AuditMetricsPanel } from "../../../components/audit/AuditMetricsPanel";
import { EmptyState } from "../../../components/emptystate/EmptyState";
import { ErrorState } from "../../../components/errorstate/ErrorState";
import { AuditLogFilters } from "../../../components/audit/AuditLogFilters.tsx";

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
        id: log.id, // ✅ stable key
        shortId: log.id.slice(0, 8), // UI only
        time: new Date(log.createdAt).toLocaleString(),
        actor: log.username ?? "System",
        action: log.action,
        client: log.metadata?.client_id ?? "—",
        result:
          log.metadata?.success === true
            ? "success"
            : log.metadata?.success === false
            ? "failure"
            : "—",
        raw: log,
      })),
    [data?.items]
  );

  const headers = [
    { key: "shortId", header: "ID" },
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
          from={from}
          to={to}
          onFromChange={setFrom}
          onToChange={setTo}
          onClientChange={setClientId}
          onActionChange={setAction}
          onSuccessChange={setSuccess}
          onClear={clearFilters}
          onExportCsv={() => window.open(buildExportUrl("csv"))}
          onExportJson={() => window.open(buildExportUrl("json"))}
        />
      </Tile>

      {/* Table */}
      <Tile>
        {loading && (
          <div style={{ padding: 16 }}>
            <InlineLoading description="Loading audit logs…" />
          </div>
        )}

        {!loading && error && (
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
          <>
            <DataTable rows={rows} headers={headers}>
              {({ rows, headers, getHeaderProps, getRowProps }) => (
                <TableContainer>
                  <Table size="lg">
                    <TableHead>
                      <TableRow>
                        {headers.map((header) => (
                          <TableHeader
                            key={header.key}
                            {...getHeaderProps({ header })}
                          >
                            {header.header}
                          </TableHeader>
                        ))}
                        <TableHeader />
                      </TableRow>
                    </TableHead>

                    <TableBody>
                      {rows.map((row) => {
                        const raw = (row as any).raw as AuditLog;

                        return (
                          <TableRow key={row.id} {...getRowProps({ row })}>
                            {row.cells.map((cell) => (
                              <TableCell key={cell.id}>
                                {cell.info.header === "result" ? (
                                  <Tag
                                    type={
                                      cell.value === "success"
                                        ? "green"
                                        : cell.value === "failure"
                                        ? "red"
                                        : "gray"
                                    }
                                  >
                                    {cell.value}
                                  </Tag>
                                ) : (
                                  cell.value
                                )}
                              </TableCell>
                            ))}

                            <TableCell style={{ textAlign: "right" }}>
                              <Button
                                size="sm"
                                kind="ghost"
                                onClick={() => setSelected(raw)}
                              >
                                View
                              </Button>
                            </TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </TableContainer>
              )}
            </DataTable>

            <Pagination
              page={1}
              pageSize={50}
              pageSizes={[50, 100, 200]}
              totalItems={data?.items?.length ?? rows.length}
              disabled
            />
          </>
        )}
      </Tile>

      <AuditLogDrawer
        log={selected}
        open={Boolean(selected)}
        onClose={() => setSelected(null)}
      />
    </div>
  );
}
