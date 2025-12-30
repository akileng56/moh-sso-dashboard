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
	ID          string              `json:"id"`
	Username    string              `json:"username"`
	Email       string              `json:"email,omitempty"`
	IsAdmin     bool                `json:"is_admin"`
	ClientRoles map[string][]string `json:"client_roles"`
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
	// Basic identity
	// ----------------------------
	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)

	username := ""
	if v, ok := claims["preferred_username"].(string); ok {
		username = v
	} else if v, ok := claims["name"].(string); ok {
		username = v
	}

	// ----------------------------
	// Realm roles → is_admin
	// ----------------------------
	isAdmin := false
	if ra, ok := claims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := ra["roles"].([]interface{}); ok {
			for _, r := range roles {
				if role, ok := r.(string); ok && role == "admin" {
					isAdmin = true
					break
				}
			}
		}
	}

	// ----------------------------
	// Client roles → app access
	// ----------------------------
	clientRoles := map[string][]string{}

	if ra, ok := claims["resource_access"].(map[string]interface{}); ok {
		for clientID, raw := range ra {
			block, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}

			roles, ok := block["roles"].([]interface{})
			if !ok {
				continue
			}

			for _, r := range roles {
				if role, ok := r.(string); ok {
					clientRoles[clientID] = append(clientRoles[clientID], role)
				}
			}
		}
	}

	return &AuthUser{
		ID:          sub,
		Username:    username,
		Email:       email,
		IsAdmin:     isAdmin,
		ClientRoles: clientRoles,
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
