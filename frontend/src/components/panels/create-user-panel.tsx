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
import {
  useCreateUserMutation,
  useUpdateUserMutation,
} from "../../store/api/users.api";
import type { User } from "../../store/types/user.types";

export type UserFormMode = "create" | "edit";

type UserFormPayload = {
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  realm_roles: string[];
  enabled: boolean;
  email_verified: boolean;
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
    first_name: initialUser?.first_name ?? "",
    last_name: initialUser?.last_name ?? "",
    realm_roles: initialUser?.realm_roles ?? [],
    enabled: initialUser?.enabled ?? true,
    email_verified: initialUser?.email_verified ?? true,
  });

  const [error, setError] = useState<string | null>(null);

  const [createUser, { isLoading: creating }] = useCreateUserMutation();

  const [updateUser, { isLoading: updating }] = useUpdateUserMutation();

  const submitting = creating || updating;

  /* -----------------------------
   * Validation
   * ----------------------------- */
  const isValid = useMemo(() => {
    return form.username && form.email && form.first_name && form.last_name;
  }, [form]);

  const handleChange = <K extends keyof UserFormPayload>(
    field: K,
    value: UserFormPayload[K]
  ) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  /* -----------------------------
   * Submit
   * ----------------------------- */
  const handleSubmit = async () => {
    if (!isValid || submitting) return;

    setError(null);

    try {
      if (mode === "create") {
        await createUser(form).unwrap();
      } else if (initialUser?.id) {
        await updateUser({
          id: initialUser.id,
          data: {
            email: form.email,
            first_name: form.first_name,
            last_name: form.last_name,
            realm_roles: form.realm_roles,
            enabled: form.enabled,
            email_verified: form.email_verified,
          },
        }).unwrap();
      }

      onSuccess?.();
    } catch (err: any) {
      setError(
        err?.data?.message ??
          (mode === "create"
            ? "Failed to create user"
            : "Failed to update user")
      );
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
              value={form.first_name}
              onChange={(e) => handleChange("first_name", e.target.value)}
            />

            <TextInput
              id="lastName"
              labelText="Last name"
              required
              value={form.last_name}
              onChange={(e) => handleChange("last_name", e.target.value)}
            />
          </Stack>
        </FormGroup>

        {/* -----------------------------
         * Access & Roles
         * ----------------------------- */}
        <FormGroup legendText="Access control">
          <Stack gap={4}>
            <MultiSelect
              label="Realm roles"
              id="roles"
              titleText="Realm roles"
              items={REALM_ROLES}
              itemToString={(item) => item?.text ?? ""}
              selectedItems={REALM_ROLES.filter((r) =>
                form.realm_roles.includes(r.id)
              )}
              onChange={({ selectedItems }) =>
                handleChange(
                  "realm_roles",
                  selectedItems.map((r) => r.id)
                )
              }
            />

            <Checkbox
              id="enabled"
              labelText="User enabled"
              checked={form.enabled}
              onChange={(_, { checked }) => handleChange("enabled", checked)}
            />

            <Checkbox
              id="emailVerified"
              labelText="Email verified"
              checked={form.email_verified}
              onChange={(_, { checked }) =>
                handleChange("email_verified", checked)
              }
            />
          </Stack>
        </FormGroup>

        {/* -----------------------------
         * Error
         * ----------------------------- */}
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
