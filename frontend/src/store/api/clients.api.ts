import { baseApi } from "./baseApi";
import type { Client, CreateClientPayload } from "../types/client.types";
import { API } from "../../lib/constants/api.constants";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

type UpdateClientPayload = {
  id: string;
  data: Partial<CreateClientPayload>;
};

export const clientsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    /* --------------------------------
     * List clients
     * -------------------------------- */
    listClients: builder.query<Client[], void>({
      query: () => ({
        url: API.clients.list(),
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<Client[]>) => res.data,

      providesTags: (result) =>
        result
          ? [
              ...result.map((client) => ({
                type: "Client" as const,
                id: client.id,
              })),
              { type: "Client", id: "LIST" },
            ]
          : [{ type: "Client", id: "LIST" }],
    }),

    /* --------------------------------
     * Get single client
     * -------------------------------- */
    getClient: builder.query<Client, string>({
      query: (id) => ({
        url: API.clients.byId(id),
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<Client>) => res.data,

      providesTags: (_result, _error, id) => [{ type: "Client", id }],
    }),

    /* --------------------------------
     * Create client
     * -------------------------------- */
    createClient: builder.mutation<Client, CreateClientPayload>({
      query: (body) => ({
        url: API.clients.create(),
        method: "POST",
        body,
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<Client>) => res.data,

      invalidatesTags: [{ type: "Client", id: "LIST" }],
    }),

    /* --------------------------------
     * Update client
     * -------------------------------- */
    updateClient: builder.mutation<Client, UpdateClientPayload>({
      query: ({ id, data }) => ({
        url: API.clients.update(id),
        method: "PATCH",
        body: data,
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<Client>) => res.data,

      invalidatesTags: (_r, _e, { id }) => [
        { type: "Client", id },
        { type: "Client", id: "LIST" },
      ],
    }),

    /* --------------------------------
     * Delete client
     * -------------------------------- */
    deleteClient: builder.mutation<void, string>({
      query: (id) => ({
        url: API.clients.delete(id),
        method: "DELETE",
        credentials: "include",
      }),

      // backend returns { success: true, data: null }
      transformResponse: () => undefined,

      invalidatesTags: (_r, _e, id) => [
        { type: "Client", id },
        { type: "Client", id: "LIST" },
      ],
    }),

    /* --------------------------------
     * Enable / Disable client
     * -------------------------------- */
    toggleClient: builder.mutation<void, { id: string; enabled: boolean }>({
      query: ({ id, enabled }) => ({
        url: API.clients.toggleClient(id),
        method: enabled ? "POST" : "DELETE",
        credentials: "include",
      }),

      transformResponse: () => undefined,

      invalidatesTags: (_r, _e, { id }) => [
        { type: "Client", id },
        { type: "Client", id: "LIST" },
      ],
    }),
  }),
});

export const {
  useListClientsQuery,
  useGetClientQuery,
  useCreateClientMutation,
  useUpdateClientMutation,
  useDeleteClientMutation,
  useToggleClientMutation,
} = clientsApi;
