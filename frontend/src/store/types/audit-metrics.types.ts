export interface AuditOverview {
  totalEvents: number;
  totalFailures: number;
  failedLogins: number;
  successfulLogins: number;
}

export interface AuditMetricsFilters {
  from: string;
  to: string;
}
