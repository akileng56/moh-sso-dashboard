import type React from "react";
import {
  Header,
  HeaderName,
  HeaderGlobalBar,
  HeaderGlobalAction,
  SkipToContent,
} from "@carbon/react";
import { UserAvatarFilled, Logout } from "@carbon/icons-react";
import { useAuth } from "../../context/useAuth";
import AppMenuAction from "../appmenu/AppMenu.component";

const PublicHeader: React.FC = () => {
  const { logout, user } = useAuth();

  return (
    <Header aria-label="MOH Integrated Health Portal">
      <SkipToContent />

      <HeaderName prefix="MOH" href="/dashboard">
        Integrated Health Portal
      </HeaderName>

      <HeaderGlobalBar>
        <HeaderGlobalAction aria-label="User">
          <UserAvatarFilled size={20} />
          <span style={{ marginLeft: 8 }}>{user?.username}</span>
        </HeaderGlobalAction>

        <HeaderGlobalAction aria-label="Logout" onClick={logout}>
          <Logout size={20} />
        </HeaderGlobalAction>

        <AppMenuAction />
      </HeaderGlobalBar>
    </Header>
  );
};

export default PublicHeader;
