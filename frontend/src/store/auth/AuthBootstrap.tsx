import { useEffect, useRef } from "react";
import { useDispatch } from "react-redux";
import { useRefreshMutation } from "../api/auth.api";
import { authLoaded } from "./auth.slice";

export default function AuthBootstrap() {
  const dispatch = useDispatch();
  const [refresh, { isLoading }] = useRefreshMutation();
  const didRun = useRef(false);

  useEffect(() => {
    if (didRun.current || isLoading) return;
    didRun.current = true;

    refresh()
      .unwrap()
      .catch(() => {
        // handled globally (redirect)
      })
      .finally(() => {
        // dispatch(authLoaded());
      });
  }, [refresh, dispatch]);

  return null;
}
