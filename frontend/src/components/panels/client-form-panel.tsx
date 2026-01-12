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
import { FormInlineAlert } from "../notifications/in-line-alerts/FormInlineAlert";
import { useToast } from "../notifications/toast/useToast";

export type ClientFormMode = "create" | "edit";

export type ClientFormPayload = {
  clientId: string;
  name: string;
  publicClient: boolean;
  enabled: boolean;
};

type Props = {
  mode: ClientFormMode;
  initialClient?: ClientFormPayload;
  onSuccess?: () => void;
};

export function ClientFormPanel({ mode, initialClient, onSuccess }: Props) {
  const toast = useToast();

  const [form, setForm] = useState<ClientFormPayload>(
    initialClient ?? {
      clientId: "",
      name: "",
      publicClient: false,
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
  const isClientIdValid = /^[a-z0-9-]+$/.test(form.clientId);

  const isValid = useMemo(() => {
    if (!form.clientId || !form.name) return false;
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

        toast.success(
          "Client created",
          `Client "${form.name}" was created successfully`
        );
      } else {
        await updateClient({
          id: form.clientId,
          data: {
            name: form.name,
            publicClient: form.publicClient,
          },
        }).unwrap();

        toast.success("Client updated", `Changes to "${form.name}" were saved`);
      }

      onSuccess?.();
    } catch (err: any) {
      const message =
        err?.data?.message ??
        (mode === "create"
          ? "Failed to create client"
          : "Failed to update client");

      setError(message);

      toast.error("Operation failed", "Please review the form and try again");
    }
  };

  return (
    <Form>
      <Stack gap={5}>
        {/* -----------------------------
         * Inline form error
         * ----------------------------- */}
        {error && (
          <FormInlineAlert title="Unable to save client" subtitle={error} />
        )}

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
              checked={form.publicClient}
              onChange={(_, { checked }) =>
                handleChange("publicClient", checked)
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
