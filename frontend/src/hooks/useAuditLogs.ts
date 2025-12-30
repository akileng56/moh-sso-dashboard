import { useEffect, useMemo, useState } from "react";
import { type AuditListResponse } from "../lib/api/audit";

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
  const [data, setData] = useState<AuditListResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const query = useMemo(() => {
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
  }, [filters]);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    fetch(`/admin/audit-logs?${query}`, { credentials: "include" })
      .then(async (r) => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return r.json();
      })
      .then((json) => {
        if (!cancelled) setData(json);
      })
      .catch((e) => {
        if (!cancelled) setError(e.message ?? "Failed");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [query]);

  return { data, loading, error };
}
