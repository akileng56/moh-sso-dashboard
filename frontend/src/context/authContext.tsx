import { createContext } from "react";

export interface AuthContextType {
  authenticated: boolean;
  accessToken: string | null;
  logout: () => void;
}

export const AuthContext = createContext<AuthContextType>({
  authenticated: false,
  accessToken: null,
  logout: () => {},
});
