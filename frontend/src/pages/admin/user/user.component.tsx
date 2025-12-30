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
  Dropdown,
  Stack,
} from "@carbon/react";
import { View, Reset, UserFollow } from "@carbon/icons-react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../../../context/useAuth";

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

  // Filters
  const [statusFilter, setStatusFilter] = useState("all");
  const [roleFilter, setRoleFilter] = useState("all");
  const [neverLoggedIn, setNeverLoggedIn] = useState(false);

  // ---------------------------
  // Load users
  // ---------------------------
  useEffect(() => {
    if (!accessToken) return;

    const controller = new AbortController();

    async function loadUsers() {
      try {
        const res = await fetch("http://localhost:9000/api/v1/users", {
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
          credentials: "include",
          signal: controller.signal,
        });

        if (!res.ok) {
          throw new Error(`Failed to load users (${res.status})`);
        }

        const data = await res.json();
        setUsers(data);
      } catch (err: any) {
        if (err.name !== "AbortError") {
          setError(err.message);
        }
      } finally {
        setLoading(false);
      }
    }

    loadUsers();
    return () => controller.abort();
  }, [accessToken]);

  // ---------------------------
  // Derived filters
  // ---------------------------
  const roles = useMemo(() => {
    const set = new Set<string>();
    users.forEach((u) => u.roles?.forEach((r) => set.add(r)));
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

  // ---------------------------
  // Actions
  // ---------------------------
  async function toggleUser(user: User) {
    await fetch(`http://localhost:9000/api/v1/users/${user.id}`, {
      method: user.enabled ? "DELETE" : "POST",
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      credentials: "include",
    });

    setUsers((prev) =>
      prev.map((u) => (u.id === user.id ? { ...u, enabled: !u.enabled } : u))
    );
  }

  function resetPassword(user: User) {
    alert(`Password reset initiated for ${user.username}`);
  }

  // ---------------------------
  // Render
  // ---------------------------
  if (loading) {
    return (
      <div style={{ padding: "2rem" }}>
        <InlineLoading description="Loading users…" />
      </div>
    );
  }

  if (error) {
    return <div style={{ padding: "2rem", color: "red" }}>{error}</div>;
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
    raw: u,
  }));

  return (
    <div style={{ padding: "2rem" }}>
      <Tile>
        {/* Header */}
        <Stack gap={3}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <div>
              <h3>Users</h3>
              <p style={{ color: "#6f6f6f" }}>
                Manage users, roles, and access
              </p>
            </div>

            <Button size="sm" onClick={() => navigate("/admin/users/import")}>
              Import users
            </Button>
          </div>

          {/* Filters */}
          <div style={{ display: "flex", gap: "1rem" }}>
            <Dropdown
              id="status-filter"
              titleText="Status"
              label="Status"
              items={["all", "active", "disabled"]}
              selectedItem={statusFilter}
              onChange={({ selectedItem }) =>
                setStatusFilter(selectedItem as string)
              }
            />

            <Dropdown
              id="role-filter"
              titleText="Role"
              label="Role"
              items={roles}
              selectedItem={roleFilter}
              onChange={({ selectedItem }) =>
                setRoleFilter(selectedItem as string)
              }
            />

            <Button
              size="sm"
              kind={neverLoggedIn ? "primary" : "secondary"}
              onClick={() => setNeverLoggedIn((v) => !v)}
            >
              Never logged in
            </Button>
          </div>

          {/* Table */}
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
                      <TableRow key={row.id} {...getRowProps({ row })}>
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
        </Stack>
      </Tile>
    </div>
  );
}
