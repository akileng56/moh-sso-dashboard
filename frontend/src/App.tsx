import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";

import { ProtectedRoute } from "./components/ProtectedRoute";
import { AdminRoute } from "./components/AdminRoute";

import PublicLayout from "./layout/public/PublicLayout";
import UserLayout from "./layout/user/UserLayout";
import AdminLayout from "./layout/admin/AdminLayout";

import NewsFeedPage from "./pages/public/newsfeed/news_feed.component";
import DataVisualizer from "./pages/public/datavisualizer/data-visualizer";

import AuditLogsPage from "./pages/admin/audit/audit_component";
import UsersPage from "./pages/admin/user/user.component";
import ClientsPage from "./pages/admin/clients/client.component";
import HomePage from "./pages/admin/home/home.component";

import { ModalProvider } from "./components/modal/modal.context";
import { UserRoute } from "./components/UserRoute";

function App() {
  return (
    <ModalProvider>
      <Router>
        <Routes>
          {/* ---------------------------------- */}
          {/* PUBLIC (no auth required) */}
          {/* ---------------------------------- */}
          <Route element={<PublicLayout />}>
            <Route index element={<NewsFeedPage />} />
            <Route path="/" element={<NewsFeedPage />} />
          </Route>

          {/* ---------------------------------- */}
          {/* USER (authenticated) */}
          {/* ---------------------------------- */}
          <Route
            path="/apps"
            element={
              <ProtectedRoute>
                <UserLayout />
              </ProtectedRoute>
            }
          >
            {/* /apps → redirect to default client home */}
            <Route
              index
              element={<Navigate to="dwh/data-visualizer" replace />}
            />

            <Route path="dwh">
              <Route
                index
                element={<Navigate to="data-visualizer" replace />}
              />
              <Route
                path="data-visualizer"
                element={
                  <UserRoute>
                    <DataVisualizer />
                  </UserRoute>
                }
              />
              {/* future */}

              {/* <Route
                path="reports"
                element={
                  <UserRoute>
                    <DwhReports />
                  </UserRoute>
                }
              /> */}
            </Route>
          </Route>

          {/* ---------------------------------- */}
          {/* ADMIN (authenticated + role) */}
          {/* ---------------------------------- */}
          <Route
            path="/admin"
            element={
              <ProtectedRoute>
                <AdminLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<Navigate to="/admin/home" replace />} />

            <Route path="home" element={<HomePage />} />

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

          {/* ---------------------------------- */}
          {/* FALLBACK */}
          {/* ---------------------------------- */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Router>
    </ModalProvider>
  );
}

export default App;
