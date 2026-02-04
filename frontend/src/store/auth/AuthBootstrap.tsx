import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useRefreshMutation } from "../api/auth.api";
import { selectAuthLoaded } from "./auth.selectors";
import { authLoaded } from "./auth.slice";

export default function AuthBootstrap() {
  const dispatch = useDispatch();
  const authLoadedFlag = useSelector(selectAuthLoaded);
  const [refresh] = useRefreshMutation();

  useEffect(() => {
    if (authLoadedFlag) return;

    refresh()
      .unwrap()
      .catch(() => {
        // refresh failure is handled globally (redirect)
      })
      .finally(() => {
        dispatch(authLoaded());
      });
  }, [authLoadedFlag, refresh, dispatch]);

  return null;
}
