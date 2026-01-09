// components/users/panels/CreateUserPanel.tsx
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
import { useState } from "react";

type CreateUserPayload = {
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

export function CreateUserPanel() {
  const [form, setForm] = useState<CreateUserPayload>({
    username: "",
    email: "",
    firstName: "",
    lastName: "",
    roles: [],
    enabled: true,
    emailVerified: true,
  });

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isValid =
    form.username && form.email && form.firstName && form.lastName;

  const handleChange = (field: keyof CreateUserPayload, value: any) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async () => {
    setSubmitting(true);
    setError(null);

    try {
      /**
       * This payload matches Keycloak user creation:
       * - credentials generated server-side
       * - temporary password
       * - email verification flow
       */
      await fetch("/api/v1/admin/users", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(form),
      });
    } catch (err: any) {
      setError("Failed to create user");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Form>
      <Stack gap={5}>
        {/* -----------------------------
         * Identity
         * ----------------------------- */}
        <FormGroup legendText="User details">
          <Stack gap={4}>
            <TextInput
              id="username"
              labelText="Username"
              value={form.username}
              required
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
              onChange={({ selectedItems }) =>
                handleChange(
                  "roles",
                  selectedItems.map((r) => r.id)
                )
              }
              label={"Realm Roles"}
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
              <InlineLoading description="Creating user…" />
            ) : (
              "Create user"
            )}
          </Button>
        </Stack>
      </Stack>
    </Form>
  );
}
