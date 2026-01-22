package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/utils"
)

type ClientInfo struct {
	ID                        string            `json:"id"`                        // Internal unique ID (UUID)
	ClientID                  string            `json:"clientId"`                  // The client ID used for OAuth/OIDC protocol
	Name                      string            `json:"name"`                      // Display name for the client
	Description               string            `json:"description,"`              // Detailed description
	BaseURL                   string            `json:"baseUrl"`                   // Base URL for the client application
	RootURL                   string            `json:"rootUrl"`                   // Root URL for relative paths
	AdminURL                  string            `json:"adminUrl"`                  // URL to the client's admin console
	Enabled                   bool              `json:"enabled"`                   // Whether the client is active
	PublicClient              bool              `json:"publicClient"`              // True for browser-based apps without a secret
	BearerOnly                bool              `json:"bearerOnly"`                // True if the client only accepts bearer tokens
	Protocol                  string            `json:"protocol"`                  // Protocol used: "openid-connect" or "saml"
	Secret                    string            `json:"secret"`                    // Shared secret for confidential clients (if not publicClient)
	ClientAuthenticatorType   string            `json:"clientAuthenticatorType"`   // e.g., "client-secret" or "client-jwt"
	RedirectURIs              []string          `json:"redirectUris"`              // Valid redirect URIs after successful authentication
	WebOrigins                []string          `json:"webOrigins"`                // List of allowed CORS origins
	StandardFlowEnabled       bool              `json:"standardFlowEnabled"`       // Authorization Code Flow
	ImplicitFlowEnabled       bool              `json:"implicitFlowEnabled"`       // Implicit Flow (legacy)
	DirectAccessGrantsEnabled bool              `json:"directAccessGrantsEnabled"` // Resource Owner Password Credentials Grant
	ServiceAccountsEnabled    bool              `json:"serviceAccountsEnabled"`    // Enables a service account for machine-to-machine
	FullScopeAllowed          bool              `json:"fullScopeAllowed"`          // If true, all realm roles and scopes are granted
	DefaultClientScopes       []string          `json:"defaultClientScopes"`       // List of required scopes (client scopes)
	OptionalClientScopes      []string          `json:"optionalClientScopes"`      // List of optional scopes
	Attributes                map[string]string `json:"attributes"`                // Custom key/value client settings
	ProtocolMappers           []ProtocolMapper  `json:"protocolMappers"`           // Defines how claims are mapped into tokens
}

type ProtocolMapper struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Protocol       string            `json:"protocol"`
	ProtocolMapper string            `json:"protocolMapper"`
	Config         map[string]string `json:"config"`
}

type CreateClientParams struct {
	ClientID               string            `json:"clientId"`
	Name                   string            `json:"name"`
	Description            string            `json:"description"`
	BaseURL                string            `json:"baseUrl"`
	RootURL                string            `json:"rootUrl"`
	RedirectURIs           []string          `json:"redirectUris"`
	WebOrigins             []string          `json:"webOrigins"`
	PublicClient           bool              `json:"publicClient"`
	Secret                 string            `json:"secret"`
	Protocol               string            `json:"protocol"`
	StandardFlowEnabled    bool              `json:"standardFlowEnabled"`
	ImplicitFlowEnabled    bool              `json:"implicitFlowEnabled"`
	DirectAccessGrants     bool              `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled bool              `json:"serviceAccountsEnabled"`
	Enabled                bool              `json:"enabled"`
	Attributes             map[string]string `json:"attributes"`
}

type UserInfo struct {
	ID                  string              `json:"id"`
	Username            string              `json:"username"`
	Email               string              `json:"email"`
	FirstName           string              `json:"firstName"`
	LastName            string              `json:"lastName"`
	Enabled             bool                `json:"enabled"`
	EmailVerified       bool                `json:"emailVerified"`
	RequiredActions     []string            `json:"requiredActions"`
	AccountStatus       string              `json:"accountStatus"`
	Roles               []string            `json:"roles"`
	ClientRoles         map[string][]string `json:"clientRoles"`
	LastLoginAt         *time.Time          `json:"lastLoginAt"`
	LastLoginIP         string              `json:"lastLoginIp"`
	FailedLoginAttempts int                 `json:"failedLoginAttempts"`
	TemporarilyLocked   bool                `json:"temporarilyLocked"`
	CreatedAt           time.Time           `json:"createdAt"`
	UpdatedAt           time.Time           `json:"updatedAt"`
	Source              string              `json:"source"`
	ImportedAt          *time.Time          `json:"importedAt"`
	Notes               string              `json:"notes"`
	DisplayName         string              `json:"displayName"`
	NeverLoggedIn       bool                `json:"neverLoggedIn"`
}

type CreateUserRequest struct {
	Username      string `json:"username"`
	Email         string `json:"email"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Enabled       bool   `json:"enabled"`
	EmailVerified bool   `json:"emailVerified"`
	Credentials   []struct {
		Type      string `json:"type"`
		Value     string `json:"value"`
		Temporary bool   `json:"temporary"`
	} `json:"credentials"`
}

