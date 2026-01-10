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
import {
  useCreateClientMutation,
  useUpdateClientMutation,
} from "../../store/api/clients.api";

export type ClientFormMode = "create" | "edit";

export type ClientFormPayload = {
  client_id: string;
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
      client_id: "",
      name: "",
      public_client: false,
      enabled: true,
    }
  );

  const [error, setError] = useState<string | null>(null);

  const [createClient, { isLoading: creating }] = useCreateClientMutation();

  const [updateClient, { isLoading: updating }] = useUpdateClientMutation();

  const submitting = creating || updating;

  /* -----------------------------
   * Validation
   * ----------------------------- */
  const isClientIdValid = /^[a-z0-9-]+$/.test(form.client_id);

  const isValid = useMemo(() => {
    if (!form.client_id || !form.name) return false;
    if (mode === "create") return isClientIdValid;
    return true;
  }, [form, mode, isClientIdValid]);

  const handleChange = <K extends keyof ClientFormPayload>(
    field: K,
    value: ClientFormPayload[K]
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
        await createClient(form).unwrap();
      } else {
        await updateClient({
          id: form.client_id,
          data: {
            name: form.name,
            public_client: form.public_client,
          },
        }).unwrap();
      }

      onSuccess?.();
    } catch (err: any) {
      setError(
        err?.data?.message ??
          (mode === "create"
            ? "Failed to create client"
            : "Failed to update client")
      );
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
              value={form.client_id}
              invalid={
                mode === "create" &&
                form.client_id.length > 0 &&
                !isClientIdValid
              }
              invalidText="Only lowercase letters, numbers, and dashes are allowed"
              onChange={(e) => handleChange("client_id", e.target.value)}
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
              onChange={(_, { checked }) =>
                handleChange("public_client", checked)
              }
            />

            <Checkbox
              id="enabled"
              labelText="Client enabled"
              checked={form.enabled}
              onChange={(_, { checked }) => handleChange("enabled", checked)}
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
