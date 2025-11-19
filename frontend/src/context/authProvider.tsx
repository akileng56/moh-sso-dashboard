// src/context/AuthProvider.tsx
import React, { useEffect, useState, useCallback, useMemo } from "react";
import { AuthContext } from "./authContext";

const REFRESH_URL = "http://localhost:9000/api/v1/auth/refresh";

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  // Try to get a new access token using the HttpOnly cookie
  const tryRefresh = useCallback(async () => {
    try {
      const res = await fetch(REFRESH_URL, { method: "POST" });

      if (res.ok) {
        const data = await res.json();
        setAccessToken(data.access_token);
        return true;
      }

      return false;
    } catch {
      return false;
    }
  }, []);

  const logout = () => {
    setAccessToken(null);
    window.location.href = "/api/v1/auth/logout";
  };

  // On initial load — check if backend can refresh token
  useEffect(() => {
    (async () => {
      const ok = await tryRefresh();
      if (!ok) {
        window.location.href = "http://localhost:9000/api/v1/auth/login";
      }
      setLoading(false);
    })();
  }, [tryRefresh]);

  const ctx = useMemo(
    () => ({
      authenticated: !!accessToken,
      accessToken,
      logout,
    }),
    [accessToken]
  );

  if (loading) return <p>Authenticating...</p>;

  return <AuthContext.Provider value={ctx}>{children}</AuthContext.Provider>;
};
