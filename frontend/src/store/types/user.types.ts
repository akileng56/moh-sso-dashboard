export interface User {
  id: string;
  username: string;
  email?: string;
  first_name?: string;
  last_name?: string;
  enabled: boolean;
  is_admin: boolean;
  is_active: boolean;
  email_verified: boolean;
  client_roles: Record<string, string[]>;
  last_login_at?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CreateUserPayload {
  username: string;
  email: string;
  first_name?: string;
  last_name?: string;
  is_admin?: boolean;
  is_active?: boolean;
  enabled?: boolean;
  email_verified: boolean;
  client_roles?: Record<string, string[]>;
}
