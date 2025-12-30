import { Grid, Column, Tile, Link, Tag } from "@carbon/react";
import { Information, Time, Add } from "@carbon/icons-react";
import { useNavigate } from "react-router-dom";
import { EmptyState } from "../../../components/emptystate/EmptyState";
import { ErrorState } from "../../../components/errorstate/ErrorState";

/* --------------------------------
 * Types
 * -------------------------------- */
interface FeedItem {
  id: string;
  title: string;
  message: string;
  tag: {
    label: string;
    type: "blue" | "red" | "green" | "gray";
  };
  timestamp: string;
}

/* --------------------------------
 * Temporary feed data
 * (replace with API later)
 * -------------------------------- */
const FEED_ITEMS: FeedItem[] = [
  {
    id: "1",
    title: "System Updates",
    message:
      "Welcome to the MOH App Portal. New features are being rolled out gradually to improve security, performance, and usability.",
    tag: { label: "New", type: "blue" },
    timestamp: "Posted today",
  },
  {
    id: "2",
    title: "Maintenance Notice",
    message:
      "Planned system maintenance will occur on Saturday from 10:00 PM – 12:00 AM. Some services may be temporarily unavailable.",
    tag: { label: "Scheduled", type: "red" },
    timestamp: "Scheduled",
  },
];

export default function NewsFeedPage() {
  const navigate = useNavigate();

  /* --------------------------------
   * Mock state (replace with API later)
   * -------------------------------- */
  const loading = false;
  const error: string | null = null;
  const isEmpty = FEED_ITEMS.length === 0;

  return (
    <Grid condensed style={{ padding: "2rem" }}>
      {/* --------------------------------
       * Page Header
       * -------------------------------- */}
      <Column lg={12} md={8} sm={4}>
        <h3 style={{ margin: 0 }}>News & Updates</h3>
        <p style={{ marginTop: 6, opacity: 0.8 }}>
          Latest system updates, announcements, and notices.
        </p>
      </Column>

      {/* --------------------------------
       * Main Feed
       * -------------------------------- */}
      <Column lg={8} md={8} sm={4}>
        {/* ---------- Error ---------- */}
        {error && (
          <ErrorState
            title="Failed to load announcements"
            description={error}
            primaryAction={{
              label: "Retry",
              onClick: () => window.location.reload(),
            }}
            secondaryAction={{
              label: "Contact support",
              onClick: () => navigate("/support"),
            }}
          />
        )}

        {/* ---------- Empty ---------- */}
        {!error && !loading && isEmpty && (
          <EmptyState
            title="No announcements yet"
            description="System updates, maintenance notices, and important messages will appear here when available."
            primaryAction={{
              label: "Create announcement",
              icon: Add,
              onClick: () => navigate("/admin/announcements/new"),
            }}
          />
        )}

        {/* ---------- Data ---------- */}
        {!error &&
          !loading &&
          !isEmpty &&
          FEED_ITEMS.map((item) => (
            <Tile key={item.id} style={{ marginTop: "1rem" }}>
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 8,
                }}
              >
                <Information size={16} />
                <h4 style={{ margin: 0 }}>{item.title}</h4>
                <Tag type={item.tag.type} size="sm">
                  {item.tag.label}
                </Tag>
              </div>

              <p style={{ marginTop: "0.75rem" }}>{item.message}</p>

              <div
                style={{
                  marginTop: "0.75rem",
                  display: "flex",
                  alignItems: "center",
                  gap: 6,
                  opacity: 0.7,
                  fontSize: "0.875rem",
                }}
              >
                <Time size={14} />
                <span>{item.timestamp}</span>
              </div>
            </Tile>
          ))}
      </Column>

      {/* --------------------------------
       * Sidebar
       * -------------------------------- */}
      <Column lg={4} md={4} sm={4}>
        <Tile style={{ marginTop: "1rem" }}>
          <h4 style={{ marginTop: 0 }}>Quick Links</h4>

          <ul style={{ paddingLeft: "1rem", marginTop: "0.75rem" }}>
            <li style={{ marginBottom: "0.5rem" }}>
              <Link href="#">Help & Support</Link>
            </li>
            <li style={{ marginBottom: "0.5rem" }}>
              <Link href="#">Privacy Policy</Link>
            </li>
            <li>
              <Link href="#">Security Notice</Link>
            </li>
          </ul>
        </Tile>
      </Column>
    </Grid>
  );
}
