export type AuditLog = {
  id: string;
  createdAt: string;
  userId?: string | null;
  username: string;
  action: string;
  metadata: Record<string, any> | null;
};

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
