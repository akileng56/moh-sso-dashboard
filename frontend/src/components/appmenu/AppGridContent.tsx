import React from "react";
import { Grid, Column } from "@carbon/react";
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

const ICON_MAP: Record<string, React.ElementType> = {
  email: Email,
  calendar: Calendar,
  document: Document,
  settings: Menu,
  user: User,
  applications: App,
};

const AppGridContent: React.FC = () => {
  const { data: clients = [], isLoading, isError } = useListClientsQuery();

  if (isLoading) {
    return (
      <Grid narrow style={{ padding: "1rem" }}>
        <Column>Loading applications…</Column>
      </Grid>
    );
  }

  if (isError) {
    return (
      <Grid narrow style={{ padding: "1rem" }}>
        <Column>Failed to load applications.</Column>
      </Grid>
    );
  }

  return (
    <Grid
      narrow
      style={{
        width: "auto",
        padding: "1rem",
        background: "white",
        display: "flex",
        flexWrap: "wrap",
        flexDirection: "row",
        justifyContent: "flex-start",
        gap: "1rem",
      }}
    >
      {clients.map((client: Client) => {
        const Icon = ICON_MAP[(client as any).attributes?.icon] || Menu;

        return (
          <Column key={client.clientId} sm={2} md={2} lg={3}>
            <AppTile
              icon={Icon}
              name={client.name}
              href={client.baseUrl ?? "/"}
            />
          </Column>
        );
      })}
    </Grid>
  );
};

export default AppGridContent;
