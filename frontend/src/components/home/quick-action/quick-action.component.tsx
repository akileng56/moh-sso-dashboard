import { Tile } from "@carbon/react";
import { useNavigate } from "react-router-dom";
import "./quick-action.css";

type QuickActionProps = {
  icon: React.ReactNode;
  label: string;
  description?: string;
  href: string;
  tone?: "default" | "warning" | "danger";
};

export function QuickAction({
  icon,
  label,
  description,
  href,
  tone = "default",
}: QuickActionProps) {
  const navigate = useNavigate();

  return (
    <Tile
      className={`quick-action quick-action--${tone}`}
      onClick={() => navigate(href)}
    >
      <div className="quick-action__icon">{icon}</div>

      <div className="quick-action__content">
        <strong>{label}</strong>
        {description && (
          <p className="quick-action__description">{description}</p>
        )}
      </div>
    </Tile>
  );
}
