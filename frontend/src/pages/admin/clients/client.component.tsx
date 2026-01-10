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
  Tag,
  Pagination,
  OverflowMenu,
  OverflowMenuItem,
} from "@carbon/react";
import { Add } from "@carbon/icons-react";

import { ClientFilters } from "../../../components/client/ClientFilters";
import { EmptyState } from "../../../components/emptystate/EmptyState";
import { ErrorState } from "../../../components/errorstate/ErrorState";

import {
  useListClientsQuery,
  useToggleClientMutation,
} from "../../../store/api/clients.api";
import type { Client } from "../../../store/types/client.types";
import { useHeaderPanel } from "../../../components/header-panel/header-panel.context";
import { ClientFormPanel } from "../../../components/panels/client-form-panel";
import { useEnableClientModal } from "../../../components/client/useEnableClientModal";

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
 * NOTE: raw is hidden but preserved
 * ----------------------------- */
const headers = [
  { key: "name", header: "Name" },
  { key: "clientId", header: "Client ID" },
  { key: "type", header: "Type" },
  { key: "status", header: "Status" },
  { key: "actions", header: "Actions" },

  { key: "raw", header: "" },
];

export default function ClientsPage() {
  const { openPanel, closePanel } = useHeaderPanel();
  const { openEnableClientModal } = useEnableClientModal();

  /* -----------------------------
   * Filters
   * ----------------------------- */
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [typeFilter, setTypeFilter] = useState<TypeFilter>("all");

  /* -----------------------------
   * Pagination
   * ----------------------------- */
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  /* -----------------------------
   * Data
   * ----------------------------- */
  const {
    data: clients = [],
    isLoading,
    isError,
    error,
    refetch,
  } = useListClientsQuery();

  const [toggleClient] = useToggleClientMutation();

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
    return filteredClients.slice(start, start + pageSize);
  }, [filteredClients, page, pageSize]);

  /* -----------------------------
   * Loading
   * ----------------------------- */
  if (isLoading) {
    return (
      <div style={{ padding: "2rem" }}>
        <InlineLoading description="Loading clients…" />
      </div>
    );
  }

  /* -----------------------------
   * Error
   * ----------------------------- */
  if (isError) {
    return (
      <ErrorState
        title="Failed to load clients"
        description={(error as any)?.data?.message ?? "Failed to load clients"}
        primaryAction={{ label: "Retry", onClick: refetch }}
      />
    );
  }

  /* -----------------------------
   * Rows (raw preserved)
   * ----------------------------- */
  const rows = paginatedClients.map((c) => ({
    id: c.clientId,
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
              onClick: () =>
                openPanel({
                  title: "Create client",
                  content: (
                    <ClientFormPanel mode="create" onSuccess={closePanel} />
                  ),
                  size: "md",
                }),
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
                      {headers.map(
                        (h) =>
                          h.key !== "raw" && (
                            <TableHeader {...getHeaderProps({ header: h })}>
                              {h.header}
                            </TableHeader>
                          )
                      )}
                    </TableRow>
                  </TableHead>

                  <TableBody>
                    {rows.map((row) => {
                      const client = row.cells.find(
                        (c) => c.info.header === "raw"
                      )?.value as Client;

                      return (
                        <TableRow {...getRowProps({ row })}>
                          {row.cells.map((cell) => {
                            if (cell.info.header === "raw") return null;

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
                                  <OverflowMenu size="sm" flipped>
                                    <OverflowMenuItem
                                      itemText="Edit client"
                                      hasDivider
                                      onClick={() =>
                                        openPanel({
                                          title: "Edit client",
                                          content: (
                                            <ClientFormPanel
                                              mode="edit"
                                              initialClient={client}
                                              onSuccess={closePanel}
                                            />
                                          ),
                                          size: "md",
                                        })
                                      }
                                    />

                                    <OverflowMenuItem
                                      itemText={
                                        client?.enabled
                                          ? "Disable client"
                                          : "Enable client"
                                      }
                                      isDelete={client?.enabled}
                                      onClick={() =>
                                        openEnableClientModal({
                                          clientName: client.name,
                                          enabled: client?.enabled,
                                          onConfirm: async () => {
                                            await toggleClient(
                                              client.clientId,
                                              client?.enabled
                                            ).unwrap();
                                          },
                                        })
                                      }
                                    />
                                  </OverflowMenu>
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
