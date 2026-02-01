import { SideNav, SideNavItems, SideNavLink, SideNavMenu } from "@carbon/react";
import { useLocation, useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";

import { selectActiveClient } from "../../store/clients/clients.selectors";
import { defaultClient } from "../../store/clients/defaultClient";
import type { Client } from "../../store/types/client.types";

type SideNavItem = {
  id: string;
  label: string;
  path: string;
  permission?: string;
};

function parseSideNav(client: Client): SideNavItem[] {
  try {
    return JSON.parse(client.attributes?.["ui.sidenav"] ?? "[]");
  } catch {
    return [];
  }
}

export function ClientSideNav() {
  const activeClient = useSelector(selectActiveClient);
  const navigate = useNavigate();
  const location = useLocation();

  const defaultItems = parseSideNav(defaultClient);
  const activeItems =
    activeClient && activeClient.clientId !== "__default__"
      ? parseSideNav(activeClient)
      : [];

  return (
    <SideNav isFixedNav expanded aria-label="Application navigation">
      <SideNavItems>
        {/* 🌍 Global / Default */}
        <SideNavMenu title="National Data WareHouse">
          {defaultItems.map((item) => (
            <SideNavLink
              key={item.id}
              isActive={location.pathname.startsWith(item.path)}
              onClick={() => navigate(item.path)}
            >
              {item.label}
            </SideNavLink>
          ))}
        </SideNavMenu>

        {/* 🧩 Active Client */}
        {activeItems.length > 0 && (
          <SideNavMenu title={activeClient?.name ?? ""}>
            {activeItems.map((item) => (
              <SideNavLink
                key={item.id}
                isActive={location.pathname.startsWith(item.path)}
                onClick={() => navigate(item.path)}
              >
                {item.label}
              </SideNavLink>
            ))}
          </SideNavMenu>
        )}
      </SideNavItems>
    </SideNav>
  );
}
