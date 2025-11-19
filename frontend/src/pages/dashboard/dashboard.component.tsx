import React, { useState, useEffect } from "react";
import DashboardHeader from "../../components/header/Header.component";
import { useAuth } from "../../context/useAuth";

interface UserProfile {
  id?: string;
  email?: string;
  preferred_username?: string;
  [key: string]: any;
}

const API_BASE = "http://localhost:9000/api/v1/auth";

const Dashboard: React.FC = () => {
  const { authenticated, accessToken, logout } = useAuth();
  const [data, setData] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [fetchError, setFetchError] = useState<string | null>(null);

  useEffect(() => {
    if (!authenticated || !accessToken) {
      setLoading(false);
      return;
    }

    const controller = new AbortController();

    const fetchUser = async () => {
      setLoading(true);
      setFetchError(null);

      try {
        const res = await fetch(`${API_BASE}/me`, {
          method: "GET",
          credentials: "include",
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
          signal: controller.signal,
        });

        if (res.status === 401) {
          logout();
          return;
        }

        if (!res.ok) throw new Error(`Server returned ${res.status}`);

        const json = await res.json();
        setData(json);
      } catch (err: any) {
        if (err.name !== "AbortError") {
          console.error(err);
          setFetchError("Failed to load profile data.");
        }
      } finally {
        setLoading(false);
      }
    };

    fetchUser();

    return () => controller.abort();
  }, [authenticated, accessToken, logout]);

  if (!authenticated) {
    return <h2>Please log in to view the dashboard.</h2>;
  }

  return (
    <div>
      <DashboardHeader />
      <h3 className="mt-4">Secure Profile Data</h3>

      {loading && <p>Loading user profile...</p>}
      {!loading && fetchError && <p style={{ color: "red" }}>{fetchError}</p>}
      {!loading && !fetchError && data && (
        <pre>{JSON.stringify(data, null, 2)}</pre>
      )}
      {!loading && !fetchError && !data && <p>No profile data available.</p>}
    </div>
  );
};

export default Dashboard;
