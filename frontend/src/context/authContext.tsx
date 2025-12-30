import { createContext } from "react";

export interface AuthUser {
  id: string;
  username: string;
  email?: string;
  is_admin: boolean;
  client_roles: Record<string, string[]>;
}

export interface AuthContextType {
  authenticated: boolean;
  accessToken: string | null;

  user: AuthUser | null;
  isAdmin: boolean;
  loading: boolean;
  logout: () => void;
}

export const AuthContext = createContext<AuthContextType>({
  authenticated: false,
  accessToken: null,
  user: null,
  isAdmin: false,
  loading: true,
  logout: () => {},
});
