import { PermissionGuard, PERMISSIONS } from "@moh-sso/auth";

const IssueDashboard = () => {
  return (
    <PermissionGuard permission={PERMISSIONS.issueTrackerRead}>
      <div
        style={{
          padding: "2rem",
          backgroundColor: "#ffffff",
          minHeight: "calc(100vh - 120px)",
          borderRadius: "4px",
          border: "1px solid #e0e0e0",
        }}
      >
        <h3 style={{ fontSize: "1.25rem", fontWeight: 600, color: "#161616", marginBottom: "0.5rem" }}>
          Issue Dashboard
        </h3>
        <p style={{ color: "#525252", fontSize: "0.875rem" }}>
          This page is currently blank and ready for custom dashboard metrics and visualizations.
        </p>
      </div>
    </PermissionGuard>
  );
};

export default IssueDashboard;
