import { baseApi } from "./baseApi";
import {
  loginSuccess,
  logout as logoutAction,
  setAccessToken,
  authLoaded,
} from "../auth/auth.slice";
import type { AuthUser } from "../auth/auth.types";
import { API } from "../../lib/constants/api.constants";

export const authApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    me: builder.query<{ user: AuthUser }, void>({
      query: () => ({
        url: API.auth.me(),
        credentials: "include",
      }),
    }),

    refresh: builder.mutation<{ access_token: string }, void>({
      query: () => ({
        url: API.auth.refresh(),
        method: "POST",
        credentials: "include",
      }),

      async onQueryStarted(_, { dispatch, queryFulfilled }) {
        try {
          const { data } = await queryFulfilled;

          // Store new access token
          dispatch(setAccessToken(data.access_token));

          // Re-fetch user profile
          const me = await dispatch(
            authApi.endpoints.me.initiate(undefined, {
              forceRefetch: true,
            })
          ).unwrap();

          dispatch(
            loginSuccess({
              accessToken: data.access_token,
              user: me.user,
            })
          );
        } catch {
          dispatch(logoutAction());
          window.location.replace(API.auth.login());
        } finally {
          dispatch(authLoaded());
        }
      },
    }),

    logout: builder.mutation<void, void>({
      query: () => ({
        url: API.auth.logout(),
        method: "GET",
        credentials: "include",
      }),
      async onQueryStarted(_, { dispatch }) {
        dispatch(logoutAction());
        window.location.replace(API.auth.logout());
      },
    }),
  }),
});

export const { useMeQuery, useRefreshMutation, useLogoutMutation } = authApi;
