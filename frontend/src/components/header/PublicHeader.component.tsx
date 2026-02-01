import type React from "react";
import {
  Header,
  HeaderName,
  HeaderGlobalBar,
  HeaderGlobalAction,
  SkipToContent,
} from "@carbon/react";
import { UserAvatarFilled, Logout } from "@carbon/icons-react";
import { useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";

import { selectUser } from "../../store/auth/auth.selectors";
import AppMenuAction from "../appmenu/AppMenu.component";
import { API } from "../../lib/constants/api.constants";
import { EnvironmentBadge } from "../../env/EnvironmentBadge";

const PublicHeader: React.FC = () => {
  const navigate = useNavigate();
  const user = useSelector(selectUser);

  const handleLogout = () => {
    window.location.replace(API.auth.logout());
  };
  return (
    <Header aria-label="MOH Integrated Health Portal">
      <SkipToContent />

      {/* Brand / Home */}
      <HeaderName
        prefix="MOH"
        onClick={() => navigate("/dashboard")}
        style={{ cursor: "pointer" }}
      >
        Integrated Health Portal
      </HeaderName>

      {/* Global Actions */}
      <HeaderGlobalBar>
        {/* 🌱 Environment */}
        {/*<EnvironmentBadge />*/}
        {/* App Launcher */}
        <AppMenuAction />
        {/* User indicator */}
        <HeaderGlobalAction
          aria-label={`Signed in as ${user?.username ?? "user"}`}
          tooltipAlignment="end"
        >
          <UserAvatarFilled size={20} />
        </HeaderGlobalAction>
        {/* Logout */}
        <HeaderGlobalAction
          aria-label="Logout"
          tooltipAlignment="end"
          onClick={() => handleLogout()}
        >
          <Logout size={20} />
        </HeaderGlobalAction>
      </HeaderGlobalBar>
    </Header>
  );
};

export default PublicHeader;
