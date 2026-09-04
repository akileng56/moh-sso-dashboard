import { baseApi } from "@moh-sso/api";
import { API } from "@moh-sso/config";

import type {
  DQACompilePreview,
  DQAFlag,
  DQARule,
  DQARuleInput,
  DQARunResult,
  DQARunSummary,
  DQASeedResult,
  DQATableMapping,
} from "../dqa.types";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const dqaApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listDQATables: builder.query<DQATableMapping[], void>({
      query: () => ({ url: API.dataValidation.dqa.tables.list(), method: "GET", credentials: "include" }),
      transformResponse: (res: ApiEnvelope<DQATableMapping[]>) => res.data,
      providesTags: ["DQATables"],
    }),

    upsertDQATable: builder.mutation<DQATableMapping, DQATableMapping>({
      query: (body) => ({ url: API.dataValidation.dqa.tables.upsert(), method: "POST", body, credentials: "include" }),
      transformResponse: (res: ApiEnvelope<DQATableMapping>) => res.data,
      invalidatesTags: ["DQATables"],
    }),

    deleteDQATable: builder.mutation<void, string>({
      query: (tableId) => ({ url: API.dataValidation.dqa.tables.remove(tableId), method: "DELETE", credentials: "include" }),
      invalidatesTags: ["DQATables"],
    }),

    listDQARules: builder.query<DQARule[], { tableId?: string } | void>({
      query: (args) => ({ url: API.dataValidation.dqa.rules.list(args?.tableId), method: "GET", credentials: "include" }),
      transformResponse: (res: ApiEnvelope<DQARule[]>) => res.data,
      providesTags: ["DQARules"],
    }),

    upsertDQARule: builder.mutation<DQARule, DQARuleInput>({
      query: (body) => ({ url: API.dataValidation.dqa.rules.upsert(), method: "POST", body, credentials: "include" }),
      transformResponse: (res: ApiEnvelope<DQARule>) => res.data,
      invalidatesTags: ["DQARules"],
    }),

    compileDQARule: builder.mutation<DQACompilePreview, DQARuleInput>({
      query: (body) => ({ url: API.dataValidation.dqa.rules.compile(), method: "POST", body, credentials: "include" }),
      transformResponse: (res: ApiEnvelope<DQACompilePreview>) => res.data,
    }),

    seedDQABuiltinRules: builder.mutation<DQASeedResult, { table_id?: string; physical_table: string }>({
      query: (body) => ({ url: API.dataValidation.dqa.rules.seed(), method: "POST", body, credentials: "include" }),
      transformResponse: (res: ApiEnvelope<DQASeedResult>) => res.data,
      invalidatesTags: ["DQARules", "DQATables"],
    }),

    deleteDQARule: builder.mutation<void, { tableId: string; code: string }>({
      query: ({ tableId, code }) => ({
        url: API.dataValidation.dqa.rules.remove(tableId, code),
        method: "DELETE",
        credentials: "include",
      }),
      invalidatesTags: ["DQARules"],
    }),

    runDQATable: builder.mutation<DQARunResult, string>({
      query: (tableId) => ({
        url: API.dataValidation.dqa.run(),
        method: "POST",
        body: { table_id: tableId },
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<DQARunResult>) => res.data,
      invalidatesTags: ["DQARuns"],
    }),

    listDQARuns: builder.query<DQARunSummary[], { tableId?: string } | void>({
      query: (args) => ({ url: API.dataValidation.dqa.runs.list(args?.tableId), method: "GET", credentials: "include" }),
      transformResponse: (res: ApiEnvelope<DQARunSummary[]>) => res.data,
      providesTags: ["DQARuns"],
    }),

    listDQAFlags: builder.query<DQAFlag[], { runId: number; severity?: string }>({
      query: ({ runId, severity }) => ({
        url: API.dataValidation.dqa.runs.flags(runId, severity),
        method: "GET",
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<DQAFlag[]>) => res.data,
      providesTags: (_result, _error, args) => [{ type: "DQAFlags", id: args.runId }],
    }),
  }),
});

export const {
  useListDQATablesQuery,
  useUpsertDQATableMutation,
  useDeleteDQATableMutation,
  useListDQARulesQuery,
  useUpsertDQARuleMutation,
  useCompileDQARuleMutation,
  useSeedDQABuiltinRulesMutation,
  useDeleteDQARuleMutation,
  useRunDQATableMutation,
  useListDQARunsQuery,
  useListDQAFlagsQuery,
} = dqaApi;
