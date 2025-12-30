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
import { useAuth } from "../../context/useAuth";
import AppMenuAction from "../appmenu/AppMenu.component";

const PublicHeader: React.FC = () => {
  const { logout, user } = useAuth();
  const navigate = useNavigate();

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
          onClick={logout}
        >
          <Logout size={20} />
        </HeaderGlobalAction>
      </HeaderGlobalBar>
    </Header>
  );
};

export default PublicHeader;
