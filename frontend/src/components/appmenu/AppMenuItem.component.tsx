import React from "react";
import { useNavigate } from "react-router-dom";
import "./AppMenu.css";

interface AppTileProps {
  icon: React.ElementType;
  name: string;
  href: string;
}

const AppTile: React.FC<AppTileProps> = ({ icon: Icon, name, href }) => {
  const navigate = useNavigate();

  return (
    <button
      type="button"
      className="app-menu-item"
      onClick={() => navigate(href)}
      aria-label={name}
    >
      <Icon size={28} className="app-menu-item__icon" />
      <span className="app-menu-item__label">{name}</span>
    </button>
  );
};

export default AppTile;
