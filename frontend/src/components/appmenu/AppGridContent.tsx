import React, { useEffect, useState } from "react";
import { Grid, Column } from "@carbon/react";
import AppTile from "./AppMenuItem.component";
import { useAuth } from "../../context/useAuth";

import {
  Document,
  Calendar,
  Email,
  Menu,
  User,
  App,
} from "@carbon/icons-react";

const API_BASE = "http://localhost:9000/api/v1/clients";

const ICON_MAP: Record<string, React.ElementType> = {
  email: Email,
  calendar: Calendar,
  document: Document,
  settings: Menu,
  user: User,
  applications: App,
};

const AppGridContent: React.FC = () => {
  const { accessToken } = useAuth();
  const [clients, setClients] = useState<any[]>([]);

  useEffect(() => {
    if (!accessToken) return;

    const controller = new AbortController();

    const fetchClients = async () => {
      try {
        const res = await fetch(`${API_BASE}/`, {
          method: "GET",
          credentials: "include",
          headers: {
            Authorization: `Bearer ${accessToken}`,
            "Content-Type": "application/json",
          },
          signal: controller.signal,
        });

        if (!res.ok) {
          console.error("Failed to fetch clients");
          return;
        }

        const data = await res.json();
        setClients(data);
      } catch (error: any) {
        console.error("Error fetching clients:", error);
      }
    };

    fetchClients();

    return () => controller.abort();
  }, [accessToken]);

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
      {clients.map((client) => {
        const Icon = ICON_MAP[client.attributes.icon] || Menu;

        return (
          <Column key={client.id} sm={2} md={2} lg={3}>
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
