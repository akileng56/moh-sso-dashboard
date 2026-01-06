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
import { useLogoutMutation } from "../../store/api/auth.api";
import AppMenuAction from "../appmenu/AppMenu.component";

const PublicHeader: React.FC = () => {
  const navigate = useNavigate();
  const user = useSelector(selectUser);
  const [logout] = useLogoutMutation();

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
          onClick={() => logout()}
        >
          <Logout size={20} />
        </HeaderGlobalAction>
      </HeaderGlobalBar>
    </Header>
  );
};

export default PublicHeader;
