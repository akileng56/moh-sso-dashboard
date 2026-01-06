import { baseApi } from "./baseApi";
import type { Client, CreateClientPayload } from "../types/client.types";
import { API } from "../../lib/constants/api.constants";

export const clientsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listClients: builder.query<Client[], void>({
      query: () => ({
        url: API.clients.list(),
        credentials: "include",
      }),
      providesTags: ["Client"],
    }),

    getClient: builder.query<Client, string>({
      query: (id) => ({
        url: API.clients.byId(id),
        credentials: "include",
      }),
      providesTags: (_r, _e, id) => [{ type: "Client", id }],
    }),

    createClient: builder.mutation<Client, CreateClientPayload>({
      query: (body) => ({
        url: API.clients.create(),
        method: "POST",
        body,
        credentials: "include",
      }),
      invalidatesTags: ["Client"],
    }),
    deleteClient: builder.mutation<void, string>({
      query: (id) => ({
        url: API.clients.delete(id),
        method: "DELETE",
        credentials: "include",
      }),
      invalidatesTags: (_r, _e, id) => [{ type: "Client", id }],
    }),

    toggleClient: builder.mutation<void, { id: string; enabled: boolean }>({
      query: ({ id, enabled }) => ({
        url: `/clients/${id}`,
        method: enabled ? "POST" : "DELETE",
      }),
      invalidatesTags: ["Client"],
    }),
  }),
});

export const {
  useListClientsQuery,
  useGetClientQuery,
  useCreateClientMutation,
  useDeleteClientMutation,
  useToggleClientMutation,
} = clientsApi;
