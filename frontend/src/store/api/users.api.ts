import { baseApi } from "./baseApi";
import type { CreateUserPayload, User } from "../types/user.types";
import { API } from "../../lib/constants/api.constants";

export const usersApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listUsers: builder.query<User[], void>({
      query: () => ({
        url: API.users.list(),
        credentials: "include",
      }),
      providesTags: ["User"],
    }),

    createUser: builder.mutation<User, CreateUserPayload>({
      query: (body) => ({
        url: API.users.create(),
        method: "POST",
        body,
        credentials: "include",
      }),
      invalidatesTags: ["User"],
    }),

    deleteUser: builder.mutation<void, string>({
      query: (id) => ({
        url: API.users.delete(id),
        method: "DELETE",
        credentials: "include",
      }),
      invalidatesTags: (_r, _e, id) => [{ type: "User", id }],
    }),

    toggleUser: builder.mutation<void, { id: string; enabled: boolean }>({
      query: ({ id, enabled }) => ({
        url: `/users/${id}`,
        method: enabled ? "POST" : "DELETE",
      }),
      invalidatesTags: ["User"],
    }),
  }),
});

export const {
  useListUsersQuery,
  useCreateUserMutation,
  useDeleteUserMutation,
  useToggleUserMutation,
} = usersApi;
