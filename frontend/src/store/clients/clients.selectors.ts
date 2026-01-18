import { createSelector } from "@reduxjs/toolkit";
import { defaultClient } from "./defaultClient";
import type { RootState } from "..";

/* -----------------------------
 * Base selectors
 * ----------------------------- */

export const selectClientsState = (state: RootState) => state.clients;

export const selectClients = createSelector(
  selectClientsState,
  (s) => [defaultClient, ...s.items] // 👈 ALWAYS prepend
);

export const selectActiveClientId = createSelector(
  selectClientsState,
  (s) => s.activeClientId
);

export const selectActiveClient = createSelector(
  [selectClients, selectActiveClientId],
  (clients, activeId) => clients.find((c) => c.clientId === activeId) ?? null
);

const parseJSON = <T>(value?: string): T | null => {
  if (!value) return null;
  try {
    return JSON.parse(value) as T;
  } catch {
    return null;
  }
};

export type SideNavItem = {
  id: string;
  label: string;
  path: string;
  permission?: string;
};

export const selectClientSideNav = createSelector(
  selectActiveClient,
  (client): SideNavItem[] => {
    const raw = client?.attributes?.["ui.sidenav"];
    return parseJSON<SideNavItem[]>(raw) ?? [];
  }
);

export const selectClientIcon = createSelector(
  selectActiveClient,
  (client) => client?.attributes?.["ui.icon"] ?? null
);

export const selectClientHome = createSelector(
  selectActiveClient,
  (client) => client?.attributes?.["ui.home"] ?? "/"
);
