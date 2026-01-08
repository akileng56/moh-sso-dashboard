import { Tile, Tag } from "@carbon/react";

export function SignalTile({
  label,
  value,
  severity,
}: {
  label: string;
  value: number;
  severity?: "warning" | "danger";
}) {
  return (
    <Tile>
      <span>{label}</span>
      <h3>{value}</h3>
      {severity && (
        <Tag type={severity === "danger" ? "red" : "yellow"}>{severity}</Tag>
      )}
    </Tile>
  );
}
