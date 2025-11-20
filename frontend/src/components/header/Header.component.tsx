import type React from "react";

import {
  Header,
  HeaderName,
  HeaderGlobalBar,
  HeaderGlobalAction,
  SkipToContent,
} from "@carbon/react";
import { UserAvatarFilled, Logout } from "@carbon/icons-react";

import AppMenuAction from "../appmenu/AppMenu.component";

const API_BASE = "http://localhost:9000/api/v1/auth";

const DashboardHeader: React.FC = () => {
  const handleLogout = () => {
    window.location.href = `${API_BASE}/logout`;
  };
  return (
    <Header aria-label="App Name">
      <SkipToContent />
      <HeaderName href="#" prefix="MOH">
        Intranet
      </HeaderName>

      <HeaderGlobalBar>
        <AppMenuAction />
        <HeaderGlobalAction aria-label="User Avatar" tooltipAlignment="end">
          <UserAvatarFilled size={20} />
        </HeaderGlobalAction>
        <HeaderGlobalAction
          aria-label="Logout"
          tooltipAlignment="end"
          onClick={handleLogout}
        >
          <Logout size={20} />
        </HeaderGlobalAction>
      </HeaderGlobalBar>
    </Header>
  );
};

export default DashboardHeader;
