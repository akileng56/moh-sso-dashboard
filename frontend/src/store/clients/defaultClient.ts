import type { Client } from "../types/client.types";

export const DEFAULT_CLIENT_ID = "__default__";

export const defaultClient: Client = {
  id: DEFAULT_CLIENT_ID,
  clientId: DEFAULT_CLIENT_ID,
  name: "National Data Warehouse",
  enabled: true,
  publicClient: false,

  // Redirects are handled centrally (auth callback + router)
  redirectUris: [],

  attributes: {
    // UI metadata
    "ui.icon": "home",

    // User landing entry (NOT a deep link)
    "ui.home": "/apps/dwh",

    // UserLayout sidenav
    "ui.sidenav": JSON.stringify([
      {
        id: "data-visualizer",
        label: "Data Visualizer",
        path: "/apps/dwh/data-visualizer",
      },
    ]),
  },
};
