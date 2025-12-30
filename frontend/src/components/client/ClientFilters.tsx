import { Dropdown } from "@carbon/react";

type StatusFilter = "all" | "enabled" | "disabled";
type TypeFilter = "all" | "public" | "confidential";

interface Option {
  id: string;
  label: string;
}

interface ClientFiltersProps {
  status: StatusFilter;
  type: TypeFilter;
  statusOptions: readonly Option[];
  typeOptions: readonly Option[];
  onStatusChange: (v: StatusFilter) => void;
  onTypeChange: (v: TypeFilter) => void;
}

export function ClientFilters({
  status,
  type,
  statusOptions,
  typeOptions,
  onStatusChange,
  onTypeChange,
}: ClientFiltersProps) {
  return (
    <div style={{ display: "flex", gap: "1rem", flexWrap: "wrap" }}>
      <Dropdown
        id="client-status-filter"
        titleText="Status"
        items={statusOptions}
        selectedItem={statusOptions.find((i) => i.id === status)}
        itemToString={(item) => item?.label ?? ""}
        onChange={({ selectedItem }) =>
          onStatusChange(selectedItem?.id as StatusFilter)
        }
      />

      <Dropdown
        id="client-type-filter"
        titleText="Type"
        items={typeOptions}
        selectedItem={typeOptions.find((i) => i.id === type)}
        itemToString={(item) => item?.label ?? ""}
        onChange={({ selectedItem }) =>
          onTypeChange(selectedItem?.id as TypeFilter)
        }
      />
    </div>
  );
}
