export interface AuthUser {
  id: string;
  username: string;
  email?: string;
  first_name?: string;
  last_name?: string;
  enabled: boolean;
  is_admin: boolean;
  client_roles: Record<string, string[]>;
  last_login_at?: string;
  created_at?: string;
  updated_at?: string;
}

export type Role = "admin" | "user" | "manager";
