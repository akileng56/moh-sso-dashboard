import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import {
  Button,
  ComboBox,
  Dropdown,
  InlineLoading,
  Popover,
  PopoverContent,
  Search,
  Tab,
  TabList,
  Tabs,
  TreeView,
} from "@carbon/react";
import {
  Add,
  ChevronDown,
  Download,
  Filter,
  List,
  Upload,
  User,
} from "@carbon/react/icons";
import * as XLSX from "xlsx";
import { useSelector } from "react-redux";
import { useLocation, useSearchParams } from "react-router-dom";
import { PERMISSIONS, PermissionGuard, selectUser } from "@moh-sso/auth";
import { PageShell } from "@moh-sso/ui";

import {
  useGetIssueByCodeQuery,
  useGetIssueDataSetsQuery,
  useGetIssueOrgHierarchyQuery,
  useGetIssuesQuery,
  useGetIssuesSummaryByProgramQuery,
  useLazyGetIssueDataSetElementsQuery,
} from "../api";
import { IssuesDataTable } from "../component/issues-data-table.component";
import { IssueModal } from "../component/issue-modal.component";
import { ImportIssuesModal } from "../component/import-issues-modal.component";
import { AssignModal } from "../component/assign-modal.component";
import {
  normalizeDistrictKey,
  normalizeRegionKey,
} from "../component/issue-choropleth-map.component";
import { OrgUnitNode, type OrgUnitNodeData } from "../component/tree-node.component";
import {
  headers,
  IMPORT_TEMPLATE_EXAMPLE_ROW,
  IMPORT_TEMPLATE_HEADERS,
  periodType,
} from "../lib/constants";
import {
  buildOrgHierarchyMaps,
  doesIssueMatchReportingPeriod,
  isIssueResolved,
  resolveIssueGeography,
  safeLower,
  safeString,
  type OrgHierarchyMaps,
} from "../lib/issueAggregation";
import {
  fetchAvailablePeriods,
  formatReportingPeriod,
} from "../lib/periodUtils";
import IssueDetail from "./issue-detail/issue-detail.component";
import type {
  Issue,
  PeriodOption,
  SelectEvent,
  ThemeElement,
} from "../types";

export {
  formatReportingPeriod,
  formatReportingPeriod as convertReportingPeriod,
} from "../lib/periodUtils";
export type { Issue } from "../types";

import "./issue-tracker.scss";

function extractIssueCodeFromPath(pathname: string): string {
  const parts = pathname.split("/").filter(Boolean);
  const lastPart = parts[parts.length - 1];
  if (lastPart && /^HMIS-\d+$/i.test(lastPart)) {
    return lastPart.trim();
  }
  return "";
}

type FilterOption = {
  id: string;
  label: string;
  value: string;
};

const STATUS_FILTER_OPTIONS: FilterOption[] = [
  { id: "all", label: "All Statuses", value: "" },
  { id: "unresolved", label: "Unresolved", value: "unresolved" },
  { id: "open", label: "Open", value: "open" },
  { id: "in_progress", label: "In Progress", value: "in_progress" },
  { id: "resolved", label: "Resolved", value: "resolved" },
  { id: "closed", label: "Closed", value: "closed" },
];

const ASSIGNMENT_FILTER_OPTIONS: FilterOption[] = [
  { id: "all", label: "All Assignments", value: "" },
  { id: "unassigned", label: "Unassigned Only", value: "unassigned" },
  { id: "assigned", label: "Assigned Only", value: "assigned" },
];

function parseStatusFromSearchParams(params: URLSearchParams): string {
  const raw = (params.get("status") ?? "").trim().toLowerCase();
  if (raw) {
    if (raw === "in progress") return "in_progress";
    return raw;
  }
  const unresolvedParam = (params.get("unresolved") ?? "").trim().toLowerCase();
  if (unresolvedParam === "true" || unresolvedParam === "1") {
    return "unresolved";
  }
  return "";
}

function parseAssignmentFromSearchParams(params: URLSearchParams): string {
  const unassignedParam = (params.get("unassigned") ?? "").trim().toLowerCase();
  if (unassignedParam === "true" || unassignedParam === "1") {
    return "unassigned";
  }
  const assignedParam = (params.get("assigned") ?? "").trim().toLowerCase();
  if (assignedParam === "true" || assignedParam === "1") {
    return "assigned";
  }
  const assignmentParam = (
    params.get("assignment") ??
    params.get("assignedTo") ??
    params.get("assigned_to") ??
    ""
  )
    .trim()
    .toLowerCase();
  if (assignmentParam === "unassigned") return "unassigned";
  if (assignmentParam === "assigned") return "assigned";
  return "";
}

function hasUrlFilterParams(params: URLSearchParams): boolean {
  return Boolean(
    params.get("region") ||
      params.get("district") ||
      params.get("orgUnit") ||
      params.get("program") ||
      params.get("dataset") ||
      params.get("dataElement") ||
      params.get("period") ||
      params.get("year") ||
      params.get("status") ||
      params.get("unresolved") ||
      params.get("unassigned") ||
      params.get("assigned") ||
      params.get("assignment") ||
      params.get("assignedTo") ||
      params.get("assigned_to") ||
      params.get("search") ||
      params.get("view") === "issues" ||
      params.get("view") === "registered-issues" ||
      params.get("tab") === "issues",
  );
}

function getFilterQuerySignature(params: URLSearchParams): string {
  const clone = new URLSearchParams(params);
  clone.delete("issueCode");
  clone.delete("issue");
  clone.delete("code");
  return clone.toString();
}

function normalizeIssueRows(
  rows: Issue[] | undefined,
  hierarchyMaps?: OrgHierarchyMaps,
): Issue[] {
  return (
    rows?.map((item) => {
      const formattedPeriod = formatReportingPeriod(
        item.time_Period ?? item.time_period,
      );
      const geo = hierarchyMaps ? resolveIssueGeography(item, hierarchyMaps) : null;
      return {
        ...item,
        id: String(item.issue_id),
        region:
          safeString(item.region) ||
          (geo && geo.regionName !== "Unspecified" ? geo.regionName : item.region),
        district:
          safeString(item.district) ||
          (geo && geo.districtName !== "Unspecified" ? geo.districtName : item.district),
        time_Period: formattedPeriod,
        time_period: formattedPeriod,
      };
    }) ?? []
  );
}

