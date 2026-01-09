import { Tile } from "@carbon/react";
import { useNavigate } from "react-router-dom";
import "./quick-action.css";

type QuickActionProps = {
  icon: React.ReactNode;
  label: string;
  description?: string;

  /** Route navigation (optional) */
  href?: string;

  /** Custom action (e.g. open header panel) */
  onClick?: () => void;

  tone?: "default" | "warning" | "danger";
};

export function QuickAction({
  icon,
  label,
  description,
  href,
  onClick,
  tone = "default",
}: QuickActionProps) {
  const navigate = useNavigate();

  const handleClick = () => {
    if (onClick) {
      onClick();
      return;
    }

    if (href) {
      navigate(href);
    }
  };

  return (
    <Tile
      role="button"
      tabIndex={0}
      className={`quick-action quick-action--${tone}`}
      onClick={handleClick}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          handleClick();
        }
      }}
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
