export interface Client {
  client_id: string;
  name: string;
  description?: string;
  enabled: boolean;
  baseUrl?: string;
  public_client: boolean;
  redirect_uris: string[];
  created_at?: string;
  updated_at?: string;
}

export interface CreateClientPayload {
  client_id: string;
  name: string;
  description?: string;
  public_client?: boolean;
  redirect_uris?: string[];
}

export interface UpdateClientPayload {
  name?: string;
  description?: string;
  public_client?: boolean;
  redirect_uris?: string[];
}
