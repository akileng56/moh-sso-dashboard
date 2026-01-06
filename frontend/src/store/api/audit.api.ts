import { API } from "../../lib/constants/api.constants";
import type { AuditFilters, AuditListResponse } from "../types/audit.types";
import { baseApi } from "./baseApi";

export const auditApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listAuditLogs: builder.query<AuditListResponse, AuditFilters>({
      query: (params) => ({
        url: API.admin.audit.list(),
        params,
        credentials: "include",
      }),
      providesTags: ["Audit"],
    }),
  }),
});

export const { useListAuditLogsQuery } = auditApi;
