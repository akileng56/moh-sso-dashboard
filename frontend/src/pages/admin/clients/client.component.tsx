import { useEffect, useMemo, useState } from "react";
import {
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  InlineLoading,
  Tile,
  Button,
  Tag,
  Pagination,
} from "@carbon/react";
import { View, UserFollow, Add } from "@carbon/icons-react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../../../context/useAuth";
import { ClientFilters } from "../../../components/client/ClientFilters";
import { EmptyState } from "../../../components/emptystate/EmptyState";
import { ErrorState } from "../../../components/errorstate/ErrorState";

interface Client {
  clientId: string;
  name: string;
  publicClient: boolean;
  enabled: boolean;
}

/* -----------------------------
 * Filters
 * ----------------------------- */
type StatusFilter = "all" | "enabled" | "disabled";
type TypeFilter = "all" | "public" | "confidential";

const STATUS_OPTIONS = [
  { id: "all", label: "All" },
  { id: "enabled", label: "Enabled" },
  { id: "disabled", label: "Disabled" },
] as const;

const TYPE_OPTIONS = [
  { id: "all", label: "All" },
  { id: "public", label: "Public" },
  { id: "confidential", label: "Confidential" },
] as const;

/* -----------------------------
 * Table headers
 * ----------------------------- */
const headers = [
  { key: "name", header: "Name" },
  { key: "clientId", header: "Client ID" },
  { key: "type", header: "Type" },
  { key: "status", header: "Status" },
  { key: "actions", header: "Actions" },
];

