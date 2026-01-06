import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";

import { ProtectedRoute } from "./components/ProtectedRoute";
import { AdminRoute } from "./components/AdminRoute";

import PublicLayout from "./layout/PublicLayout";
import AdminLayout from "./layout/AdminLayout";

import NewsFeedPage from "./pages/public/newsfeed/news_feed.component";
import AppLauncherPage from "./pages/public/applauncher/app_launcher.component";

import AuditLogsPage from "./pages/admin/audit/audit_component";
import UsersPage from "./pages/admin/user/user.component";
import ClientsPage from "./pages/admin/clients/client.component";

function App() {
  return (
    <Router>
      <Routes>
        {/* ---------------------------------- */}
        {/* PUBLIC (authenticated users) */}
        {/* ---------------------------------- */}
        <Route element={<ProtectedRoute />}>
          <Route element={<PublicLayout />}>
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={<NewsFeedPage />} />
            <Route path="/apps" element={<AppLauncherPage />} />
          </Route>
        </Route>

        {/* ---------------------------------- */}
        {/* ADMIN */}
        {/* ---------------------------------- */}
        <Route element={<ProtectedRoute />}>
          <Route path="/admin" element={<AdminLayout />}>
            <Route
              path="users"
              element={
                <AdminRoute>
                  <UsersPage />
                </AdminRoute>
              }
            />
            <Route
              path="clients"
              element={
                <AdminRoute>
                  <ClientsPage />
                </AdminRoute>
              }
            />
            <Route
              path="audit-logs"
              element={
                <AdminRoute>
                  <AuditLogsPage />
                </AdminRoute>
              }
            />
          </Route>
        </Route>

        {/* ---------------------------------- */}
        {/* FALLBACK */}
        {/* ---------------------------------- */}
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </Router>
  );
}

export default App;
