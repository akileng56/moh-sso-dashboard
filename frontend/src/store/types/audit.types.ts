export type Cursor = {
  cursor_created_at?: string | null;
  cursor_id?: string | null;
};

export type AuditListResponse = {
  items: AuditLog[];
  next_cursor: Cursor;
  has_more: boolean;
};

export type MetricsOverview = {
  totalEvents: number;
  totalFailures: number;
  failedLogins: number;
  successfulLogins: number;
};

export type AuditFilters = {
  from: string;
  to: string;
  action?: string;
  user_id?: string;
  client_id?: string;
  ip?: string;
  success?: "true" | "false";
  page?: number;
  limit?: number;
};

export interface AuditLog {
  id: string;
  action: string;
  username?: string;
  user_id: string;
  client_id?: string;
  ip?: string;
  success: boolean;
  created_at: string;
  metadata?: Record<string, any>;
}
