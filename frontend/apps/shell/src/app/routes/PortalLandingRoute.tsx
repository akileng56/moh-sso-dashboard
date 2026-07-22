import { InlineLoading } from "@carbon/react";
import { useSelector } from "react-redux";
import { Navigate } from "react-router-dom";

import {
  PERMISSIONS,
  selectAuthenticated,
  selectAuthLoaded,
  selectAuthLoading,
  selectUser,
} from "@moh-sso/auth";

import NewsFeedPage from "@/app/newsfeed/pages/news_feed.component";

export function PortalLandingRoute() {
  const authenticated = useSelector(selectAuthenticated);
  const loaded = useSelector(selectAuthLoaded);
  const loading = useSelector(selectAuthLoading);
  const user = useSelector(selectUser);

  if (!loaded || loading) {
    return <InlineLoading description="Checking session..." />;
  }

  if (!authenticated || !user) {
    return <NewsFeedPage />;
  }

  if (user.isAdmin) {
    return <Navigate to="/admin/home" replace />;
  }

  if (user.permissions.includes(PERMISSIONS.portalAccess)) {
    return <Navigate to="/apps/news" replace />;
  }

  return <NewsFeedPage />;
}
