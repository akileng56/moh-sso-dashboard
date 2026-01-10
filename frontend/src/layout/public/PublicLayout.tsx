import { Content } from "@carbon/react";
import { Outlet } from "react-router-dom";
import PublicHeader from "../../components/header/PublicHeader.component";

export default function PublicLayout() {
  return (
    <>
      <PublicHeader />

      <Content style={{ paddingTop: "3rem" }}>
        <Outlet />
      </Content>
    </>
  );
}
