import { ContentSwitcher, Switch } from "@carbon/react";
import { useState } from "react";

import { DQARulesPanel } from "./dqa-rules-panel";
import { DQARunsPanel } from "./dqa-runs-panel";
import { DQATablesPanel } from "./dqa-tables-panel";

type View = "tables" | "rules" | "runs";

export function DQAEnginePanel() {
  const [view, setView] = useState<View>("tables");

  return (
    <div className="dqa-engine-panel">
      <p style={{ marginBottom: "1rem", color: "var(--cds-text-secondary)" }}>
        The DQA engine compiles declarative rules to SQL and runs them directly against the DWH.
        Register a table, add or seed rules, then trigger a scan.
      </p>

      <ContentSwitcher
        selectedIndex={["tables", "rules", "runs"].indexOf(view)}
        onChange={({ index }) => setView((["tables", "rules", "runs"] as View[])[index ?? 0])}
        style={{ marginBottom: "1.5rem", maxWidth: "24rem" }}
      >
        <Switch name="tables" text="Tables" />
        <Switch name="rules" text="Rules" />
        <Switch name="runs" text="Runs & flags" />
      </ContentSwitcher>

      {view === "tables" ? <DQATablesPanel /> : null}
      {view === "rules" ? <DQARulesPanel /> : null}
      {view === "runs" ? <DQARunsPanel /> : null}
    </div>
  );
}
