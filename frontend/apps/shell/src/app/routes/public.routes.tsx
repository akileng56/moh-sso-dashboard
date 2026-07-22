import { Route } from "react-router-dom";

import PublicLayout from "../layouts/public/public-layout.component";
import { ForbiddenPage } from "./ForbiddenPage";
import { PortalLandingRoute } from "./PortalLandingRoute";

export const publicRoutes = (
  <Route element={<PublicLayout />}>
    <Route index element={<PortalLandingRoute />} />
    <Route path="/" element={<PortalLandingRoute />} />
    <Route path="/forbidden" element={<ForbiddenPage />} />
  </Route>
);
