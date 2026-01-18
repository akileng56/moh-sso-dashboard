import type { Client } from "../types/client.types";

export const DEFAULT_CLIENT_ID = "__default__";

export const defaultClient: Client = {
  id: DEFAULT_CLIENT_ID,
  clientId: DEFAULT_CLIENT_ID,
  name: "MOH Portal",
  enabled: true,
  publicClient: false,
  redirectUris: [],
  attributes: {
    "ui.icon": "home",
    "ui.home": "/dashboard",
    "ui.sidenav": JSON.stringify([
      {
        id: "dashboard",
        label: "Dashboard",
        path: "/dashboard",
      },
    ]),
  },
};
