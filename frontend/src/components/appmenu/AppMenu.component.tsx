import type React from "react";

import { HeaderGlobalAction, OverflowMenu } from "@carbon/react";
import { Menu as MenuIcon } from "@carbon/icons-react";
import AppGridContent from "./AppGridContent";

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
