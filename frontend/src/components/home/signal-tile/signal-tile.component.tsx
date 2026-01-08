import { Tile, Tag, Stack } from "@carbon/react";

type Severity = "success" | "warning" | "danger";

type Props = {
  label: string;
  value: number | string;
  severity?: Severity;
  helperText?: string;
};

export function SignalTile({ label, value, severity, helperText }: Props) {
  return (
    <Tile
      className={`signal-tile ${severity ? `signal-tile--${severity}` : ""}`}
    >
      <Stack gap={2}>
        <span className="signal-tile__label">{label}</span>

        <div className="signal-tile__value">{value}</div>

        {(severity || helperText) && (
          <Stack orientation="horizontal" gap={2}>
            {severity && (
              <Tag size="sm" type={mapSeverity(severity)}>
                {severity}
              </Tag>
            )}

            {helperText && (
              <span className="signal-tile__helper">{helperText}</span>
            )}
          </Stack>
        )}
      </Stack>
    </Tile>
  );
}

/* -----------------------------
 * Severity → Carbon mapping
 * ----------------------------- */
function mapSeverity(severity: Severity): "green" | "yellow" | "red" {
  switch (severity) {
    case "success":
      return "green";
    case "warning":
      return "yellow";
    case "danger":
      return "red";
  }
}
