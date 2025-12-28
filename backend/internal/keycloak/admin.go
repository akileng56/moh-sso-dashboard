package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/moh-sso-dashboard/internal/model"
)

// -------------------------------------------------------------------
// DTOs
// -------------------------------------------------------------------

type ClientInfo struct {
	// --- Base Identification & URLs ---
	ID          string `json:"id,omitempty"`          // Internal unique ID (UUID)
	ClientID    string `json:"clientId"`              // The client ID used for OAuth/OIDC protocol
	Name        string `json:"name,omitempty"`        // Display name for the client
	Description string `json:"description,omitempty"` // Detailed description
	BaseURL     string `json:"baseUrl,omitempty"`     // Base URL for the client application
	RootURL     string `json:"rootUrl,omitempty"`     // Root URL for relative paths
	AdminURL    string `json:"adminUrl,omitempty"`    // URL to the client's admin console

	// --- Configuration & Status ---
	Enabled      bool   `json:"enabled"`              // Whether the client is active
	PublicClient bool   `json:"publicClient"`         // True for browser-based apps without a secret
	BearerOnly   bool   `json:"bearerOnly,omitempty"` // True if the client only accepts bearer tokens
	Protocol     string `json:"protocol"`             // Protocol used: "openid-connect" or "saml"

	// --- Security & Credentials ---
	Secret                  string `json:"secret,omitempty"`                  // Shared secret for confidential clients (if not publicClient)
	ClientAuthenticatorType string `json:"clientAuthenticatorType,omitempty"` // e.g., "client-secret" or "client-jwt"

	// --- Flows & Settings ---
	RedirectURIs              []string `json:"redirectUris,omitempty"`              // Valid redirect URIs after successful authentication
	WebOrigins                []string `json:"webOrigins,omitempty"`                // List of allowed CORS origins
	StandardFlowEnabled       bool     `json:"standardFlowEnabled,omitempty"`       // Authorization Code Flow
	ImplicitFlowEnabled       bool     `json:"implicitFlowEnabled,omitempty"`       // Implicit Flow (legacy)
	DirectAccessGrantsEnabled bool     `json:"directAccessGrantsEnabled,omitempty"` // Resource Owner Password Credentials Grant
	ServiceAccountsEnabled    bool     `json:"serviceAccountsEnabled,omitempty"`    // Enables a service account for machine-to-machine

	// --- Scopes & Mappers ---
	FullScopeAllowed     bool     `json:"fullScopeAllowed,omitempty"`     // If true, all realm roles and scopes are granted
	DefaultClientScopes  []string `json:"defaultClientScopes,omitempty"`  // List of required scopes (client scopes)
	OptionalClientScopes []string `json:"optionalClientScopes,omitempty"` // List of optional scopes

	// --- Customization ---
	Attributes map[string]string `json:"attributes,omitempty"` // Custom key/value client settings

	// --- Complex Fields (often require separate structs/endpoints) ---
	ProtocolMappers []ProtocolMapper `json:"protocolMappers,omitempty"` // Defines how claims are mapped into tokens
}

// --- Auxiliary Struct for ProtocolMappers (as defined in the previous response) ---
type ProtocolMapper struct {
	ID             string            `json:"id,omitempty"`
	Name           string            `json:"name,omitempty"`
	Protocol       string            `json:"protocol,omitempty"`
	ProtocolMapper string            `json:"protocolMapper,omitempty"`
	Config         map[string]string `json:"config,omitempty"`
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
	c.BaseURL = "http://keycloak:8080"
	if opts.Protocol == "" {
		opts.Protocol = "openid-connect"
	}
	if !opts.Enabled {
		opts.Enabled = true
	}

	res, err := c.Post("clients", opts)
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
	c.BaseURL = "http://keycloak:8080"
	query := url.Values{}
	query.Set("clientId", clientID)

	res, err := c.Get("clients?" + query.Encode())
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
		return nil, nil
	}

	return &clients[0], nil
}

