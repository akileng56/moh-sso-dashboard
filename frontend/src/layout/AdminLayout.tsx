import {
  Header,
  HeaderName,
  HeaderGlobalBar,
  HeaderGlobalAction,
  SideNav,
  SideNavItems,
  SideNavLink,
  Content,
} from "@carbon/react";
import { Logout, UserAvatarFilled } from "@carbon/icons-react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../context/useAuth";

export default function AdminLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuth();

  return (
    <>
      {/* Top Header */}
      <Header aria-label="MOH Integrated Health Portal">
        <HeaderName
          prefix="MOH"
          onClick={() => navigate("/admin")}
          style={{ cursor: "pointer" }}
        >
          Integrated Health Portal
        </HeaderName>

        {/* Global actions (icons only) */}
        <HeaderGlobalBar>
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

      {/* Side Navigation */}
      <SideNav isFixedNav expanded aria-label="Admin navigation">
        <SideNavItems>
          <SideNavLink
            isActive={location.pathname === "/admin"}
            onClick={() => navigate("/admin")}
          >
            Dashboard
          </SideNavLink>

          <SideNavLink
            isActive={location.pathname.startsWith("/admin/audit-logs")}
            onClick={() => navigate("/admin/audit-logs")}
          >
            Audit Logs
          </SideNavLink>

          <SideNavLink
            isActive={location.pathname.startsWith("/admin/users")}
            onClick={() => navigate("/admin/users")}
          >
            Users
          </SideNavLink>

          <SideNavLink
            isActive={location.pathname.startsWith("/admin/clients")}
            onClick={() => navigate("/admin/clients")}
          >
            Clients
          </SideNavLink>
        </SideNavItems>
      </SideNav>

      {/* Page Content */}
      <Content style={{ marginLeft: 256, paddingTop: "3rem" }}>
        <Outlet />
      </Content>
    </>
  );
}
