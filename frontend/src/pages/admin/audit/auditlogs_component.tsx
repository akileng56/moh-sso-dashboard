import React, { useMemo, useState } from "react";
import {
  Button,
  ComboBox,
  DataTable,
  DatePicker,
  DatePickerInput,
  InlineLoading,
  Pagination,
  Search,
  Tag,
  Tile,
} from "@carbon/react";
import { useAuditLogs } from "../../../hooks/useAuditLogs";
import { type AuditLog } from "../../../lib/api/audit";
import { AuditLogDrawer } from "../../../components/AuditLogDrawer";
import { AuditMetricsPanel } from "../../../components/AuditMetricsPanel";

function toRFC3339(d: Date) {
  return d.toISOString();
}

type SuccessFilter = "true" | "false";

export const AuditLogs: React.FC = () => {
  const now = new Date();
  const start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);

  const [from, setFrom] = useState(toRFC3339(start));
  const [to, setTo] = useState(toRFC3339(now));

  const [action, setAction] = useState<string>();
  const [userId, setUserId] = useState<string>();
  const [clientId, setClientId] = useState<string>();
  const [ip, setIp] = useState<string>();
  const [success, setSuccess] = useState<SuccessFilter>();

  const [selected, setSelected] = useState<AuditLog | null>(null);

  const filters = useMemo(
    () => ({
      from,
      to,
      action,
      user_id: userId,
      client_id: clientId,
      ip,
      success,
    }),
    [from, to, action, userId, clientId, ip, success]
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
        ip: log.metadata?.ip ?? "",
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
    { key: "ip", header: "IP" },
    { key: "result", header: "Result" },
  ];

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
      <div>
        <h3 style={{ margin: 0 }}>Audit Logs</h3>
        <p style={{ marginTop: 6, opacity: 0.8 }}>
          Security and administrative activity across the platform.
        </p>
      </div>

      <AuditMetricsPanel from={from} to={to} />

      <Tile>
        <div style={{ display: "grid", gap: 12 }}>
          <div style={{ display: "flex", gap: 12, flexWrap: "wrap" }}>
            <DatePicker
              datePickerType="range"
              onChange={(dates) => {
                if (Array.isArray(dates)) {
                  if (dates[0] instanceof Date) setFrom(toRFC3339(dates[0]));
                  if (dates[1] instanceof Date) setTo(toRFC3339(dates[1]));
                }
              }}
            >
              <DatePickerInput id="from" labelText="From" />
              <DatePickerInput id="to" labelText="To" />
            </DatePicker>

            <Search
              labelText="Client ID"
              onChange={(e) => setClientId(e.target.value || undefined)}
            />
            <Search
              labelText="User ID"
              onChange={(e) => setUserId(e.target.value || undefined)}
            />
            <Search
              labelText="Action"
              onChange={(e) => setAction(e.target.value || undefined)}
            />

            <ComboBox
              id="success"
              titleText="Result"
              items={[
                { id: "true", label: "Success" },
                { id: "false", label: "Failure" },
              ]}
              itemToString={(item) => item?.label ?? ""}
              onChange={({ selectedItem }) =>
                setSuccess(selectedItem?.id as SuccessFilter | undefined)
              }
            />

            <Button
              kind="secondary"
              onClick={() => {
                setAction(undefined);
                setUserId(undefined);
                setClientId(undefined);
                setIp(undefined);
                setSuccess(undefined);
              }}
            >
              Clear filters
            </Button>

            <Button
              kind="primary"
              onClick={() => window.open(buildExportUrl("csv"))}
            >
              Export CSV
            </Button>

            <Button
              kind="tertiary"
              onClick={() => window.open(buildExportUrl("json"))}
            >
              Export JSON
            </Button>
          </div>

          {loading && <InlineLoading description="Loading audit logs..." />}
          {error && <p style={{ color: "crimson" }}>{error}</p>}

          <DataTable rows={rows} headers={headers}>
            {({
              rows,
              headers,
              getHeaderProps,
              getRowProps,
              getTableProps,
            }) => (
              <table {...getTableProps()} style={{ width: "100%" }}>
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
                        style={{ cursor: "pointer" }}
                        onClick={() => setSelected(raw)}
                      >
                        {row.cells.map((cell) => (
                          <td key={cell.id}>
                            {cell.info.header === "result" && cell.value ? (
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

          <Pagination
            page={1}
            pageSize={50}
            pageSizes={[50, 100, 200]}
            totalItems={data?.items?.length ?? 0}
            onChange={() => {}}
          />
        </div>
      </Tile>

      <AuditLogDrawer
        log={selected}
        open={Boolean(selected)}
        onClose={() => setSelected(null)}
      />
    </div>
  );
};

export default AuditLogs;
