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
  TextInput,
} from "@carbon/react";
import { Add, TrashCan, Renew } from "@carbon/react/icons";
import { useState } from "react";

import { TableStatusTag, useHeaderPanel, useToast } from "@moh-sso/ui";

import {
  useDeleteDQARuleMutation,
  useListDQARulesQuery,
  useListDQATablesQuery,
  useSeedDQABuiltinRulesMutation,
  useUpsertDQARuleMutation,
} from "../../api";
import type { DQARule, DQARuleInput } from "../../dqa.types";
import { DQARuleForm } from "./dqa-rule-form";

const headers = [
  { key: "enabled", header: "" },
  { key: "code", header: "Code" },
  { key: "table_id", header: "Table" },
  { key: "type", header: "Type" },
  { key: "category", header: "Category" },
  { key: "name", header: "Name" },
  { key: "actions", header: "" },
];

export function DQARulesPanel() {
  const { openPanel, closePanel } = useHeaderPanel();
  const toast = useToast();
  const { data: tables = [] } = useListDQATablesQuery();
  const [tableFilter, setTableFilter] = useState("");
  const { data: rules = [], isLoading } = useListDQARulesQuery({ tableId: tableFilter || undefined });
  const [upsertRule, { isLoading: isSaving }] = useUpsertDQARuleMutation();
  const [deleteRule] = useDeleteDQARuleMutation();
  const [seedBuiltinRules, { isLoading: isSeeding }] = useSeedDQABuiltinRulesMutation();

  const tableOptions = tables.map((t) => t.table_id);

  const handleSave = async (input: DQARuleInput) => {
    try {
      await upsertRule(input).unwrap();
      toast.success("Rule saved", `${input.code} is now active on ${input.table_id}.`);
      closePanel();
    } catch {
      toast.error("Rule not saved", "Check the definition JSON and try again.");
    }
  };

  const openCreatePanel = () => {
    openPanel({
      title: "Add DQA rule",
      size: "md",
      content: (
        <DQARuleForm
          tableOptions={tableOptions}
          isSubmitting={isSaving}
          onSubmit={handleSave}
          onClose={closePanel}
        />
      ),
    });
  };

  const openEditPanel = (rule: DQARule) => {
    openPanel({
      title: `Edit rule: ${rule.code}`,
      size: "md",
      content: (
        <DQARuleForm
          tableOptions={tableOptions}
          initialRule={rule}
          isSubmitting={isSaving}
          onSubmit={handleSave}
          onClose={closePanel}
        />
      ),
    });
  };

  const [seedTableId, setSeedTableId] = useState("vht_monthly");
  const [seedPhysicalTable, setSeedPhysicalTable] = useState("");

  const handleSeed = async () => {
    if (!seedPhysicalTable.trim()) {
      toast.error("Physical table required", "Enter the real DWH table the built-in checks should scan.");
      return;
    }
    try {
      const result = await seedBuiltinRules({
        table_id: seedTableId.trim() || "vht_monthly",
        physical_table: seedPhysicalTable.trim(),
      }).unwrap();
      toast.success("Built-in checks seeded", `${result.seeded} rule(s) added, ${result.failed} failed.`);
    } catch {
      toast.error("Seeding failed", "Please try again.");
    }
  };

  const rows = rules.map((r) => ({
    id: `${r.table_id}:${r.code}`,
    enabled: r.enabled,
    code: r.code,
    table_id: r.table_id,
    type: r.type,
    category: r.category,
    name: r.name || "—",
    raw: r,
  }));

  return (
    <div>
      <div
        style={{
          display: "flex",
          gap: "1rem",
          alignItems: "flex-end",
          marginBottom: "1rem",
          flexWrap: "wrap",
          padding: "1rem",
          border: "1px solid var(--cds-border-subtle)",
        }}
      >
        <TextInput id="dqa-seed-table-id" labelText="Seed table_id" value={seedTableId} onChange={(e) => setSeedTableId(e.target.value)} />
        <TextInput
          id="dqa-seed-physical-table"
          labelText="Physical table for the built-in eCHIS checks"
          placeholder="e.g. report.echis_vht_monthly"
          value={seedPhysicalTable}
          onChange={(e) => setSeedPhysicalTable(e.target.value)}
        />
        <Button kind="tertiary" renderIcon={Renew} onClick={handleSeed} disabled={isSeeding}>
          {isSeeding ? "Seeding…" : "Seed built-in eCHIS checks"}
        </Button>
      </div>

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", marginBottom: "1rem" }}>
        <div style={{ minWidth: "16rem" }}>
          <Select id="dqa-rules-table-filter" labelText="Filter by table" value={tableFilter} onChange={(e) => setTableFilter(e.target.value)}>
            <SelectItem value="" text="All tables" />
            {tableOptions.map((t) => (
              <SelectItem key={t} value={t} text={t} />
            ))}
          </Select>
        </div>
        <Button renderIcon={Add} onClick={openCreatePanel}>
          Add rule
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
                      <TableCell colSpan={headers.length}>No rules yet.</TableCell>
                    </TableRow>
                  ) : (
                    rows.map((row) => {
                      const ruleRaw = rules.find((r) => `${r.table_id}:${r.code}` === row.id);
                      const { key, ...rowProps } = getRowProps({ row });
                      return (
                        <TableRow key={key} {...rowProps}>
                          {row.cells.map((cell) => {
                            if (cell.info.header === "enabled") {
                              return (
                                <TableCell key={cell.id}>
                                  <TableStatusTag status={cell.value ? "on" : "off"} kind={cell.value ? "green" : "gray"} />
                                </TableCell>
                              );
                            }
                            if (cell.info.header === "actions") {
                              return (
                                <TableCell key={cell.id}>
                                  <div style={{ display: "flex", gap: "0.5rem" }}>
                                    <Button kind="ghost" size="sm" onClick={() => ruleRaw && openEditPanel(ruleRaw)}>
                                      Edit
                                    </Button>
                                    <Button
                                      kind="ghost"
                                      size="sm"
                                      hasIconOnly
                                      iconDescription="Delete"
                                      renderIcon={TrashCan}
                                      onClick={() => ruleRaw && deleteRule({ tableId: ruleRaw.table_id, code: ruleRaw.code })}
                                    />
                                  </div>
                                </TableCell>
                              );
                            }
                            return <TableCell key={cell.id}>{String(cell.value)}</TableCell>;
                          })}
                        </TableRow>
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
