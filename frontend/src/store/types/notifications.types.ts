export type Notification = {
  id: string;
  title: string;
  message: string;
  severity: "info" | "warning" | "critical";
  created_at: string;
  read: boolean;
};

export type GetNotificationsParams = {
  unread?: boolean;
  limit?: number;
  offset?: number;
};
