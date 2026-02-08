import React from "react";
import { useDispatch } from "react-redux";

import { setActiveClient } from "../../store/clients/clients.slice";
import "./AppMenu.css";

interface AppTileProps {
  icon: React.ElementType;
  name: string;
  href: string;
  clientId?: string;
}

const AppTile: React.FC<AppTileProps> = ({
  icon: Icon,
  name,
  href,
  clientId,
}) => {
  const dispatch = useDispatch();

  const handleClick = () => {
    // 🔑 Activate client if provided
    if (clientId) {
      dispatch(setActiveClient(clientId));
    }

    // 🌍 Open in new tab (safe defaults)
    window.open(href, "_blank", "noopener,noreferrer");
  };

  return (
    <button
      type="button"
      className="app-menu-item"
      onClick={handleClick}
      aria-label={name}
    >
      <Icon size={28} className="app-menu-item__icon" />
      <span className="app-menu-item__label">{name}</span>
    </button>
  );
};

export default AppTile;
