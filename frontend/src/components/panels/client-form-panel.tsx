// components/clients/panels/ClientFormPanel.tsx
import {
  Stack,
  TextInput,
  Button,
  Checkbox,
  InlineLoading,
  Form,
  FormGroup,
} from "@carbon/react";
import { useMemo, useState } from "react";

export type ClientFormMode = "create" | "edit";

export type ClientFormPayload = {
  clientId: string;
  name: string;
  public_client: boolean;
  enabled: boolean;
};

type Props = {
  mode: ClientFormMode;
  initialClient?: ClientFormPayload;
  onSuccess?: () => void;
};

export function ClientFormPanel({ mode, initialClient, onSuccess }: Props) {
  const [form, setForm] = useState<ClientFormPayload>(
    initialClient ?? {
      clientId: "",
      name: "",
      public_client: false,
      enabled: true,
    }
  );

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  /* -----------------------------
   * Validation
   * ----------------------------- */
  const isClientIdValid = /^[a-z0-9-]+$/.test(form.clientId);

  const isValid = useMemo(() => {
    if (!form.clientId || !form.name) return false;
    if (mode === "create") return isClientIdValid;
    return true;
  }, [form, mode, isClientIdValid]);

  const handleChange = (field: keyof ClientFormPayload, value: any) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  /* -----------------------------
   * Submit
   * ----------------------------- */
  const handleSubmit = async () => {
    if (!isValid) return;

    setSubmitting(true);
    setError(null);

    try {
      const url =
        mode === "create"
          ? "/api/v1/admin/clients"
          : `/api/v1/admin/clients/${form.clientId}`;

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
        mode === "create"
          ? "Failed to create client"
          : "Failed to update client"
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Form>
      <Stack gap={5}>
        {/* -----------------------------
         * Client details
         * ----------------------------- */}
        <FormGroup legendText="Client details">
          <Stack gap={4}>
            <TextInput
              id="clientId"
              labelText="Client ID"
              helperText={
                mode === "create"
                  ? "Lowercase letters, numbers, and dashes only"
                  : "Client ID cannot be changed"
              }
              placeholder="moh-dashboard"
              required
              disabled={mode === "edit"}
              value={form.clientId}
              invalid={
                mode === "create" &&
                form.clientId.length > 0 &&
                !isClientIdValid
              }
              invalidText="Only lowercase letters, numbers, and dashes are allowed"
              onChange={(e) => handleChange("clientId", e.target.value)}
            />

            <TextInput
              id="name"
              labelText="Client name"
              placeholder="MOH SSO Dashboard"
              required
              value={form.name}
              onChange={(e) => handleChange("name", e.target.value)}
            />
          </Stack>
        </FormGroup>

        {/* -----------------------------
         * Access & Security
         * ----------------------------- */}
        <FormGroup legendText="Access & security">
          <Stack gap={4}>
            <Checkbox
              id="publicClient"
              labelText="Public client (no client secret)"
              checked={form.public_client}
              onChange={(checked) => handleChange("publicClient", checked)}
            />

            <Checkbox
              id="enabled"
              labelText="Client enabled"
              checked={form.enabled}
              onChange={(checked) => handleChange("enabled", checked)}
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
                  mode === "create" ? "Creating client…" : "Updating client…"
                }
              />
            ) : mode === "create" ? (
              "Create client"
            ) : (
              "Save changes"
            )}
          </Button>
        </Stack>
      </Stack>
    </Form>
  );
}
