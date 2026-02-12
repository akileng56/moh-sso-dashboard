import { baseApi } from "./baseApi";
import type { AuditOverview, AuditMetricsFilters } from "../types/audit-metrics.types";
import { API } from "../../lib/constants/api.constants";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const metricsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    auditOverview: builder.query<AuditOverview, AuditMetricsFilters>({
      query: ({ from, to }) => ({
        url: API.admin.audit.metrics.overview(),
        params: { from, to },
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<AuditOverview>) => res.data,

      providesTags: ["Audit"],
    }),
  }),
});

export const { useAuditOverviewQuery } = metricsApi;
