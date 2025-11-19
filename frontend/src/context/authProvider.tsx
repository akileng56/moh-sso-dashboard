// src/context/AuthProvider.tsx
import React, { useEffect, useState, useCallback, useMemo } from "react";
import { AuthContext } from "./authContext";

const API_BASE = "http://localhost:9000/api/v1/auth";

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const tryRefresh = useCallback(async () => {
    try {
      const res = await fetch(`${API_BASE}/refresh`, {
        method: "POST",
        credentials: "include",
      });

      if (!res.ok) return false;

      const data = await res.json();
      setAccessToken(data.access_token);
      return true;
    } catch (error) {
      return false;
    }
  }, []);

  const logout = useCallback(() => {
    setAccessToken(null);
    window.location.href = `${API_BASE}/logout`;
  }, []);

  // Initial startup refresh
  useEffect(() => {
    (async () => {
      const refreshed = await tryRefresh();
      if (!refreshed) {
        window.location.href = `${API_BASE}/login`;
      }
      setLoading(false);
    })();
  }, [tryRefresh]);

  useEffect(() => {
    const interval = setInterval(() => {
      tryRefresh();
    }, 240000); // 4 min
    return () => clearInterval(interval);
  }, [tryRefresh]);

  const ctx = useMemo(
    () => ({
      authenticated: !!accessToken,
      accessToken,
      logout,
    }),
    [accessToken, logout]
  );

  if (loading) return <p>Authenticating...</p>;

  return <AuthContext.Provider value={ctx}>{children}</AuthContext.Provider>;
};
