import { Dropdown, Button } from "@carbon/react";

interface UserFiltersProps {
  status: string;
  roles: string[];
  selectedRole: string;
  neverLoggedIn: boolean;
  onStatusChange: (v: string) => void;
  onRoleChange: (v: string) => void;
  onToggleNeverLoggedIn: () => void;
}

export function UserFilters({
  status,
  roles,
  selectedRole,
  neverLoggedIn,
  onStatusChange,
  onRoleChange,
  onToggleNeverLoggedIn,
}: UserFiltersProps) {
  return (
    <div style={{ display: "flex", gap: "1rem", flexWrap: "wrap" }}>
      <Dropdown
        label="Status Filter"
        id="status-filter"
        titleText="Status"
        items={["all", "active", "disabled"]}
        selectedItem={status}
        onChange={({ selectedItem }) => onStatusChange(selectedItem as string)}
      />

      <Dropdown
        label="Role Filter"
        id="role-filter"
        titleText="Role"
        items={roles}
        selectedItem={selectedRole}
        onChange={({ selectedItem }) => onRoleChange(selectedItem as string)}
      />

      <Button
        size="sm"
        kind={neverLoggedIn ? "primary" : "secondary"}
        onClick={onToggleNeverLoggedIn}
      >
        Never logged in
      </Button>
    </div>
  );
}
