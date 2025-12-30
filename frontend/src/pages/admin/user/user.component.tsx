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
} from "@carbon/react";
import { View, Reset, UserFollow, Add } from "@carbon/icons-react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../../../context/useAuth";
import { EmptyState } from "../../../components/emptystate/EmptyState";
import { ErrorState } from "../../../components/errorstate/ErrorState";
import { UserFilters } from "../../../components/user/UserFilters";

interface User {
  id: string;
  username: string;
  email?: string;
  enabled: boolean;
  roles: string[];
  lastLoginAt?: string;
}

const headers = [
  { key: "username", header: "Username" },
  { key: "email", header: "Email" },
  { key: "status", header: "Status" },
  { key: "roles", header: "Roles" },
  { key: "lastLogin", header: "Last Login" },
  { key: "actions", header: "" },
];

export default function UsersPage() {
  const { accessToken } = useAuth();
  const navigate = useNavigate();

  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [statusFilter, setStatusFilter] = useState("all");
  const [roleFilter, setRoleFilter] = useState("all");
  const [neverLoggedIn, setNeverLoggedIn] = useState(false);

  /* ---------------------------
   * Load users
   * --------------------------- */
  useEffect(() => {
    if (!accessToken) return;

    const controller = new AbortController();

    async function loadUsers() {
      try {
        const res = await fetch("http://localhost:9000/api/v1/users", {
          headers: { Authorization: `Bearer ${accessToken}` },
          credentials: "include",
          signal: controller.signal,
        });

        if (!res.ok) {
          throw new Error(`Failed to load users (${res.status})`);
        }

        setUsers(await res.json());
      } catch (err: any) {
        if (err.name !== "AbortError") {
          setError(err.message ?? "Failed to load users");
        }
      } finally {
        setLoading(false);
      }
    }

    loadUsers();
    return () => controller.abort();
  }, [accessToken]);

  /* ---------------------------
   * Derived filters
   * --------------------------- */
  const roles = useMemo(() => {
    const set = new Set<string>();
    users.forEach((u) => u.roles.forEach((r) => set.add(r)));
    return ["all", ...Array.from(set)];
  }, [users]);

  const filteredUsers = useMemo(() => {
    return users.filter((u) => {
      if (statusFilter === "active" && !u.enabled) return false;
      if (statusFilter === "disabled" && u.enabled) return false;
      if (neverLoggedIn && u.lastLoginAt) return false;
      if (roleFilter !== "all" && !u.roles.includes(roleFilter)) return false;
      return true;
    });
  }, [users, statusFilter, roleFilter, neverLoggedIn]);

  /* ---------------------------
   * Actions
   * --------------------------- */
  async function toggleUser(user: User) {
    await fetch(`http://localhost:9000/api/v1/users/${user.id}`, {
      method: user.enabled ? "DELETE" : "POST",
      headers: { Authorization: `Bearer ${accessToken}` },
      credentials: "include",
    });

    setUsers((prev) =>
      prev.map((u) => (u.id === user.id ? { ...u, enabled: !u.enabled } : u))
    );
  }

  function resetPassword(user: User) {
    alert(`Password reset initiated for ${user.username}`);
  }

  /* ---------------------------
   * Loading
   * --------------------------- */
  if (loading) {
    return (
      <div style={{ padding: "2rem" }}>
        <InlineLoading description="Loading users…" />
      </div>
    );
  }

  /* ---------------------------
   * Error
   * --------------------------- */
  if (error) {
    return (
      <ErrorState
        title="Failed to load users"
        description={error}
        primaryAction={{
          label: "Retry",
          onClick: () => window.location.reload(),
        }}
      />
    );
  }

  const rows = filteredUsers.map((u) => ({
    id: u.id,
    username: u.username,
    email: u.email ?? "—",
    status: u.enabled ? "Active" : "Disabled",
    roles: u.roles,
    lastLogin: u.lastLoginAt
      ? new Date(u.lastLoginAt).toLocaleString()
      : "Never",
    actions: "",
    raw: u,
  }));

  return (
    <div style={{ padding: 16, display: "grid", gap: 16 }}>
      {/* --------------------------------
       * Page Header
       * -------------------------------- */}
      <div style={{ display: "flex", justifyContent: "space-between" }}>
        <div>
          <h3 style={{ margin: 0 }}>Users</h3>
          <p style={{ marginTop: 6, opacity: 0.8 }}>
            Manage users, roles, and access to applications.
          </p>
        </div>

        <Button size="sm" onClick={() => navigate("/admin/users/import")}>
          Import users
        </Button>
      </div>

      {/* --------------------------------
       * Filters
       * -------------------------------- */}
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

      {/* --------------------------------
       * Empty / Table
       * -------------------------------- */}
      <Tile>
        {filteredUsers.length === 0 ? (
          <EmptyState
            title="No users found"
            description="No users match the selected filters, or no users have been added yet."
            primaryAction={{
              label: "Import users",
              icon: Add,
              onClick: () => navigate("/admin/users/import"),
            }}
          />
        ) : (
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
                    const user = (row as any).raw as User;

                    return (
                      <TableRow {...getRowProps({ row })}>
                        {row.cells.map((cell) => {
                          if (cell.info.header === "status") {
                            return (
                              <TableCell key={cell.id}>
                                <Tag type={user.enabled ? "green" : "red"}>
                                  {cell.value}
                                </Tag>
                              </TableCell>
                            );
                          }

                          if (cell.info.header === "roles") {
                            return (
                              <TableCell key={cell.id}>
                                <div style={{ display: "flex", gap: 4 }}>
                                  {user.roles.map((r) => (
                                    <Tag key={r} size="sm">
                                      {r}
                                    </Tag>
                                  ))}
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
                                      user.enabled
                                        ? "Disable user"
                                        : "Enable user"
                                    }
                                    onClick={() => toggleUser(user)}
                                  />

                                  <Button
                                    size="sm"
                                    kind="ghost"
                                    hasIconOnly
                                    renderIcon={Reset}
                                    iconDescription="Reset password"
                                    onClick={() => resetPassword(user)}
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
        )}
      </Tile>
    </div>
  );
}
