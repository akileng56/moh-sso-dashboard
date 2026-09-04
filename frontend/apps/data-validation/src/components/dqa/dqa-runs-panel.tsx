import {
  Button,
  DataTable,
  Select,
  SelectItem,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow,
} from "@carbon/react";
import { Play, ChevronDown, ChevronUp } from "@carbon/react/icons";
import { useState } from "react";

import { useToast } from "@moh-sso/ui";

import { useListDQARunsQuery, useListDQATablesQuery, useRunDQATableMutation } from "../../api";
import { DQAFlagsPanel } from "./dqa-flags-panel";

const headers = [
  { key: "expand", header: "" },
  { key: "table_id", header: "Table" },
  { key: "run_at", header: "Run at" },
  { key: "rules_evaluated", header: "Rules" },
  { key: "total_flags", header: "Flags" },
  { key: "errors", header: "Fails" },
  { key: "warnings", header: "Warns" },
  { key: "triggered_by", header: "Triggered by" },
];

export function DQARunsPanel() {
  const { data: tables = [] } = useListDQATablesQuery();
  const [selectedTable, setSelectedTable] = useState("");
  const { data: runs = [], isLoading } = useListDQARunsQuery({ tableId: selectedTable || undefined });
  const [runDQATable, { isLoading: isRunning }] = useRunDQATableMutation();
  const [expandedRunId, setExpandedRunId] = useState<number | null>(null);
  const toast = useToast();

  const handleRun = async () => {
    if (!selectedTable) {
      toast.error("Select a table", "Choose a table to scan first.");
      return;
    }
    try {
      const result = await runDQATable(selectedTable).unwrap();
      toast.success(
        "Scan complete",
        `${result.rules_evaluated} rule(s) evaluated, ${result.total_flags} flag(s) (${result.fails} fail, ${result.warns} warn).`,
      );
    } catch {
      toast.error("Scan failed", "Check the table mapping and rule definitions, then try again.");
    }
  };

  const rows = runs.map((r) => ({
    id: String(r.id),
    expand: "",
    table_id: r.table_id,
    run_at: new Date(r.run_at).toLocaleString(),
    rules_evaluated: r.rules_evaluated,
    total_flags: r.total_flags,
    errors: r.errors,
    warnings: r.warnings,
    triggered_by: r.triggered_by || "—",
  }));

  return (
    <div>
      <div style={{ display: "flex", gap: "1rem", alignItems: "flex-end", marginBottom: "1.5rem", flexWrap: "wrap" }}>
        <div style={{ minWidth: "16rem" }}>
          <Select
            id="dqa-runs-table-filter"
            labelText="Table"
            value={selectedTable}
            onChange={(e) => setSelectedTable(e.target.value)}
          >
            <SelectItem value="" text="All tables" />
            {tables.map((t) => (
              <SelectItem key={t.table_id} value={t.table_id} text={t.table_id} />
            ))}
          </Select>
        </div>
        <Button renderIcon={Play} onClick={handleRun} disabled={isRunning || !selectedTable}>
          {isRunning ? "Running…" : "Run now"}
        </Button>
      </div>

      {isLoading ? (
        <p>Loading…</p>
      ) : (
        <DataTable rows={rows} headers={headers}>
          {({ rows, headers, getHeaderProps, getRowProps, getTableProps }) => (
            <TableContainer>
              <Table {...getTableProps()} size="sm">
                <TableHead>
                  <TableRow>
                    {headers.map((header) => {
                      const { key, ...headerProps } = getHeaderProps({ header });
                      return (
                        <TableHeader key={key} {...headerProps}>
                          {header.header}
                        </TableHeader>
                      );
                    })}
                  </TableRow>
                </TableHead>
                <TableBody>
                  {rows.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={headers.length}>No runs yet.</TableCell>
                    </TableRow>
                  ) : (
                    rows.map((row) => {
                      const { key, ...rowProps } = getRowProps({ row });
                      const runId = Number(row.id);
                      const expanded = expandedRunId === runId;
                      return (
                        <>
                          <TableRow
                            key={key}
                            {...rowProps}
                            style={{ cursor: "pointer" }}
                            onClick={() => setExpandedRunId(expanded ? null : runId)}
                          >
                            {row.cells.map((cell) =>
                              cell.info.header === "expand" ? (
                                <TableCell key={cell.id}>
                                  {expanded ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
                                </TableCell>
                              ) : (
                                <TableCell key={cell.id}>{cell.value}</TableCell>
                              ),
                            )}
                          </TableRow>
                          {expanded ? (
                            <TableRow key={`${key}-flags`}>
                              <TableCell colSpan={headers.length}>
                                <DQAFlagsPanel runId={runId} />
                              </TableCell>
                            </TableRow>
                          ) : null}
                        </>
                      );
                    })
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </DataTable>
      )}
    </div>
  );
}
