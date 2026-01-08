package keycloak

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Client represents the Keycloak API client
type Client struct {
	BaseURL    string
	Realm      string
	ClientID   string
	Secret     string
	Token      string
	httpClient *http.Client
}

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	IDToken          string `json:"id_token"`
}

type AuthUser struct {
	ID               string              `json:"id"`
	Username         string              `json:"username"`
	Email            string              `json:"email,omitempty"`
	FirstName        string              `json:"first_name,omitempty"`
	LastName         string              `json:"last_name,omitempty"`
	FullName         string              `json:"full_name,omitempty"`
	IsAdmin          bool                `json:"is_admin"`     // derived (realm admin)
	RealmRoles       []string            `json:"realm_roles"`  // e.g. ["admin"]
	ClientRoles      map[string][]string `json:"client_roles"` // client_id → roles
	Enabled          bool                `json:"enabled"`
	EmailVerified    bool                `json:"email_verified"`
	RequirePwdChange bool                `json:"require_pwd_change"`
	LastLoginAt      *time.Time          `json:"last_login_at,omitempty"`
	LoginIP          string              `json:"login_ip,omitempty"`
	UserAgent        string              `json:"user_agent,omitempty"`
	CreatedAt        time.Time           `json:"created_at,omitempty"`
}

// NewClient creates a new Keycloak client
func NewClient(baseURL, realm, clientID, secret string) *Client {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}

	baseURL = strings.TrimSuffix(baseURL, "/")

	httpClient := &http.Client{}

	return &Client{
		BaseURL:    baseURL,
		Realm:      realm,
		ClientID:   clientID,
		Secret:     secret,
		httpClient: httpClient,
	}
}

// Authenticate retrieves an access token using client credentials
func (c *Client) Authenticate() error {
	form := url.Values{}
	form.Add("grant_type", "client_credentials")
	form.Add("client_id", c.ClientID)
	form.Add("client_secret", c.Secret)

	url := fmt.Sprintf("http://keycloak:8080/realms/%s/protocol/openid-connect/token", c.Realm)
	res, err := c.httpClient.PostForm(url, form)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("authentication failed: %s", string(body))
	}

	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return err
	}
	c.Token = body.AccessToken
	return nil
}

// ExchangeCodeForToken exchanges the authorization code for an access token
func (c *Client) ExchangeCodeForToken(code, redirectURI string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("http://keycloak:8080/realms/%s/protocol/openid-connect/token", c.Realm)

	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("client_id", c.ClientID)
	form.Add("client_secret", c.Secret)
	form.Add("code", code)
	form.Add("redirect_uri", redirectURI)

	res, err := c.httpClient.PostForm(tokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed [%d]: %s", res.StatusCode, string(bodyBytes))
	}

	var tokenRes TokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenRes); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenRes, nil
}

// Get access token
func (c *Client) AccessToken(refreshToken string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("http://keycloak:8080/realms/%s/protocol/openid-connect/token", c.Realm)

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.Secret)

	res, err := c.httpClient.PostForm(tokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("gettting access  failed: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gettting access  failed [%d]: %s", res.StatusCode, string(bodyBytes))
	}

	var tokenRes TokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenRes); err != nil {
		return nil, fmt.Errorf("failed to decode access token response: %w", err)
	}

	return &tokenRes, nil
}

func (c *Client) LogOut(refreshToken string) error {
	if refreshToken == "" {
		return fmt.Errorf("missing refresh token for logout")
	}

	logoutURL := fmt.Sprintf(
		"http://keycloak:8080/realms/%s/protocol/openid-connect/logout",
		c.Realm,
	)

	// Prepare form data
	form := url.Values{}
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.Secret)
	form.Set("refresh_token", refreshToken)

	// Ensure httpClient exists
	httpClient := c.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	// Send logout request
	res, err := httpClient.PostForm(logoutURL, form)
	if err != nil {
		return fmt.Errorf("logout request failed: %w", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	// Keycloak returns 204 No Content on successful logout
	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"logout failed: status=%d response=%s",
			res.StatusCode,
			string(body),
		)
	}

	return nil
}

// get User profile / Info
func (c *Client) Me(accessToken string) (*AuthUser, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("token parse error: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// ----------------------------
	// Helpers (panic-safe)
	// ----------------------------
	getString := func(key string) string {
		if v, ok := claims[key].(string); ok {
			return v
		}
		return ""
	}

	getBool := func(key string) bool {
		if v, ok := claims[key].(bool); ok {
			return v
		}
		return false
	}

	// ----------------------------
	// Core identity
	// ----------------------------
	id := getString("sub")
	username := getString("preferred_username")
	email := getString("email")

	firstName := getString("given_name")
	lastName := getString("family_name")

	fullName := strings.TrimSpace(
		fmt.Sprintf("%s %s", firstName, lastName),
	)
	if fullName == "" {
		fullName = getString("name")
	}

	// ----------------------------
	// Realm roles
	// ----------------------------
	realmRoles := []string{}
	isAdmin := false

	if ra, ok := claims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := ra["roles"].([]interface{}); ok {
			for _, r := range roles {
				if role, ok := r.(string); ok {
					realmRoles = append(realmRoles, role)
					if role == "admin" {
						isAdmin = true
					}
				}
			}
		}
	}

	// ----------------------------
	// Client roles (resource_access)
	// ----------------------------
	clientRoles := map[string][]string{}

	if ra, ok := claims["resource_access"].(map[string]interface{}); ok {
		for clientID, raw := range ra {
			block, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}

			if roles, ok := block["roles"].([]interface{}); ok {
				for _, r := range roles {
					if role, ok := r.(string); ok {
						clientRoles[clientID] = append(
							clientRoles[clientID],
							role,
						)
					}
				}
			}
		}
	}

	// ----------------------------
	// Session & account metadata
	// ----------------------------
	var lastLoginAt *time.Time
	if v, ok := claims["auth_time"].(float64); ok {
		t := time.Unix(int64(v), 0)
		lastLoginAt = &t
	}

	emailVerified := getBool("email_verified")

	// Keycloak only issues tokens for enabled users
	enabled := true

	return &AuthUser{
		// Identity
		ID:        id,
		Username:  username,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		FullName:  fullName,

		// Authorization
		IsAdmin:     isAdmin,
		RealmRoles:  realmRoles,
		ClientRoles: clientRoles,

		// Account state
		Enabled:       enabled,
		EmailVerified: emailVerified,

		// Session
		LastLoginAt: lastLoginAt,
	}, nil
}

// doRequest performs an authenticated HTTP request to Keycloak
func (c *Client) doRequest(method, path string, body any) (*http.Response, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/%s", c.BaseURL, c.Realm, path)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

// Get performs a GET request to the Keycloak Admin API
func (c *Client) Get(path string) (*http.Response, error) {
	return c.doRequest(http.MethodGet, path, nil)
}

// Post performs a POST request to the Keycloak Admin API
func (c *Client) Post(path string, body any) (*http.Response, error) {
	return c.doRequest(http.MethodPost, path, body)
}

// Put performs a PUT request to the Keycloak Admin API
func (c *Client) Put(path string, body any) (*http.Response, error) {
	return c.doRequest(http.MethodPut, path, body)
}

// Delete performs a DELETE request to the Keycloak Admin API
func (c *Client) Delete(path string) (*http.Response, error) {
	return c.doRequest(http.MethodDelete, path, nil)
}
