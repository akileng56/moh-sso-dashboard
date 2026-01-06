import { Navigate, Outlet } from "react-router-dom";
import { useSelector } from "react-redux";
import {
  selectAuthenticated,
  selectAuthLoading,
} from "../store/auth/auth.selectors";
import type { JSX } from "react";

export const ProtectedRoute = ({ children }: { children?: JSX.Element }) => {
  const authenticated = useSelector(selectAuthenticated);
  const loading = useSelector(selectAuthLoading);

  if (loading) {
    return <p>Authenticating…</p>; // or spinner
  }

  if (!authenticated) {
    window.location.href = "http://localhost:9000/api/v1/auth/login";
    return null;
  }

  return children ? children : <Outlet />;
};
