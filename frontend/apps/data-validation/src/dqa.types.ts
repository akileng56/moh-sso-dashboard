// Types for the ported DQA v2 engine (backend: internal/features/data_quality/dqa).

export type DQARuleType =
  | "metric"
  | "row_expression"
  | "row_join"
  | "aggregate_compare"
  | "reference"
  | "raw_sql";

export const DQA_RULE_TYPES: { value: DQARuleType; label: string; description: string }[] = [
  { value: "row_expression", label: "Row expression", description: "Flag rows failing a boolean predicate, e.g. tested >= positive." },
  { value: "metric", label: "Metric", description: "Measure an aggregate (sum, avg, missing %, …) and compare it to thresholds." },
  { value: "row_join", label: "Row join", description: "Evaluate a predicate across two row-aligned tables." },
  { value: "aggregate_compare", label: "Aggregate compare", description: "Compare totals across two tables within a tolerance." },
  { value: "reference", label: "Reference", description: "Check referential integrity via an anti-join." },
  { value: "raw_sql", label: "Raw SQL", description: "Escape hatch: a SELECT whose rows are the violations." },
];

export type DQAZones = {
  warn: string;
  fail: string;
};

export type DQATableMapping = {
  table_id: string;
  physical_table: string;
  description?: string;
  is_active: boolean;
};

export type DQARule = {
  id: number;
  identity: string;
  code: string;
  table_id: string;
  type: DQARuleType;
  category: string;
  name: string;
  enabled: boolean;
  row_filter?: string;
  group_by?: string[];
  zones: DQAZones;
  definition: Record<string, unknown>;
  compiled_sql?: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
};

export type DQARuleInput = {
  code: string;
  table_id: string;
  type: DQARuleType;
  category?: string;
  name?: string;
  enabled?: boolean;
  row_filter?: string;
  group_by?: string[];
  zones: DQAZones;
  definition: Record<string, unknown>;
};

export type DQACompilePreview = {
  kind: string;
  measure_sql: string;
  sample_sql?: string;
};

export type DQARunResult = {
  run_id: number;
  table_id: string;
  rules_evaluated: number;
  total_flags: number;
  fails: number;
  warns: number;
  compile_errors?: { code: string; error: string }[];
};

export type DQARunSummary = {
  id: number;
  table_id: string;
  run_at: string;
  rules_evaluated: number;
  total_flags: number;
  errors: number;
  warnings: number;
  triggered_by?: string;
};

export type DQAFlag = {
  id: number;
  run_id: number;
  table_id: string;
  rule_code: string;
  severity: "warn" | "fail";
  category: string;
  detail: string;
  entity_id?: string;
  period_date?: string;
  district?: string;
  entity_name?: string;
  row_id?: string;
  dims?: Record<string, unknown>;
  comment?: string;
};

export type DQASeedResult = {
  table_id: string;
  seeded: number;
  failed: number;
  rules: DQARule[];
  errors: { code: string; error: string }[];
};
