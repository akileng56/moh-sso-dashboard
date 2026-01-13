// src/components/footer/PublicFooter.tsx
import { Grid, Column } from "@carbon/react";
import { BuildMeta } from "./BuildMeta";
import "./PublicFooter.css";

export function PublicFooter() {
  return (
    <footer className="public-footer">
      <Grid>
        <Column sm={4} md={4} lg={8}>
          <strong>MOH Intergrated Health Portal</strong>
          <p>Secure access to Ministry of Health systems</p>
        </Column>
      </Grid>

      <div className="footer-bottom">
        © {new Date().getFullYear()} Ministry of Health – Uganda
        <BuildMeta />
      </div>
    </footer>
  );
}