type UserRep struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Enabled  bool   `json:"enabled"`
}

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Composite   bool   `json:"composite"`
	ClientRole  bool   `json:"clientRole"`
}

type UserClientRoleAssignment struct {
	ClientID   string   `json:"clientId"`
	ClientName string   `json:"clientName"`
	Roles      []string `json:"roles"`
}

type ClientRoleRep struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoleRep struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"descritpion"`
}

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

//------------------------------------------
// client management
//------------------------------------------

func (c *Client) CreateClient(opts CreateClientParams) (string, error) {
	if opts.Protocol == "" {
		opts.Protocol = "openid-connect"
	}
	if !opts.Enabled {
		opts.Enabled = true
	}

	res, err := c.Post("clients", opts)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("failed to create client: %s", string(body))
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

func (c *Client) GetClientByClientID(clientID string) (*ClientInfo, error) {
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

func (c *Client) GetClientByID(id string) (*ClientInfo, error) {
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

func (c *Client) ListClients() ([]ClientInfo, error) {
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

func (c *Client) UpdateClient(id string, payload *model.Client) error {
	res, err := c.Put("clients/"+id, payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf(
			"failed to update client (status %d): %s",
			res.StatusCode,
			string(body),
		)
	}

	return nil
}

func (c *Client) UpdateClientEnabled(
	ctx context.Context,
	clientId string, // this is Keycloak "clientId" (e.g. dashboard), NOT UUID
	enabled bool,
) error {

	clientUUID, err := c.resolveClientUUID(ctx, clientId)
	if err != nil {
		return fmt.Errorf("failed to resolve client UUID: %w", err)
	}

	u := fmt.Sprintf("clients/%s", clientUUID)

	res, err := c.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("get client failed [%d]: %s", res.StatusCode, string(b))
	}

	var kcClient ClientInfo
	if err := json.NewDecoder(res.Body).Decode(&kcClient); err != nil {
		return fmt.Errorf("decode client failed: %w", err)
	}

	kcClient.Enabled = enabled

	putRes, err := c.Put(u, kcClient)
	if err != nil {
		return err
	}
	defer putRes.Body.Close()

	if putRes.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(putRes.Body)
		return fmt.Errorf("update client enabled failed [%d]: %s", putRes.StatusCode, string(b))
	}

	return nil
}

func (c *Client) DeleteClient(id string) error {
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

// -----------------------------------------
// user management
// -----------------------------------------

func (c *Client) CreateUser(user *model.User) (string, error) {
	payload := map[string]any{
		"username":      user.Username,
		"email":         user.Email,
		"firstName":     user.FirstName,
		"lastName":      user.LastName,
		"enabled":       user.Enabled,
		"emailVerified": user.EmailVerified,
		"credentials": []map[string]any{
			{
				"type":      "password",
				"value":     uuid.NewString(),
				"temporary": true,
			},
		},
	}

	res, err := c.Post("users", payload)
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
	u := fmt.Sprintf("realms/%s/users", c.Realm)

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
	res, err := c.Get("users")
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

func (c *Client) GetUser(userID string) (*UserInfo, error) {
	res, err := c.Get("users/" + userID)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to get user: %s", string(body))
	}

	var user UserInfo
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

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

	res, err := c.Put(("users" + user.ID),
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

func (c *Client) DeleteUser(userID string) error {
	if userID == "" {
		return fmt.Errorf("invalid user ID")
	}

	res, err := c.Delete("users/" + userID)
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

// -------------------------------------------------------------------
// Realm roles Managemet
// -------------------------------------------------------------------

func (c *Client) CreateRealmRole(ctx context.Context, roleName, description string) error {
	u := fmt.Sprintf("roles")
	payload := CreateRoleRequest{
		Name:        roleName,
		Description: description,
		Composite:   false,
		ClientRole:  false,
	}

	res, err := c.Post(u, payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 409 = already exists (safe for bootstrap)
	if res.StatusCode == http.StatusConflict {
		return nil
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("create realm role failed: %s", string(b))
	}

	return nil
}

func (c *Client) ListRealmRoles(ctx context.Context) ([]RoleRep, error) {
	u := fmt.Sprintf("roles")

	res, err := c.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("list realm roles failed: %s", string(b))
	}

	var roles []RoleRep
	return roles, json.NewDecoder(res.Body).Decode(&roles)
}

func (c *Client) GetRealmRoleByName(ctx context.Context, roleName string) (*RoleRep, error) {
	res, err := c.Get(fmt.Sprintf("roles/%s", url.PathEscape(roleName)))
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

	u := fmt.Sprintf("users/%s/role-mappings/realm", userID)

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

func (c *Client) RemoveRealmRoleFromUser(
	ctx context.Context,
	userID string,
	role RoleRep,
) error {

	u := fmt.Sprintf(
		"users/%s/role-mappings/realm",
		userID,
	)

	bs, _ := json.Marshal([]RoleRep{role})

	bytesData := bytes.NewReader(bs)

	res, err := c.Post(u, bytesData)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("remove realm role failed: %s", string(b))
	}

	return nil
}

func (c *Client) BootstrapRealmRoles(ctx context.Context) error {
	roles := []string{
		"admin",
		"user",
		"manager",
	}

	for _, r := range roles {
		if err := c.CreateRealmRole(ctx, r, "system role"); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------------------------------------------------
// ROLES: Client roles
// -------------------------------------------------------------------

func (c *Client) CreateClientRole(
	ctx context.Context,
	clientId string,
	req *model.CreateClientRoleRequest,
) error {
	clientUUID, err := c.resolveClientUUID(ctx, clientId)
	if err != nil {
		return fmt.Errorf("failed to resolve client UUID: %w", err)
	}

	path := fmt.Sprintf(
		"clients/%s/roles",
		clientUUID,
	)

	log.Printf("[KEYCLOAK] create client role | clientId=%s uuid=%s role=%s",
		clientId,
		clientUUID,
		req.Role,
	)

	payload := map[string]any{
		"name":        req.Role,
		"description": req.Description,
		"clientRole":  true,
	}

	res, err := c.Post(path, payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusConflict {
		return nil
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf(
			"create client role failed [%d]: %s",
			res.StatusCode,
			string(b),
		)
	}

	return nil
}

func (c *Client) GetClientRoleByName(
	ctx context.Context,
	clientID string,
	roleName string,
) (*ClientRoleRep, error) {
	u := fmt.Sprintf(
		"clients/%s/roles/%s",
		clientID,
		url.PathEscape(roleName),
	)

	res, err := c.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("get client role failed: %s", string(b))
	}

	var out ClientRoleRep
	return &out, json.NewDecoder(res.Body).Decode(&out)
}

func (c *Client) GetUserClientRoles(
	ctx context.Context,
	userID string,
	clientUUID string,
) ([]ClientRoleRep, error) {

	u := fmt.Sprintf(
		"users/%s/role-mappings/clients/%s",
		userID,
		clientUUID,
	)

	res, err := c.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf(
			"get user client roles failed [%d]: %s",
			res.StatusCode,
			string(b),
		)
	}

	var out []ClientRoleRep
	return out, json.NewDecoder(res.Body).Decode(&out)
}

func (c *Client) AssignClientRolesToUser(
	ctx context.Context,
	userID string,
	clientUUID string,
	roles []string,
) error {

	if len(roles) == 0 {
		return nil
	}

	var payload []ClientRoleRep

	for _, roleName := range roles {
		role, err := c.GetClientRoleByName(ctx, clientUUID, roleName)
		if err != nil {
			return err
		}
		payload = append(payload, *role)
	}

	u := fmt.Sprintf(
		"users/%s/role-mappings/clients/%s",
		userID,
		clientUUID,
	)

	bs, _ := json.Marshal(payload)

	res, err := c.Post(u, bytes.NewReader(bs))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("assign client roles failed: %s", string(b))
	}

	return nil
}

func (c *Client) RemoveClientRoleFromUser(
	ctx context.Context,
	userID, clientID string,
	role ClientRoleRep,
) error {

	u := fmt.Sprintf(
		"users/%s/role-mappings/clients/%s",
		userID, clientID,
	)

	bs, _ := json.Marshal([]ClientRoleRep{role})

	res, err := c.Post(u, bytes.NewReader(bs))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("remove client role failed: %s", string(b))
	}
	return nil
}

func (c *Client) RemoveClientRolesFromUser(
	ctx context.Context,
	userID string,
	clientUUID string,
	roles []string,
) error {

	if len(roles) == 0 {
		return nil
	}

	payload := make([]ClientRoleRep, 0, len(roles))

	for _, roleName := range roles {
		role, err := c.GetClientRoleByName(ctx, clientUUID, roleName)
		if err != nil {
			return err
		}
		payload = append(payload, *role)
	}

	u := fmt.Sprintf(
		"users/%s/role-mappings/clients/%s",
		userID,
		clientUUID,
	)

	bs, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	res, err := c.DeleteWithBody(
		u,
		bytes.NewReader(bs),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("remove client roles failed [%d]: %s", res.StatusCode, string(b))
	}

	return nil
}

func (c *Client) ListClientRoles(
	ctx context.Context,
	clientId string, // logical clientId (e.g. "dashboard")
) ([]ClientRoleRep, error) {

	clientUUID, err := c.resolveClientUUID(ctx, clientId)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve client UUID: %w", err)
	}

	path := fmt.Sprintf("clients/%s/roles", clientUUID)

	res, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf(
			"list client roles failed [%d]: %s",
			res.StatusCode,
			string(b),
		)
	}

	var roles []ClientRoleRep
	if err := json.NewDecoder(res.Body).Decode(&roles); err != nil {
		return nil, err
	}

	return roles, nil
}

func (c *Client) DeleteClientRole(
	ctx context.Context,
	clientId string, // logical clientId e.g. "dashboard-web"
	role string,
) error {

	clientUUID, err := c.resolveClientUUID(ctx, clientId)
	if err != nil {
		return fmt.Errorf("failed to resolve client UUID: %w", err)
	}

	path := fmt.Sprintf(
		"clients/%s/roles/%s",
		clientUUID,
		url.PathEscape(role),
	)

	log.Printf(
		"[KEYCLOAK] delete client role | clientId=%s uuid=%s role=%s",
		clientId,
		clientUUID,
		role,
	)

	res, err := c.Delete(path)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf(
			"delete client role failed [%d]: %s",
			res.StatusCode,
			string(b),
		)
	}

	return nil
}

func (c *Client) resolveClientUUID(
	ctx context.Context,
	clientId string,
) (string, error) {

	path := "clients?clientId=" + url.QueryEscape(clientId)

	res, err := c.Get(path)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf(
			"failed to resolve client UUID [%d]: %s",
			res.StatusCode,
			string(b),
		)
	}

	var clients []struct {
		ID       string `json:"id"`
		ClientID string `json:"clientId"`
	}

	if err := json.NewDecoder(res.Body).Decode(&clients); err != nil {
		return "", err
	}

	if len(clients) == 0 {
		return "", fmt.Errorf("client not found in keycloak: %s", clientId)
	}

	return clients[0].ID, nil
}

// support helper functions
func (c *Client) SendUserOnboardingEmail(
	ctx context.Context,
	userID string,
) error {

	actions := []string{
		"UPDATE_PASSWORD",
		"VERIFY_EMAIL",
	}

	url := fmt.Sprintf(
		"users/%s/execute-actions-email",
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

// -------------------------------------------------------------------
// USERS: Reset password (ADMIN)
// -------------------------------------------------------------------

func (c *Client) ResetUserPassword(
	ctx context.Context,
	userID string, // Keycloak user UUID
) error {

	tmpPassword := generateTemporaryPassword()

	payload := map[string]any{
		"type":      "password",
		"value":     tmpPassword,
		"temporary": true,
	}

	u := fmt.Sprintf("users/%s/reset-password", userID)

	res, err := c.Put(u, payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf(
			"reset user password failed [%d]: %s",
			res.StatusCode,
			string(b),
		)
	}

	actions := []string{"UPDATE_PASSWORD"}

	actionsRes, err := c.Put(
		fmt.Sprintf("users/%s/execute-actions-email", userID),
		actions,
	)
	if err != nil {
		return err
	}
	defer actionsRes.Body.Close()

	if actionsRes.StatusCode != http.StatusNoContent &&
		actionsRes.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(actionsRes.Body)
		return fmt.Errorf(
			"execute required actions failed [%d]: %s",
			actionsRes.StatusCode,
			string(b),
		)
	}

	return nil
}

func generateTemporaryPassword() string {
	return utils.RandomString(16) + "!A1"
}
