import { Navigate } from "react-router-dom";
import { useSelector } from "react-redux";
import { InlineLoading } from "@carbon/react";

import {
  selectAuthenticated,
  selectAuthLoaded,
  selectUser,
  selectIsAdmin,
} from "../store/auth/auth.selectors";
import type { JSX } from "react";

export const AdminRoute = ({ children }: { children: JSX.Element }) => {
  const loaded = useSelector(selectAuthLoaded);
  const authenticated = useSelector(selectAuthenticated);
  const user = useSelector(selectUser);
  const isAdmin = useSelector(selectIsAdmin);

  if (!loaded) {
    return <InlineLoading description="Checking permissions…" />;
  }

  if (!authenticated || !user) {
    return <Navigate to="/login" replace />;
  }

  if (!isAdmin) {
    return <Navigate to="/dashboard" replace />;
  }

  return children;
};
