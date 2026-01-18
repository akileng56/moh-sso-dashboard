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
  OverflowMenu,
  OverflowMenuItem,
  Tag,
  Pagination,
  Stack,
} from "@carbon/react";
import { Add } from "@carbon/icons-react";

import { EmptyState } from "../../../components/emptystate/EmptyState";
import { ErrorState } from "../../../components/errorstate/ErrorState";
import { UserFilters } from "../../../components/user/UserFilters";
import { useHeaderPanel } from "../../../components/header-panel/header-panel.context";

import {
  useListUsersQuery,
  useToggleUserMutation,
} from "../../../store/api/users.api";
import type { User } from "../../../store/types/user.types";
import { UserFormPanel } from "../../../components/panels/create-user-panel";
import { useEnableUserModal } from "../../../components/user/useEnableUserModal";
import { useResetPasswordModal } from "../../../components/user/useResetPasswordModal";
import { UserClientRolesPanel } from "../../../components/panels/user-client-roles-panel";

/* -----------------------------
 * Table headers (raw hidden)
 * ----------------------------- */
const headers = [
  { key: "username", header: "Username" },
  { key: "email", header: "Email" },
  { key: "status", header: "Status" },
  { key: "verified", header: "Email verified" },
  { key: "lastLogin", header: "Last login" },
  { key: "actions", header: "" },

  // hidden technical column
  { key: "raw", header: "" },
];

export default function UsersPage() {
  const { openPanel, closePanel } = useHeaderPanel();
  const { openEnableUserModal } = useEnableUserModal();
  const { openResetPasswordModal } = useResetPasswordModal();

  /* -----------------------------
   * Filters
   * ----------------------------- */
  const [statusFilter, setStatusFilter] = useState("all");
  const [roleFilter, setRoleFilter] = useState("all");
  const [neverLoggedIn, setNeverLoggedIn] = useState(false);

  /* -----------------------------
   * Pagination
   * ----------------------------- */
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  /* -----------------------------
   * Data
   * ----------------------------- */
  const {
    data: users = [],
    isLoading,
    isError,
    error,
    refetch,
  } = useListUsersQuery();

  const [toggleUser] = useToggleUserMutation();

  /* -----------------------------
   * Reset page on filter change
   * ----------------------------- */
  useEffect(() => {
    setPage(1);
  }, [statusFilter, roleFilter, neverLoggedIn]);

  /* -----------------------------
   * Derived roles
   * ----------------------------- */
  const roles = useMemo(() => {
    const set = new Set<string>();
    users.forEach((u) => u?.realmRoles?.forEach((r) => set.add(r)));
    return ["all", ...Array.from(set)];
  }, [users]);

  /* -----------------------------
   * Filtering
   * ----------------------------- */
  const filteredUsers = useMemo(() => {
    return users.filter((u) => {
      if (statusFilter === "active" && !u.isActive) return false;
      if (statusFilter === "disabled" && u.isActive) return false;
      if (neverLoggedIn && u.lastLoginAt) return false;
      if (roleFilter !== "all" && !u?.realmRoles?.includes(roleFilter))
        return false;
      return true;
    });
  }, [users, statusFilter, roleFilter, neverLoggedIn]);

  /* -----------------------------
   * Pagination slice
   * ----------------------------- */
  const paginatedUsers = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filteredUsers.slice(start, start + pageSize);
  }, [filteredUsers, page, pageSize]);

  /* -----------------------------
   * Loading / Error
   * ----------------------------- */
  if (isLoading) {
    return (
      <div style={{ padding: "2rem" }}>
        <InlineLoading description="Loading users…" />
      </div>
    );
  }

  if (isError) {
    return (
      <ErrorState
        title="Failed to load users"
        description={(error as any)?.data?.message ?? "Failed to load users"}
        primaryAction={{ label: "Retry", onClick: refetch }}
      />
    );
  }

  /* -----------------------------
   * Rows (raw preserved)
   * ----------------------------- */
  const rows = paginatedUsers.map((u) => ({
    id: u.id,
    username: u.username,
    email: u.email ?? "—",
    status: u.isActive ? "Active" : "Disabled",
    verified: u.emailVerified ? "Verified" : "Not verified",
    lastLogin: u.lastLoginAt
      ? new Date(u.lastLoginAt).toLocaleString()
      : "Never",
    actions: "",
    raw: u,
  }));

  return (
    <div style={{ padding: 16, display: "grid", gap: 16 }}>
      {/* Header */}
      <div>
        <h3 style={{ margin: 0 }}>Users</h3>
        <p style={{ marginTop: 6, opacity: 0.8 }}>
          Manage users, roles, and access to applications.
        </p>
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
              label: "Create user",
              icon: Add,
              onClick: () =>
                openPanel({
                  title: "Create user",
                  content: (
                    <UserFormPanel mode="create" onSuccess={closePanel} />
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
                      const user = row.cells.find(
                        (c) => c.info.header === "raw"
                      )?.value as User;

                      return (
                        <TableRow {...getRowProps({ row })}>
                          {row.cells.map((cell) => {
                            if (cell.info.header === "raw") return null;

                            if (cell.info.header === "status") {
                              return (
                                <TableCell key={cell.id}>
                                  <Tag type={!user?.isAdmin ? "green" : "red"}>
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
                                      itemText="Edit user"
                                      hasDivider
                                      onClick={() =>
                                        openPanel({
                                          title: "Edit user",
                                          content: (
                                            <UserFormPanel
                                              mode="edit"
                                              initialUser={user}
                                              onSuccess={closePanel}
                                            />
                                          ),
                                          size: "md",
                                        })
                                      }
                                    />

                                    <OverflowMenuItem
                                      itemText="Manage roles"
                                      hasDivider
                                      onClick={() =>
                                        openPanel({
                                          title: `Roles: ${user.username}`,
                                          size: "lg",
                                          content: (
                                            <Stack gap={6}>
                                              <UserFormPanel
                                                mode="edit"
                                                initialUser={user}
                                                onSuccess={closePanel}
                                              />
                                              <UserClientRolesPanel
                                                userId={user.id}
                                              />
                                            </Stack>
                                          ),
                                        })
                                      }
                                    />

                                    <OverflowMenuItem
                                      itemText={
                                        user?.isActive
                                          ? "Disable user"
                                          : "Enable user"
                                      }
                                      isDelete={user.enabled}
                                      onClick={() =>
                                        openEnableUserModal({
                                          username: user.username,
                                          enabled: user?.isActive,
                                          onConfirm: async () => {
                                            await toggleUser({
                                              id: user.id,
                                              enabled: !user?.isActive,
                                            }).unwrap();
                                          },
                                        })
                                      }
                                    />

                                    <OverflowMenuItem
                                      itemText="Reset password"
                                      hasDivider
                                      onClick={() =>
                                        openResetPasswordModal({
                                          username: user.username,
                                          email: user.email,
                                          onConfirm: async () => {
                                            console.log(
                                              "Reset password for",
                                              user.username
                                            );
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
