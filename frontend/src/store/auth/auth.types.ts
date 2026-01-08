export interface AuthUser {
  id: string;
  username: string;
  email?: string;

  first_name?: string;
  last_name?: string;
  full_name?: string;
  is_admin: boolean;
  realm_roles: string[];
  client_roles: Record<string, string[]>;
  enabled: boolean;
  email_verified: boolean;
  require_pwd_change: boolean;
  last_login_at?: string;
  created_at?: string;
  updated_at?: string;
}

export type Role = "admin" | "user" | "manager";
