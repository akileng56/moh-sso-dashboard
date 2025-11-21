import { useEffect, useState } from "react";
import { useAuth } from "../../context/useAuth";

export default function ClientsPage() {
  const { accessToken } = useAuth();
  const [clients, setClients] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("http://localhost:9000/api/v1/clients", {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
      .then((res) => res.json())
      .then((data) => {
        setClients(data);
        setLoading(false);
      });
  }, []);

  if (loading) return <p>Loading...</p>;

  return (
    <div style={{ padding: "2rem" }}>
      <h1>Registered Clients</h1>
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Client ID</th>
            <th>Type</th>
            <th>Status</th>
          </tr>
        </thead>

        <tbody>
          {clients.map((c: any) => (
            <tr key={c.id}>
              <td>{c.name}</td>
              <td>{c.clientId}</td>
              <td>{c.publicClient ? "Public" : "Confidential"}</td>
              <td>{c.enabled ? "Enabled" : "Disabled"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
