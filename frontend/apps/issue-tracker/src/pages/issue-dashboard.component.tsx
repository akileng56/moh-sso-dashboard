import { useEffect, useMemo, useRef, useState, Fragment } from "react";
import {
  Button,
  ComboBox,
  Dropdown,
  Modal,
  Popover,
  PopoverContent,
  Search,
  TextArea,
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableExpandHeader,
  TableExpandRow,
  TableExpandedRow,
  InlineLoading,
} from "@carbon/react";
import {
  ChevronDown,
  ChevronRight,
  Download,
  Filter,
  ArrowRight,
  Renew,
  CheckmarkOutline,
  UserFollow,
  WarningAlt,
} from "@carbon/react/icons";
import { GroupedBarChart, PieChart, SimpleBarChart } from "@carbon/charts-react";
import { Alignments, ScaleTypes, ToolbarControlTypes } from "@carbon/charts";
import "@carbon/charts/styles.css";
import * as XLSX from "xlsx";
import { useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";
import { PermissionGuard, PERMISSIONS, selectUser } from "@moh-sso/auth";

import {
  useCreateTransactionMutation,
  useGetIssuesQuery,
  useGetIssuesSummaryByProgramQuery,
} from "../api";
import { AssignModal } from "../component/assign-modal.component.tsx";
import type { Issue } from "../types";
import { periodType } from "../../../data-visualizer/src/pages/Constants.tsx";
import {
  convertPeriodToCode,
  fetchAvailablePeriods,
  formatReportingPeriod,
  isPeriodMatching,
} from "../lib/periodUtils.ts";
import {
  useGetDataSetsQuery,
  type Dataset,
} from "../../../data-visualizer/src/pages/modals/data-model/data-model.ts";
import { useGetHierarchyQuery } from "../../../data-visualizer/src/pages/modals/orgunit/org-unit.ts";

import "./issue-dashboard.scss";

type FacilitySummary = {
  facility: string;
  levelOfCare: string;
  critical: number;
  open: number;
  resolved: number;
  total: number;
};

type DistrictSummary = {
  district: string;
  critical: number;
  open: number;
  resolved: number;
  total: number;
  facilities: FacilitySummary[];
};

type RegionSummaryRow = {
  id: string;
  region: string;
  critical: number;
  open: number;
  resolved: number;
  total: number;
  percentage: number;
  districts: DistrictSummary[];
  [key: string]: unknown;
};

type SelectEvent<T> = {
  selectedItem?: T | null;
};

type PeriodOption = {
  id?: string;
  label: string;
  value?: string;
  disabled?: boolean;
};

type SortHeader = "region" | "critical" | "open" | "resolved" | "total" | "percentage" | string;

type AgingBucketId = "0-7" | "8-14" | "15-30" | "31-60" | "61-90" | "90+";

type AgingBucketDefinition = {
  id: AgingBucketId;
  label: string;
  shortLabel: string;
  minDays: number;
  maxDays: number;
  color: string;
  badgeClass: string;
};

const AGING_BUCKETS: AgingBucketDefinition[] = [
  {
    id: "0-7",
    label: "0–7 Days",
    shortLabel: "0–7d",
    minDays: 0,
    maxDays: 7,
    color: "#0f62fe",
    badgeClass: "aging-pill--fresh",
  },
  {
    id: "8-14",
    label: "8–14 Days",
    shortLabel: "8–14d",
    minDays: 8,
    maxDays: 14,
    color: "#007d79",
    badgeClass: "aging-pill--recent",
  },
  {
    id: "15-30",
    label: "15–30 Days",
    shortLabel: "15–30d",
    minDays: 15,
    maxDays: 30,
    color: "#d2a106",
    badgeClass: "aging-pill--aging",
  },
  {
    id: "31-60",
    label: "31–60 Days",
    shortLabel: "31–60d",
    minDays: 31,
    maxDays: 60,
    color: "#eb6200",
    badgeClass: "aging-pill--overdue",
  },
  {
    id: "61-90",
    label: "61–90 Days",
    shortLabel: "61–90d",
    minDays: 61,
    maxDays: 90,
    color: "#da1e28",
    badgeClass: "aging-pill--stale",
  },
  {
    id: "90+",
    label: "> 90 Days",
    shortLabel: "> 90d",
    minDays: 91,
    maxDays: Number.POSITIVE_INFINITY,
    color: "#a2191f",
    badgeClass: "aging-pill--critical",
  },
];

type ActionableAgingIssue = {
  issue: Issue;
  issueCode: string;
  ageDays: number;
  bucket: AgingBucketDefinition;
  regionName: string;
  districtName: string;
  facilityName: string;
  isCritical: boolean;
  isUnassigned: boolean;
  reportedDateLabel: string;
};

function parseIssueDate(issue: Issue): Date | null {
  const candidates = [safeString(issue.date_reported), safeString(issue.updated_date)];

  for (const raw of candidates) {
    if (!raw) continue;

    const isoMatch = raw.match(/^(\d{4})-(0[1-9]|1[0-2])-(\d{2})/);
    if (isoMatch) {
      const parsed = new Date(
        Number(isoMatch[1]),
        Number(isoMatch[2]) - 1,
        Number(isoMatch[3]),
      );
      if (!Number.isNaN(parsed.getTime())) return parsed;
    }

    const dmyMatch = raw.match(/^(\d{1,2})[/-](0?[1-9]|1[0-2])[/-](\d{4})$/);
    if (dmyMatch) {
      const parsed = new Date(
        Number(dmyMatch[3]),
        Number(dmyMatch[2]) - 1,
        Number(dmyMatch[1]),
      );
      if (!Number.isNaN(parsed.getTime())) return parsed;
    }

    const direct = new Date(raw);
    if (!Number.isNaN(direct.getTime())) return direct;
  }

  const rawPeriod = safeString(issue.time_Period || issue.time_period);
  if (rawPeriod) {
    const code = convertPeriodToCode(rawPeriod);
    const mMatch = code.match(/^(\d{4})(0[1-9]|1[0-2])$/);
    if (mMatch) {
      return new Date(Number(mMatch[1]), Number(mMatch[2]) - 1, 1);
    }
    const qMatch = code.match(/^(\d{4})Q([1-4])$/i);
    if (qMatch) {
      const startMonth = (Number(qMatch[2]) - 1) * 3;
      return new Date(Number(qMatch[1]), startMonth, 1);
    }
    const yMatch = rawPeriod.match(/\b(19\d\d|20\d\d)\b/);
    if (yMatch) {
      return new Date(Number(yMatch[1]), 0, 1);
    }
  }

  return null;
}

function getAgingBucketForDays(ageDays: number): AgingBucketDefinition {
  for (const bucket of AGING_BUCKETS) {
    if (ageDays >= bucket.minDays && ageDays <= bucket.maxDays) {
      return bucket;
    }
  }
  return AGING_BUCKETS[AGING_BUCKETS.length - 1];
}

function safeString(val: unknown): string {
  if (val == null) return "";
  return String(val).trim();
}

function safeLower(val: unknown): string {
  return safeString(val).toLowerCase();
}

function isIssueResolved(status?: string | null): boolean {
  if (!status) return false;
  const s = safeLower(status);
  return s === "resolved" || s === "closed" || s === "done";
}

function isCriticalIssue(issue: Issue): boolean {
  const sev = safeLower(issue.severity);
  const prio = safeLower(issue.priority);
  return sev === "critical" || prio === "critical";
}

function doesIssueMatchReportingPeriod(
  issue: Issue,
  selectedPeriod: string,
  selectedYear: number,
  isYearFilterActive: boolean,
): boolean {
  const rawPeriod = safeString(issue.time_Period || issue.time_period);

  if (selectedPeriod) {
    if (!rawPeriod) return false;

    if (isPeriodMatching(selectedPeriod, rawPeriod, selectedYear)) {
      return true;
    }

    const formattedTarget = safeLower(formatReportingPeriod(selectedPeriod));
    const formattedIssue = safeLower(formatReportingPeriod(rawPeriod));
    if (formattedTarget && formattedIssue && formattedTarget === formattedIssue) {
      return true;
    }

    const targetCode = convertPeriodToCode(selectedPeriod, selectedYear).toUpperCase();
    const issueCode = convertPeriodToCode(rawPeriod, selectedYear).toUpperCase();

    // If filtering by Quarter (e.g. 2026Q1), also match monthly or ISO-date issues within that quarter
    const qMatch = targetCode.match(/^(\d{4})Q([1-4])$/);
    if (qMatch) {
      const targetYear = qMatch[1];
      const targetQuarter = Number(qMatch[2]);

      const mMatch = issueCode.match(/^(\d{4})(0[1-9]|1[0-2])$/);
      if (mMatch && mMatch[1] === targetYear) {
        const issueQuarter = Math.ceil(Number(mMatch[2]) / 3);
        if (issueQuarter === targetQuarter) return true;
      }

      const dateMatch = rawPeriod.match(/^(\d{4})-(0[1-9]|1[0-2])-\d{2}/);
      if (dateMatch && dateMatch[1] === targetYear) {
        const issueQuarter = Math.ceil(Number(dateMatch[2]) / 3);
        if (issueQuarter === targetQuarter) return true;
      }
    }

    // If filtering by Month (e.g. 202602), also match ISO-date issues in that month
    const targetMonthMatch = targetCode.match(/^(\d{4})(0[1-9]|1[0-2])$/);
    if (targetMonthMatch) {
      const dateMatch = rawPeriod.match(/^(\d{4})-(0[1-9]|1[0-2])-\d{2}/);
      if (
        dateMatch &&
        dateMatch[1] === targetMonthMatch[1] &&
        dateMatch[2] === targetMonthMatch[2]
      ) {
        return true;
      }
    }

    return false;
  }

  if (isYearFilterActive) {
    if (!rawPeriod) return false;
    const yearStr = String(selectedYear);
    const formattedIssue = formatReportingPeriod(rawPeriod);
    const issueCode = convertPeriodToCode(rawPeriod);
    return (
      rawPeriod.includes(yearStr) ||
      formattedIssue.includes(yearStr) ||
      issueCode.includes(yearStr)
    );
  }

  return true;
}

function resolveLevelOfCare(rawLevel?: string | null, facilityName?: string | null): string {
  const explicit = safeLower(rawLevel);
  if (explicit) {
    if (explicit === "nrh" || explicit.includes("national referral")) return "NRH";
    if (explicit === "rrh" || explicit.includes("regional referral")) return "RRH";
    if (explicit.includes("hospital") || explicit === "gh") return "General Hospital";
    if (explicit === "hciv" || explicit === "hc iv" || explicit === "hc 4") return "HC IV";
    if (explicit === "hciii" || explicit === "hc iii" || explicit === "hc 3") return "HC III";
    if (explicit === "hcii" || explicit === "hc ii" || explicit === "hc 2") return "HC II";
    if (explicit.includes("clinic")) return "Clinic";
    if (explicit.includes("drug shop")) return "Drug Shop";
    if (explicit === "rbb") return "RBB";
    if (explicit === "nbb") return "NBB";
    if (explicit === "bcdp") return "BCDP";
    return safeString(rawLevel);
  }

  const combined = `${safeString(rawLevel)} ${safeString(facilityName)}`;
  if (/\bnrh\b|national\s+referral/i.test(combined)) return "NRH";
  if (/\brrh\b|regional\s+referral/i.test(combined)) return "RRH";
  if (/\bhospital\b|\bgh\b/i.test(combined)) return "General Hospital";
  if (/\bhc\s*(?:iv|4)\b|health\s+cent(?:re|er)\s+(?:iv|4)\b/i.test(combined)) return "HC IV";
  if (/\bhc\s*(?:iii|3)\b|health\s+cent(?:re|er)\s+(?:iii|3)\b/i.test(combined)) return "HC III";
  if (/\bhc\s*(?:ii|2)\b|health\s+cent(?:re|er)\s+(?:ii|2)\b/i.test(combined)) return "HC II";
  if (/\bclinic\b|medical\s+cent(?:re|er)|maternity|dispensary|health\s+unit/i.test(combined)) {
    return "Clinic";
  }
  if (/\bdrug\s*shop\b|pharmacy/i.test(combined)) return "Drug Shop";
  if (/\brbb\b|regional\s+blood\s+bank/i.test(combined)) return "RBB";
  if (/\bnbb\b|national\s+blood\s+bank/i.test(combined)) return "NBB";
  if (/\bbcdp\b/i.test(combined)) return "BCDP";

  return "Other / Unspecified";
}

function aggregateIssuesByRegion(
  issues: Issue[],
  orgToRegionMap: Map<string, string>,
  orgToDistrictMap: Map<string, string>,
  orgIdToNameMap: Map<string, string>,
) {
  const regionGroups = new Map<
    string,
    {
      critical: number;
      open: number;
      resolved: number;
      total: number;
      districtGroups: Map<
        string,
        {
          critical: number;
          open: number;
          resolved: number;
          total: number;
          facilityGroups: Map<
            string,
            {
              levelOfCare: string;
              critical: number;
              open: number;
              resolved: number;
              total: number;
            }
          >;
        }
      >;
    }
  >();

  let totalCritical = 0;
  let totalOpen = 0;
  let totalResolved = 0;
  let totalAll = 0;

  for (const issue of issues) {
    // 1. Resolve Region
    let regionName = safeString(issue.region);
    if (regionName && orgIdToNameMap.has(safeLower(regionName))) {
      regionName = orgIdToNameMap.get(safeLower(regionName))!;
    }
    if (!regionName && issue.org_unit) {
      regionName = orgToRegionMap.get(safeLower(issue.org_unit)) || "";
    }
    if (!regionName && issue.district) {
      regionName = orgToRegionMap.get(safeLower(issue.district)) || "";
    }
    if (!regionName) {
      regionName = "Unspecified";
    }

    // 2. Resolve District
    let districtName = safeString(issue.district);
    if (districtName && orgIdToNameMap.has(safeLower(districtName))) {
      districtName = orgIdToNameMap.get(safeLower(districtName))!;
    }
    if (!districtName && issue.org_unit) {
      districtName =
        orgToDistrictMap.get(safeLower(issue.org_unit)) || safeString(issue.org_unit);
    }
    if (!districtName) {
      districtName = "Unspecified";
    }

    // 3. Resolve Facility
    let facilityName = safeString(issue.org_unit);
    if (!facilityName) {
      facilityName = "Unspecified Facility";
    } else {
      const mappedName = orgIdToNameMap.get(safeLower(facilityName));
      if (mappedName) {
        facilityName = mappedName;
      }
    }

    // 4. Resolve Status & Critical (Open)
    const resolved = isIssueResolved(issue.status);
    const isCriticalOpen = !resolved && isCriticalIssue(issue);

    // Init Region group if absent
    if (!regionGroups.has(regionName)) {
      regionGroups.set(regionName, {
        critical: 0,
        open: 0,
        resolved: 0,
        total: 0,
        districtGroups: new Map(),
      });
    }

    const rGroup = regionGroups.get(regionName)!;
    rGroup.total += 1;
    totalAll += 1;

    if (resolved) {
      rGroup.resolved += 1;
      totalResolved += 1;
    } else {
      rGroup.open += 1;
      totalOpen += 1;
    }

    if (isCriticalOpen) {
      rGroup.critical += 1;
      totalCritical += 1;
    }

    // District aggregation inside Region
    if (!rGroup.districtGroups.has(districtName)) {
      rGroup.districtGroups.set(districtName, {
        critical: 0,
        open: 0,
        resolved: 0,
        total: 0,
        facilityGroups: new Map(),
      });
    }

    const dGroup = rGroup.districtGroups.get(districtName)!;
    dGroup.total += 1;
    if (resolved) {
      dGroup.resolved += 1;
    } else {
      dGroup.open += 1;
    }
    if (isCriticalOpen) {
      dGroup.critical += 1;
    }

    // Facility aggregation inside District
    const levelOfCare = resolveLevelOfCare(issue.level_of_care, facilityName);
    if (!dGroup.facilityGroups.has(facilityName)) {
      dGroup.facilityGroups.set(facilityName, {
        levelOfCare,
        critical: 0,
        open: 0,
        resolved: 0,
        total: 0,
      });
    }

    const fGroup = dGroup.facilityGroups.get(facilityName)!;
    if (
      fGroup.levelOfCare === "Other / Unspecified" &&
      levelOfCare !== "Other / Unspecified"
    ) {
      fGroup.levelOfCare = levelOfCare;
    }
    fGroup.total += 1;
    if (resolved) {
      fGroup.resolved += 1;
    } else {
      fGroup.open += 1;
    }
    if (isCriticalOpen) {
      fGroup.critical += 1;
    }
  }

  // Build rows list
  const rows: RegionSummaryRow[] = [];
  regionGroups.forEach((val, region) => {
    const districtsList: DistrictSummary[] = [];
    val.districtGroups.forEach((dVal, dName) => {
      const facilitiesList: FacilitySummary[] = [];
      dVal.facilityGroups.forEach((fVal, fName) => {
        facilitiesList.push({
          facility: fName,
          levelOfCare: fVal.levelOfCare,
          critical: fVal.critical,
          open: fVal.open,
          resolved: fVal.resolved,
          total: fVal.total,
        });
      });

      // Sort facilities by total descending, then critical descending
      facilitiesList.sort((a, b) => b.total - a.total || b.critical - a.critical);

      districtsList.push({
        district: dName,
        critical: dVal.critical,
        open: dVal.open,
        resolved: dVal.resolved,
        total: dVal.total,
        facilities: facilitiesList,
      });
    });

    // Sort districts by total descending
    districtsList.sort((a, b) => b.total - a.total);

    const percentage = totalAll > 0 ? (val.total / totalAll) * 100 : 0;

    rows.push({
      id: `region-${region}`,
      region,
      critical: val.critical,
      open: val.open,
      resolved: val.resolved,
      total: val.total,
      percentage,
      districts: districtsList,
    });
  });

  return {
    regionalSummary: rows,
    grandTotals: {
      critical: totalCritical,
      open: totalOpen,
      resolved: totalResolved,
      total: totalAll,
      activeRegionsCount: regionGroups.size,
    },
  };
}

const IssueDashboard = () => {
  const navigate = useNavigate();
  const user = useSelector(selectUser);
  const currentYear = new Date().getFullYear();

  /* -------------------------------------------------------------
   * API Queries & Mutations
   * ------------------------------------------------------------- */
  const [selectedProgram, setSelectedProgram] = useState<string | undefined>();
  const [createTransaction, { isLoading: isResolvingIssue }] = useCreateTransactionMutation();

  const {
    data: issuesData,
    isLoading: isLoadingIssues,
    error: issuesError,
    refetch: refetchIssues,
  } = useGetIssuesQuery({
    limit: 10000,
    offset: 0,
  });

  const {
    data: programIssuesData,
    isFetching: isFetchingProgramIssues,
    refetch: refetchProgramIssues,
  } = useGetIssuesQuery(
    {
      limit: 10000,
      offset: 0,
      program: selectedProgram,
    },
    { skip: !selectedProgram },
  );

  const { data: summaryPrograms = [] } = useGetIssuesSummaryByProgramQuery();
  const { data: datasets = [] } = useGetDataSetsQuery();
  const { data: hierarchyData = [] } = useGetHierarchyQuery();

  /* -------------------------------------------------------------
   * Filters State
   * ------------------------------------------------------------- */
  const [selectedYear, setSelectedYear] = useState<number>(currentYear);
  const [isYearFilterApplied, setIsYearFilterApplied] = useState<boolean>(false);
  const [selectedPeriodType, setSelectedPeriodType] = useState<string>("Monthly");
  const [selectedPeriod, setSelectedPeriod] = useState<string>("");
  const [availablePeriods, setAvailablePeriods] = useState<PeriodOption[]>(
    fetchAvailablePeriods("Monthly", currentYear.toString()),
  );
  const [isPeriodPopoverOpen, setIsPeriodPopoverOpen] = useState<boolean>(false);
  const periodPopoverRef = useRef<HTMLDivElement>(null);

  const [selectedDataset, setSelectedDataset] = useState<string>("");
  const [regionSearchTerm, setRegionSearchTerm] = useState<string>("");

  /* Sorting state */
  const [sortHeader, setSortHeader] = useState<SortHeader>("total");
  const [sortDirection, setSortDirection] = useState<"ASC" | "DESC">("DESC");

  /* Chart view controls */
  const DEFAULT_CHART_REGION = "South Buganda";
  const [chartViewMode, setChartViewMode] = useState<"total" | "status">("total");
  const [selectedChartRegion, setSelectedChartRegion] =
    useState<string>(DEFAULT_CHART_REGION);

  /* Unresolved aging issues visual & action state */
  const [selectedAgingBucket, setSelectedAgingBucket] = useState<"all" | AgingBucketId>("all");
  const [agingTriageFilter, setAgingTriageFilter] = useState<
    "all" | "overdue" | "critical" | "unassigned"
  >("all");
  const [assigningIssueCodes, setAssigningIssueCodes] = useState<string[]>([]);
  const [isAssignReassign, setIsAssignReassign] = useState<boolean>(false);
  const [resolvingAgingIssue, setResolvingAgingIssue] = useState<ActionableAgingIssue | null>(null);
  const [resolutionComment, setResolutionComment] = useState<string>("");

  /* Expanded districts state for facility breakdown */
  const [expandedDistricts, setExpandedDistricts] = useState<Record<string, boolean>>({});

  const isDistrictExpanded = (region: string, d: DistrictSummary) => {
    const districtKey = `${region}-${d.district}`;
    if (expandedDistricts[districtKey] !== undefined) {
      return expandedDistricts[districtKey];
    }
    if (regionSearchTerm.trim()) {
      const term = safeLower(regionSearchTerm);
      return (
        safeLower(d.district).includes(term) ||
        d.facilities.some((f) => safeLower(f.facility).includes(term))
      );
    }
    return false;
  };

  const toggleDistrict = (region: string, d: DistrictSummary) => {
    const districtKey = `${region}-${d.district}`;
    const currentlyExpanded = isDistrictExpanded(region, d);
    setExpandedDistricts((prev) => ({
      ...prev,
      [districtKey]: !currentlyExpanded,
    }));
  };

  const isEveryDistrictExpanded = (region: string, districts: DistrictSummary[]) => {
    if (districts.length === 0) return false;
    return districts.every((d) => isDistrictExpanded(region, d));
  };

  const toggleAllDistrictsInRegion = (region: string, districts: DistrictSummary[]) => {
    const allExpanded = isEveryDistrictExpanded(region, districts);
    setExpandedDistricts((prev) => {
      const next = { ...prev };
      for (const d of districts) {
        next[`${region}-${d.district}`] = !allExpanded;
      }
      return next;
    });
  };

  const years = useMemo(
    () => Array.from({ length: 10 }, (_, index) => currentYear - index),
    [currentYear],
  );

  const datasetNames = useMemo(
    () => (datasets as Dataset[]).map((d) => d.display_name).filter(Boolean),
    [datasets],
  );

  const programNames = useMemo(
    () => summaryPrograms.map((p) => p.program).filter(Boolean),
    [summaryPrograms],
  );

  const hasActiveFilters =
    selectedYear !== currentYear ||
    isYearFilterApplied ||
    selectedPeriodType !== "Monthly" ||
    Boolean(selectedPeriod) ||
    Boolean(selectedDataset) ||
    Boolean(selectedProgram);

  /* Close popover on outside click */
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node;
      if (periodPopoverRef.current && !periodPopoverRef.current.contains(target)) {
        setIsPeriodPopoverOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  /* -------------------------------------------------------------
   * Hierarchy Org-to-Region & Org-to-District mapping
   * ------------------------------------------------------------- */
  const { orgToRegionMap, orgToDistrictMap, orgIdToNameMap } = useMemo(() => {
    const regionMap = new Map<string, string>();
    const districtMap = new Map<string, string>();
    const idToNameMap = new Map<string, string>();

    if (!hierarchyData) {
      return {
        orgToRegionMap: regionMap,
        orgToDistrictMap: districtMap,
        orgIdToNameMap: idToNameMap,
      };
    }

    const rootList: any[] = Array.isArray(hierarchyData)
      ? hierarchyData
      : typeof hierarchyData === "object"
        ? [hierarchyData]
        : [];

    function traverseTree(
      nodes: any[],
      currentRegion: string | null,
      currentDistrict: string | null,
    ) {
      if (!Array.isArray(nodes)) return;

      for (const node of nodes) {
        if (!node || typeof node !== "object") continue;

        const nodeName = safeString(node.name);
        const nodeId = safeString(node.id);

        if (nodeName && nodeId) {
          idToNameMap.set(nodeId.toLowerCase(), nodeName);
        }

        const nodeRegion = currentRegion;
        const nodeDistrict = currentDistrict;

        // If top root (Level 1 e.g. "MOH - Uganda"), its children are Level 2 Regions
        if (!currentRegion) {
          if (node.children && Array.isArray(node.children) && node.children.length > 0) {
            for (const regionNode of node.children) {
              if (!regionNode || typeof regionNode !== "object") continue;

              const regName = safeString(regionNode.name);
              const regId = safeString(regionNode.id);

              if (regName) {
                regionMap.set(regName.toLowerCase(), regName);
                if (regId) {
                  regionMap.set(regId.toLowerCase(), regName);
                  idToNameMap.set(regId.toLowerCase(), regName);
                }
              }

              // Children of region are Level 3 Districts
              if (
                regionNode.children &&
                Array.isArray(regionNode.children) &&
                regionNode.children.length > 0
              ) {
                for (const distNode of regionNode.children) {
                  if (!distNode || typeof distNode !== "object") continue;

                  const distName = safeString(distNode.name);
                  const distId = safeString(distNode.id);

                  if (distName) {
                    districtMap.set(distName.toLowerCase(), distName);
                    if (regName) regionMap.set(distName.toLowerCase(), regName);
                    if (distId) {
                      districtMap.set(distId.toLowerCase(), distName);
                      if (regName) regionMap.set(distId.toLowerCase(), regName);
                      idToNameMap.set(distId.toLowerCase(), distName);
                    }
                  }
                  // Any facility or sub-county under district
                  if (
                    distNode.children &&
                    Array.isArray(distNode.children) &&
                    distNode.children.length > 0
                  ) {
                    traverseTree(distNode.children, regName, distName);
                  }
                }
              }
            }
            continue;
          }
        }

        if (nodeName) {
          if (nodeRegion) regionMap.set(nodeName.toLowerCase(), nodeRegion);
          if (nodeDistrict) districtMap.set(nodeName.toLowerCase(), nodeDistrict);
        }
        if (nodeId) {
          if (nodeRegion) regionMap.set(nodeId.toLowerCase(), nodeRegion);
          if (nodeDistrict) districtMap.set(nodeId.toLowerCase(), nodeDistrict);
        }

        if (node.children && Array.isArray(node.children) && node.children.length > 0) {
          traverseTree(node.children, nodeRegion, nodeDistrict);
        }
      }
    }

    traverseTree(rootList, null, null);
    return {
      orgToRegionMap: regionMap,
      orgToDistrictMap: districtMap,
      orgIdToNameMap: idToNameMap,
    };
  }, [hierarchyData]);

  /* -------------------------------------------------------------
   * Page-Level Issues & Aggregation (Unfiltered by Table Filters)
   * ------------------------------------------------------------- */
  const rawIssues: Issue[] = useMemo(() => {
    return issuesData?.items ?? [];
  }, [issuesData]);

  const {
    regionalSummary: pageRegionalSummary,
    grandTotals: pageGrandTotals,
  } = useMemo(
    () =>
      aggregateIssuesByRegion(
        rawIssues,
        orgToRegionMap,
        orgToDistrictMap,
        orgIdToNameMap,
      ),
    [rawIssues, orgToRegionMap, orgToDistrictMap, orgIdToNameMap],
  );

  /* -------------------------------------------------------------
   * Table-Level Filtered Issues & Aggregation
   * ------------------------------------------------------------- */
  const tableBaseIssues: Issue[] = useMemo(() => {
    if (selectedProgram) {
      return programIssuesData?.items ?? [];
    }
    return rawIssues;
  }, [selectedProgram, programIssuesData, rawIssues]);

  const filteredIssues = useMemo(() => {
    const isYearActive = isYearFilterApplied || selectedYear !== currentYear;

    return tableBaseIssues.filter((issue) => {
      // Dataset filter
      if (selectedDataset) {
        if (safeLower(issue.dataset) !== safeLower(selectedDataset)) {
          return false;
        }
      }

      // Period filter
      if (
        !doesIssueMatchReportingPeriod(
          issue,
          selectedPeriod,
          selectedYear,
          isYearActive,
        )
      ) {
        return false;
      }

      return true;
    });
  }, [
    tableBaseIssues,
    selectedDataset,
    selectedPeriod,
    selectedYear,
    currentYear,
    isYearFilterApplied,
  ]);

  const { regionalSummary, grandTotals: filteredGrandTotals } = useMemo(
    () =>
      aggregateIssuesByRegion(
        filteredIssues,
        orgToRegionMap,
        orgToDistrictMap,
        orgIdToNameMap,
      ),
    [filteredIssues, orgToRegionMap, orgToDistrictMap, orgIdToNameMap],
  );

  /* -------------------------------------------------------------
   * Table Columns Definition
   * ------------------------------------------------------------- */
  const tableHeaders = useMemo(
    () => [
      { key: "region", header: "Region" },
      { key: "critical", header: "Critical Issues" },
      { key: "open", header: "Open Issues" },
      { key: "resolved", header: "Resolved Issues" },
      { key: "total", header: "Total Issues" },
      { key: "percentage", header: "Share of Total" },
    ],
    [],
  );

  /* -------------------------------------------------------------
   * Sorted & Filtered Rows for Table Display
   * ------------------------------------------------------------- */
  const displayedRows = useMemo(() => {
    let result = [...regionalSummary];

    // Filter by Region, District, or Facility search
    if (regionSearchTerm.trim()) {
      const term = safeLower(regionSearchTerm);
      result = result.filter(
        (r) =>
          safeLower(r.region).includes(term) ||
          r.districts.some(
            (d) =>
              safeLower(d.district).includes(term) ||
              d.facilities.some((f) => safeLower(f.facility).includes(term)),
          ),
      );
    }

    // Sort
    result.sort((a, b) => {
      let aVal = a[sortHeader];
      let bVal = b[sortHeader];

      if (typeof aVal === "string") {
        aVal = (aVal as string).toLowerCase();
        bVal = String(bVal ?? "").toLowerCase();
        if (sortDirection === "ASC") {
          return (aVal as string).localeCompare(bVal as string);
        }
        return (bVal as string).localeCompare(aVal as string);
      }

      const aNum = Number(aVal ?? 0);
      const bNum = Number(bVal ?? 0);
      if (sortDirection === "ASC") {
        return aNum - bNum;
      }
      return bNum - aNum;
    });

    return result;
  }, [regionalSummary, regionSearchTerm, sortHeader, sortDirection]);

  const tableGrandTotals = useMemo(() => {
    if (!regionSearchTerm.trim()) {
      return filteredGrandTotals;
    }
    return displayedRows.reduce(
      (acc, row) => ({
        critical: acc.critical + row.critical,
        open: acc.open + row.open,
        resolved: acc.resolved + row.resolved,
        total: acc.total + row.total,
        activeRegionsCount: displayedRows.length,
      }),
      {
        critical: 0,
        open: 0,
        resolved: 0,
        total: 0,
        activeRegionsCount: 0,
      },
    );
  }, [regionSearchTerm, filteredGrandTotals, displayedRows]);

  /* -------------------------------------------------------------
   * Column Charts Data & Options (Page-Level, Independent of Table Filters)
   * ------------------------------------------------------------- */
  const defaultRegionLabel = useMemo(() => {
    const match = pageRegionalSummary.find((r) =>
      safeLower(r.region).includes("south buganda"),
    );
    return match ? match.region : DEFAULT_CHART_REGION;
  }, [pageRegionalSummary]);

  const chartRegionOptions = useMemo(() => {
    const names = pageRegionalSummary.map((r) => r.region);
    const hasDefaultRegion = names.some(
      (name) => safeLower(name) === safeLower(defaultRegionLabel),
    );
    return [
      "All Regions",
      ...(hasDefaultRegion ? [] : [defaultRegionLabel]),
      ...names,
    ];
  }, [pageRegionalSummary, defaultRegionLabel]);

  const activeChartRegion = useMemo(() => {
    const exactMatch = chartRegionOptions.find(
      (opt) => safeLower(opt) === safeLower(selectedChartRegion),
    );
    if (exactMatch) return exactMatch;
    if (safeLower(selectedChartRegion) === safeLower(DEFAULT_CHART_REGION)) {
      return defaultRegionLabel;
    }
    return defaultRegionLabel;
  }, [chartRegionOptions, selectedChartRegion, defaultRegionLabel]);

  const chartToolbarConfig = useMemo(
    () => ({
      enabled: true,
      numberOfIcons: 3,
      controls: [
        { type: ToolbarControlTypes.MAKE_FULLSCREEN },
        { type: ToolbarControlTypes.SHOW_AS_DATATABLE },
        { type: ToolbarControlTypes.EXPORT_CSV },
        { type: ToolbarControlTypes.EXPORT_PNG },
        { type: ToolbarControlTypes.EXPORT_JPG },
      ],
    }),
    [],
  );

  const regionChartData = useMemo(() => {
    const sortedRegions = [...pageRegionalSummary].sort((a, b) => b.total - a.total);
    if (chartViewMode === "status") {
      return sortedRegions.flatMap((r) => [
        { group: "Critical Issues", key: r.region, value: r.critical },
        { group: "Open Issues", key: r.region, value: r.open },
        { group: "Resolved Issues", key: r.region, value: r.resolved },
      ]);
    }
    return sortedRegions.map((r) => ({
      group: "Total Issues",
      key: r.region,
      value: r.total,
    }));
  }, [pageRegionalSummary, chartViewMode]);

  const districtChartData = useMemo(() => {
    const districtMap = new Map<
      string,
      { district: string; critical: number; open: number; resolved: number; total: number }
    >();

    for (const regionRow of pageRegionalSummary) {
      if (
        activeChartRegion !== "All Regions" &&
        safeLower(regionRow.region) !== safeLower(activeChartRegion) &&
        !(
          safeLower(activeChartRegion) === safeLower(DEFAULT_CHART_REGION) &&
          safeLower(regionRow.region).includes("south buganda")
        )
      ) {
        continue;
      }

      for (const d of regionRow.districts) {
        const existing = districtMap.get(d.district);
        if (existing) {
          existing.critical += d.critical;
          existing.open += d.open;
          existing.resolved += d.resolved;
          existing.total += d.total;
        } else {
          districtMap.set(d.district, {
            district: d.district,
            critical: d.critical,
            open: d.open,
            resolved: d.resolved,
            total: d.total,
          });
        }
      }
    }

    const sortedDistricts = Array.from(districtMap.values()).sort(
      (a, b) => b.total - a.total,
    );

    if (chartViewMode === "status") {
      return sortedDistricts.flatMap((d) => [
        { group: "Critical Issues", key: d.district, value: d.critical },
        { group: "Open Issues", key: d.district, value: d.open },
        { group: "Resolved Issues", key: d.district, value: d.resolved },
      ]);
    }

    return sortedDistricts.map((d) => ({
      group: "Total Issues",
      key: d.district,
      value: d.total,
    }));
  }, [pageRegionalSummary, activeChartRegion, chartViewMode]);

  const regionChartOptions = useMemo(
    () => ({
      title: "Issues by Region",
      axes: {
        left: {
          mapsTo: "value",
          title: "Number of Issues",
          scaleType: ScaleTypes.LINEAR,
        },
        bottom: {
          mapsTo: "key",
          title: "Region",
          scaleType: ScaleTypes.LABELS,
        },
      },
      color: {
        scale: {
          "Total Issues": "#0f62fe",
          "Critical Issues": "#a2191f",
          "Open Issues": "#da1e28",
          "Resolved Issues": "#198038",
        },
      },
      height: "480px",
      toolbar: chartToolbarConfig,
      legend: {
        enabled: chartViewMode === "status",
      },
    }),
    [chartViewMode, chartToolbarConfig],
  );

  const districtChartOptions = useMemo(
    () => ({
      title:
        activeChartRegion && activeChartRegion !== "All Regions"
          ? `Issues by District (${activeChartRegion})`
          : "Issues by District",
      axes: {
        left: {
          mapsTo: "value",
          title: "Number of Issues",
          scaleType: ScaleTypes.LINEAR,
        },
        bottom: {
          mapsTo: "key",
          title: "District",
          scaleType: ScaleTypes.LABELS,
        },
      },
      color: {
        scale: {
          "Total Issues": "#007d79",
          "Critical Issues": "#a2191f",
          "Open Issues": "#da1e28",
          "Resolved Issues": "#198038",
        },
      },
      height: "480px",
      toolbar: chartToolbarConfig,
      legend: {
        enabled: chartViewMode === "status",
      },
    }),
    [activeChartRegion, chartViewMode, chartToolbarConfig],
  );

  const levelOfCareChartData = useMemo(() => {
    const careMap = new Map<
      string,
      { levelOfCare: string; critical: number; open: number; resolved: number; total: number }
    >();

    const hasActiveRegionInRows =
      activeChartRegion === "All Regions" ||
      pageRegionalSummary.some(
        (r) =>
          safeLower(r.region) === safeLower(activeChartRegion) ||
          (safeLower(activeChartRegion) === safeLower(DEFAULT_CHART_REGION) &&
            safeLower(r.region).includes("south buganda")),
      );

    for (const regionRow of pageRegionalSummary) {
      if (
        hasActiveRegionInRows &&
        activeChartRegion !== "All Regions" &&
        safeLower(regionRow.region) !== safeLower(activeChartRegion) &&
        !(
          safeLower(activeChartRegion) === safeLower(DEFAULT_CHART_REGION) &&
          safeLower(regionRow.region).includes("south buganda")
        )
      ) {
        continue;
      }

      for (const d of regionRow.districts) {
        for (const f of d.facilities) {
          const careLevel = f.levelOfCare || "Other / Unspecified";
          const existing = careMap.get(careLevel);
          if (existing) {
            existing.critical += f.critical;
            existing.open += f.open;
            existing.resolved += f.resolved;
            existing.total += f.total;
          } else {
            careMap.set(careLevel, {
              levelOfCare: careLevel,
              critical: f.critical,
              open: f.open,
              resolved: f.resolved,
              total: f.total,
            });
          }
        }
      }
    }

    const sortedCareLevels = Array.from(careMap.values()).sort(
      (a, b) => b.total - a.total,
    );

    return sortedCareLevels.map((c) => ({
      group: c.levelOfCare,
      value: c.total,
    }));
  }, [pageRegionalSummary, activeChartRegion]);

  const levelOfCareChartOptions = useMemo(
    () => ({
      title:
        activeChartRegion && activeChartRegion !== "All Regions"
          ? `Issues by Level of Care (${activeChartRegion})`
          : "Issues by Level of Care",
      resizable: true,
      pie: {
        alignment: Alignments.CENTER,
      },
      height: "480px",
      toolbar: chartToolbarConfig,
      legend: {
        enabled: true,
        alignment: Alignments.CENTER,
      },
    }),
    [activeChartRegion, chartToolbarConfig],
  );

  /* -------------------------------------------------------------
   * Unresolved Issue Aging Data, Chart Options & Actionable Queue
   * ------------------------------------------------------------- */
  const unresolvedAgingIssues: ActionableAgingIssue[] = useMemo(() => {
    const nowMs = Date.now();
    const dayMs = 1000 * 60 * 60 * 24;

    const hasActiveRegionInRows =
      activeChartRegion === "All Regions" ||
      pageRegionalSummary.some(
        (r) =>
          safeLower(r.region) === safeLower(activeChartRegion) ||
          (safeLower(activeChartRegion) === safeLower(DEFAULT_CHART_REGION) &&
            safeLower(r.region).includes("south buganda")),
      );

    const items: ActionableAgingIssue[] = [];

    for (const issue of rawIssues) {
      if (isIssueResolved(issue.status)) {
        continue;
      }

      // Resolve Region
      let regionName = safeString(issue.region);
      if (regionName && orgIdToNameMap.has(safeLower(regionName))) {
        regionName = orgIdToNameMap.get(safeLower(regionName))!;
      }
      if (!regionName && issue.org_unit) {
        regionName = orgToRegionMap.get(safeLower(issue.org_unit)) || "";
      }
      if (!regionName && issue.district) {
        regionName = orgToRegionMap.get(safeLower(issue.district)) || "";
      }
      if (!regionName) {
        regionName = "Unspecified";
      }

      if (
        hasActiveRegionInRows &&
        activeChartRegion !== "All Regions" &&
        safeLower(regionName) !== safeLower(activeChartRegion) &&
        !(
          safeLower(activeChartRegion) === safeLower(DEFAULT_CHART_REGION) &&
          safeLower(regionName).includes("south buganda")
        )
      ) {
        continue;
      }

      // Resolve District
      let districtName = safeString(issue.district);
      if (districtName && orgIdToNameMap.has(safeLower(districtName))) {
        districtName = orgIdToNameMap.get(safeLower(districtName))!;
      }
      if (!districtName && issue.org_unit) {
        districtName =
          orgToDistrictMap.get(safeLower(issue.org_unit)) || safeString(issue.org_unit);
      }
      if (!districtName) {
        districtName = "Unspecified";
      }

      // Resolve Facility
      let facilityName = safeString(issue.org_unit);
      if (!facilityName) {
        facilityName = "Unspecified Facility";
      } else {
        const mappedName = orgIdToNameMap.get(safeLower(facilityName));
        if (mappedName) {
          facilityName = mappedName;
        }
      }

      const parsedDate = parseIssueDate(issue);
      const ageDays = parsedDate
        ? Math.max(0, Math.floor((nowMs - parsedDate.getTime()) / dayMs))
        : 0;
      const bucket = getAgingBucketForDays(ageDays);
      const isCritical = isCriticalIssue(issue);
      const isUnassigned = !safeString(issue.assigned_to);
      const issueCode =
        safeString(issue.issue_code) ||
        (issue.issue_id ? `HMIS-${String(issue.issue_id).padStart(4, "0")}` : "ISSUE");
      const reportedDateLabel =
        safeString(issue.date_reported).slice(0, 10) ||
        formatReportingPeriod(issue.time_Period || issue.time_period) ||
        "N/A";

      items.push({
        issue,
        issueCode,
        ageDays,
        bucket,
        regionName,
        districtName,
        facilityName,
        isCritical,
        isUnassigned,
        reportedDateLabel,
      });
    }

    // Sort actionable backlog: critical + oldest first
    items.sort((a, b) => {
      if (b.ageDays !== a.ageDays) return b.ageDays - a.ageDays;
      if (a.isCritical !== b.isCritical) return a.isCritical ? -1 : 1;
      return a.isUnassigned === b.isUnassigned ? 0 : a.isUnassigned ? -1 : 1;
    });

    return items;
  }, [
    rawIssues,
    activeChartRegion,
    pageRegionalSummary,
    orgToRegionMap,
    orgToDistrictMap,
    orgIdToNameMap,
  ]);

  const agingMetrics = useMemo(() => {
    const bucketMap = new Map<
      AgingBucketId,
      { total: number; critical: number; unassigned: number; assigned: number }
    >();

    for (const b of AGING_BUCKETS) {
      bucketMap.set(b.id, { total: 0, critical: 0, unassigned: 0, assigned: 0 });
    }

    let overdueCount = 0;
    let criticalCount = 0;
    let unassignedCount = 0;
    let totalAgeDays = 0;
    let maxAgeDays = 0;

    for (const item of unresolvedAgingIssues) {
      const bStat = bucketMap.get(item.bucket.id)!;
      bStat.total += 1;
      if (item.isCritical) {
        bStat.critical += 1;
        criticalCount += 1;
      }
      if (item.isUnassigned) {
        bStat.unassigned += 1;
        unassignedCount += 1;
      } else {
        bStat.assigned += 1;
      }
      if (item.ageDays > 30) {
        overdueCount += 1;
      }
      totalAgeDays += item.ageDays;
      if (item.ageDays > maxAgeDays) {
        maxAgeDays = item.ageDays;
      }
    }

    const totalUnresolved = unresolvedAgingIssues.length;
    const avgAgeDays = totalUnresolved > 0 ? Math.round(totalAgeDays / totalUnresolved) : 0;

    return {
      totalUnresolved,
      overdueCount,
      criticalCount,
      unassignedCount,
      avgAgeDays,
      maxAgeDays,
      bucketMap,
    };
  }, [unresolvedAgingIssues]);

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
          { group: "Critical Open", key: bucket.shortLabel, value: stats.critical },
          { group: "Unassigned Open", key: bucket.shortLabel, value: stats.unassigned },
          { group: "Assigned Open", key: bucket.shortLabel, value: stats.assigned },
        ];
      });
    }

    return AGING_BUCKETS.map((bucket) => {
      const stats = agingMetrics.bucketMap.get(bucket.id);
      return {
        group: bucket.label,
        key: bucket.shortLabel,
        value: stats?.total ?? 0,
      };
    });
  }, [agingMetrics, chartViewMode]);

  const agingChartOptions = useMemo(
    () => ({
      title:
        activeChartRegion && activeChartRegion !== "All Regions"
          ? `Unresolved Issue Aging (${activeChartRegion})`
          : "Unresolved Issue Aging",
      axes: {
        left: {
          mapsTo: "value",
          title: "Unresolved",
          scaleType: ScaleTypes.LINEAR,
        },
        bottom: {
          mapsTo: "key",
          title: "Age Bracket",
          scaleType: ScaleTypes.LABELS,
        },
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
                "0–7 Days": "#0f62fe",
                "8–14 Days": "#007d79",
                "15–30 Days": "#d2a106",
                "31–60 Days": "#eb6200",
                "61–90 Days": "#da1e28",
                "> 90 Days": "#a2191f",
              },
      },
      height: "215px",
      toolbar: chartToolbarConfig,
      legend: {
        enabled: chartViewMode === "status",
      },
    }),
    [activeChartRegion, chartViewMode, chartToolbarConfig],
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

  /* -------------------------------------------------------------
   * Handlers
   * ------------------------------------------------------------- */
  const handleSort = (headerKey: string) => {
    if (sortHeader === headerKey) {
      setSortDirection((prev) => (prev === "ASC" ? "DESC" : "ASC"));
    } else {
      setSortHeader(headerKey);
      setSortDirection("DESC");
    }
  };

  const handleYearChange = ({ selectedItem }: SelectEvent<number>) => {
    if (selectedItem == null) return;
    setSelectedYear(selectedItem);
    setIsYearFilterApplied(true);
    setAvailablePeriods(fetchAvailablePeriods(selectedPeriodType, selectedItem.toString()));
    setSelectedPeriod("");
  };

  const handlePeriodTypeChange = ({
    selectedItem,
  }: SelectEvent<{ label: string; value: string }>) => {
    if (!selectedItem) return;
    setSelectedPeriodType(selectedItem.value);
    setAvailablePeriods(fetchAvailablePeriods(selectedItem.value, selectedYear.toString()));
    setSelectedPeriod("");
  };

  const handlePeriodChange = ({ selectedItem }: SelectEvent<PeriodOption>) => {
    setSelectedPeriod(selectedItem?.label ?? "");
  };

  const handleResetFilters = () => {
    setSelectedYear(currentYear);
    setIsYearFilterApplied(false);
    setSelectedPeriodType("Monthly");
    setSelectedPeriod("");
    setAvailablePeriods(fetchAvailablePeriods("Monthly", currentYear.toString()));
    setSelectedDataset("");
    setSelectedProgram(undefined);
    setRegionSearchTerm("");
    setIsPeriodPopoverOpen(false);
  };

  /* -------------------------------------------------------------
   * Excel Export
   * ------------------------------------------------------------- */
  const handleExportExcel = () => {
    const exportHeaders = [
      "Region",
      "Critical Issues (Open)",
      "Open Issues",
      "Resolved Issues",
      "Total Issues",
      "Share (%)",
    ];

    const dataRows = displayedRows.map((row) => [
      row.region,
      row.critical,
      row.open,
      row.resolved,
      row.total,
      `${row.percentage.toFixed(1)}%`,
    ]);

    // Grand total row
    const totalRow = [
      "TOTAL",
      tableGrandTotals.critical,
      tableGrandTotals.open,
      tableGrandTotals.resolved,
      tableGrandTotals.total,
      "100%",
    ];

    const sheetData = [exportHeaders, ...dataRows, [], totalRow];
    const worksheet = XLSX.utils.aoa_to_sheet(sheetData);

    worksheet["!cols"] = exportHeaders.map(() => ({ wch: 18 }));

    const workbook = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(workbook, worksheet, "Regional Issue Summary");

    // District Breakdown Sheet
    const districtHeaders = [
      "Region",
      "District",
      "Facilities Affected",
      "Critical Issues (Open)",
      "Open Issues",
      "Resolved Issues",
      "Total Issues",
    ];

    const districtDataRows: (string | number)[][] = [];
    for (const r of displayedRows) {
      for (const d of r.districts) {
        districtDataRows.push([
          r.region,
          d.district,
          d.facilities.length,
          d.critical,
          d.open,
          d.resolved,
          d.total,
        ]);
      }
    }

    const districtWorksheet = XLSX.utils.aoa_to_sheet([districtHeaders, ...districtDataRows]);
    districtWorksheet["!cols"] = districtHeaders.map(() => ({ wch: 22 }));
    XLSX.utils.book_append_sheet(workbook, districtWorksheet, "District Breakdown");

    // District & Facility Breakdown Sheet
    const facilityHeaders = [
      "Region",
      "District",
      "Health Facility / Org Unit",
      "Critical Issues (Open)",
      "Open Issues",
      "Resolved Issues",
      "Total Issues",
    ];

    const facilityDataRows: (string | number)[][] = [];
    for (const r of displayedRows) {
      for (const d of r.districts) {
        for (const f of d.facilities) {
          facilityDataRows.push([
            r.region,
            d.district,
            f.facility,
            f.critical,
            f.open,
            f.resolved,
            f.total,
          ]);
        }
      }
    }

    const facilityWorksheet = XLSX.utils.aoa_to_sheet([facilityHeaders, ...facilityDataRows]);
    facilityWorksheet["!cols"] = facilityHeaders.map(() => ({ wch: 24 }));
    XLSX.utils.book_append_sheet(workbook, facilityWorksheet, "Facility Breakdown");

    const dateStr = new Date().toISOString().slice(0, 10);
    XLSX.writeFile(workbook, `issue-regional-summary-${dateStr}.xlsx`);
  };

  const handleNavigateToIssues = (filter?: string) => {
    if (filter && !filter.toLowerCase().startsWith("unspecified")) {
      navigate(`/issues?region=${encodeURIComponent(filter)}`);
    } else {
      navigate("/issues");
    }
  };

  const handleNavigateToIssueCode = (issueCode: string) => {
    if (issueCode) {
      navigate(`/issues?issueCode=${encodeURIComponent(issueCode)}`);
    } else {
      navigate("/issues");
    }
  };

  const handleOpenAssignSingle = (item: ActionableAgingIssue) => {
    const code = safeString(item.issue.issue_code) || item.issueCode;
    if (!code) return;
    setAssigningIssueCodes([code]);
    setIsAssignReassign(!item.isUnassigned);
  };

  const handleOpenBulkAssignUnassigned = () => {
    if (actionableUnassignedCodes.length === 0) return;
    setAssigningIssueCodes(actionableUnassignedCodes);
    setIsAssignReassign(false);
  };

  const handleConfirmQuickResolve = async () => {
    if (!resolvingAgingIssue || !resolutionComment.trim()) return;
    const code =
      safeString(resolvingAgingIssue.issue.issue_code) || resolvingAgingIssue.issueCode;
    if (!code) return;

    try {
      await createTransaction({
        id: code,
        body: {
          resolution_action: resolutionComment.trim(),
          resolved_by: user?.email || user?.username || "",
          status: "resolved",
        },
      }).unwrap();

      setResolvingAgingIssue(null);
      setResolutionComment("");
      handleRefreshAll();
    } catch (err) {
      console.error("Failed to resolve aging issue:", err);
    }
  };

  const handleRefreshAll = () => {
    refetchIssues();
    if (selectedProgram) {
      refetchProgramIssues();
    }
  };

  /* Page-level resolution rate & table loading state */
  const resolutionRate =
    pageGrandTotals.total > 0
      ? Math.round((pageGrandTotals.resolved / pageGrandTotals.total) * 100)
      : 0;
  const isTableLoading =
    isLoadingIssues || (Boolean(selectedProgram) && isFetchingProgramIssues);

  return (
    <PermissionGuard permission={PERMISSIONS.issueTrackerRead}>
      <div className="issue-dashboard-page">
        {/* ================= Header ================= */}
        <div className="dashboard-header">
          <div className="dashboard-header-text">
            <h4 className="dashboard-title">Issue Summary</h4>
            <p className="dashboard-subtitle">
              Aggregated overview of reported data quality issues categorized by resolution status
              and grouped by region.
            </p>
          </div>

          <div className="dashboard-actions">
            <Button
              size="sm"
              kind="ghost"
              renderIcon={Renew}
              onClick={handleRefreshAll}
              disabled={isLoadingIssues}
            >
              Refresh
            </Button>

            <Button
              size="sm"
              kind="secondary"
              renderIcon={Download}
              onClick={handleExportExcel}
              disabled={displayedRows.length === 0 || isTableLoading}
            >
              Export Summary (.xlsx)
            </Button>
          </div>
        </div>

        {/* ================= KPI Metric Cards ================= */}
        <div className="kpi-grid">
          <div className="kpi-card kpi-card--total">
            <div className="kpi-header">
              <span className="kpi-label">Total Issues</span>
            </div>
            <div className="kpi-value">{isLoadingIssues ? "..." : pageGrandTotals.total}</div>
            <p className="kpi-subtext">
              Across <strong>{pageGrandTotals.activeRegionsCount}</strong> affected regions
            </p>
          </div>

          <div className="kpi-card kpi-card--critical">
            <div className="kpi-header">
              <span className="kpi-label">Critical Issues</span>
              <span className="kpi-badge kpi-badge--critical">Critical</span>
            </div>
            <div className="kpi-value" style={{ color: "#a2191f" }}>
              {isLoadingIssues ? "..." : pageGrandTotals.critical}
            </div>
            <p className="kpi-subtext">
              <strong>
                {pageGrandTotals.open > 0
                  ? `${((pageGrandTotals.critical / pageGrandTotals.open) * 100).toFixed(1)}%`
                  : "0%"}
              </strong>{" "}
              of open issues
            </p>
          </div>

          <div className="kpi-card kpi-card--open">
            <div className="kpi-header">
              <span className="kpi-label">Open Issues</span>
              <span className="kpi-badge kpi-badge--open">Open</span>
            </div>
            <div className="kpi-value" style={{ color: "#da1e28" }}>
              {isLoadingIssues ? "..." : pageGrandTotals.open}
            </div>
            <p className="kpi-subtext">
              <strong>
                {pageGrandTotals.total > 0
                  ? `${((pageGrandTotals.open / pageGrandTotals.total) * 100).toFixed(1)}%`
                  : "0%"}
              </strong>{" "}
              of total issues
            </p>
          </div>

          <div className="kpi-card kpi-card--resolved">
            <div className="kpi-header">
              <span className="kpi-label">Resolved Issues</span>
              <span className="kpi-badge kpi-badge--resolved">Resolved</span>
            </div>
            <div className="kpi-value" style={{ color: "#198038" }}>
              {isLoadingIssues ? "..." : pageGrandTotals.resolved}
            </div>
            <p className="kpi-subtext">
              <strong>
                {pageGrandTotals.total > 0
                  ? `${((pageGrandTotals.resolved / pageGrandTotals.total) * 100).toFixed(1)}%`
                  : "0%"}
              </strong>{" "}
              of total issues
            </p>
          </div>

          <div className="kpi-card kpi-card--rate">
            <div className="kpi-header">
              <span className="kpi-label">Resolution Rate</span>
              <span className="kpi-badge kpi-badge--rate">{resolutionRate}%</span>
            </div>
            <div className="kpi-value" style={{ color: "#0f62fe" }}>
              {isLoadingIssues ? "..." : `${resolutionRate}%`}
            </div>
            <p className="kpi-subtext">
              Resolved vs <strong>{pageGrandTotals.open}</strong> open issues
            </p>
          </div>
        </div>

        {/* ================= Column Charts Section ================= */}
        {!isLoadingIssues && !issuesError && pageRegionalSummary.length > 0 && (
          <div className="dashboard-charts-section">
            <div className="dashboard-charts-header">
              <div>
                <h3 className="dashboard-charts-title">Visual Issue Distribution</h3>
                <p className="dashboard-charts-subtitle">
                  Comparison of reported issues across regions, districts, level of care, and
                  actionable unresolved issue aging.
                </p>
              </div>

              <div className="dashboard-charts-controls">
                <div className="chart-region-filter">
                  <Dropdown
                    id="chart-region-filter-dropdown"
                    size="sm"
                    titleText=""
                    label={DEFAULT_CHART_REGION}
                    items={chartRegionOptions}
                    selectedItem={activeChartRegion}
                    itemToString={(item) => (item ? String(item) : "")}
                    onChange={({ selectedItem }: SelectEvent<string>) =>
                      setSelectedChartRegion(selectedItem || DEFAULT_CHART_REGION)
                    }
                  />
                </div>

                <div className="chart-mode-toggle" role="group" aria-label="Chart metric mode">
                  <Button
                    size="sm"
                    kind={chartViewMode === "total" ? "primary" : "tertiary"}
                    onClick={() => setChartViewMode("total")}
                  >
                    Total Issues
                  </Button>
                  <Button
                    size="sm"
                    kind={chartViewMode === "status" ? "primary" : "tertiary"}
                    onClick={() => setChartViewMode("status")}
                  >
                    By Status
                  </Button>
                </div>
              </div>
            </div>

            <div className="dashboard-charts-grid">
              <div className="dashboard-chart-card">
                {chartViewMode === "status" ? (
                  <GroupedBarChart data={regionChartData} options={regionChartOptions} />
                ) : (
                  <SimpleBarChart data={regionChartData} options={regionChartOptions} />
                )}
              </div>

              <div className="dashboard-chart-card">
                {chartViewMode === "status" ? (
                  <GroupedBarChart data={districtChartData} options={districtChartOptions} />
                ) : (
                  <SimpleBarChart data={districtChartData} options={districtChartOptions} />
                )}
              </div>

              <div className="dashboard-chart-card">
                <PieChart
                  data={levelOfCareChartData}
                  options={levelOfCareChartOptions}
                />
              </div>

              <div className="dashboard-chart-card aging-visual-card">
                <div className="aging-visual-chart-wrapper">
                  {chartViewMode === "status" ? (
                    <GroupedBarChart data={agingChartData} options={agingChartOptions} />
                  ) : (
                    <SimpleBarChart data={agingChartData} options={agingChartOptions} />
                  )}
                </div>

                {/* Summary KPI Strip */}
                <div className="aging-kpi-strip">
                  <div className="aging-kpi-item">
                    <span className="aging-kpi-label">Unresolved</span>
                    <strong className="aging-kpi-val">{agingMetrics.totalUnresolved}</strong>
                  </div>
                  <div className="aging-kpi-item aging-kpi-item--overdue">
                    <span className="aging-kpi-label">Overdue (&gt;30d)</span>
                    <strong className="aging-kpi-val">{agingMetrics.overdueCount}</strong>
                  </div>
                  <div className="aging-kpi-item aging-kpi-item--critical">
                    <span className="aging-kpi-label">Critical</span>
                    <strong className="aging-kpi-val">{agingMetrics.criticalCount}</strong>
                  </div>
                  <div className="aging-kpi-item aging-kpi-item--unassigned">
                    <span className="aging-kpi-label">Unassigned</span>
                    <strong className="aging-kpi-val">{agingMetrics.unassignedCount}</strong>
                  </div>
                  <div className="aging-kpi-item">
                    <span className="aging-kpi-label">Avg / Oldest</span>
                    <strong className="aging-kpi-val">
                      {agingMetrics.avgAgeDays}d / {agingMetrics.maxAgeDays}d
                    </strong>
                  </div>
                </div>

                {/* Interactive Bucket & Triage Filters */}
                <div className="aging-controls-bar">
                  <div className="aging-bucket-pills" role="group" aria-label="Filter by aging bracket">
                    <button
                      type="button"
                      className={`aging-bucket-chip ${selectedAgingBucket === "all" ? "active" : ""}`}
                      onClick={() => setSelectedAgingBucket("all")}
                    >
                      All ({agingMetrics.totalUnresolved})
                    </button>
                    {AGING_BUCKETS.map((bucket) => {
                      const count = agingMetrics.bucketMap.get(bucket.id)?.total ?? 0;
                      const isActive = selectedAgingBucket === bucket.id;
                      return (
                        <button
                          key={bucket.id}
                          type="button"
                          className={`aging-bucket-chip ${isActive ? "active" : ""}`}
                          style={
                            isActive
                              ? {
                                  backgroundColor: bucket.color,
                                  borderColor: bucket.color,
                                  color: "#ffffff",
                                }
                              : undefined
                          }
                          onClick={() =>
                            setSelectedAgingBucket((prev) =>
                              prev === bucket.id ? "all" : bucket.id,
                            )
                          }
                        >
                          <span
                            className="aging-bucket-dot"
                            style={{
                              backgroundColor: isActive ? "#ffffff" : bucket.color,
                            }}
                          />
                          {bucket.shortLabel} ({count})
                        </button>
                      );
                    })}
                  </div>

                  <div className="aging-triage-row">
                    <div className="aging-triage-filters" role="group" aria-label="Quick triage filters">
                      <button
                        type="button"
                        className={`aging-triage-btn ${agingTriageFilter === "all" ? "active" : ""}`}
                        onClick={() => setAgingTriageFilter("all")}
                      >
                        All Backlog
                      </button>
                      <button
                        type="button"
                        className={`aging-triage-btn ${agingTriageFilter === "overdue" ? "active" : ""}`}
                        onClick={() =>
                          setAgingTriageFilter((prev) => (prev === "overdue" ? "all" : "overdue"))
                        }
                      >
                        Overdue &gt;30d ({agingMetrics.overdueCount})
                      </button>
                      <button
                        type="button"
                        className={`aging-triage-btn ${agingTriageFilter === "critical" ? "active" : ""}`}
                        onClick={() =>
                          setAgingTriageFilter((prev) =>
                            prev === "critical" ? "all" : "critical",
                          )
                        }
                      >
                        Critical ({agingMetrics.criticalCount})
                      </button>
                      <button
                        type="button"
                        className={`aging-triage-btn ${agingTriageFilter === "unassigned" ? "active" : ""}`}
                        onClick={() =>
                          setAgingTriageFilter((prev) =>
                            prev === "unassigned" ? "all" : "unassigned",
                          )
                        }
                      >
                        Unassigned ({agingMetrics.unassignedCount})
                      </button>
                    </div>

                    {actionableUnassignedCodes.length > 0 && (
                      <PermissionGuard permission={PERMISSIONS.issueTrackerAssign}>
                        <Button
                          size="sm"
                          kind="ghost"
                          renderIcon={UserFollow}
                          onClick={handleOpenBulkAssignUnassigned}
                          className="aging-bulk-assign-btn"
                        >
                          Assign {actionableUnassignedCodes.length} Unassigned
                        </Button>
                      </PermissionGuard>
                    )}
                  </div>
                </div>

                {/* Actionable Unresolved Issues List */}
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
                          }}
                        >
                          Clear Aging Filters
                        </Button>
                      )}
                    </div>
                  ) : (
                    <div className="aging-queue-table-wrap">
                      <table className="aging-queue-table" aria-label="Actionable aging unresolved issues">
                        <thead>
                          <tr>
                            <th>Issue &amp; Facility</th>
                            <th>Age</th>
                            <th>Assignee</th>
                            <th style={{ textAlign: "right" }}>Quick Actions</th>
                          </tr>
                        </thead>
                        <tbody>
                          {actionableAgingIssues.slice(0, 25).map((item) => (
                            <tr key={`${item.issue.issue_id}-${item.issueCode}`}>
                              <td>
                                <div className="aging-issue-main">
                                  <div className="aging-issue-top">
                                    <button
                                      type="button"
                                      className="aging-issue-code-link"
                                      onClick={() => handleNavigateToIssueCode(item.issueCode)}
                                      title={`Open issue ${item.issueCode}`}
                                    >
                                      {item.issueCode}
                                    </button>
                                    {item.isCritical && (
                                      <span className="aging-critical-tag">
                                        <WarningAlt size={12} /> Critical
                                      </span>
                                    )}
                                    <span className="aging-issue-element">
                                      {safeString(item.issue.data_element) ||
                                        safeString(item.issue.dataset) ||
                                        safeString(item.issue.issue_type) ||
                                        "Data Quality Issue"}
                                    </span>
                                  </div>
                                  <div className="aging-issue-meta">
                                    <span>{item.facilityName}</span>
                                    <span>•</span>
                                    <span>{item.districtName}</span>
                                    <span>•</span>
                                    <span>Reported: {item.reportedDateLabel}</span>
                                  </div>
                                </div>
                              </td>
                              <td>
                                <span className={`aging-pill ${item.bucket.badgeClass}`}>
                                  {item.ageDays}d
                                </span>
                              </td>
                              <td>
                                {item.isUnassigned ? (
                                  <span className="aging-assignee-unassigned">Unassigned</span>
                                ) : (
                                  <span
                                    className="aging-assignee-assigned"
                                    title={safeString(item.issue.assigned_to)}
                                  >
                                    {safeString(item.issue.assigned_to)}
                                  </span>
                                )}
                              </td>
                              <td>
                                <div className="aging-row-actions">
                                  <PermissionGuard permission={PERMISSIONS.issueTrackerAssign}>
                                    <button
                                      type="button"
                                      className="aging-action-btn aging-action-btn--assign"
                                      onClick={() => handleOpenAssignSingle(item)}
                                      title={
                                        item.isUnassigned
                                          ? `Assign ${item.issueCode}`
                                          : `Re-assign ${item.issueCode}`
                                      }
                                    >
                                      <UserFollow size={14} />
                                      <span>{item.isUnassigned ? "Assign" : "Re-assign"}</span>
                                    </button>
                                  </PermissionGuard>

                                  <PermissionGuard permission={PERMISSIONS.issueTrackerManage}>
                                    <button
                                      type="button"
                                      className="aging-action-btn aging-action-btn--resolve"
                                      onClick={() => {
                                        setResolvingAgingIssue(item);
                                        setResolutionComment("");
                                      }}
                                      title={`Resolve ${item.issueCode}`}
                                    >
                                      <CheckmarkOutline size={14} />
                                      <span>Resolve</span>
                                    </button>
                                  </PermissionGuard>

                                  <button
                                    type="button"
                                    className="aging-action-btn aging-action-btn--view"
                                    onClick={() => handleNavigateToIssueCode(item.issueCode)}
                                    title={`View full details for ${item.issueCode}`}
                                  >
                                    <ArrowRight size={14} />
                                  </button>
                                </div>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* ================= Filters Bar ================= */}
        <div className="dashboard-filter-container">
          {/* Period Filter Popover */}
          <div ref={periodPopoverRef} className="issue-filter-wrapper">
            <Popover open={isPeriodPopoverOpen} align="bottom-left" dropShadow>
              <Button
                size="sm"
                kind="tertiary"
                renderIcon={ChevronDown}
                onClick={() => setIsPeriodPopoverOpen((prev) => !prev)}
              >
                {selectedPeriod
                  ? `Period: ${selectedPeriod}`
                  : isYearFilterApplied || selectedYear !== currentYear
                  ? `Year: ${selectedYear}`
                  : "Reporting Period"}
              </Button>

              <PopoverContent className="filter-popover-content">
                <div className="popover-inner">
                  <Dropdown
                    id="dashboard-filter-year"
                    titleText="Year"
                    label="Select year"
                    items={years}
                    selectedItem={selectedYear}
                    itemToString={(item) => (item == null ? "" : String(item))}
                    onChange={handleYearChange}
                  />

                  <Dropdown
                    id="dashboard-filter-period-type"
                    titleText="Period Type"
                    label="Select period type"
                    items={periodType}
                    selectedItem={periodType.find((item) => item.value === selectedPeriodType)}
                    itemToString={(item) => item?.label ?? ""}
                    onChange={handlePeriodTypeChange}
                  />

                  <ComboBox
                    id="dashboard-filter-period"
                    titleText="Period"
                    placeholder="Search or select period..."
                    items={availablePeriods}
                    selectedItem={
                      availablePeriods.find((item) => item.label === selectedPeriod) ?? null
                    }
                    itemToString={(item) => item?.label ?? ""}
                    shouldFilterItem={({ item, inputValue }) => {
                      if (!inputValue) return true;
                      const query = inputValue.toLowerCase().trim();
                      const label = (item?.label ?? "").toLowerCase();
                      const id = (item?.id ?? "").toLowerCase();
                      return label.includes(query) || id.includes(query);
                    }}
                    onChange={handlePeriodChange}
                  />

                  <div className="popover-footer">
                    <Button
                      size="sm"
                      kind="ghost"
                      onClick={() => {
                        setSelectedPeriod("");
                        setIsYearFilterApplied(false);
                        setSelectedYear(currentYear);
                        setAvailablePeriods(
                          fetchAvailablePeriods(selectedPeriodType, currentYear.toString()),
                        );
                        setIsPeriodPopoverOpen(false);
                      }}
                    >
                      Clear
                    </Button>
                    <Button
                      size="sm"
                      onClick={() => {
                        if (!selectedPeriod) {
                          setIsYearFilterApplied(true);
                        }
                        setIsPeriodPopoverOpen(false);
                      }}
                    >
                      Apply
                    </Button>
                  </div>
                </div>
              </PopoverContent>
            </Popover>
          </div>

          {/* Dataset Filter */}
          <div style={{ minWidth: "15rem" }}>
            <ComboBox
              id="dashboard-dataset-filter"
              size="sm"
              titleText=""
              placeholder="Filter by Dataset"
              items={datasetNames}
              selectedItem={selectedDataset || null}
              shouldFilterItem={({ item, inputValue }) =>
                !inputValue ||
                String(item ?? "")
                  .toLowerCase()
                  .includes(inputValue.toLowerCase().trim())
              }
              onChange={({ selectedItem }: SelectEvent<string>) =>
                setSelectedDataset(selectedItem ?? "")
              }
            />
          </div>

          {/* Program Filter */}
          {programNames.length > 0 && (
            <div style={{ minWidth: "15rem" }}>
              <ComboBox
                id="dashboard-program-filter"
                size="sm"
                titleText=""
                placeholder="Filter by Program"
                items={programNames}
                selectedItem={selectedProgram || null}
                shouldFilterItem={({ item, inputValue }) =>
                  !inputValue ||
                  String(item ?? "")
                    .toLowerCase()
                    .includes(inputValue.toLowerCase().trim())
                }
                onChange={({ selectedItem }: SelectEvent<string>) =>
                  setSelectedProgram(selectedItem ?? undefined)
                }
              />
            </div>
          )}

          {/* Reset Filters */}
          <Button
            size="sm"
            kind="ghost"
            renderIcon={Filter}
            disabled={!hasActiveFilters}
            onClick={handleResetFilters}
          >
            Reset Filters
          </Button>

          {hasActiveFilters && (
            <span className="filter-active-indicator">
              Active Filters:{" "}
              {[
                selectedPeriod && `Period: ${selectedPeriod}`,
                (isYearFilterApplied || selectedYear !== currentYear) &&
                  !selectedPeriod &&
                  `Year: ${selectedYear}`,
                selectedDataset && `Dataset: ${selectedDataset}`,
                selectedProgram && `Program: ${selectedProgram}`,
              ]
                .filter(Boolean)
                .join(", ")}
            </span>
          )}
        </div>

        {/* ================= Search & Toolbar Row ================= */}
        <div className="dashboard-toolbar-row">
          <div className="dashboard-search">
            <Search
              labelText="Search regions, districts, or facilities"
              placeholder="Search by region, district, or facility name..."
              value={regionSearchTerm}
              onChange={(e) => setRegionSearchTerm(e.target.value)}
              size="sm"
              closeButtonLabelText="Clear search"
            />
          </div>
        </div>

        {/* ================= Data Table Section ================= */}
        {isTableLoading ? (
          <div className="dashboard-loading-state">
            <InlineLoading status="active" description="Loading regional issue summaries..." />
          </div>
        ) : issuesError ? (
          <div className="dashboard-empty-state">
            <h4>Failed to Load Issue Data</h4>
            <p>An error occurred while retrieving issue data from the server.</p>
            <Button size="sm" kind="tertiary" onClick={handleRefreshAll}>
              Retry
            </Button>
          </div>
        ) : displayedRows.length === 0 ? (
          <div className="dashboard-empty-state">
            <h4>No Regional Issues Found</h4>
            <p>
              {hasActiveFilters || regionSearchTerm
                ? "No issues match the selected filters or search criteria. Try adjusting or resetting the filters."
                : "There are currently no recorded issues to summarize."}
            </p>
            {hasActiveFilters && (
              <Button size="sm" kind="tertiary" onClick={handleResetFilters}>
                Reset Filters
              </Button>
            )}
          </div>
        ) : (
          <div className="summary-table-container">
            <DataTable rows={displayedRows} headers={tableHeaders}>
              {({
                rows,
                headers,
                getTableProps,
                getHeaderProps,
                getRowProps,
                getExpandHeaderProps,
              }) => (
                <Table {...getTableProps()} aria-label="Regional issue summary table">
                  <TableHead>
                    <TableRow>
                      <TableExpandHeader {...getExpandHeaderProps()} />
                      {headers.map((header) => {
                        const isSorted = sortHeader === header.key;
                        const { key: headerKey, ...headerProps } = getHeaderProps({
                          header,
                          isSortable: true,
                          onClick: () => handleSort(header.key),
                        });
                        return (
                          <TableHeader
                            key={headerKey}
                            {...headerProps}
                            isSortable
                            isSortHeader={isSorted}
                            sortDirection={isSorted ? sortDirection : "NONE"}
                          >
                            {header.header}
                          </TableHeader>
                        );
                      })}
                    </TableRow>
                  </TableHead>

                  <TableBody>
                    {rows.map((row) => {
                      const summaryItem = displayedRows.find((item) => item.id === row.id);
                      const { key: rowKey, ...rowProps } = getRowProps({ row });

                      return (
                        <Fragment key={row.id}>
                          <TableExpandRow key={rowKey} {...rowProps}>
                            {row.cells.map((cell) => {
                              const headerKey = cell.info.header;

                              if (headerKey === "region") {
                                return (
                                  <TableCell key={cell.id}>
                                    <div className="region-cell">
                                      <span>{cell.value}</span>
                                    </div>
                                  </TableCell>
                                );
                              }

                              if (headerKey === "critical") {
                                const count = Number(cell.value || 0);
                                return (
                                  <TableCell key={cell.id}>
                                    <span
                                      className={`severity-pill ${
                                        count > 0
                                          ? "severity-pill--critical"
                                          : "severity-pill--zero"
                                      }`}
                                    >
                                      {count > 0 ? count : "-"}
                                    </span>
                                  </TableCell>
                                );
                              }

                              if (headerKey === "open") {
                                const count = Number(cell.value || 0);
                                return (
                                  <TableCell key={cell.id}>
                                    <span className="stat-cell">
                                      <span className="stat-dot stat-dot--open" />
                                      {count}
                                    </span>
                                  </TableCell>
                                );
                              }

                              if (headerKey === "resolved") {
                                const count = Number(cell.value || 0);
                                return (
                                  <TableCell key={cell.id}>
                                    <span className="stat-cell">
                                      <span className="stat-dot stat-dot--resolved" />
                                      {count}
                                    </span>
                                  </TableCell>
                                );
                              }

                              if (headerKey === "total") {
                                const count = Number(cell.value || 0);
                                return (
                                  <TableCell key={cell.id}>
                                    <span className="total-pill">{count}</span>
                                  </TableCell>
                                );
                              }

                              if (headerKey === "percentage") {
                                const pct = Number(cell.value || 0);
                                return (
                                  <TableCell key={cell.id}>
                                    <div className="share-bar-wrapper">
                                      <div className="share-bar-bg">
                                        <div
                                          className="share-bar-fill"
                                          style={{ width: `${Math.min(100, pct)}%` }}
                                        />
                                      </div>
                                      <span className="share-percent-text">{pct.toFixed(1)}%</span>
                                    </div>
                                  </TableCell>
                                );
                              }

                              return <TableCell key={cell.id}>{cell.value ?? "-"}</TableCell>;
                            })}
                          </TableExpandRow>

                          <TableExpandedRow key={`expanded-${row.id}`} colSpan={headers.length + 1}>
                            {row.isExpanded && summaryItem && (
                              <div className="expanded-panel">
                                <div className="expanded-panel-header">
                                  <div>
                                    <h4 className="expanded-panel-title">
                                      District & Facility Breakdown for {summaryItem.region} Region
                                    </h4>
                                    <p
                                      style={{
                                        fontSize: "0.8125rem",
                                        color: "#525252",
                                        margin: "0.25rem 0 0",
                                      }}
                                    >
                                      {summaryItem.districts.length} district(s) |{" "}
                                      {summaryItem.districts.reduce(
                                        (acc, d) => acc + d.facilities.length,
                                        0,
                                      )}{" "}
                                      facility(ies) affected | <strong>{summaryItem.total}</strong>{" "}
                                      total issues (
                                      {summaryItem.critical > 0 && (
                                        <>
                                          <span style={{ color: "#a2191f", fontWeight: 600 }}>
                                            {summaryItem.critical} Critical
                                          </span>
                                          {", "}
                                        </>
                                      )}
                                      <span style={{ color: "#da1e28", fontWeight: 600 }}>
                                        {summaryItem.open} Open
                                      </span>
                                      ,{" "}
                                      <span style={{ color: "#198038", fontWeight: 600 }}>
                                        {summaryItem.resolved} Resolved
                                      </span>
                                      )
                                    </p>
                                  </div>

                                  <div
                                    style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}
                                  >
                                    <Button
                                      size="sm"
                                      kind="ghost"
                                      onClick={() =>
                                        toggleAllDistrictsInRegion(
                                          summaryItem.region,
                                          summaryItem.districts,
                                        )
                                      }
                                    >
                                      {isEveryDistrictExpanded(
                                        summaryItem.region,
                                        summaryItem.districts,
                                      )
                                        ? "Collapse All Facilities"
                                        : "Expand All Facilities"}
                                    </Button>

                                    <Button
                                      size="sm"
                                      kind="tertiary"
                                      renderIcon={ArrowRight}
                                      onClick={() => handleNavigateToIssues(summaryItem.region)}
                                    >
                                      View Issues in {summaryItem.region}
                                    </Button>
                                  </div>
                                </div>

                                <div className="district-table-wrapper">
                                  <table>
                                    <thead>
                                      <tr>
                                        <th>District</th>
                                        <th>Critical Issues</th>
                                        <th>Open</th>
                                        <th>Resolved</th>
                                        <th>Total Issues</th>
                                        <th style={{ width: "4rem", textAlign: "center" }}>
                                          Actions
                                        </th>
                                      </tr>
                                    </thead>
                                    <tbody>
                                      {summaryItem.districts.map((d) => {
                                        const expanded = isDistrictExpanded(summaryItem.region, d);

                                        return (
                                          <Fragment key={d.district}>
                                            <tr
                                              className={`district-row ${expanded ? "district-row--expanded" : ""}`}
                                            >
                                              <td>
                                                <button
                                                  type="button"
                                                  aria-expanded={expanded}
                                                  className="district-expand-btn"
                                                  onClick={() =>
                                                    toggleDistrict(summaryItem.region, d)
                                                  }
                                                  title={
                                                    expanded
                                                      ? "Collapse facility breakdown"
                                                      : "Expand facility breakdown"
                                                  }
                                                >
                                                  {expanded ? (
                                                    <ChevronDown
                                                      size={16}
                                                      className="district-expand-icon"
                                                    />
                                                  ) : (
                                                    <ChevronRight
                                                      size={16}
                                                      className="district-expand-icon"
                                                    />
                                                  )}
                                                  <span className="district-name-text">
                                                    {d.district}
                                                  </span>
                                                  <span className="district-facility-badge">
                                                    {d.facilities.length}{" "}
                                                    {d.facilities.length === 1
                                                      ? "facility"
                                                      : "facilities"}
                                                  </span>
                                                </button>
                                              </td>
                                              <td>
                                                <span
                                                  className={`severity-pill ${
                                                    d.critical > 0
                                                      ? "severity-pill--critical"
                                                      : "severity-pill--zero"
                                                  }`}
                                                >
                                                  {d.critical > 0 ? d.critical : "-"}
                                                </span>
                                              </td>
                                              <td>
                                                <span className="stat-cell">
                                                  <span className="stat-dot stat-dot--open" />
                                                  {d.open}
                                                </span>
                                              </td>
                                              <td>
                                                <span className="stat-cell">
                                                  <span className="stat-dot stat-dot--resolved" />
                                                  {d.resolved}
                                                </span>
                                              </td>
                                              <td>
                                                <strong>{d.total}</strong>
                                              </td>
                                              <td className="table-action-cell">
                                                <Button
                                                  size="sm"
                                                  kind="ghost"
                                                  hasIconOnly
                                                  renderIcon={ArrowRight}
                                                  iconDescription={`View issues in ${d.district}`}
                                                  tooltipPosition="left"
                                                  onClick={() => handleNavigateToIssues(d.district)}
                                                />
                                              </td>
                                            </tr>

                                            {expanded && (
                                              <tr className="facility-breakdown-row">
                                                <td colSpan={6}>
                                                  <div className="facility-breakdown-panel">
                                                    <div className="facility-breakdown-header">
                                                      <span className="facility-breakdown-title">
                                                        Facilities in {d.district} (
                                                        {d.facilities.length})
                                                      </span>
                                                    </div>
                                                    {d.facilities.length === 0 ? (
                                                      <p
                                                        style={{
                                                          margin: "0.5rem 0",
                                                          color: "#6f6f6f",
                                                          fontSize: "0.8125rem",
                                                        }}
                                                      >
                                                        No health facilities recorded for this
                                                        district.
                                                      </p>
                                                    ) : (
                                                      <table className="facility-table">
                                                        <thead>
                                                          <tr>
                                                            <th>Health Facility / Org Unit</th>
                                                            <th>Critical Issues</th>
                                                            <th>Open</th>
                                                            <th>Resolved</th>
                                                            <th>Total Issues</th>
                                                            <th
                                                              style={{
                                                                width: "4rem",
                                                                textAlign: "center",
                                                              }}
                                                            >
                                                              Actions
                                                            </th>
                                                          </tr>
                                                        </thead>
                                                        <tbody>
                                                          {d.facilities.map((f) => (
                                                            <tr key={f.facility}>
                                                              <td>
                                                                <span className="facility-name-cell">
                                                                  {f.facility}
                                                                </span>
                                                              </td>
                                                              <td>
                                                                <span
                                                                  className={`severity-pill ${
                                                                    f.critical > 0
                                                                      ? "severity-pill--critical"
                                                                      : "severity-pill--zero"
                                                                  }`}
                                                                >
                                                                  {f.critical > 0
                                                                    ? f.critical
                                                                    : "-"}
                                                                </span>
                                                              </td>
                                                              <td>
                                                                <span className="stat-cell">
                                                                  <span className="stat-dot stat-dot--open" />
                                                                  {f.open}
                                                                </span>
                                                              </td>
                                                              <td>
                                                                <span className="stat-cell">
                                                                  <span className="stat-dot stat-dot--resolved" />
                                                                  {f.resolved}
                                                                </span>
                                                              </td>
                                                              <td>
                                                                <strong>{f.total}</strong>
                                                              </td>
                                                              <td className="table-action-cell">
                                                                <Button
                                                                  size="sm"
                                                                  kind="ghost"
                                                                  hasIconOnly
                                                                  renderIcon={ArrowRight}
                                                                  iconDescription={`View issues in ${f.facility}`}
                                                                  tooltipPosition="left"
                                                                  onClick={() =>
                                                                    handleNavigateToIssues(
                                                                      f.facility,
                                                                    )
                                                                  }
                                                                />
                                                              </td>
                                                            </tr>
                                                          ))}
                                                        </tbody>
                                                      </table>
                                                    )}
                                                  </div>
                                                </td>
                                              </tr>
                                            )}
                                          </Fragment>
                                        );
                                      })}
                                    </tbody>
                                  </table>
                                </div>
                              </div>
                            )}
                          </TableExpandedRow>
                        </Fragment>
                      );
                    })}

                    {/* ================= Grand Total Summary Row ================= */}
                    <TableRow className="grand-total-row">
                      <TableCell />
                      <TableCell>
                        <strong>TOTAL (All Regions)</strong>
                      </TableCell>
                      <TableCell>
                        <span
                          className={`severity-pill ${
                            tableGrandTotals.critical > 0
                              ? "severity-pill--critical"
                              : "severity-pill--zero"
                          }`}
                        >
                          <strong>{tableGrandTotals.critical}</strong>
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="stat-cell">
                          <span className="stat-dot stat-dot--open" />
                          <strong>{tableGrandTotals.open}</strong>
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="stat-cell">
                          <span className="stat-dot stat-dot--resolved" />
                          <strong>{tableGrandTotals.resolved}</strong>
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="total-pill">{tableGrandTotals.total}</span>
                      </TableCell>
                      <TableCell>
                        <span style={{ fontSize: "0.8125rem", fontWeight: 700 }}>100.0%</span>
                      </TableCell>
                    </TableRow>
                  </TableBody>
                </Table>
              )}
            </DataTable>
          </div>
        )}

        {/* ================= Action Modals for Aging Issues ================= */}
        {assigningIssueCodes.length > 0 && (
          <PermissionGuard permission={PERMISSIONS.issueTrackerAssign}>
            <AssignModal
              issueCodes={assigningIssueCodes}
              isReassign={isAssignReassign}
              onClose={() => setAssigningIssueCodes([])}
              onSuccess={() => {
                setAssigningIssueCodes([]);
                handleRefreshAll();
              }}
            />
          </PermissionGuard>
        )}

        {resolvingAgingIssue && (
          <PermissionGuard permission={PERMISSIONS.issueTrackerManage}>
            <Modal
              open
              modalHeading={`Resolve Issue [${resolvingAgingIssue.issueCode}]`}
              primaryButtonText={isResolvingIssue ? "Resolving..." : "Confirm Resolution"}
              secondaryButtonText="Cancel"
              primaryButtonDisabled={!resolutionComment.trim() || isResolvingIssue}
              onRequestClose={() => {
                if (!isResolvingIssue) {
                  setResolvingAgingIssue(null);
                  setResolutionComment("");
                }
              }}
              onRequestSubmit={handleConfirmQuickResolve}
            >
              <div style={{ display: "flex", flexDirection: "column", gap: "0.75rem" }}>
                <div className="aging-resolve-modal-summary">
                  <div>
                    <strong>Facility:</strong> {resolvingAgingIssue.facilityName} (
                    {resolvingAgingIssue.districtName})
                  </div>
                  <div>
                    <strong>Age:</strong> {resolvingAgingIssue.ageDays} days unresolved (Reported:{" "}
                    {resolvingAgingIssue.reportedDateLabel})
                  </div>
                  <div>
                    <strong>Issue:</strong> {safeString(resolvingAgingIssue.issue.issue) || "N/A"}
                  </div>
                </div>

                <TextArea
                  data-modal-primary-focus
                  id="dashboard-aging-resolve-comment"
                  labelText="Resolution Action / Notes"
                  placeholder="Describe the corrective action taken to resolve this aging issue..."
                  value={resolutionComment}
                  onChange={(e) => setResolutionComment(e.target.value)}
                  rows={4}
                  disabled={isResolvingIssue}
                />
              </div>
            </Modal>
          </PermissionGuard>
        )}
      </div>
    </PermissionGuard>
  );
};

export default IssueDashboard;