export default function ClientsPage() {
  const { accessToken } = useAuth();
  const navigate = useNavigate();

  const [clients, setClients] = useState<Client[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [typeFilter, setTypeFilter] = useState<TypeFilter>("all");

  // Pagination
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  /* -----------------------------
   * Load clients
   * ----------------------------- */
  useEffect(() => {
    if (!accessToken) return;

    const controller = new AbortController();

    async function loadClients() {
      try {
        const res = await fetch("http://localhost:9000/api/v1/clients", {
          headers: { Authorization: `Bearer ${accessToken}` },
          credentials: "include",
          signal: controller.signal,
        });

        if (!res.ok) {
          throw new Error(`Failed to load clients (${res.status})`);
        }

        setClients(await res.json());
      } catch (err: any) {
        if (err.name !== "AbortError") {
          setError(err.message ?? "Failed to load clients");
        }
      } finally {
        setLoading(false);
      }
    }

    loadClients();
    return () => controller.abort();
  }, [accessToken]);

  /* -----------------------------
   * Reset page on filter change
   * ----------------------------- */
  useEffect(() => {
    setPage(1);
  }, [statusFilter, typeFilter]);

  /* -----------------------------
   * Filtering
   * ----------------------------- */
  const filteredClients = useMemo(() => {
    return clients.filter((c) => {
      if (statusFilter === "enabled" && !c.enabled) return false;
      if (statusFilter === "disabled" && c.enabled) return false;
      if (typeFilter === "public" && !c.publicClient) return false;
      if (typeFilter === "confidential" && c.publicClient) return false;
      return true;
    });
  }, [clients, statusFilter, typeFilter]);

  /* -----------------------------
   * Pagination slice
   * ----------------------------- */
  const paginatedClients = useMemo(() => {
    const start = (page - 1) * pageSize;
    const end = start + pageSize;
    return filteredClients.slice(start, end);
  }, [filteredClients, page, pageSize]);

  /* -----------------------------
   * Toggle (placeholder)
   * ----------------------------- */
  function toggleClient(client: Client) {
    setClients((prev) =>
      prev.map((c) =>
        c.clientId === client.clientId ? { ...c, enabled: !c.enabled } : c
      )
    );
  }

  /* -----------------------------
   * Loading
   * ----------------------------- */
  if (loading) {
    return (
      <div style={{ padding: "2rem" }}>
        <InlineLoading description="Loading clients…" />
      </div>
    );
  }

  /* -----------------------------
   * Error
   * ----------------------------- */
  if (error) {
    return (
      <ErrorState
        title="Failed to load clients"
        description={error}
        primaryAction={{
          label: "Retry",
          onClick: () => window.location.reload(),
        }}
      />
    );
  }

  const rows = paginatedClients.map((c) => ({
    id: `${c.clientId}`,
    name: c.name,
    clientId: c.clientId,
    type: c.publicClient ? "Public" : "Confidential",
    status: c.enabled ? "Enabled" : "Disabled",
    actions: "",
    raw: c,
  }));

  return (
    <div style={{ padding: 16, display: "grid", gap: 16 }}>
      {/* Header */}
      <div>
        <h3 style={{ margin: 0 }}>Clients</h3>
        <p style={{ marginTop: 6, opacity: 0.8 }}>
          Registered applications and services integrated with the platform.
        </p>
      </div>

      {/* Filters */}
      <Tile>
        <ClientFilters
          status={statusFilter}
          type={typeFilter}
          statusOptions={STATUS_OPTIONS}
          typeOptions={TYPE_OPTIONS}
          onStatusChange={setStatusFilter}
          onTypeChange={setTypeFilter}
        />
      </Tile>

      {/* Table */}
      <Tile>
        {filteredClients.length === 0 ? (
          <EmptyState
            title="No clients registered"
            description="Start by registering an application to make it available in the platform."
            primaryAction={{
              label: "Register application",
              icon: Add,
              onClick: () => navigate("/admin/clients/new"),
            }}
          />
        ) : (
          <>
            <DataTable rows={rows} headers={headers}>
              {({
                rows,
                headers,
                getHeaderProps,
                getRowProps,
                getTableProps,
              }) => (
                <Table {...getTableProps()}>
                  <TableHead>
                    <TableRow>
                      {headers.map((h) => (
                        <TableHeader
                          key={h.key}
                          {...getHeaderProps({ header: h })}
                        >
                          {h.header}
                        </TableHeader>
                      ))}
                    </TableRow>
                  </TableHead>

                  <TableBody>
                    {rows.map((row) => {
                      const client = (row as any).raw as Client;

                      return (
                        <TableRow {...getRowProps({ row })}>
                          {row.cells.map((cell) => {
                            if (cell.info.header === "status") {
                              return (
                                <TableCell key={cell.id}>
                                  <Tag type={client?.enabled ? "green" : "red"}>
                                    {cell.value}
                                  </Tag>
                                </TableCell>
                              );
                            }

                            if (cell.info.header === "actions") {
                              return (
                                <TableCell key={cell.id}>
                                  <div style={{ display: "flex", gap: 4 }}>
                                    <Button
                                      size="sm"
                                      kind="ghost"
                                      hasIconOnly
                                      renderIcon={View}
                                      iconDescription="View audit logs"
                                      onClick={() =>
                                        navigate(
                                          `/admin/audit-logs?client_id=${client.clientId}`
                                        )
                                      }
                                    />

                                    <Button
                                      size="sm"
                                      kind="ghost"
                                      hasIconOnly
                                      renderIcon={UserFollow}
                                      iconDescription={
                                        client?.enabled
                                          ? "Disable client"
                                          : "Enable client"
                                      }
                                      onClick={() => toggleClient(client)}
                                    />
                                  </div>
                                </TableCell>
                              );
                            }

                            return (
                              <TableCell key={cell.id}>{cell.value}</TableCell>
                            );
                          })}
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              )}
            </DataTable>

            <Pagination
              page={page}
              pageSize={pageSize}
              pageSizes={[10, 20, 30, 50]}
              totalItems={filteredClients.length}
              onChange={({ page, pageSize }) => {
                setPage(page);
                setPageSize(pageSize);
              }}
            />
          </>
        )}
      </Tile>
    </div>
  );
}
