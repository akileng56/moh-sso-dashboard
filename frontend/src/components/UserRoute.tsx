// components/UserRoute.tsx
import { Navigate } from "react-router-dom";
import { useSelector } from "react-redux";
import { InlineLoading } from "@carbon/react";

import { selectAuthenticated, selectAuthLoaded, selectIsUser } from "../store/auth/auth.selectors";
import type { JSX } from "react";

export const UserRoute = ({ children }: { children: JSX.Element }) => {
  const loaded = useSelector(selectAuthLoaded);
  const authenticated = useSelector(selectAuthenticated);
  const isUser = useSelector(selectIsUser);

  if (!loaded) {
    return <InlineLoading description="Checking session…" />;
  }

  if (!authenticated || !isUser) {
    return <Navigate to="/" replace />;
  }

  return children;
};