const EXPORT_COLUMNS: { key: string; header: string }[] = [
  { key: "program", header: "Program" },
  ...headers,
  { key: "region", header: "Region" },
  { key: "time_period", header: "Reporting Period" },
  { key: "issue_type", header: "Issue Type" },
  { key: "priority", header: "Priority" },
  { key: "severity", header: "Severity" },
  { key: "reported_by", header: "Reported By" },
];

const IssueTracker = () => {
  const user = useSelector(selectUser);
  const location = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();
  const currentYear = new Date().getFullYear();

  const [showModal, setShowModal] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [showAssignModal, setShowAssignModal] = useState(false);
  const [selectedRowIds, setSelectedRowIds] = useState<string[]>([]);

  const [activeTab, setActiveTab] = useState<"assigned_to_me" | "all_issues">(() => {
    return hasUrlFilterParams(searchParams) ? "all_issues" : "assigned_to_me";
  });

  const [selectedIssue, setSelectedIssue] = useState<Issue>();
  const [isViewIssueDetail, setIsViewIssueDetail] = useState(false);

  const [urlIssueCode, setUrlIssueCode] = useState<string>(() => {
    const codeParam =
      searchParams.get("issueCode") ||
      searchParams.get("issue") ||
      searchParams.get("code");
    if (codeParam && codeParam.trim()) return codeParam.trim();
    return extractIssueCodeFromPath(location.pathname);
  });

  /*
   * Filter State
   */
  const [tableSearchTerm, setTableSearchTerm] = useState(
    () => searchParams.get("search")?.trim() || "",
  );
  const [selectedProgram, setSelectedProgram] = useState<string | undefined>(() => {
    const prog = searchParams.get("program");
    return prog ? prog.trim() : undefined;
  });

  const [selectedYear, setSelectedYear] = useState<number>(() => {
    const rawYear = Number(searchParams.get("year"));
    return Number.isInteger(rawYear) && rawYear > 1900 ? rawYear : currentYear;
  });
  const [isYearFilterApplied, setIsYearFilterApplied] = useState<boolean>(() =>
    Boolean(searchParams.get("year")?.trim()),
  );
  const [selectedPeriodType, setSelectedPeriodType] = useState<string>(
    () => searchParams.get("periodType")?.trim() || "Monthly",
  );
  const [selectedPeriod, setSelectedPeriod] = useState<string>(
    () => searchParams.get("period")?.trim() || "",
  );
  const [availablePeriods, setAvailablePeriods] = useState<PeriodOption[]>(() => {
    const pType = searchParams.get("periodType")?.trim() || "Monthly";
    const rawYear = Number(searchParams.get("year"));
    const yr = Number.isInteger(rawYear) && rawYear > 1900 ? rawYear : currentYear;
    return fetchAvailablePeriods(pType, yr.toString());
  });
  const [isPeriodPopoverOpen, setIsPeriodPopoverOpen] = useState(false);
  const periodPopoverRef = useRef<HTMLDivElement>(null);

  const [selectedDataset, setSelectedDataset] = useState(
    () => searchParams.get("dataset")?.trim() || "",
  );
  const [selectedDataElement, setSelectedDataElement] = useState(
    () => searchParams.get("dataElement")?.trim() || "",
  );
  const [dataElements, setDataElements] = useState<ThemeElement[]>([]);
  const [isDataPopoverOpen, setIsDataPopoverOpen] = useState(false);
  const dataPopoverRef = useRef<HTMLDivElement>(null);

  const [selectedRegion, setSelectedRegion] = useState<string>(
    () => searchParams.get("region")?.trim() || "",
  );
  const [selectedDistrict, setSelectedDistrict] = useState<string>(
    () => searchParams.get("district")?.trim() || "",
  );
  const [selectedOrgUnit, setSelectedOrgUnit] = useState<string>(
    () => searchParams.get("orgUnit")?.trim() || "",
  );
  const [orgSearchTerm, setOrgSearchTerm] = useState("");
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());
  const [isOrgPopoverOpen, setIsOrgPopoverOpen] = useState(false);
  const orgPopoverRef = useRef<HTMLDivElement>(null);

  const [selectedStatus, setSelectedStatus] = useState<string>(() =>
    parseStatusFromSearchParams(searchParams),
  );
  const [selectedAssignment, setSelectedAssignment] = useState<string>(() =>
    parseAssignmentFromSearchParams(searchParams),
  );

  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const lastFilterSignatureRef = useRef<string>(
    getFilterQuerySignature(searchParams),
  );

  /* Sync URL Search Params on Navigation (e.g. from Issue Summary Dashboard) */
  useEffect(() => {
    const codeParam =
      searchParams.get("issueCode") ||
      searchParams.get("issue") ||
      searchParams.get("code") ||
      extractIssueCodeFromPath(location.pathname);

    if (codeParam && codeParam.trim()) {
      setUrlIssueCode(codeParam.trim());
    }

    const currentSignature = getFilterQuerySignature(searchParams);
    if (currentSignature === lastFilterSignatureRef.current) {
      return;
    }
    lastFilterSignatureRef.current = currentSignature;

    const regionParam = searchParams.get("region")?.trim() ?? "";
    const districtParam = searchParams.get("district")?.trim() ?? "";
    const orgUnitParam = searchParams.get("orgUnit")?.trim() ?? "";
    const programParam = searchParams.get("program")?.trim();
    const datasetParam = searchParams.get("dataset")?.trim() ?? "";
    const dataElementParam = searchParams.get("dataElement")?.trim() ?? "";
    const periodParam = searchParams.get("period")?.trim() ?? "";
    const periodTypeParam = searchParams.get("periodType")?.trim() || "Monthly";
    const yearParamRaw = searchParams.get("year")?.trim();
    const parsedYear = yearParamRaw ? Number(yearParamRaw) : NaN;
    const effectiveYear =
      Number.isInteger(parsedYear) && parsedYear > 1900
        ? parsedYear
        : currentYear;
    const statusParam = parseStatusFromSearchParams(searchParams);
    const assignmentParam = parseAssignmentFromSearchParams(searchParams);
    const searchParam = searchParams.get("search")?.trim() ?? "";
    const viewParam = searchParams.get("view") || searchParams.get("tab");

    setSelectedRegion(regionParam);
    setSelectedDistrict(districtParam);
    setSelectedOrgUnit(orgUnitParam);
    setSelectedProgram(programParam || undefined);
    setSelectedDataset(datasetParam);
    setSelectedDataElement(dataElementParam);
    setSelectedPeriodType(periodTypeParam);
    setSelectedYear(effectiveYear);
    setIsYearFilterApplied(Boolean(yearParamRaw));
    setAvailablePeriods(
      fetchAvailablePeriods(periodTypeParam, effectiveYear.toString()),
    );
    setSelectedPeriod(periodParam);
    setSelectedStatus(statusParam);
    setSelectedAssignment(assignmentParam);
    setTableSearchTerm(searchParam);

    if (
      hasUrlFilterParams(searchParams) ||
      viewParam === "issues" ||
      viewParam === "registered-issues" ||
      viewParam === "dashboard"
    ) {
      setActiveTab("all_issues");
    }
  }, [searchParams, location.pathname, currentYear]);

  /*
   * API Hooks
   */
  const { data: directIssueData, isLoading: isLoadingDirectIssue } =
    useGetIssueByCodeQuery(urlIssueCode, { skip: !urlIssueCode });

  useEffect(() => {
    if (directIssueData && urlIssueCode) {
      setSelectedIssue(directIssueData);
      setIsViewIssueDetail(true);
    }
  }, [directIssueData, urlIssueCode]);

  const { data: summaryData = [] } = useGetIssuesSummaryByProgramQuery();
  const { data: datasets = [] } = useGetIssueDataSetsQuery();
  const [triggerGetDataSetElements] = useLazyGetIssueDataSetElementsQuery();
  const { data: hierarchyData = [] } = useGetIssueOrgHierarchyQuery();

  const hierarchyMaps = useMemo(
    () => buildOrgHierarchyMaps(hierarchyData),
    [hierarchyData],
  );

  const programOptions = useMemo(() => {
    const names = summaryData
      .map((item) => item.program || "Unspecified")
      .filter(Boolean);
    return Array.from(new Set(names));
  }, [summaryData]);

  const years = useMemo(
    () => Array.from({ length: 10 }, (_, index) => currentYear - index),
    [currentYear],
  );

  const datasetNames = useMemo(
    () => datasets.map((dataset) => dataset.display_name).filter(Boolean),
    [datasets],
  );

  const dataElementNames = useMemo(
    () =>
      dataElements
        .map((element) => element.data_element_short_name ?? "")
        .filter(Boolean),
    [dataElements],
  );

  const hasActiveFilters =
    selectedYear !== currentYear ||
    isYearFilterApplied ||
    selectedPeriodType !== "Monthly" ||
    Boolean(selectedPeriod) ||
    Boolean(selectedDataset) ||
    Boolean(selectedDataElement) ||
    Boolean(selectedRegion) ||
    Boolean(selectedDistrict) ||
    Boolean(selectedOrgUnit) ||
    Boolean(selectedProgram) ||
    Boolean(selectedStatus) ||
    Boolean(selectedAssignment);

  useEffect(() => {
    setPage(1);
  }, [
    selectedProgram,
    selectedYear,
    isYearFilterApplied,
    selectedPeriod,
    selectedDataset,
    selectedDataElement,
    selectedRegion,
    selectedDistrict,
    selectedOrgUnit,
    selectedStatus,
    selectedAssignment,
    tableSearchTerm,
    activeTab,
  ]);

  const userIdentifier = user?.email || user?.username || user?.id || "";

  const {
    data: allIssuesData,
    isLoading: isLoadingAllIssues,
    error: allIssuesError,
  } = useGetIssuesQuery({
    limit: 10000,
    offset: 0,
  });

  const {
    data: programIssuesData,
    isFetching: isFetchingProgramIssues,
    error: programIssuesError,
  } = useGetIssuesQuery(
    {
      limit: 10000,
      offset: 0,
      program: selectedProgram,
    },
    { skip: !selectedProgram },
  );

  const isLoading =
    isLoadingAllIssues || (Boolean(selectedProgram) && isFetchingProgramIssues);
  const error = allIssuesError || programIssuesError;

  const baseIssues = useMemo(() => {
    if (selectedProgram) {
      return programIssuesData?.items ?? [];
    }
    return allIssuesData?.items ?? [];
  }, [selectedProgram, programIssuesData?.items, allIssuesData?.items]);

  const assignedToMeCount = useMemo(() => {
    if (!userIdentifier) return 0;
    const allItems = allIssuesData?.items ?? [];
    return allItems.filter(
      (item) => safeLower(item.assigned_to) === safeLower(userIdentifier),
    ).length;
  }, [allIssuesData?.items, userIdentifier]);

  const allIssuesCount = useMemo(() => {
    if (allIssuesData?.items && allIssuesData.items.length > 0) {
      return allIssuesData.totalCount || allIssuesData.items.length;
    }
    if (summaryData.length > 0) {
      return summaryData.reduce((sum, item) => sum + (item.issue_count || 0), 0);
    }
    return 0;
  }, [allIssuesData, summaryData]);

  const closeAllPopovers = useCallback(() => {
    setIsPeriodPopoverOpen(false);
    setIsDataPopoverOpen(false);
    setIsOrgPopoverOpen(false);
  }, []);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node;

      if (periodPopoverRef.current && !periodPopoverRef.current.contains(target)) {
        setIsPeriodPopoverOpen(false);
      }
      if (dataPopoverRef.current && !dataPopoverRef.current.contains(target)) {
        setIsDataPopoverOpen(false);
      }
      if (orgPopoverRef.current && !orgPopoverRef.current.contains(target)) {
        setIsOrgPopoverOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  useEffect(() => {
    if (error) {
      console.error("Error encountered while fetching issues:", error);
    }
  }, [error]);

  const filteredIssues = useMemo(() => {
    const isYearActive = isYearFilterApplied || selectedYear !== currentYear;
    const searchTerm = safeLower(tableSearchTerm);

    const matching = baseIssues.filter((issue) => {
      if (activeTab === "assigned_to_me") {
        if (
          !userIdentifier ||
          safeLower(issue.assigned_to) !== safeLower(userIdentifier)
        ) {
          return false;
        }
      }

      if (selectedProgram !== undefined && safeString(issue.program)) {
        const itemProg = safeString(issue.program);
        const targetProg = safeString(selectedProgram) || "Unspecified";
        if (safeLower(itemProg) !== safeLower(targetProg)) {
          return false;
        }
      }

      if (
        selectedDataset &&
        safeLower(issue.dataset) !== safeLower(selectedDataset)
      ) {
        return false;
      }

      if (
        selectedDataElement &&
        safeLower(issue.data_element) !== safeLower(selectedDataElement)
      ) {
        return false;
      }

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

      const geo = resolveIssueGeography(issue, hierarchyMaps);

      if (selectedRegion) {
        if (safeLower(selectedRegion) === "unspecified") {
          if (safeLower(geo.regionName) !== "unspecified") return false;
        } else {
          const regMatches =
            safeLower(geo.regionName) === safeLower(selectedRegion) ||
            normalizeRegionKey(geo.regionName) ===
              normalizeRegionKey(selectedRegion) ||
            safeLower(issue.region) === safeLower(selectedRegion);
          if (!regMatches) return false;
        }
      }

      if (selectedDistrict) {
        if (safeLower(selectedDistrict) === "unspecified") {
          if (safeLower(geo.districtName) !== "unspecified") return false;
        } else {
          const distMatches =
            safeLower(geo.districtName) === safeLower(selectedDistrict) ||
            normalizeDistrictKey(geo.districtName) ===
              normalizeDistrictKey(selectedDistrict) ||
            safeLower(issue.district) === safeLower(selectedDistrict);
          if (!distMatches) return false;
        }
      }

      if (selectedOrgUnit) {
        if (safeLower(selectedOrgUnit) === "unspecified facility") {
          if (safeLower(geo.facilityName) !== "unspecified facility") {
            return false;
          }
        } else if (selectedDistrict || selectedRegion) {
          const facMatches =
            safeLower(geo.facilityName) === safeLower(selectedOrgUnit) ||
            safeLower(issue.org_unit) === safeLower(selectedOrgUnit);
          if (!facMatches) return false;
        } else {
          const orgMatches =
            safeLower(geo.facilityName) === safeLower(selectedOrgUnit) ||
            safeLower(issue.org_unit) === safeLower(selectedOrgUnit) ||
            safeLower(geo.districtName) === safeLower(selectedOrgUnit) ||
            normalizeDistrictKey(geo.districtName) ===
              normalizeDistrictKey(selectedOrgUnit) ||
            safeLower(geo.regionName) === safeLower(selectedOrgUnit) ||
            normalizeRegionKey(geo.regionName) ===
              normalizeRegionKey(selectedOrgUnit);
          if (!orgMatches) return false;
        }
      }

      if (selectedStatus) {
        if (selectedStatus === "unresolved") {
          if (isIssueResolved(issue.status)) return false;
        } else if (selectedStatus === "in_progress") {
          const st = safeLower(issue.status);
          if (st !== "in_progress" && st !== "in progress") return false;
        } else if (safeLower(issue.status) !== safeLower(selectedStatus)) {
          return false;
        }
      }

      if (selectedAssignment) {
        const isUnassigned = !safeString(issue.assigned_to);
        if (selectedAssignment === "unassigned" && !isUnassigned) {
          return false;
        }
        if (selectedAssignment === "assigned" && isUnassigned) {
          return false;
        }
      }

      if (searchTerm) {
        const formattedPeriod = formatReportingPeriod(
          issue.time_Period ?? issue.time_period,
        );
        const matchesSearch =
          safeLower(issue.issue_code).includes(searchTerm) ||
          safeLower(issue.dataset).includes(searchTerm) ||
          safeLower(issue.data_element).includes(searchTerm) ||
          safeLower(issue.issue).includes(searchTerm) ||
          safeLower(issue.org_unit).includes(searchTerm) ||
          safeLower(geo.facilityName).includes(searchTerm) ||
          safeLower(geo.districtName).includes(searchTerm) ||
          safeLower(geo.regionName).includes(searchTerm) ||
          safeLower(issue.status).includes(searchTerm) ||
          safeLower(issue.assigned_to).includes(searchTerm) ||
          safeLower(issue.program).includes(searchTerm) ||
          safeLower(formattedPeriod).includes(searchTerm);
        if (!matchesSearch) return false;
      }

      return true;
    });

    return normalizeIssueRows(matching, hierarchyMaps);
  }, [
    baseIssues,
    activeTab,
    userIdentifier,
    selectedProgram,
    selectedDataset,
    selectedDataElement,
    selectedPeriod,
    selectedYear,
    currentYear,
    isYearFilterApplied,
    selectedRegion,
    selectedDistrict,
    selectedOrgUnit,
    selectedStatus,
    selectedAssignment,
    tableSearchTerm,
    hierarchyMaps,
  ]);

  const issues = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filteredIssues.slice(start, start + pageSize);
  }, [filteredIssues, page, pageSize]);

  /* -----------------------------
   * Excel template & export
   * ----------------------------- */
  const downloadTemplate = () => {
    const workbook = XLSX.utils.book_new();
    const worksheet = XLSX.utils.aoa_to_sheet([
      IMPORT_TEMPLATE_HEADERS,
      IMPORT_TEMPLATE_EXAMPLE_ROW,
    ]);
    worksheet["!cols"] = IMPORT_TEMPLATE_HEADERS.map(() => ({ wch: 22 }));
    XLSX.utils.book_append_sheet(workbook, worksheet, "Issues Template");
    XLSX.writeFile(workbook, "issue-import-template.xlsx");
  };

  const downloadIssues = () => {
    try {
      const issuesToDownload = filteredIssues;
      const headerRow = EXPORT_COLUMNS.map((col) => col.header);

      const dataRows = issuesToDownload.map((issue) =>
        EXPORT_COLUMNS.map((col) => {
          if (col.key === "program") {
            const rawProgram = issue.program;
            if (typeof rawProgram === "string" && rawProgram.trim()) {
              return rawProgram.trim();
            }
            if (selectedProgram !== undefined) {
              return selectedProgram || "Unspecified";
            }
            return "Unspecified";
          }
          if (col.key === "time_period") {
            return issue.time_period ?? issue.time_Period ?? "";
          }
          return (issue as Record<string, unknown>)[col.key] ?? "";
        }),
      );

      const workbook = XLSX.utils.book_new();
      const worksheet = XLSX.utils.aoa_to_sheet([headerRow, ...dataRows]);
      worksheet["!cols"] = EXPORT_COLUMNS.map(() => ({ wch: 22 }));
      XLSX.utils.book_append_sheet(workbook, worksheet, "Issues");
      XLSX.writeFile(
        workbook,
        `issues-${new Date().toISOString().slice(0, 10)}.xlsx`,
      );
    } catch (err) {
      console.error("Failed to download issues:", err);
    }
  };

  /* -----------------------------
   * Selection & Detail Handlers
   * ----------------------------- */
  const handleSelectRow = useCallback((rowId: string, checked: boolean) => {
    setSelectedRowIds((prev) => {
      if (checked) {
        return prev.includes(rowId) ? prev : [...prev, rowId];
      }
      return prev.filter((id) => id !== rowId);
    });
  }, []);

  const handleSelectAll = useCallback(
    (checked: boolean) => {
      if (checked) {
        const allIds = issues.map((item) => String(item.id ?? item.issue_id));
        setSelectedRowIds(allIds);
      } else {
        setSelectedRowIds([]);
      }
    },
    [issues],
  );

  const selectedIssueObjects = useMemo(
    () =>
      filteredIssues.filter((item) =>
        selectedRowIds.includes(String(item.id ?? item.issue_id)),
      ),
    [filteredIssues, selectedRowIds],
  );

  const selectedIssueCodes = useMemo(
    () => selectedIssueObjects.map((item) => item.issue_code).filter(Boolean),
    [selectedIssueObjects],
  );

  const areAllSelectedAssigned = useMemo(
    () =>
      selectedIssueObjects.length > 0 &&
      selectedIssueObjects.every((item) => Boolean(item.assigned_to)),
    [selectedIssueObjects],
  );

  const handleIssueClick = (row: { id?: string }) => {
    const selectedItem = filteredIssues.find(
      (item) => String(item.issue_id) === String(row.id ?? ""),
    );
    if (!selectedItem) return;

    setSelectedIssue(selectedItem);
    setIsViewIssueDetail(true);

    if (selectedItem.issue_code) {
      const next = new URLSearchParams(searchParams);
      next.set("issueCode", selectedItem.issue_code);
      setSearchParams(next, { replace: true });
    }
  };

  const handleBackFromDetail = () => {
    setSelectedIssue(undefined);
    setIsViewIssueDetail(false);
    setUrlIssueCode("");
    const next = new URLSearchParams(searchParams);
    next.delete("issueCode");
    next.delete("issue");
    next.delete("code");
    setSearchParams(next, { replace: true });
  };

  /* -----------------------------
   * Filter Control Handlers
   * ----------------------------- */
  const handleYearChange = ({ selectedItem }: SelectEvent<number>) => {
    if (selectedItem == null) return;
    setSelectedYear(selectedItem);
    setIsYearFilterApplied(true);
    setAvailablePeriods(
      fetchAvailablePeriods(selectedPeriodType, selectedItem.toString()),
    );
    setSelectedPeriod("");
  };

  const handlePeriodTypeChange = ({
    selectedItem,
  }: SelectEvent<{ label: string; value: string }>) => {
    if (!selectedItem) return;
    setSelectedPeriodType(selectedItem.value);
    setAvailablePeriods(
      fetchAvailablePeriods(selectedItem.value, selectedYear.toString()),
    );
    setSelectedPeriod("");
  };

  const handlePeriodChange = ({ selectedItem }: SelectEvent<PeriodOption>) => {
    setSelectedPeriod(selectedItem?.label ?? "");
  };

  const handleDatasetChange = async ({ selectedItem }: SelectEvent<string>) => {
    const datasetName = selectedItem ?? "";
    setSelectedDataset(datasetName);
    setSelectedDataElement("");

    if (!datasetName) {
      setDataElements([]);
      return;
    }

    const dataset = datasets.find((item) => item.display_name === datasetName);
    if (!dataset) {
      setDataElements([]);
      return;
    }

    try {
      const elements = await triggerGetDataSetElements(dataset.dataset_id).unwrap();
      setDataElements(elements);
    } catch (requestError) {
      console.error("Failed to fetch data elements:", requestError);
      setDataElements([]);
    }
  };

  const nodeMatchesSearch = useCallback(
    (node: OrgUnitNodeData, term: string): boolean => {
      if (node.name.toLowerCase().includes(term.toLowerCase())) {
        return true;
      }
      return Boolean(
        node.children?.some((child) => nodeMatchesSearch(child, term)),
      );
    },
    [],
  );

  const handleOrgSelect = (name: string) => {
    setSelectedOrgUnit((prev) => (prev === name ? "" : name));
    setSelectedRegion("");
    setSelectedDistrict("");
  };

  const handleToggleNode = useCallback((nodeId: string, isExpanded: boolean) => {
    setExpandedNodes((prev) => {
      const next = new Set(prev);
      if (isExpanded) {
        next.add(nodeId);
      } else {
        next.delete(nodeId);
      }
      return next;
    });
  }, []);

  const renderRecursive = (
    nodes: OrgUnitNodeData[],
    idPrefix = "filter-org",
  ): ReactNode[] => {
    if (!Array.isArray(nodes)) return [];

    return nodes
      .filter((node) => !orgSearchTerm || nodeMatchesSearch(node, orgSearchTerm))
      .map((node) => (
        <OrgUnitNode
          key={node.id}
          node={node}
          searchTerm={orgSearchTerm}
          selectedOrgUnit={selectedOrgUnit}
          expandedNodes={expandedNodes}
          onSelect={handleOrgSelect}
          onToggleNode={handleToggleNode}
          renderRecursive={renderRecursive}
          idPrefix={idPrefix}
        />
      ));
  };

  const handleFilter = (
    overrides: {
      period?: string;
      year?: number;
      dataset?: string;
      dataElement?: string;
      orgUnit?: string;
    } = {},
  ) => {
    closeAllPopovers();
    setPage(1);

    if (overrides.period !== undefined) setSelectedPeriod(overrides.period);
    if (overrides.year !== undefined) setSelectedYear(overrides.year);
    if (overrides.dataset !== undefined) setSelectedDataset(overrides.dataset);
    if (overrides.dataElement !== undefined) {
      setSelectedDataElement(overrides.dataElement);
    }
    if (overrides.orgUnit !== undefined) {
      setSelectedOrgUnit(overrides.orgUnit);
      if (!overrides.orgUnit) {
        setSelectedRegion("");
        setSelectedDistrict("");
      }
    }
  };

  const handleReset = () => {
    setSelectedYear(currentYear);
    setIsYearFilterApplied(false);
    setSelectedPeriodType("Monthly");
    setSelectedPeriod("");
    setAvailablePeriods(fetchAvailablePeriods("Monthly", currentYear.toString()));

    setSelectedDataset("");
    setSelectedDataElement("");
    setDataElements([]);

    setSelectedRegion("");
    setSelectedDistrict("");
    setSelectedOrgUnit("");
    setOrgSearchTerm("");
    setTableSearchTerm("");
    setSelectedProgram(undefined);
    setSelectedStatus("");
    setSelectedAssignment("");
    setPage(1);

    const next = new URLSearchParams(searchParams);
    next.delete("region");
    next.delete("district");
    next.delete("orgUnit");
    next.delete("program");
    next.delete("dataset");
    next.delete("dataElement");
    next.delete("period");
    next.delete("periodType");
    next.delete("year");
    next.delete("status");
    next.delete("unresolved");
    next.delete("unassigned");
    next.delete("assigned");
    next.delete("assignment");
    next.delete("assignedTo");
    next.delete("assigned_to");
    next.delete("search");
    lastFilterSignatureRef.current = getFilterQuerySignature(next);
    setSearchParams(next, { replace: true });

    closeAllPopovers();
  };

  if (isLoadingDirectIssue && urlIssueCode && !selectedIssue) {
    return (
      <PermissionGuard permission={PERMISSIONS.issueTrackerRead}>
        <div className="issue-direct-loading">
          <InlineLoading
            status="active"
            description={`Fetching issue [${urlIssueCode}] from server...`}
          />
        </div>
      </PermissionGuard>
    );
  }

  if (isViewIssueDetail && selectedIssue) {
    return (
      <PermissionGuard permission={PERMISSIONS.issueTrackerRead}>
        <IssueDetail selectedIssue={selectedIssue} goToBack={handleBackFromDetail} />
      </PermissionGuard>
    );
  }

  const activeLocationLabel =
    selectedOrgUnit || selectedDistrict || selectedRegion || "";

  const selectedStatusLabel =
    STATUS_FILTER_OPTIONS.find((opt) => opt.value === selectedStatus)?.label ||
    selectedStatus;
  const selectedAssignmentLabel =
    ASSIGNMENT_FILTER_OPTIONS.find((opt) => opt.value === selectedAssignment)
      ?.label || selectedAssignment;

  const activeFilterSummary = [
    selectedProgram !== undefined &&
      `Program: ${selectedProgram || "Unspecified"}`,
    selectedPeriod && `Period: ${selectedPeriod}`,
    (isYearFilterApplied || selectedYear !== currentYear) &&
      !selectedPeriod &&
      `Year: ${selectedYear}`,
    selectedDataset && `Dataset: ${selectedDataset}`,
    selectedDataElement && `Element: ${selectedDataElement}`,
    selectedRegion && `Region: ${selectedRegion}`,
    selectedDistrict && `District: ${selectedDistrict}`,
    selectedOrgUnit && `Org Unit: ${selectedOrgUnit}`,
    selectedStatus && `Status: ${selectedStatusLabel}`,
    selectedAssignment && `Assignment: ${selectedAssignmentLabel}`,
  ]
    .filter(Boolean)
    .join(" • ");

  return (
    <PermissionGuard permission={PERMISSIONS.issueTrackerRead}>
      <PageShell
        title="Registered Issues"
        subtitle="Review, filter, assign, and resolve reported data quality issues."
        actions={
          <div className="issue-toolbar-actions">
            <PermissionGuard permission={PERMISSIONS.issueTrackerAssign}>
              <Button
                size="sm"
                kind="tertiary"
                disabled={selectedIssueCodes.length === 0}
                onClick={() => setShowAssignModal(true)}
              >
                {areAllSelectedAssigned
                  ? "Re-assign Selected"
                  : "Assign Selected"}{" "}
                {selectedIssueCodes.length > 0
                  ? `(${selectedIssueCodes.length})`
                  : ""}
              </Button>
            </PermissionGuard>

            <PermissionGuard permission={PERMISSIONS.issueTrackerWrite}>
              <Button
                size="sm"
                kind="ghost"
                renderIcon={Download}
                onClick={downloadTemplate}
              >
                Template
              </Button>
            </PermissionGuard>

            <PermissionGuard permission={PERMISSIONS.issueTrackerWrite}>
              <Button
                size="sm"
                kind="secondary"
                renderIcon={Upload}
                onClick={() => setShowImportModal(true)}
              >
                Import Issues
              </Button>
            </PermissionGuard>

            <PermissionGuard permission={PERMISSIONS.issueTrackerWrite}>
              <Button
                size="sm"
                kind="primary"
                renderIcon={Add}
                className="dwh-btn-width"
                onClick={() => setShowModal(true)}
              >
                New Issue
              </Button>
            </PermissionGuard>
          </div>
        }
      >
        <div className="issue-tracker-page">
          <div className="issue-tracker-assignments">
            <Tabs
              selectedIndex={activeTab === "assigned_to_me" ? 0 : 1}
              onChange={({ selectedIndex }) => {
                setActiveTab(
                  selectedIndex === 0 ? "assigned_to_me" : "all_issues",
                );
                setSelectedRowIds([]);
              }}
            >
              <TabList aria-label="Issue tracker view options" contained>
                <Tab renderIcon={User}>
                  Assigned to me ({assignedToMeCount})
                </Tab>
                <Tab renderIcon={List}>All Issues ({allIssuesCount})</Tab>
              </TabList>
            </Tabs>
          </div>

          <div className="issue-filter-container">
            {programOptions.length > 0 && (
              <div className="issue-filter-combobox">
                <ComboBox
                  id="issue-filter-program"
                  size="md"
                  titleText=""
                  aria-label="Filter by Program"
                  placeholder="Filter by Program"
                  items={programOptions}
                  selectedItem={
                    selectedProgram === undefined
                      ? null
                      : selectedProgram || "Unspecified"
                  }
                  shouldFilterItem={({ item, inputValue }) =>
                    !inputValue ||
                    String(item ?? "")
                      .toLowerCase()
                      .includes(inputValue.toLowerCase().trim())
                  }
                  onChange={({ selectedItem }: SelectEvent<string>) => {
                    if (!selectedItem) {
                      setSelectedProgram(undefined);
                    } else {
                      setSelectedProgram(
                        selectedItem === "Unspecified" ? "Unspecified" : selectedItem,
                      );
                    }
                  }}
                />
              </div>
            )}

            <div ref={periodPopoverRef} className="issue-filter-wrapper">
              <Popover open={isPeriodPopoverOpen} align="bottom-left" dropShadow>
                <Button
                  size="md"
                  kind="tertiary"
                  renderIcon={ChevronDown}
                  onClick={() => {
                    setIsPeriodPopoverOpen((current) => !current);
                    setIsDataPopoverOpen(false);
                    setIsOrgPopoverOpen(false);
                  }}
                >
                  {selectedPeriod
                    ? `Period: ${selectedPeriod}`
                    : isYearFilterApplied || selectedYear !== currentYear
                      ? `Year: ${selectedYear}`
                      : "Period"}
                </Button>

                <PopoverContent className="filter-popover-content">
                  <div className="popover-inner">
                    <Dropdown
                      id="issue-filter-year"
                      titleText="Year"
                      label="Select year"
                      items={years}
                      selectedItem={selectedYear}
                      itemToString={(item) =>
                        item == null ? "" : String(item)
                      }
                      onChange={handleYearChange}
                    />

                    <Dropdown
                      id="issue-filter-period-type"
                      titleText="Period type"
                      label="Select period type"
                      items={periodType}
                      selectedItem={periodType.find(
                        (item) => item.value === selectedPeriodType,
                      )}
                      itemToString={(item) => item?.label ?? ""}
                      onChange={handlePeriodTypeChange}
                    />

                    <ComboBox
                      id="issue-filter-period"
                      titleText="Period"
                      placeholder="Select period"
                      items={availablePeriods}
                      selectedItem={
                        availablePeriods.find(
                          (item) => item.label === selectedPeriod,
                        ) ?? null
                      }
                      itemToString={(item) => item?.label ?? ""}
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
                          handleFilter({ period: "", year: currentYear });
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
                          handleFilter();
                        }}
                      >
                        Update
                      </Button>
                    </div>
                  </div>
                </PopoverContent>
              </Popover>
            </div>

            <div ref={dataPopoverRef} className="issue-filter-wrapper">
              <Popover open={isDataPopoverOpen} align="bottom-left" dropShadow>
                <Button
                  size="md"
                  kind="tertiary"
                  renderIcon={ChevronDown}
                  onClick={() => {
                    setIsDataPopoverOpen((current) => !current);
                    setIsPeriodPopoverOpen(false);
                    setIsOrgPopoverOpen(false);
                  }}
                >
                  {selectedDataset ? `Data: ${selectedDataset}` : "Data"}
                </Button>

                <PopoverContent className="filter-popover-content">
                  <div className="popover-inner">
                    <ComboBox
                      id="issue-filter-dataset"
                      titleText="Dataset"
                      placeholder="Select dataset"
                      items={datasetNames}
                      selectedItem={selectedDataset || null}
                      onChange={handleDatasetChange}
                    />

                    <ComboBox
                      id="issue-filter-data-element"
                      titleText="Data element"
                      placeholder="Select data element"
                      items={dataElementNames}
                      selectedItem={selectedDataElement || null}
                      disabled={!selectedDataset || dataElementNames.length === 0}
                      onChange={({ selectedItem }: SelectEvent<string>) =>
                        setSelectedDataElement(selectedItem ?? "")
                      }
                    />

                    <div className="popover-footer">
                      <Button
                        size="sm"
                        kind="ghost"
                        onClick={() => {
                          setSelectedDataset("");
                          setSelectedDataElement("");
                          setDataElements([]);
                          handleFilter({ dataset: "", dataElement: "" });
                        }}
                      >
                        Clear
                      </Button>

                      <Button size="sm" onClick={() => handleFilter()}>
                        Update
                      </Button>
                    </div>
                  </div>
                </PopoverContent>
              </Popover>
            </div>

            <div ref={orgPopoverRef} className="issue-filter-wrapper">
              <Popover open={isOrgPopoverOpen} align="bottom-left" dropShadow>
                <Button
                  size="md"
                  kind="tertiary"
                  renderIcon={ChevronDown}
                  onClick={() => {
                    setIsOrgPopoverOpen((current) => !current);
                    setIsPeriodPopoverOpen(false);
                    setIsDataPopoverOpen(false);
                  }}
                >
                  {activeLocationLabel
                    ? `Location: ${activeLocationLabel}`
                    : "Organisation Unit"}
                </Button>

                <PopoverContent className="filter-popover-content">
                  <div className="popover-inner org-filter-inner">
                    <Search
                      labelText="Search organisation unit"
                      placeholder="Search..."
                      value={orgSearchTerm}
                      onChange={(event) => setOrgSearchTerm(event.target.value)}
                      size="sm"
                    />

                    <div className="org-tree-container">
                      <TreeView label="Organisation Units" hideLabel>
                        {renderRecursive(hierarchyData, "filter-org")}
                      </TreeView>
                    </div>

                    <div className="popover-footer">
                      <Button
                        size="sm"
                        kind="ghost"
                        onClick={() => {
                          setSelectedOrgUnit("");
                          setSelectedRegion("");
                          setSelectedDistrict("");
                          setOrgSearchTerm("");
                          handleFilter({ orgUnit: "" });
                        }}
                      >
                        Clear
                      </Button>

                      <Button size="sm" onClick={() => handleFilter()}>
                        Update
                      </Button>
                    </div>
                  </div>
                </PopoverContent>
              </Popover>
            </div>

            <div className="issue-filter-dropdown">
              <Dropdown
                id="issue-filter-status"
                size="md"
                titleText=""
                label="Filter by Status"
                aria-label="Filter by Status"
                items={STATUS_FILTER_OPTIONS}
                selectedItem={
                  STATUS_FILTER_OPTIONS.find(
                    (opt) => opt.value === selectedStatus,
                  ) ?? STATUS_FILTER_OPTIONS[0]
                }
                itemToString={(item) => item?.label ?? ""}
                onChange={({ selectedItem }: SelectEvent<FilterOption>) =>
                  setSelectedStatus(selectedItem?.value ?? "")
                }
              />
            </div>

            {activeTab === "all_issues" && (
              <div className="issue-filter-dropdown">
                <Dropdown
                  id="issue-filter-assignment"
                  size="md"
                  titleText=""
                  label="Filter by Assignment"
                  aria-label="Filter by Assignment"
                  items={ASSIGNMENT_FILTER_OPTIONS}
                  selectedItem={
                    ASSIGNMENT_FILTER_OPTIONS.find(
                      (opt) => opt.value === selectedAssignment,
                    ) ?? ASSIGNMENT_FILTER_OPTIONS[0]
                  }
                  itemToString={(item) => item?.label ?? ""}
                  onChange={({ selectedItem }: SelectEvent<FilterOption>) =>
                    setSelectedAssignment(selectedItem?.value ?? "")
                  }
                />
              </div>
            )}

            <Button
              size="md"
              kind="ghost"
              renderIcon={Filter}
              disabled={!hasActiveFilters}
              onClick={handleReset}
            >
              Reset Filters
            </Button>

            {hasActiveFilters && activeFilterSummary && (
              <span className="issue-active-filters-badge">
                {activeFilterSummary}
              </span>
            )}
          </div>

          <div className="issue-table-search">
            <div className="issue-table-search-bar">
              <Search
                labelText="Search issues"
                placeholder="Search by issue code, dataset, data element, issue, status, or organisation unit"
                value={tableSearchTerm}
                onChange={(event) => setTableSearchTerm(event.target.value)}
                size="lg"
                closeButtonLabelText="Clear search"
              />
              <PermissionGuard permission={PERMISSIONS.issueTrackerWrite}>
                <Button
                  size="md"
                  kind="ghost"
                  renderIcon={Download}
                  onClick={downloadIssues}
                >
                  Download
                </Button>
              </PermissionGuard>
            </div>
          </div>

          {activeTab === "assigned_to_me" &&
            issues.length === 0 &&
            !isLoading && (
              <div className="issue-assigned-empty-banner">
                <strong>No issues currently assigned to you.</strong> Switch to
                the <em>All Issues</em> tab to view all registered issues.
              </div>
            )}

          <div
            className="issue-tracker-table"
            role="region"
            aria-label="Registered issues table"
            tabIndex={0}
          >
            <IssuesDataTable
              columns={headers}
              data={issues}
              totalItems={filteredIssues.length}
              currentPage={page}
              currentPageSize={pageSize}
              onPageChange={(newPage, newPageSize) => {
                setPage(newPage);
                setPageSize(newPageSize);
              }}
              handleIssueClick={handleIssueClick}
              closeView={() => {
                setSelectedIssue(undefined);
                setIsViewIssueDetail(false);
              }}
              selectedRowIds={selectedRowIds}
              onSelectRow={handleSelectRow}
              onSelectAll={handleSelectAll}
            />
          </div>

          {showModal && (
            <PermissionGuard permission={PERMISSIONS.issueTrackerWrite}>
              <IssueModal onClose={() => setShowModal(false)} />
            </PermissionGuard>
          )}

          {showImportModal && (
            <PermissionGuard permission={PERMISSIONS.issueTrackerWrite}>
              <ImportIssuesModal onClose={() => setShowImportModal(false)} />
            </PermissionGuard>
          )}

          {showAssignModal && selectedIssueCodes.length > 0 && (
            <PermissionGuard permission={PERMISSIONS.issueTrackerAssign}>
              <AssignModal
                issueCodes={selectedIssueCodes}
                isReassign={areAllSelectedAssigned}
                onClose={() => setShowAssignModal(false)}
                onSuccess={() => setSelectedRowIds([])}
              />
            </PermissionGuard>
          )}
        </div>
      </PageShell>
    </PermissionGuard>
  );
};

export default IssueTracker;
