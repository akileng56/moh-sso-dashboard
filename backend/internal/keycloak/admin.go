package keycloak

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// --- Realm Management ---

// EnsureRealmExists checks if a realm exists, and creates it if missing.
func (c *Client) EnsureRealmExists(realmName string) error {
	res, err := c.Get(fmt.Sprintf("admin/realms/%s", realmName))
	if err == nil && res.StatusCode == http.StatusOK {
		// realm already exists
		res.Body.Close()
		return nil
	}

	// create realm if not found
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

// --- Client Management ---

type ClientInfo struct {
	ID       string `json:"id"`
	ClientID string `json:"clientId"`
	Name     string `json:"name,omitempty"`
	BaseURL  string `json:"baseUrl,omitempty"`
}

// CreateClient creates a new Keycloak client (application)
func (c *Client) CreateClient(name, redirectUri, baseUrl string) error {
	payload := map[string]any{
		"clientId":            name,
		"name":                name,
		"enabled":             true,
		"publicClient":        true,
		"redirectUris":        []string{redirectUri},
		"baseUrl":             baseUrl,
		"protocol":            "openid-connect",
		"standardFlowEnabled": true,
	}

	res, err := c.Post(fmt.Sprintf("admin/realms/%s/clients", c.Realm), payload)
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

// ListClients retrieves all clients from the realm
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

// DeleteClient removes a Keycloak client by ID
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

// --- User Management ---

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Enabled  bool   `json:"enabled"`
}

// CreateUser creates a new user with password credentials
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

// ListUsers retrieves all users in the realm
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
