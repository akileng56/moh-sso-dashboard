import React from "react";
import { useNavigate } from "react-router-dom";
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
  const navigate = useNavigate();
  const dispatch = useDispatch();

  const handleClick = () => {
    // 🔑 Activate client if provided
    if (clientId) {
      dispatch(setActiveClient(clientId));
    }

    // 🧭 Navigate to route
    navigate(href);
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
