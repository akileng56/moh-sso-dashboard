import { useAuth } from "../context/useAuth";
import type { JSX } from "react";

export function ProtectedRoute({ children }: { children: JSX.Element }) {
  const { authenticated, loading } = useAuth();

  if (loading) return null;
  if (!authenticated) {
    window.location.href = "http://localhost:9000/api/v1/auth/login";
    return null;
  }

  return children;
}
