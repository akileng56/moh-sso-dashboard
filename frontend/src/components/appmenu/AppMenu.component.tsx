import React, { useState, useRef, useEffect } from "react";
import { HeaderGlobalAction } from "@carbon/react";
import { Menu as MenuIcon } from "@carbon/icons-react";
import AppGridContent from "./AppGridContent";
import "./AppMenu.css";

const AppMenuAction: React.FC = () => {
  const [expanded, setExpanded] = useState(false);
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const panelRef = useRef<HTMLDivElement | null>(null);

  /* -----------------------------
   * Close helpers
   * ----------------------------- */
  const closeMenu = () => setExpanded(false);

  /* -----------------------------
   * Click outside & ESC to close
   * ----------------------------- */
  useEffect(() => {
    if (!expanded) return;

    const handleClickOutside = (e: MouseEvent) => {
      if (
        panelRef.current &&
        !panelRef.current.contains(e.target as Node) &&
        triggerRef.current &&
        !triggerRef.current.contains(e.target as Node)
      ) {
        closeMenu();
      }
    };

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        closeMenu();
        triggerRef.current?.focus();
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [expanded]);

  /* -----------------------------
   * Focus management
   * ----------------------------- */
  useEffect(() => {
    if (expanded) {
      panelRef.current?.focus();
    }
  }, [expanded]);

  return (
    <>
      <HeaderGlobalAction
        ref={triggerRef}
        aria-label="Open application menu"
        aria-haspopup="dialog"
        aria-expanded={expanded}
        isActive={expanded}
        onClick={() => setExpanded((prev) => !prev)}
      >
        <MenuIcon size={20} />
      </HeaderGlobalAction>

      {expanded && (
        <div
          ref={panelRef}
          id="app_menu_container"
          role="dialog"
          aria-label="Application menu"
          tabIndex={-1}
        >
          <AppGridContent />
        </div>
      )}
    </>
  );
};

export default AppMenuAction;
