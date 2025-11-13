import type React from "react";

import {
  Header,
  HeaderName,
  HeaderGlobalBar,
  HeaderGlobalAction,
  SkipToContent,
} from "@carbon/react";
import { Notification, UserAvatarFilled } from "@carbon/icons-react";

import AppMenuAction from "../appmenu/AppMenu.component";

const DashboardHeader: React.FC = () => (
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
    </HeaderGlobalBar>
  </Header>
);

export default DashboardHeader;
