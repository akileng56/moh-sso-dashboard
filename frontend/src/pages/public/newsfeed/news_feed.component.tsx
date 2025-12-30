import { Tile, Grid, Column } from "@carbon/react";

export default function NewsFeedPage() {
  return (
    <Grid condensed>
      <Column lg={8} md={8} sm={4}>
        <Tile>
          <h4>System Updates</h4>
          <p>
            Welcome to the MOH App Portal. New features are being rolled out
            gradually.
          </p>
        </Tile>

        <Tile style={{ marginTop: "1rem" }}>
          <h4>Maintenance Notice</h4>
          <p>Planned maintenance on Saturday from 10pm–12am.</p>
        </Tile>
      </Column>

      <Column lg={4} md={4} sm={4}>
        <Tile>
          <h4>Quick Links</h4>
          <ul>
            <li>Help & Support</li>
            <li>Privacy Policy</li>
            <li>Security Notice</li>
          </ul>
        </Tile>
      </Column>
    </Grid>
  );
}
