import type React from "react";

import { HeaderGlobalAction, OverflowMenu, Grid, Column } from "@carbon/react";
import {
  Menu as MenuIcon,
  Calendar,
  Email,
  Document,
} from "@carbon/icons-react";

import AppTile from "./AppMenuItem.component";

const AppGridContent = () => (
  <Grid narrow style={{ width: "250px", padding: "1rem", background: "white" }}>
    <Column sm={2} md={2} lg={4}>
      <AppTile icon={Email} name="Mail" href="/mail" />
    </Column>
    <Column sm={2} md={2} lg={4}>
      <AppTile icon={Calendar} name="Calendar" href="/calendar" />
    </Column>
    <Column sm={2} md={2} lg={4}>
      <AppTile icon={Document} name="Docs" href="/docs" />
    </Column>
    <Column sm={2} md={2} lg={4}>
      <AppTile icon={MenuIcon} name="Settings" href="/settings" />
    </Column>
  </Grid>
);

const AppMenuAction: React.FC = () => (
  <HeaderGlobalAction aria-label="App Menu" tooltipAlignment="end">
    <OverflowMenu
      iconDescription="App Menu"
      renderIcon={MenuIcon}
      direction="bottom"
    >
      <AppGridContent />
    </OverflowMenu>
  </HeaderGlobalAction>
);

export default AppMenuAction;