// GetClientByID retrieves the actual client object
func (c *Client) GetClientByID(id string) (*ClientInfo, error) {
	c.BaseURL = "http://keycloak:8080"
	res, err := c.Get("clients/" + id)
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
	c.BaseURL = "http://keycloak:8080"
	res, err := c.Get("clients")
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
	c.BaseURL = "http://keycloak:8080"
	res, err := c.Delete("clients/" + id)
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
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Email       string    `json:"email"`
	Enabled     bool      `json:"enabled"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	LastLoginAt time.Time `json:"lastLoginAt"`
}

type CreateUserRequest struct {
	Username      string `json:"username"`
	Email         string `json:"email"`
	FirstName     string `json:"firstName,omitempty"`
	LastName      string `json:"lastName,omitempty"`
	Enabled       bool   `json:"enabled"`
	EmailVerified bool   `json:"emailVerified"`
	Credentials   []struct {
		Type      string `json:"type"`
		Value     string `json:"value"`
		Temporary bool   `json:"temporary"`
	} `json:"credentials,omitempty"`
}

type UserRep struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Enabled  bool   `json:"enabled"`
}

type RoleRep struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) CreateUser(user *model.User) (string, error) {
	payload := map[string]any{
		"username":      user.Username,
		"email":         user.Email,
		"firstName":     user.FirstName,
		"lastName":      user.LastName,
		"enabled":       user.Enabled,
		"emailVerified": true,
		"credentials": []map[string]any{
			{
				"type":      "password",
				"value":     "",
				"temporary": false,
			},
		},
	}

	res, err := c.Post(fmt.Sprintf("admin/realms/%s/users", c.Realm), payload)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("failed to create user: %s", string(body))
	}

	location := res.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no Location header returned by Keycloak")
	}

	parts := strings.Split(strings.TrimSpace(location), "/")
	kcID := parts[len(parts)-1]

	if kcID == "" {
		return "", fmt.Errorf("failed to parse Keycloak user ID from Location header")
	}

	return kcID, nil
}

func (c *Client) FindUsers(ctx context.Context, q string, exact bool) ([]UserRep, error) {
	u := fmt.Sprintf("%s/admin/realms/%s/users", c.BaseURL, c.Realm)

	v := url.Values{}
	if q != "" {
		v.Set("search", q)
	}
	if exact {
		v.Set("exact", "true")
	}
	u = u + "?" + v.Encode()

	res, err := c.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("keycloak find users failed: status=%d body=%s", res.StatusCode, string(b))
	}

	var out []UserRep
	return out, json.NewDecoder(res.Body).Decode(&out)
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

// -------------------------------------------------------------------
// Update User in Keycloak
// -------------------------------------------------------------------
func (c *Client) UpdateUser(user *model.User) error {
	if user.ID == "" {
		return fmt.Errorf("missing Keycloak user ID")
	}

	payload := map[string]any{
		"username":  user.Username,
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"enabled":   user.Enabled,
	}

	res, err := c.Put(
		fmt.Sprintf("admin/realms/%s/users/%s", c.Realm, user.ID),
		payload,
	)
	if err != nil {
		return fmt.Errorf("keycloak update request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to update user in keycloak: %s", string(body))
	}

	return nil
}

// -------------------------------------------------------------------
// Delete User in Keycloak
// -------------------------------------------------------------------
func (c *Client) DeleteUser(userID string) error {
	if userID == "" {
		return fmt.Errorf("invalid user ID")
	}

	res, err := c.Delete(
		fmt.Sprintf("admin/realms/%s/users/%s", c.Realm, userID),
	)
	if err != nil {
		return fmt.Errorf("keycloak delete request failed: %w", err)
	}
	defer res.Body.Close()

	// Keycloak returns 204 No Content on success
	if res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to delete user in keycloak: %s", string(body))
	}

	return nil
}

// realm
func (c *Client) GetRealmRoleByName(ctx context.Context, roleName string) (*RoleRep, error) {

	res, err := c.Get(fmt.Sprintf("%s/admin/realms/%s/roles/%s", c.BaseURL, c.Realm, url.PathEscape(roleName)))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("keycloak get role failed: status=%d body=%s", res.StatusCode, string(b))
	}

	var out RoleRep
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) AddRealmRoleToUser(ctx context.Context, userID string, role RoleRep) error {

	payload := []RoleRep{role}
	bs, _ := json.Marshal(payload)

	bytesData := bytes.NewReader(bs)

	u := fmt.Sprintf("%s/admin/realms/%s/users/%s/role-mappings/realm", c.BaseURL, c.Realm, userID)

	res, err := c.Post(u, bytesData)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("keycloak add role failed: status=%d body=%s", res.StatusCode, string(b))
	}
	return nil
}

func (c *Client) SendUserOnboardingEmail(
	ctx context.Context,
	userID string,
) error {

	actions := []string{
		"UPDATE_PASSWORD",
		"VERIFY_EMAIL",
	}

	url := fmt.Sprintf(
		"%s/admin/realms/%s/users/%s/execute-actions-email",
		c.BaseURL,
		c.Realm,
		userID,
	)

	req, _ := json.Marshal(actions)

	resp, err := c.Put(url, bytes.NewBuffer(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("failed to send email: %s", resp.Status)
	}

	return nil
}
