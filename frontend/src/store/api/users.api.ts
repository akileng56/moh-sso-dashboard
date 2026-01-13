import { baseApi } from "./baseApi";
import type { CreateUserPayload, User } from "../types/user.types";
import { API } from "../../lib/constants/api.constants";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

type UpdateUserPayload = {
  id: string;
  data: Partial<CreateUserPayload>;
};

export const usersApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    /* --------------------------------
     * List users
     * -------------------------------- */
    listUsers: builder.query<User[], void>({
      query: () => ({
        url: API.users.list(),
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<User[]>) => res.data,

      providesTags: (result) =>
        result
          ? [
              ...result.map((user) => ({
                type: "User" as const,
                id: user.id,
              })),
              { type: "User", id: "LIST" },
            ]
          : [{ type: "User", id: "LIST" }],
    }),

    /* --------------------------------
     * Get single user
     * -------------------------------- */
    getUser: builder.query<User, string>({
      query: (id) => ({
        url: API.users.byId(id),
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<User>) => res.data,

      providesTags: (_r, _e, id) => [{ type: "User", id }],
    }),

    /* --------------------------------
     * Create user
     * -------------------------------- */
    createUser: builder.mutation<User, CreateUserPayload>({
      query: (body) => ({
        url: API.users.create(),
        method: "POST",
        body,
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<User>) => res.data,

      invalidatesTags: [{ type: "User", id: "LIST" }],
    }),

    /* --------------------------------
     * Update user
     * -------------------------------- */
    updateUser: builder.mutation<User, UpdateUserPayload>({
      query: ({ id, data }) => ({
        url: API.users.update(id),
        method: "PATCH",
        body: data,
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<User>) => res.data,

      invalidatesTags: (_r, _e, { id }) => [
        { type: "User", id },
        { type: "User", id: "LIST" },
      ],
    }),

    /* --------------------------------
     * Reset password
     * -------------------------------- */
    resetUserPassword: builder.mutation<void, string>({
      query: (id) => ({
        url: API.users.resetPassword(id),
        method: "POST",
        credentials: "include",
      }),

      transformResponse: () => undefined,

      invalidatesTags: (_r, _e, id) => [{ type: "User", id }],
    }),

    /* --------------------------------
     * Delete user
     * -------------------------------- */
    deleteUser: builder.mutation<void, string>({
      query: (id) => ({
        url: API.users.delete(id),
        method: "DELETE",
        credentials: "include",
      }),

      transformResponse: () => undefined,

      invalidatesTags: (_r, _e, id) => [
        { type: "User", id },
        { type: "User", id: "LIST" },
      ],
    }),

    /* --------------------------------
     * Enable / Disable user
     * -------------------------------- */
    toggleUser: builder.mutation<void, { id: string; enabled: boolean }>({
      query: ({ id, enabled }) => ({
        url: API.users.toggerUser(id),
        method: enabled ? "POST" : "DELETE",
        credentials: "include",
      }),

      transformResponse: () => undefined,

      invalidatesTags: (_r, _e, { id }) => [
        { type: "User", id },
        { type: "User", id: "LIST" },
      ],
    }),
  }),
});

export const {
  useListUsersQuery,
  useGetUserQuery,
  useCreateUserMutation,
  useUpdateUserMutation,
  useResetUserPasswordMutation,
  useDeleteUserMutation,
  useToggleUserMutation,
} = usersApi;
