// components/clients/panels/CreateClientPanel.tsx
import {
  Stack,
  TextInput,
  Button,
  Checkbox,
  InlineLoading,
  Form,
  FormGroup,
} from "@carbon/react";
import { useState } from "react";

type CreateClientPayload = {
  clientId: string;
  name: string;
  publicClient: boolean;
  enabled: boolean;
};

export function CreateClientPanel() {
  const [form, setForm] = useState<CreateClientPayload>({
    clientId: "",
    name: "",
    publicClient: false,
    enabled: true,
  });

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  /* -----------------------------
   * Validation
   * ----------------------------- */
  const isClientIdValid = /^[a-z0-9-]+$/.test(form.clientId);

  const isValid = form.clientId && form.name && isClientIdValid;

  const handleChange = (field: keyof CreateClientPayload, value: any) => {
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
      /**
       * Payload mirrors Keycloak client creation:
       * - publicClient controls secret generation
       * - enabled toggles client access
       */
      await fetch("/api/v1/admin/clients", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(form),
      });
    } catch (err) {
      setError("Failed to create client");
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
              helperText="Lowercase letters, numbers, and dashes only"
              placeholder="moh-dashboard"
              required
              value={form.clientId}
              invalid={form.clientId.length > 0 && !isClientIdValid}
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
              checked={form.publicClient}
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
              <InlineLoading description="Creating client…" />
            ) : (
              "Create client"
            )}
          </Button>
        </Stack>
      </Stack>
    </Form>
  );
}
