package keycloak

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// -------------------------------------------------------------------
// DTOs
// -------------------------------------------------------------------

type ClientInfo struct {
	ID           string   `json:"id"`
	ClientID     string   `json:"clientId"`
	Name         string   `json:"name,omitempty"`
	Description  string   `json:"description,omitempty"`
	BaseURL      string   `json:"baseUrl,omitempty"`
	RootURL      string   `json:"rootUrl,omitempty"`
	RedirectURIs []string `json:"redirectUris,omitempty"`
	WebOrigins   []string `json:"webOrigins,omitempty"`
	PublicClient bool     `json:"publicClient"`
	Enabled      bool     `json:"enabled"`
	Protocol     string   `json:"protocol"`
}

type CreateClientParams struct {
	ClientID               string   `json:"clientId"`
	Name                   string   `json:"name,omitempty"`
	Description            string   `json:"description,omitempty"`
	BaseURL                string   `json:"baseUrl,omitempty"`
	RootURL                string   `json:"rootUrl,omitempty"`
	RedirectURIs           []string `json:"redirectUris,omitempty"`
	WebOrigins             []string `json:"webOrigins,omitempty"`
	PublicClient           bool     `json:"publicClient"`
	Secret                 string   `json:"secret,omitempty"`
	Protocol               string   `json:"protocol"`
	StandardFlowEnabled    bool     `json:"standardFlowEnabled"`
	ImplicitFlowEnabled    bool     `json:"implicitFlowEnabled"`
	DirectAccessGrants     bool     `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled bool     `json:"serviceAccountsEnabled"`
	Enabled                bool     `json:"enabled"`
}

// -------------------------------------------------------------------
// Realm Management
// -------------------------------------------------------------------

func (c *Client) EnsureRealmExists(realmName string) error {
	res, err := c.Get(fmt.Sprintf("admin/realms/%s", realmName))
	if err == nil && res.StatusCode == http.StatusOK {
		res.Body.Close()
		return nil
	}

	if res != nil {
		res.Body.Close()
	}

	payload := map[string]any{
		"realm":   realmName,
		"enabled": true,
	}
	res, err = c.Post("admin/realms", payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to create realm: %s", string(body))
	}
	return nil
}

// -------------------------------------------------------------------
// Client Management
// -------------------------------------------------------------------

// CreateClient with full configuration support
func (c *Client) CreateClient(opts CreateClientParams) error {
	if opts.Protocol == "" {
		opts.Protocol = "openid-connect"
	}
	if opts.Enabled == false {
		opts.Enabled = true
	}

	res, err := c.Post(fmt.Sprintf("admin/realms/%s/clients", c.Realm), opts)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to create client: %s", string(body))
	}

	return nil
}

// GetClientByClientID using Keycloak's search API
func (c *Client) GetClientByClientID(clientID string) (*ClientInfo, error) {
	query := url.Values{}
	query.Set("clientId", clientID)

	res, err := c.Get(fmt.Sprintf("admin/realms/%s/clients?%s", c.Realm, query.Encode()))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed checking clientId '%s': %s", clientID, string(body))
	}

	var clients []ClientInfo
	if err := json.NewDecoder(res.Body).Decode(&clients); err != nil {
		return nil, err
	}

	if len(clients) == 0 {
		return nil, nil // not found
	}

	return &clients[0], nil
}

// GetClientByID retrieves the actual client object
func (c *Client) GetClientByID(id string) (*ClientInfo, error) {
	res, err := c.Get(fmt.Sprintf("admin/realms/%s/clients/%s", c.Realm, id))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to get client: %s", string(body))
	}

	var cli ClientInfo
	if err := json.NewDecoder(res.Body).Decode(&cli); err != nil {
		return nil, err
	}

	return &cli, nil
}

// List all clients in the realm
func (c *Client) ListClients() ([]ClientInfo, error) {
	res, err := c.Get(fmt.Sprintf("admin/realms/%s/clients", c.Realm))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to list clients: %s", string(body))
	}

	var clients []ClientInfo
	if err := json.NewDecoder(res.Body).Decode(&clients); err != nil {
		return nil, err
	}

	return clients, nil
}

func (c *Client) DeleteClient(id string) error {
	res, err := c.Delete(fmt.Sprintf("admin/realms/%s/clients/%s", c.Realm, id))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to delete client: %s", string(body))
	}

	return nil
}

// -------------------------------------------------------------------
// User Management
// -------------------------------------------------------------------

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Enabled  bool   `json:"enabled"`
}

func (c *Client) CreateUser(username, password, role string) error {
	payload := map[string]any{
		"username": username,
		"enabled":  true,
		"credentials": []map[string]any{
			{
				"type":      "password",
				"value":     password,
				"temporary": false,
			},
		},
	}

	res, err := c.Post(fmt.Sprintf("admin/realms/%s/users", c.Realm), payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to create user: %s", string(body))
	}
	return nil
}

func (c *Client) ListUsers() ([]UserInfo, error) {
	res, err := c.Get(fmt.Sprintf("admin/realms/%s/users", c.Realm))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to list users: %s", string(body))
	}

	var users []UserInfo
	if err := json.NewDecoder(res.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users, nil
}
