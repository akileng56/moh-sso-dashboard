import React, { useEffect, useState } from "react";
import type { ReactNode } from "react";
import keycloak from "./config/keycloak";

interface KeycloakProviderProps {
  children: ReactNode;
}

export const KeycloakProvider: React.FC<KeycloakProviderProps> = ({
  children,
}) => {
  const [authenticated, setAuthenticated] = useState(false);
  const redirectUri = import.meta.env.VITE_KEYCLOAK_REDIRECT_URI;

  if (!redirectUri) {
    console.error(
      "VITE_KEYCLOAK_REDIRECT_URI is not set. Cannot initiate login."
    );
    return;
  }

  useEffect(() => {
    keycloak
      .init({
        redirectUri: redirectUri,
        onLoad: "login-required",
        checkLoginIframe: false,
      })
      .then((auth) => {
        setAuthenticated(auth);
        if (auth) {
          console.log("Authenticated ✅");
          console.log("Access Token:", keycloak.token);
        } else {
          keycloak.login();
        }
      })
      .catch((err) => console.error("Keycloak init failed", err));
  }, []);

  if (!authenticated) return <div>Loading...</div>;

  return <>{children}</>;
};
