import React from "react";
import "./AppMenu.css";

interface AppTileProps {
  icon: React.ElementType;
  name: string;
  href: string;
}

const AppTile: React.FC<AppTileProps> = ({ icon: Icon, name, href }) => (
  <a id="app_menu_it" href={href}>
    <Icon size={32} style={{ marginBottom: "0.25rem" }} />
    <span style={{ fontSize: "0.75rem" }}>{name}</span>
  </a>
);

export default AppTile;
