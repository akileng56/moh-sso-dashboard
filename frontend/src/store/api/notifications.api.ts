import { API } from "../../lib/constants/api.constants";
import type {
  GetNotificationsParams,
  Notification,
} from "../types/notifications.types";
import { baseApi } from "./baseApi";

export const notificationsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getNotifications: builder.query<Notification[], GetNotificationsParams>({
      query: ({ unread, limit = 20, offset = 0 }) => {
        const params = new URLSearchParams();

        if (unread !== undefined) {
          params.set("unread", String(unread));
        }

        params.set("limit", String(limit));
        params.set("offset", String(offset));

        return {
          url: `${API.admin.notifications.list()}?${params.toString()}`,
          credentials: "include",
        };
      },

      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({
                type: "Notification" as const,
                id,
              })),
              { type: "Notification", id: "LIST" },
            ]
          : [{ type: "Notification", id: "LIST" }],
    }),

    markNotificationAsRead: builder.mutation<void, string>({
      query: (id) => ({
        url: API.admin.notifications.markAsRead(id),
        method: "PATCH",
        credentials: "include",
      }),
      invalidatesTags: (result, error, id) => [
        { type: "Notification", id },
        { type: "Notification", id: "LIST" },
        { type: "Notification", id: "COUNT" },
      ],
    }),

    deleteNotification: builder.mutation<void, string>({
      query: (id) => ({
        url: API.admin.notifications.delete(id),
        credentials: "include",
        method: "DELETE",
      }),
      invalidatesTags: ["Notification"],
    }),

    deleteOldNotifications: builder.mutation<void, void>({
      query: () => ({
        url: API.admin.notifications.deleteOld(),
        credentials: "include",
        method: "DELETE",
      }),
      invalidatesTags: ["Notification"],
    }),

    countNotifications: builder.query<number, void>({
      query: () => ({
        url: API.admin.notifications.count(),
        credentials: "include",
      }),
      providesTags: ["Notification"],
    }),

    notify: builder.mutation<void, void>({
      query: () => ({
        url: API.admin.notifications.notify(),
        credentials: "include",
        method: "POST",
      }),
      invalidatesTags: ["Notification"],
    }),

    getNotification: builder.query<Notification, string>({
      query: (id) => ({
        url: API.admin.notifications.byId(id),
        credentials: "include",
      }),
      providesTags: (_r, _e, id) => [{ type: "Notification", id }],
    }),

    getUnreadNotificationsCount: builder.query<number, void>({
      query: () => ({
        url: API.admin.notifications.countUnread(),
        credentials: "include",
      }),
      providesTags: [{ type: "Notification", id: "COUNT" }],
    }),
  }),
});

export const {
  useGetNotificationsQuery,
  useMarkNotificationAsReadMutation,
  useDeleteNotificationMutation,
  useDeleteOldNotificationsMutation,
  useCountNotificationsQuery,
  useNotifyMutation,
  useGetNotificationQuery,
  useGetUnreadNotificationsCountQuery,
} = notificationsApi;
