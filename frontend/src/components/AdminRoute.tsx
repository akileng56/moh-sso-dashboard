import { Navigate } from "react-router-dom";
import { useSelector } from "react-redux";
import { selectIsAdmin } from "../store/auth/auth.selectors";
import type { JSX } from "react";

export const AdminRoute = ({ children }: { children: JSX.Element }) => {
  const isAdmin = useSelector(selectIsAdmin);

  if (!isAdmin) {
    return <Navigate to="/dashboard" replace />;
  }

  return children;
};
