import { useEffect, useMemo, useState } from "react";
import { type AuditListResponse } from "../lib/api/audit";
import { useAuth } from "../context/useAuth";

type Filters = {
  from: string;
  to: string;
  action?: string;
  user_id?: string;
  client_id?: string;
  ip?: string;
  success?: "true" | "false";
};

export function useAuditLogs(filters: Filters) {
  const { accessToken } = useAuth();

  const [data, setData] = useState<AuditListResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  /* ---------------------------
   * Build query (stable + guarded)
   * --------------------------- */
  const query = useMemo(() => {
    if (!filters.from || !filters.to) return null;

    const p = new URLSearchParams();
    p.set("from", filters.from);
    p.set("to", filters.to);

    if (filters.action) p.set("action", filters.action);
    if (filters.user_id) p.set("user_id", filters.user_id);
    if (filters.client_id) p.set("client_id", filters.client_id);
    if (filters.ip) p.set("ip", filters.ip);
    if (filters.success) p.set("success", filters.success);

    p.set("limit", "50");

    return p.toString();
  }, [
    filters.from,
    filters.to,
    filters.action,
    filters.user_id,
    filters.client_id,
    filters.ip,
    filters.success,
  ]);

  /* ---------------------------
   * Fetch audit logs
   * --------------------------- */
  useEffect(() => {
    if (!query || !accessToken) return;

    const controller = new AbortController();

    setLoading(true);
    setError(null);

    fetch(`http://localhost:9000/api/v1/admin/audit-logs?${query}`, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      credentials: "include",
      signal: controller.signal,
    })
      .then(async (res) => {
        if (!res.ok) {
          const text = await res.text();
          throw new Error(text || `HTTP ${res.status}`);
        }
        return res.json();
      })
      .then((json: AuditListResponse) => {
        setData(json);
      })
      .catch((err: any) => {
        if (err.name !== "AbortError") {
          setError(err.message ?? "Failed to load audit logs");
        }
      })
      .finally(() => {
        setLoading(false);
      });

    return () => controller.abort();
  }, [query, accessToken]);

  return { data, loading, error };
}
