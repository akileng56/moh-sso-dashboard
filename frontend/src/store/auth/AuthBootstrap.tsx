import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useRefreshMutation } from "../api/auth.api";
import { authLoaded } from "./auth.slice";
import { selectAuthLoaded } from "./auth.selectors";

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
