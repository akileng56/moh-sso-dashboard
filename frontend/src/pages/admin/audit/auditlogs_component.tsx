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

export const AuditLogs: React.FC = () => {
  const now = new Date();
  const start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);

  const [from, setFrom] = useState(toRFC3339(start));
  const [to, setTo] = useState(toRFC3339(now));

  const [action, setAction] = useState<string | undefined>(undefined);
  const [userId, setUserId] = useState<string | undefined>(undefined);
  const [clientId, setClientId] = useState<string | undefined>(undefined);
  const [ip, setIp] = useState<string | undefined>(undefined);
  const [success, setSuccess] = useState<"true" | "false" | undefined>(
    undefined
  );

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

  const rows = (data?.items ?? []).map((l) => ({
    id: l.id,
    time: new Date(l.createdAt).toLocaleString(),
    actor: l.username ?? "System",
    action: l.action,
    ip: l.metadata?.ip ?? "",
    client: l.metadata?.client_id ?? "",
    result:
      typeof l.metadata?.success === "boolean"
        ? l.metadata.success
          ? "success"
          : "failure"
        : "",
    raw: l,
  }));

  const headers = [
    { key: "time", header: "Time" },
    { key: "actor", header: "Actor" },
    { key: "action", header: "Action" },
    { key: "client", header: "Client" },
    { key: "ip", header: "IP" },
    { key: "result", header: "Result" },
  ];

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
              onChange={(dates: Date[]) => {
                if (dates?.[0]) setFrom(toRFC3339(dates[0]));
                if (dates?.[1]) setTo(toRFC3339(dates[1]));
              }}
            >
              <DatePickerInput
                id="from"
                labelText="From"
                placeholder="mm/dd/yyyy"
              />
              <DatePickerInput
                id="to"
                labelText="To"
                placeholder="mm/dd/yyyy"
              />
            </DatePicker>

            <Search
              labelText="IP"
              placeholder="Filter by IP"
              onChange={(e) => setIp(e.currentTarget.value || undefined)}
            />

            <Search
              labelText="Client ID"
              placeholder="Filter by client_id"
              onChange={(e) => setClientId(e.currentTarget.value || undefined)}
            />

            <ComboBox
              id="success"
              titleText="Result"
              items={[
                { id: "true", label: "Success" },
                { id: "false", label: "Failure" },
              ]}
              itemToString={(it) => (it ? it.label : "")}
              onChange={({ selectedItem }) =>
                setSuccess((selectedItem?.id as any) || undefined)
              }
            />

            <Search
              labelText="User ID"
              placeholder="Filter by user UUID"
              onChange={(e) => setUserId(e.currentTarget.value || undefined)}
            />

            <Search
              labelText="Action"
              placeholder="Filter by action"
              onChange={(e) => setAction(e.currentTarget.value || undefined)}
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
              onClick={() => {
                const p = new URLSearchParams(filters as any);
                p.set("format", "csv");
                window.open(
                  `/admin/audit-logs/export?${p.toString()}`,
                  "_blank"
                );
              }}
            >
              Export CSV
            </Button>

            <Button
              kind="tertiary"
              onClick={() => {
                const p = new URLSearchParams(filters as any);
                p.set("format", "json");
                window.open(
                  `/admin/audit-logs/export?${p.toString()}`,
                  "_blank"
                );
              }}
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
                {/* <thead>
                  <tr>
                    {headers.map((header) => (
                      <th key={header.key} {...getHeaderProps({ header })}>
                        {header.header}
                      </th>
                    ))}
                    <th>Details</th>
                  </tr>
                </thead> */}
                <tbody>
                  {rows.map((r) => {
                    const raw = (r as any).raw as AuditLog;
                    const result = (r as any).cells.find(
                      (c: any) => c.info.header === "result"
                    )?.value;
                    return (
                      <tr
                        {...getRowProps({ row: r })}
                        style={{ cursor: "pointer" }}
                        onClick={() => setSelected(raw)}
                      >
                        {r.cells.map((cell) => (
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
            onChange={() => {
              // MVP: cursor paging “Load more” is better than Carbon's page numbers.
              // Implement a Load More button using next_cursor for stable pagination.
            }}
          />
        </div>
      </Tile>

      <AuditLogDrawer
        log={selected}
        open={!!selected}
        onClose={() => setSelected(null)}
      />
    </div>
  );
};

export default AuditLogs;
