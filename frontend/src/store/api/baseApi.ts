import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import type { RootState } from "..";
import { selectAccessToken } from "../auth/auth.selectors";
import { API } from "../../lib/constants/api.constants";

export const baseApi = createApi({
  reducerPath: "api",
  baseQuery: fetchBaseQuery({
    baseUrl: API.base,
    credentials: "include",
    prepareHeaders: (headers, { getState }) => {
      const token = selectAccessToken(getState() as RootState);

      if (token) {
        headers.set("Authorization", `Bearer ${token}`);
      }

      return headers;
    },
  }),
  tagTypes: ["Auth", "User", "Client", "Audit", "Metrics", "Notification"],
  endpoints: () => ({}),
});
