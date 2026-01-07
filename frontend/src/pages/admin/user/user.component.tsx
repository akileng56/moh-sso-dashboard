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
import { View, Reset, UserFollow, Add } from "@carbon/icons-react";
import { useNavigate } from "react-router-dom";

import { EmptyState } from "../../../components/emptystate/EmptyState";
import { ErrorState } from "../../../components/errorstate/ErrorState";
import { UserFilters } from "../../../components/user/UserFilters";

import {
  useListUsersQuery,
  useToggleUserMutation,
} from "../../../store/api/users.api";
import type { User } from "../../../store/types/user.types";

const headers = [
  { key: "username", header: "Username" },
  { key: "email", header: "Email" },
  { key: "status", header: "Status" },
  { key: "roles", header: "Roles" },
  { key: "lastLogin", header: "Last Login" },
  { key: "actions", header: "" },
];

export default function UsersPage() {
  const navigate = useNavigate();

  /* ---------------------------
   * Filters
   * --------------------------- */
  const [statusFilter, setStatusFilter] = useState("all");
  const [roleFilter, setRoleFilter] = useState("all");
  const [neverLoggedIn, setNeverLoggedIn] = useState(false);

  /* ---------------------------
   * Pagination
   * --------------------------- */
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  /* ---------------------------
   * Data (SSOT)
   * --------------------------- */
  const {
    data: users = [],
    isLoading,
    isError,
    error,
    refetch,
  } = useListUsersQuery();

  const [toggleUser] = useToggleUserMutation();

  /* ---------------------------
   * Reset page on filter change
   * --------------------------- */
  useEffect(() => {
    setPage(1);
  }, [statusFilter, roleFilter, neverLoggedIn]);

  /* ---------------------------
   * Derived filters
   * --------------------------- */
  const roles = useMemo(() => {
    const set = new Set<string>();
    users.forEach((u) => u?.client_roles?.admin?.forEach((r) => set.add(r)));
    return ["all", ...Array.from(set)];
  }, [users]);

  const filteredUsers = useMemo(() => {
    return users.filter((u) => {
      if (statusFilter === "active" && !u?.enabled) return false;
      if (statusFilter === "disabled" && u?.enabled) return false;
      if (neverLoggedIn && u.last_login_at) return false;
      if (roleFilter !== "all" && !u?.client_roles?.admin?.includes(roleFilter))
        return false;
      return true;
    });
  }, [users, statusFilter, roleFilter, neverLoggedIn]);

  /* ---------------------------
   * Pagination slice
   * --------------------------- */
  const paginatedUsers = useMemo(() => {
    const start = (page - 1) * pageSize;
    const end = start + pageSize;
    return filteredUsers.slice(start, end);
  }, [filteredUsers, page, pageSize]);

  /* ---------------------------
   * Loading
   * --------------------------- */
  if (isLoading) {
    return (
      <div style={{ padding: "2rem" }}>
        <InlineLoading description="Loading users…" />
      </div>
    );
  }

  /* ---------------------------
   * Error
   * --------------------------- */
  if (isError) {
    return (
      <ErrorState
        title="Failed to load users"
        description={(error as any)?.data?.message ?? "Failed to load users"}
        primaryAction={{
          label: "Retry",
          onClick: refetch,
        }}
      />
    );
  }

  const rows = paginatedUsers.map((u) => ({
    id: u.id,
    username: u.username,
    email: u.email ?? "—",
    status: u?.enabled ? "Active" : "Disabled",
    roles: u?.client_roles?.admin ?? [],
    lastLogin: u.last_login_at
      ? new Date(u.last_login_at).toLocaleString()
      : "Never",
    actions: "",
    raw: u,
  }));

  return (
    <div style={{ padding: 16, display: "grid", gap: 16 }}>
      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between" }}>
        <div>
          <h3 style={{ margin: 0 }}>Users</h3>
          <p style={{ marginTop: 6, opacity: 0.8 }}>
            Manage users, roles, and access to applications.
          </p>
        </div>
      </div>

      {/* Filters */}
      <Tile>
        <UserFilters
          status={statusFilter}
          roles={roles}
          selectedRole={roleFilter}
          neverLoggedIn={neverLoggedIn}
          onStatusChange={setStatusFilter}
          onRoleChange={setRoleFilter}
          onToggleNeverLoggedIn={() => setNeverLoggedIn((v) => !v)}
        />
      </Tile>

      {/* Table */}
      <Tile>
        {filteredUsers.length === 0 ? (
          <EmptyState
            title="No users found"
            description="No users match the selected filters."
            primaryAction={{
              label: "Import users",
              icon: Add,
              onClick: () => navigate("/admin/users/import"),
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
                        <TableHeader {...getHeaderProps({ header: h })}>
                          {h.header}
                        </TableHeader>
                      ))}
                    </TableRow>
                  </TableHead>

                  <TableBody>
                    {rows.map((row) => {
                      const user = (row as any).raw as User;

                      return (
                        <TableRow {...getRowProps({ row })}>
                          {row.cells.map((cell) => {
                            if (cell.info.header === "status") {
                              return (
                                <TableCell key={cell.id}>
                                  <Tag type={user?.enabled ? "green" : "red"}>
                                    {cell.value}
                                  </Tag>
                                </TableCell>
                              );
                            }

                            if (cell.info.header === "roles") {
                              return (
                                <TableCell key={cell.id}>
                                  <div style={{ display: "flex", gap: 4 }}>
                                    {(user?.client_roles?.admin ?? []).map(
                                      (r) => (
                                        <Tag key={r} size="sm">
                                          {r}
                                        </Tag>
                                      )
                                    )}
                                  </div>
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
                                          `/admin/audit-logs?user_id=${user.id}`
                                        )
                                      }
                                    />

                                    <Button
                                      size="sm"
                                      kind="ghost"
                                      hasIconOnly
                                      renderIcon={UserFollow}
                                      iconDescription={
                                        user?.enabled
                                          ? "Disable user"
                                          : "Enable user"
                                      }
                                      onClick={() =>
                                        toggleUser({
                                          id: user.id,
                                          enabled: !user?.enabled,
                                        })
                                      }
                                    />

                                    <Button
                                      size="sm"
                                      kind="ghost"
                                      hasIconOnly
                                      renderIcon={Reset}
                                      iconDescription="Reset password"
                                      onClick={() =>
                                        alert(
                                          `Password reset initiated for ${user.username}`
                                        )
                                      }
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
              totalItems={filteredUsers.length}
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
