import React from "react";
import { Grid, Column, InlineLoading, Tile } from "@carbon/react";
import {
  Document,
  Calendar,
  Email,
  Menu,
  User,
  App,
} from "@carbon/icons-react";

import AppTile from "./AppMenuItem.component";
import { useListClientsQuery } from "../../store/api/clients.api";
import type { Client } from "../../store/types/client.types";
import "./AppMenu.css";

const ICON_MAP: Record<string, React.ElementType> = {
  email: Email,
  calendar: Calendar,
  document: Document,
  settings: Menu,
  user: User,
  applications: App,
};

const AppGridContent: React.FC = () => {
  const {
    data: clients = [],
    isLoading,
    isError,
    refetch,
  } = useListClientsQuery();

  /* -----------------------------
   * Loading state
   * ----------------------------- */
  if (isLoading) {
    return (
      <div className="app-grid-state">
        <InlineLoading description="Loading applications…" />
      </div>
    );
  }

  /* -----------------------------
   * Error state
   * ----------------------------- */
  if (isError) {
    return (
      <div className="app-grid-state">
        <Tile>
          <p style={{ marginBottom: "0.5rem" }}>Failed to load applications.</p>
          <button
            type="button"
            className="app-grid-retry"
            onClick={() => refetch()}
          >
            Retry
          </button>
        </Tile>
      </div>
    );
  }

  /* -----------------------------
   * Empty state
   * ----------------------------- */
  if (clients.length === 0) {
    return (
      <div className="app-grid-state">
        <p style={{ opacity: 0.7 }}>No applications available.</p>
      </div>
    );
  }

  return (
    <Grid narrow className="app-grid">
      {clients.map((client: Client) => {
        const Icon = ICON_MAP[(client as any).attributes?.icon] ?? App;

        return (
          <Column
            key={client.client_id}
            sm={2}
            md={2}
            lg={3}
            className="app-grid__column"
          >
            <div className="app-tile-wrapper">
              <AppTile
                icon={Icon}
                name={client.name}
                href={client.baseUrl ?? "/"}
              />
            </div>
          </Column>
        );
      })}
    </Grid>
  );
};

export default AppGridContent;
