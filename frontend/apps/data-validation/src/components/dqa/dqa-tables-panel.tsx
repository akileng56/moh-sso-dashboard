import {
  Button,
  DataTable,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow,
  TextInput,
} from "@carbon/react";
import { TrashCan } from "@carbon/react/icons";
import { useState } from "react";

import { useToast } from "@moh-sso/ui";

import { useDeleteDQATableMutation, useListDQATablesQuery, useUpsertDQATableMutation } from "../../api";

const headers = [
  { key: "table_id", header: "Table ID" },
  { key: "physical_table", header: "Physical table" },
  { key: "description", header: "Description" },
  { key: "actions", header: "" },
];

export function DQATablesPanel() {
  const { data: tables = [], isLoading } = useListDQATablesQuery();
  const [upsertTable, { isLoading: isSaving }] = useUpsertDQATableMutation();
  const [deleteTable] = useDeleteDQATableMutation();
  const toast = useToast();

  const [tableId, setTableId] = useState("");
  const [physicalTable, setPhysicalTable] = useState("");
  const [description, setDescription] = useState("");

  const handleAdd = async () => {
    if (!tableId.trim() || !physicalTable.trim()) {
      toast.error("Missing fields", "Table ID and physical table are required.");
      return;
    }
    try {
      await upsertTable({
        table_id: tableId.trim(),
        physical_table: physicalTable.trim(),
        description: description.trim(),
        is_active: true,
      }).unwrap();
      toast.success("Table registered", `${tableId} now maps to ${physicalTable}.`);
      setTableId("");
      setPhysicalTable("");
      setDescription("");
    } catch {
      toast.error("Could not save table mapping", "Please try again.");
    }
  };

  const rows = tables.map((t) => ({
    id: t.table_id,
    table_id: t.table_id,
    physical_table: t.physical_table,
    description: t.description || "—",
  }));

  return (
    <div>
      <p style={{ marginBottom: "1rem", color: "var(--cds-text-secondary)" }}>
        Register a logical table_id used by rules to the real DWH table it scans, e.g.{" "}
        <code>vht_monthly</code> → <code>report.echis_vht_monthly</code>.
      </p>

      <div style={{ display: "flex", gap: "1rem", alignItems: "flex-end", marginBottom: "1.5rem", flexWrap: "wrap" }}>
        <TextInput id="dqa-new-table-id" labelText="Table ID" value={tableId} onChange={(e) => setTableId(e.target.value)} />
        <TextInput
          id="dqa-new-physical-table"
          labelText="Physical table (schema.table)"
          value={physicalTable}
          onChange={(e) => setPhysicalTable(e.target.value)}
        />
        <TextInput
          id="dqa-new-table-desc"
          labelText="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
        <Button onClick={handleAdd} disabled={isSaving}>
          {isSaving ? "Saving…" : "Register table"}
        </Button>
      </div>

      {isLoading ? (
        <p>Loading…</p>
      ) : (
        <DataTable rows={rows} headers={headers}>
          {({ rows, headers, getHeaderProps, getRowProps, getTableProps }) => (
            <TableContainer>
              <Table {...getTableProps()}>
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
                      <TableCell colSpan={headers.length}>No tables registered yet.</TableCell>
                    </TableRow>
                  ) : (
                    rows.map((row) => {
                      const { key, ...rowProps } = getRowProps({ row });
                      return (
                        <TableRow key={key} {...rowProps}>
                          {row.cells.map((cell) =>
                            cell.info.header === "actions" ? (
                              <TableCell key={cell.id}>
                                <Button
                                  kind="ghost"
                                  size="sm"
                                  hasIconOnly
                                  iconDescription="Remove"
                                  renderIcon={TrashCan}
                                  onClick={() => deleteTable(row.id)}
                                />
                              </TableCell>
                            ) : (
                              <TableCell key={cell.id}>{cell.value}</TableCell>
                            ),
                          )}
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
