import { Outlet } from "react-router-dom";
import { useSelector } from "react-redux";
import { InlineLoading } from "@carbon/react";
import {
  selectAuthenticated,
  selectAuthLoaded,
} from "../store/auth/auth.selectors";
import { API } from "../lib/constants/api.constants";
import type { JSX } from "react";

export const ProtectedRoute = ({ children }: { children?: JSX.Element }) => {
  const authenticated = useSelector(selectAuthenticated);
  const loaded = useSelector(selectAuthLoaded);

  if (!loaded) {
    return <InlineLoading description="Checking session…" />;
  }

  if (!authenticated) {
    window.location.replace(API.auth.login());
    return null;
  }

  return children ? children : <Outlet />;
};
