import {
  Button,
  ComboBox,
  DatePicker,
  DatePickerInput,
  Search,
} from "@carbon/react";

type SuccessFilter = "true" | "false";

interface AuditLogFiltersProps {
  from: string;
  to: string;
  onFromChange: (v: string) => void;
  onToChange: (v: string) => void;
  onClientChange: (v?: string) => void;
  onActionChange: (v?: string) => void;
  onSuccessChange: (v?: SuccessFilter) => void;
  onClear: () => void;
  onExportCsv: () => void;
  onExportJson: () => void;
}

function toRFC3339(d: Date) {
  return d.toISOString();
}

export function AuditLogFilters({
  onFromChange,
  onToChange,
  onClientChange,
  onActionChange,
  onSuccessChange,
  onClear,
  onExportCsv,
  onExportJson,
}: AuditLogFiltersProps) {
  return (
    <div style={{ display: "flex", gap: 12, flexWrap: "wrap" }}>
      <DatePicker
        datePickerType="range"
        onChange={(dates) => {
          if (Array.isArray(dates)) {
            if (dates[0] instanceof Date) onFromChange(toRFC3339(dates[0]));
            if (dates[1] instanceof Date) onToChange(toRFC3339(dates[1]));
          }
        }}
      >
        <DatePickerInput id="from" labelText="From" />
        <DatePickerInput id="to" labelText="To" />
      </DatePicker>

      <Search
        labelText="Client ID"
        onChange={(e) => onClientChange(e.target.value || undefined)}
      />

      <Search
        labelText="Action"
        onChange={(e) => onActionChange(e.target.value || undefined)}
      />

      <ComboBox
        id="success"
        titleText="Result"
        items={[
          { id: "true", label: "Success" },
          { id: "false", label: "Failure" },
        ]}
        itemToString={(item) => item?.label ?? ""}
        onChange={({ selectedItem }) =>
          onSuccessChange(selectedItem?.id as SuccessFilter | undefined)
        }
      />

      <Button kind="secondary" onClick={onClear}>
        Clear filters
      </Button>

      <Button kind="primary" onClick={onExportCsv}>
        Export CSV
      </Button>

      <Button kind="tertiary" onClick={onExportJson}>
        Export JSON
      </Button>
    </div>
  );
}
