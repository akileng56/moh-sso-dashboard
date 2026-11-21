import { useMemo, useState } from "react";
import {
  Button,
  DataTable,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  Tag,
} from "@carbon/react";
import {
  ArrowRight,
  CheckmarkOutline,
  UserFollow,
  WarningAlt,
} from "@carbon/react/icons";
import { AreaChart, LineChart } from "@carbon/charts-react";
import { ScaleTypes } from "@carbon/charts";
import { DataTablePagination } from "@moh-sso/ui";
import { PermissionGuard, PERMISSIONS } from "@moh-sso/auth";

import type { ActionableAgingIssue, AgingBucketId } from "../types";
import {
  AGING_BUCKETS,
  type AgingMetricsSummary,
  safeString,
} from "../lib/issueAggregation";

export interface DashboardAgingTabProps {
  unresolvedAgingIssues: ActionableAgingIssue[];
  agingMetrics: AgingMetricsSummary;
  chartViewMode: "total" | "status";
  activeLocationFilter: string;
  defaultLocationLabel: string;
  chartToolbarConfig: Record<string, unknown>;
  onAssignSingle: (item: ActionableAgingIssue) => void;
  onBulkAssignUnassigned: (codes: string[]) => void;
  onQuickResolve: (item: ActionableAgingIssue) => void;
  onNavigateToIssueCode: (issueCode: string) => void;
}

const AGING_TABLE_HEADERS = [
  { key: "issueCode", header: "Issue Code" },
  { key: "element", header: "Data Element / Dataset" },
  { key: "facilityName", header: "Facility" },
  { key: "location", header: "District (Region)" },
  { key: "ageDays", header: "Age" },
  { key: "reportedDateLabel", header: "Reported" },
  { key: "assignee", header: "Assigned To" },
  { key: "actions", header: "Actions" },
];

