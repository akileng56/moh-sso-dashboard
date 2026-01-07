import {
  HeaderPanel,
  HeaderGlobalAction,
  Content,
  Header,
  HeaderGlobalBar,
  HeaderName,
  SideNav,
  SideNavItems,
  SideNavLink,
} from "@carbon/react";
import { Notification, Logout, UserAvatarFilled } from "@carbon/icons-react";
import { useState } from "react";
import { useSelector } from "react-redux";
import { useNavigate, useLocation, Outlet } from "react-router-dom";
import { useLogoutMutation } from "../store/api/auth.api";
import { selectUser } from "../store/auth/auth.selectors";
import { NotificationsPanel } from "../components/notifications/notifications-panel.component";
import {
  useGetNotificationsQuery,
  useGetUnreadNotificationsCountQuery,
} from "../store/api/notifications.api";

export default function AdminLayout() {
  const navigate = useNavigate();
  const location = useLocation();

  const user = useSelector(selectUser);
  const [logout] = useLogoutMutation();

  const [showNotifications, setShowNotifications] = useState(false);

  const { data: notifications = [] } = useGetNotificationsQuery({
    unread: false,
    limit: 10,
    offset: 0,
  });

  const { data: unreadCount = 0 } = useGetUnreadNotificationsCountQuery();

  return (
    <>
      <Header aria-label="MOH Integrated Health Portal">
        <HeaderName
          prefix="MOH"
          onClick={() => navigate("/admin")}
          style={{ cursor: "pointer" }}
        >
          Integrated Health Portal
        </HeaderName>

        <HeaderGlobalBar>
          {/* 🔔 Notifications */}
          <HeaderGlobalAction
            aria-label="Notifications"
            tooltipAlignment="end"
            onClick={() => setShowNotifications(!showNotifications)}
          >
            <Notification size={20} />
            {unreadCount > 0 && (
              <span className="notification-badge">{unreadCount}</span>
            )}
          </HeaderGlobalAction>

          {/* User */}
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

        {/* 🔔 Notifications Panel */}
        {showNotifications && (
          <HeaderPanel expanded>
            <NotificationsPanel
              notifications={notifications}
              onClose={() => setShowNotifications(false)}
            />
          </HeaderPanel>
        )}
      </Header>

      {/* SideNav unchanged */}
      <SideNav isFixedNav expanded aria-label="Admin navigation">
        <SideNavItems>
          <SideNavLink
            isActive={location.pathname === "/admin"}
            onClick={() => navigate("/admin")}
          >
            Home
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
          <SideNavLink
            isActive={location.pathname.startsWith("/admin/audit-logs")}
            onClick={() => navigate("/admin/audit-logs")}
          >
            Audits
          </SideNavLink>
        </SideNavItems>
      </SideNav>

      <Content style={{ marginLeft: 256, paddingTop: "3rem" }}>
        <Outlet />
      </Content>
    </>
  );
}
