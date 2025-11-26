import React, { useState, useRef, useEffect } from "react";
import { HeaderGlobalAction } from "@carbon/react";
import { Menu as MenuIcon } from "@carbon/icons-react";
import AppGridContent from "./AppGridContent";
import "./AppMenu.css";

const AppMenuAction: React.FC = () => {
  const [expanded, setExpanded] = useState(false);
  const panelRef = useRef<HTMLDivElement>(null);

  // Close menu when clicking outside
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (
        panelRef.current &&
        !panelRef.current.contains(e.target as Node) &&
        !(e.target as Element).closest("[data-appmenu-trigger]")
      ) {
        setExpanded(false);
      }
    };
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  return (
    <>
      <HeaderGlobalAction
        data-appmenu-trigger
        aria-label="App Menu"
        onClick={() => setExpanded((prev) => !prev)}
        isActive={expanded}
      >
        <MenuIcon size={20} />
      </HeaderGlobalAction>

      {expanded && (
        <div ref={panelRef} id="app_menu_container">
          <AppGridContent />
        </div>
      )}
    </>
  );
};

export default AppMenuAction;
