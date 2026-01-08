import { Tile } from "@carbon/react";
import { useNavigate } from "react-router-dom";

export function QuickAction({
  icon,
  label,
  href,
}: {
  icon: React.ReactNode;
  label: string;
  href: string;
}) {
  const navigate = useNavigate();

  return (
    <Tile className="quick-action" onClick={() => navigate(href)}>
      <div className="qa-icon">{icon}</div>
      <span>{label}</span>
    </Tile>
  );
}
