import { Navigate } from "react-router-dom";
import { useAuth } from "./context/useAuth";
import type { JSX } from "react";

export const ProtectedRoute = ({ children }: { children: JSX.Element }) => {
  const { authenticated } = useAuth();

  if (!authenticated) {
    return <Navigate to="http://localhost:9000/api/v1/auth/login" />;
  }

  return children;
};