export const DashboardAgingTab = ({
  unresolvedAgingIssues,
  agingMetrics,
  chartViewMode,
  activeLocationFilter,
  defaultLocationLabel,
  chartToolbarConfig,
  onAssignSingle,
  onBulkAssignUnassigned,
  onQuickResolve,
  onNavigateToIssueCode,
}: DashboardAgingTabProps) => {
  const [selectedAgingBucket, setSelectedAgingBucket] = useState<"all" | AgingBucketId>(
    "all",
  );
  const [agingTriageFilter, setAgingTriageFilter] = useState<
    "all" | "overdue" | "critical" | "unassigned"
  >("all");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  const agingChartData = useMemo(() => {
    if (chartViewMode === "status") {
      return AGING_BUCKETS.flatMap((bucket) => {
        const stats = agingMetrics.bucketMap.get(bucket.id) ?? {
          total: 0,
          critical: 0,
          unassigned: 0,
          assigned: 0,
        };
        return [
          { group: "Critical Open", key: bucket.label, value: stats.critical },
          { group: "Unassigned Open", key: bucket.label, value: stats.unassigned },
          { group: "Assigned Open", key: bucket.label, value: stats.assigned },
        ];
      });
    }

    return AGING_BUCKETS.flatMap((bucket) => {
      const stats = agingMetrics.bucketMap.get(bucket.id);
      return [
        {
          group: "Total Unresolved",
          key: bucket.label,
          value: stats?.total ?? 0,
        },
        {
          group: "Critical Open",
          key: bucket.label,
          value: stats?.critical ?? 0,
        },
      ];
    });
  }, [agingMetrics, chartViewMode]);

  const agingChartOptions = useMemo(
    () => ({
      title:
        activeLocationFilter && activeLocationFilter !== defaultLocationLabel
          ? `Unresolved Issue Aging Series (${activeLocationFilter})`
          : "Unresolved Issue Aging Series",
      axes: {
        left: {
          mapsTo: "value",
          title: "Unresolved Issues",
          scaleType: ScaleTypes.LINEAR,
        },
        bottom: {
          mapsTo: "key",
          title: "Age Bracket",
          scaleType: ScaleTypes.LABELS,
        },
      },
      curve: "curveMonotoneX",
      points: {
        enabled: true,
        radius: 5,
        filled: true,
      },
      color: {
        scale:
          chartViewMode === "status"
            ? {
                "Critical Open": "#a2191f",
                "Unassigned Open": "#eb6200",
                "Assigned Open": "#0f62fe",
              }
            : {
                "Total Unresolved": "#0f62fe",
                "Critical Open": "#a2191f",
              },
      },
      height: "340px",
      toolbar: chartToolbarConfig,
      legend: {
        enabled: true,
      },
    }),
    [activeLocationFilter, defaultLocationLabel, chartViewMode, chartToolbarConfig],
  );

  const actionableAgingIssues = useMemo(() => {
    return unresolvedAgingIssues.filter((item) => {
      if (selectedAgingBucket !== "all" && item.bucket.id !== selectedAgingBucket) {
        return false;
      }
      if (agingTriageFilter === "overdue" && item.ageDays <= 30) {
        return false;
      }
      if (agingTriageFilter === "critical" && !item.isCritical) {
        return false;
      }
      if (agingTriageFilter === "unassigned" && !item.isUnassigned) {
        return false;
      }
      return true;
    });
  }, [unresolvedAgingIssues, selectedAgingBucket, agingTriageFilter]);

  const actionableUnassignedCodes = useMemo(
    () =>
      actionableAgingIssues
        .filter((item) => item.isUnassigned && Boolean(item.issue.issue_code))
        .map((item) => item.issue.issue_code),
    [actionableAgingIssues],
  );

  const paginatedItems = useMemo(() => {
    const start = (page - 1) * pageSize;
    return actionableAgingIssues.slice(start, start + pageSize);
  }, [actionableAgingIssues, page, pageSize]);

  const tableRows = useMemo(
    () =>
      paginatedItems.map((item, idx) => ({
        id: `${item.issue.issue_id || idx}-${item.issueCode}`,
        issueCode: item.issueCode,
        element:
          safeString(item.issue.data_element) ||
          safeString(item.issue.dataset) ||
          safeString(item.issue.issue_type) ||
          "Data Quality Issue",
        facilityName: item.facilityName,
        location: `${item.districtName} (${item.regionName})`,
        ageDays: item.ageDays,
        reportedDateLabel: item.reportedDateLabel,
        assignee: item.isUnassigned ? "" : safeString(item.issue.assigned_to),
        actions: item.issueCode,
      })),
    [paginatedItems],
  );

  return (
    <div className="dashboard-chart-card aging-visual-card aging-visual-card--tab">
      <div className="aging-visual-chart-wrapper">
        {chartViewMode === "status" ? (
          <LineChart data={agingChartData} options={agingChartOptions} />
        ) : (
          <AreaChart data={agingChartData} options={agingChartOptions} />
        )}
      </div>

      <div className="aging-controls-bar">
        <div
          className="aging-bucket-pills"
          role="group"
          aria-label="Filter by aging bracket"
        >
          <button
            type="button"
            className={`aging-bucket-chip ${selectedAgingBucket === "all" ? "active" : ""}`}
            onClick={() => {
              setSelectedAgingBucket("all");
              setPage(1);
            }}
          >
            All Brackets ({agingMetrics.totalUnresolved})
          </button>
          {AGING_BUCKETS.map((bucket) => {
            const count = agingMetrics.bucketMap.get(bucket.id)?.total ?? 0;
            const isActive = selectedAgingBucket === bucket.id;
            return (
              <button
                key={bucket.id}
                type="button"
                className={`aging-bucket-chip aging-bucket-chip--${bucket.id.replace("+", "plus")} ${
                  isActive ? "active" : ""
                }`}
                onClick={() => {
                  setSelectedAgingBucket((prev) =>
                    prev === bucket.id ? "all" : bucket.id,
                  );
                  setPage(1);
                }}
              >
                <span
                  className={`aging-bucket-dot aging-bucket-dot--${bucket.id.replace("+", "plus")}`}
                />
                {bucket.label} ({count})
              </button>
            );
          })}
        </div>

        <div className="aging-triage-row">
          <div
            className="aging-triage-filters"
            role="group"
            aria-label="Quick triage filters"
          >
            <button
              type="button"
              className={`aging-triage-btn ${agingTriageFilter === "all" ? "active" : ""}`}
              onClick={() => {
                setAgingTriageFilter("all");
                setPage(1);
              }}
            >
              All Backlog ({agingMetrics.totalUnresolved})
            </button>
            <button
              type="button"
              className={`aging-triage-btn ${agingTriageFilter === "overdue" ? "active" : ""}`}
              onClick={() => {
                setAgingTriageFilter((prev) => (prev === "overdue" ? "all" : "overdue"));
                setPage(1);
              }}
            >
              Overdue &gt;30d ({agingMetrics.overdueCount})
            </button>
            <button
              type="button"
              className={`aging-triage-btn ${agingTriageFilter === "critical" ? "active" : ""}`}
              onClick={() => {
                setAgingTriageFilter((prev) =>
                  prev === "critical" ? "all" : "critical",
                );
                setPage(1);
              }}
            >
              Critical ({agingMetrics.criticalCount})
            </button>
            <button
              type="button"
              className={`aging-triage-btn ${agingTriageFilter === "unassigned" ? "active" : ""}`}
              onClick={() => {
                setAgingTriageFilter((prev) =>
                  prev === "unassigned" ? "all" : "unassigned",
                );
                setPage(1);
              }}
            >
              Unassigned ({agingMetrics.unassignedCount})
            </button>
          </div>

          <div className="aging-triage-meta">
            <span className="aging-avg-text">
              Avg / Oldest Age:{" "}
              <strong>
                {agingMetrics.avgAgeDays}d / {agingMetrics.maxAgeDays}d
              </strong>
            </span>

            {actionableUnassignedCodes.length > 0 && (
              <PermissionGuard permission={PERMISSIONS.issueTrackerAssign}>
                <Button
                  size="sm"
                  kind="secondary"
                  renderIcon={UserFollow}
                  onClick={() => onBulkAssignUnassigned(actionableUnassignedCodes)}
                  className="aging-bulk-assign-btn"
                >
                  Assign {actionableUnassignedCodes.length} Unassigned
                </Button>
              </PermissionGuard>
            )}
          </div>
        </div>
      </div>

      <div className="aging-actionable-queue">
        {actionableAgingIssues.length === 0 ? (
          <div className="aging-empty-queue">
            <p>
              No unresolved issues match the selected aging bracket or triage filter.
            </p>
            {(selectedAgingBucket !== "all" || agingTriageFilter !== "all") && (
              <Button
                size="sm"
                kind="ghost"
                onClick={() => {
                  setSelectedAgingBucket("all");
                  setAgingTriageFilter("all");
                  setPage(1);
                }}
              >
                Clear Aging Filters
              </Button>
            )}
          </div>
        ) : (
          <div className="summary-table-container">
            <DataTable rows={tableRows} headers={AGING_TABLE_HEADERS}>
              {({ rows, headers, getTableProps, getHeaderProps, getRowProps }) => (
                <Table
                  {...getTableProps()}
                  size="md"
                  aria-label="Unresolved aging issues queue"
                >
                  <TableHead>
                    <TableRow>
                      {headers.map((header) => {
                        const { key: headerKey, ...headerProps } = getHeaderProps({
                          header,
                        });
                        return (
                          <TableHeader key={headerKey} {...headerProps}>
                            {header.header}
                          </TableHeader>
                        );
                      })}
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {rows.map((row, rowIdx) => {
                      const item = paginatedItems[rowIdx];
                      const { key: rowKey, ...rowProps } = getRowProps({ row });
                      if (!item) return null;

                      return (
                        <TableRow key={rowKey} {...rowProps}>
                          <TableCell>
                            <div className="aging-code-cell">
                              <button
                                type="button"
                                className="aging-issue-code-link"
                                onClick={() => onNavigateToIssueCode(item.issueCode)}
                                title={`Open issue ${item.issueCode} in Registered Issues`}
                              >
                                {item.issueCode}
                              </button>
                              {item.isCritical && (
                                <Tag type="red" size="sm" renderIcon={WarningAlt}>
                                  Critical
                                </Tag>
                              )}
                            </div>
                          </TableCell>
                          <TableCell>{tableRows[rowIdx]?.element}</TableCell>
                          <TableCell>{item.facilityName}</TableCell>
                          <TableCell>
                            {item.districtName} ({item.regionName})
                          </TableCell>
                          <TableCell>
                            <span className={`aging-pill ${item.bucket.badgeClass}`}>
                              {item.ageDays}d open
                            </span>
                          </TableCell>
                          <TableCell>{item.reportedDateLabel}</TableCell>
                          <TableCell>
                            {item.isUnassigned ? (
                              <span className="aging-assignee-unassigned">
                                Unassigned
                              </span>
                            ) : (
                              <span className="aging-assignee-assigned">
                                {safeString(item.issue.assigned_to)}
                              </span>
                            )}
                          </TableCell>
                          <TableCell>
                            <div className="aging-row-actions">
                              <PermissionGuard permission={PERMISSIONS.issueTrackerAssign}>
                                <button
                                  type="button"
                                  className="aging-action-btn aging-action-btn--assign"
                                  onClick={() => onAssignSingle(item)}
                                  title={
                                    item.isUnassigned
                                      ? `Assign ${item.issueCode}`
                                      : `Re-assign ${item.issueCode}`
                                  }
                                >
                                  <UserFollow size={16} />
                                  <span>{item.isUnassigned ? "Assign" : "Re-assign"}</span>
                                </button>
                              </PermissionGuard>

                              <PermissionGuard permission={PERMISSIONS.issueTrackerManage}>
                                <button
                                  type="button"
                                  className="aging-action-btn aging-action-btn--resolve"
                                  onClick={() => onQuickResolve(item)}
                                  title={`Resolve ${item.issueCode}`}
                                >
                                  <CheckmarkOutline size={16} />
                                  <span>Resolve</span>
                                </button>
                              </PermissionGuard>

                              <button
                                type="button"
                                className="aging-action-btn aging-action-btn--view"
                                onClick={() => onNavigateToIssueCode(item.issueCode)}
                                title={`View full details for ${item.issueCode}`}
                              >
                                <span>Details</span>
                                <ArrowRight size={16} />
                              </button>
                            </div>
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              )}
            </DataTable>

            <DataTablePagination
              page={page}
              pageSize={pageSize}
              totalItems={actionableAgingIssues.length}
              onChange={({ page: newPage, pageSize: newPageSize }) => {
                setPage(newPage);
                setPageSize(newPageSize);
              }}
            />
          </div>
        )}
      </div>
    </div>
  );
};
