import { baseApi } from "./baseApi";
import {
  loginSuccess,
  logout as logoutAction,
  authLoaded,
  setAccessToken,
} from "../auth/auth.slice";
import type { AuthUser } from "../auth/auth.types";
import { API } from "../../lib/constants/api.constants";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

type RefreshResponse = ApiEnvelope<{
  access_token: string;
  user: AuthUser;
}>;

export const authApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    /* -----------------------------
     * Get current user
     * ----------------------------- */
    me: builder.query<AuthUser, void>({
      query: () => ({
        url: API.auth.me(),
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<{ user: AuthUser }>) =>
        res.data.user,
    }),

    /* -----------------------------
     * Refresh token
     * ----------------------------- */
    refresh: builder.mutation<RefreshResponse, void>({
      query: () => ({
        url: API.auth.refresh(),
        method: "POST",
        credentials: "include",
      }),

      async onQueryStarted(_, { dispatch, queryFulfilled }) {
        try {
          const { data } = await queryFulfilled;

          // Store new access token
          dispatch(setAccessToken(data.data.access_token));

          const me = await dispatch(
            authApi.endpoints.me.initiate(undefined, {
              forceRefetch: true,
            })
          ).unwrap();

          dispatch(
            loginSuccess({
              accessToken: data.data.access_token,
              user: me,
            })
          );
        } catch {
          dispatch(logoutAction());
          window.location.replace(API.auth.login());
        } finally {
          // dispatch(authLoaded());
        }
      },
    }),
  }),
});

export const { useMeQuery, useRefreshMutation } = authApi;
