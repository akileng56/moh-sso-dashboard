import React from "react";
import "./AppMenu.css";

interface AppTileProps {
  icon: React.ElementType;
  name: string;
  href: string;
}

const AppTile: React.FC<AppTileProps> = ({ icon: Icon, name, href }) => (
  <a id="app_menu_it" href={href}>
    <Icon size={36} style={{ marginBottom: "0.4rem" }} />
    <span style={{ fontSize: "0.8rem" }}>{name}</span>
  </a>
);

export default AppTile;
