import React, { useEffect, useMemo, useState } from "react";
import {
  DataTable,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TableExpandHeader,
  TableExpandRow,
  TableExpandedRow,
} from "@carbon/react";
import { DataTablePagination } from "@moh-sso/ui";
import IssueDetail from "../../../../../issue-tracker/src/pages/issue-detail/issue-detail.component.tsx";


interface ListProps {
  columns: any;
  data: any;
  selectedIssue?: any;
  closeView: () => void;
  handleIssueClick: (row: any) => void;
  totalItems?: number;
  currentPage?: number;
  currentPageSize?: number;
  onPageChange?: (page: number, pageSize: number) => void;
}

const DataList: React.FC<ListProps> = ({
  columns,
  data,
  handleIssueClick,
  closeView,
  totalItems,
  currentPage,
  currentPageSize,
  onPageChange,
}) => {
  /* -----------------------------
   * Pagination
   * ----------------------------- */
  const [localPage, setLocalPage] = useState(1);
  const [localPageSize, setLocalPageSize] = useState(10);

  const isServerSide = onPageChange !== undefined;

  const page = isServerSide ? (currentPage ?? 1) : localPage;
  const pageSize = isServerSide ? (currentPageSize ?? 10) : localPageSize;

  /* -----------------------------
   * Pagination slice
   * ----------------------------- */
  const paginatedData = useMemo(() => {
    if (isServerSide) {
      return data;
    }
    const start = (page - 1) * pageSize;
    const end = start + pageSize;
    return data.slice(start, end);
  }, [data, page, pageSize, isServerSide]);

  /* -----------------------------
   * Clamp page when data shrinks
   * ----------------------------- */
  useEffect(() => {
    if (isServerSide) return;
    const lastPage = Math.ceil(data.length / pageSize) || 1;
    if (page > lastPage) {
      setLocalPage(lastPage);
    }
  }, [data.length, pageSize, page, isServerSide]);

  const total = isServerSide ? (totalItems ?? 0) : data.length;

  return (
    <>
      <DataTable rows={paginatedData} headers={columns}>
        {({
            rows,
            headers,
            getTableProps,
            getHeaderProps,
            getRowProps,
            getExpandHeaderProps,
          }) => (
            <Table {...getTableProps()}>
              <TableHead>
                <TableRow>
                  <TableExpandHeader {...getExpandHeaderProps()} />
                  {headers.map((header) => {
                    const { key: headerKey, ...headerProps } = getHeaderProps({ header });
                    return (
                      <TableHeader key={headerKey} {...headerProps}>
                        {header.header}
                      </TableHeader>
                    );
                  })}
                </TableRow>
              </TableHead>
              <TableBody>
                {rows.map((row) => {
                  const issue = data.find((item: any) => item.id === row.id);
                  const { key: rowKey, ...rowProps } = getRowProps({ row });
                  return (
                      <React.Fragment key={row.id}>
                        <TableExpandRow key={rowKey} {...rowProps}>
                          {row.cells.map((cell) => {
                            if (cell.info.header === "issue") {
                              return (
                                  <TableCell key={cell.id}>
                                    <button
                                        type="button"
                                        className="issue-clickable-cell"
                                        onClick={() => handleIssueClick(row)}
                                    >
                                      {cell.value}
                                    </button>
                                  </TableCell>
                              );
                            }
                            return <TableCell key={cell.id}>{cell.value}</TableCell>;
                          })}
                        </TableExpandRow>
                        <TableExpandedRow colSpan={headers.length + 1}>
                          {row.isExpanded && (
                              <section>
                                <IssueDetail selectedIssue={issue} goToBack={closeView} showBack={false}/>
                              </section>
                          )}
                        </TableExpandedRow>
                      </React.Fragment>
                  )
                })}
              </TableBody>
            </Table>
        )}
      </DataTable>

      {/* ================= PAGINATION ================= */}
      <DataTablePagination
        page={page}
        pageSize={pageSize}
        totalItems={total}
        onChange={({ page: newPage, pageSize: newPageSize }) => {
          if (isServerSide) {
            onPageChange(newPage, newPageSize);
          } else {
            setLocalPage(newPage);
            setLocalPageSize(newPageSize);
          }
        }}
      />
    </>
  );
};

export default DataList;