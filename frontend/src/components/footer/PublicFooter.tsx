// src/components/footer/PublicFooter.tsx
import { BuildMeta } from "./BuildMeta";
import "./PublicFooter.css";

export function PublicFooter() {
  return (
    <footer className="public-footer">
      <div className="footer-bottom">
        © {new Date().getFullYear()} Ministry of Health – Uganda
        {/*<BuildMeta />*/}
      </div>
    </footer>
  );
}
