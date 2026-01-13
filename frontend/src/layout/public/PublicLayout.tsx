import { Content } from "@carbon/react";
import { Outlet } from "react-router-dom";
import PublicHeader from "../../components/header/PublicHeader.component";
import { PublicFooter } from "../../components/footer/PublicFooter";

export default function PublicLayout() {
  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        flexDirection: "column",
      }}
    >
      <PublicHeader />

      {/* Main content */}
      <Content
        style={{
          paddingTop: "3rem",
          flex: 1, // 🔑 pushes footer to bottom
        }}
      >
        <Outlet />
      </Content>

      {/* Footer OUTSIDE Content */}
      <PublicFooter />
    </div>
  );
}
