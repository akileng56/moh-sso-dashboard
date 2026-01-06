export interface AuthUser {
  id: string;
  username: string;
  email?: string;
  is_admin: boolean;
  client_roles: Record<string, string[]>;
}

export type Role = "admin" | "user" | "manager";
