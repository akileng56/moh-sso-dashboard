import React, { useEffect, useState, useCallback, useMemo } from "react";
import { AuthContext, type AuthUser } from "./authContext";

const API_BASE = "http://localhost:9000/api/v1/auth";

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);
  const [loggingOut, setLoggingOut] = useState(false);

  const fetchMe = useCallback(async (token: string) => {
    const res = await fetch(`${API_BASE}/me`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      credentials: "include",
    });

    if (!res.ok) return null;
    return res.json();
  }, []);

  const tryRefresh = useCallback(async () => {
    if (loggingOut) return false;

    try {
      const res = await fetch(`${API_BASE}/refresh`, {
        method: "POST",
        credentials: "include",
      });

      if (!res.ok) return false;

      const data = await res.json();
      setAccessToken(data.access_token);

      const me = await fetchMe(data.access_token);
      if (me?.user) setUser(me.user);

      return true;
    } catch {
      return false;
    }
  }, [fetchMe, loggingOut]);

  const logout = useCallback(() => {
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

  // Silent refresh loop
  useEffect(() => {
    if (loggingOut) return;

    const interval = setInterval(() => {
      tryRefresh();
    }, 240000);

    return () => clearInterval(interval);
  }, [tryRefresh, loggingOut]);

  const isAdmin = user?.is_admin === true;

  const ctx = useMemo(
    () => ({
      authenticated: !!accessToken,
      accessToken,
      user,
      isAdmin,
      loading,
      logout,
    }),
    [accessToken, user, isAdmin, loading, logout]
  );

  if (loading) return <p>Authenticating…</p>;

  return <AuthContext.Provider value={ctx}>{children}</AuthContext.Provider>;
};
