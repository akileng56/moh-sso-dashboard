import type React from "react";

import {
  Header,
  HeaderName,
  HeaderGlobalBar,
  HeaderGlobalAction,
  SkipToContent,
} from "@carbon/react";
import { Notification, UserAvatarFilled, Logout } from "@carbon/icons-react";
import keycloak from "../../config/keycloak";

import AppMenuAction from "../appmenu/AppMenu.component";

const DashboardHeader: React.FC = () => {
  const handleLogout = () => {
    keycloak.logout({ redirectUri: window.location.origin });
  };
  return (
    <Header aria-label="App Name">
      <SkipToContent />
      <HeaderName href="#" prefix="MOH">
        Intranet
      </HeaderName>

      <HeaderGlobalBar>
        <AppMenuAction />
        <HeaderGlobalAction aria-label="Notifications" tooltipAlignment="end">
          <Notification size={20} />
        </HeaderGlobalAction>
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
