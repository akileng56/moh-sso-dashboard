// components/users/panels/UserFormPanel.tsx
import {
  Stack,
  TextInput,
  Button,
  Checkbox,
  InlineLoading,
  MultiSelect,
  Form,
  FormGroup,
} from "@carbon/react";
import { useMemo, useState } from "react";
import type { User } from "../../store/types/user.types";

export type UserFormMode = "create" | "edit";

type UserFormPayload = {
  username: string;
  email: string;
  firstName: string;
  lastName: string;
  roles: string[];
  enabled: boolean;
  emailVerified: boolean;
};

const REALM_ROLES = [
  { id: "admin", text: "Admin" },
  { id: "manager", text: "Manager" },
  { id: "user", text: "User" },
];

type Props = {
  mode: UserFormMode;
  initialUser?: User;
  onSuccess?: () => void;
};

export function UserFormPanel({ mode, initialUser, onSuccess }: Props) {
  const [form, setForm] = useState<UserFormPayload>({
    username: initialUser?.username ?? "",
    email: initialUser?.email ?? "",
    firstName: initialUser?.first_name ?? "",
    lastName: initialUser?.last_name ?? "",
    roles: initialUser?.client_roles?.admin ?? [],
    enabled: initialUser?.enabled ?? true,
    emailVerified: initialUser?.email_verified ?? true,
  });

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isValid = useMemo(() => {
    return form.username && form.email && form.firstName && form.lastName;
  }, [form]);

  const handleChange = (field: keyof UserFormPayload, value: any) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async () => {
    if (!isValid) return;

    setSubmitting(true);
    setError(null);

    try {
      const url =
        mode === "create"
          ? "/api/v1/admin/users"
          : `/api/v1/admin/users/${initialUser?.id}`;

      const method = mode === "create" ? "POST" : "PUT";

      await fetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(form),
      });

      onSuccess?.();
    } catch {
      setError(
        mode === "create" ? "Failed to create user" : "Failed to update user"
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Form>
      <Stack gap={5}>
        {/* -----------------------------
         * User details
         * ----------------------------- */}
        <FormGroup legendText="User details">
          <Stack gap={4}>
            <TextInput
              id="username"
              labelText="Username"
              required
              disabled={mode === "edit"}
              helperText={
                mode === "edit" ? "Username cannot be changed" : undefined
              }
              value={form.username}
              onChange={(e) => handleChange("username", e.target.value)}
            />

            <TextInput
              id="email"
              labelText="Email"
              type="email"
              required
              value={form.email}
              onChange={(e) => handleChange("email", e.target.value)}
            />

            <TextInput
              id="firstName"
              labelText="First name"
              required
              value={form.firstName}
              onChange={(e) => handleChange("firstName", e.target.value)}
            />

            <TextInput
              id="lastName"
              labelText="Last name"
              required
              value={form.lastName}
              onChange={(e) => handleChange("lastName", e.target.value)}
            />
          </Stack>
        </FormGroup>

        {/* -----------------------------
         * Access & Roles
         * ----------------------------- */}
        <FormGroup legendText="Access control">
          <Stack gap={4}>
            <MultiSelect
              id="roles"
              titleText="Realm roles"
              items={REALM_ROLES}
              itemToString={(item) => item?.text ?? ""}
              initialSelectedItems={REALM_ROLES.filter((r) =>
                form.roles.includes(r.id)
              )}
              onChange={({ selectedItems }) =>
                handleChange(
                  "roles",
                  selectedItems.map((r) => r.id)
                )
              }
            />

            <Checkbox
              id="enabled"
              labelText="User enabled"
              checked={form.enabled}
              onChange={(checked) => handleChange("enabled", checked)}
            />

            <Checkbox
              id="emailVerified"
              labelText="Email verified"
              checked={form.emailVerified}
              onChange={(checked) => handleChange("emailVerified", checked)}
            />
          </Stack>
        </FormGroup>

        {error && <p style={{ color: "var(--cds-text-error)" }}>{error}</p>}

        {/* -----------------------------
         * Actions
         * ----------------------------- */}
        <Stack orientation="horizontal" gap={3}>
          <Button
            kind="primary"
            disabled={!isValid || submitting}
            onClick={handleSubmit}
          >
            {submitting ? (
              <InlineLoading
                description={
                  mode === "create" ? "Creating user…" : "Saving changes…"
                }
              />
            ) : mode === "create" ? (
              "Create user"
            ) : (
              "Save changes"
            )}
          </Button>
        </Stack>
      </Stack>
    </Form>
  );
}
