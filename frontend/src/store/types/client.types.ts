export interface Client {
  id: string;
  clientId: string;
  name: string;
  description?: string;
  enabled: boolean;
  publicClient: boolean;
  baseUrl?: string;
  redirectUris: string[];
  createdAt?: string;
  updatedAt?: string;
}

export interface CreateClientPayload {
  clientId: string;
  name: string;
  description?: string;
  publicClient?: boolean;
  redirectUris?: string[];
}

export interface UpdateClientPayload {
  name?: string;
  description?: string;
  publicClient?: boolean;
  redirectUris?: string[];
}
