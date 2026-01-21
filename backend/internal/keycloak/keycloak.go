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
)

// Client represents the Keycloak API client
// - Admin operations use dashboard-admin (client_credentials)
// - Web auth/refresh/logout use dashboard-web (authorization_code + refresh_token)
type Client struct {
	BaseURL string
	Realm   string

	AdminClientID     string
	AdminClientSecret string

	WebClientID     string
	WebClientSecret string

	AdminToken string // access token for admin API calls
	httpClient *http.Client
}

type KCUserInfo struct {
	Sub               string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Name              string `json:"name"`
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
	ID            string              `json:"id"`
	Username      string              `json:"username"`
	Email         string              `json:"email"`
	FirstName     string              `json:"firstName"`
	LastName      string              `json:"lastName"`
	FullName      string              `json:"fullName"`
	IsAdmin       bool                `json:"isAdmin"`
	RealmRoles    []string            `json:"realmRoles"`
	ClientRoles   map[string][]string `json:"clientRoles"`
	Enabled       bool                `json:"enabled"`
	EmailVerified bool                `json:"emailVerified"`
	LastLoginAt   *time.Time          `json:"lastLoginAt"`
}

// ----------------------------------------------------
// CLIENT
// ----------------------------------------------------
func NewClient(
	baseURL string,
	realm string,
	adminClientID string,
	adminClientSecret string,
	webClientID string,
	webClientSecret string,
) *Client {

	baseURL = strings.TrimSuffix(baseURL, "/")

	return &Client{
		BaseURL:           baseURL,
		Realm:             realm,
		AdminClientID:     adminClientID,
		AdminClientSecret: adminClientSecret,
		WebClientID:       webClientID,
		WebClientSecret:   webClientSecret,
		httpClient:        &http.Client{Timeout: 10 * time.Second},
	}
}

// ----------------------------------------------------
// ADMIN AUTH (client_credentials)
// ----------------------------------------------------
func (c *Client) Authenticate() error {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.AdminClientID)
	form.Set("client_secret", c.AdminClientSecret)

	tokenURL := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/token",
		c.BaseURL,
		c.Realm,
	)

	res, err := c.httpClient.PostForm(tokenURL, form)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("admin authentication failed [%d]: %s", res.StatusCode, string(bodyBytes))
	}

	var body struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return err
	}

	c.AdminToken = body.AccessToken
	return nil
}

// ----------------------------------------------------
// WEB AUTH (authorization_code)
// ----------------------------------------------------
func (c *Client) ExchangeCodeForToken(
	code string,
	redirectURI string,
	codeVerifier string,
) (*TokenResponse, error) {

	tokenURL := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/token",
		c.BaseURL,
		c.Realm,
	)

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", c.WebClientID)
	form.Set("client_secret", c.WebClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("code_verifier", codeVerifier)

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

// ----------------------------------------------------
// REFRESH TOKEN (WEB CLIENT)
// ----------------------------------------------------
func (c *Client) AccessToken(refreshToken string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/token",
		c.BaseURL,
		c.Realm,
	)

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", c.WebClientID)
	form.Set("client_secret", c.WebClientSecret)

	res, err := c.httpClient.PostForm(tokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("refresh token failed: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh token failed [%d]: %s", res.StatusCode, string(bodyBytes))
	}

	var tokenRes TokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenRes); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	return &tokenRes, nil
}

// ----------------------------------------------------
// LOGOUT (WEB CLIENT)
// ----------------------------------------------------
func (c *Client) LogOut(refreshToken string) error {
	if refreshToken == "" {
		return fmt.Errorf("missing refresh token for logout")
	}

	logoutURL := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/logout",
		c.BaseURL,
		c.Realm,
	)

	form := url.Values{}
	form.Set("client_id", c.WebClientID)
	form.Set("client_secret", c.WebClientSecret)
	form.Set("refresh_token", refreshToken)

	res, err := c.httpClient.PostForm(logoutURL, form)
	if err != nil {
		return fmt.Errorf("logout request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("logout failed [%d]: %s", res.StatusCode, string(body))
	}

	return nil
}

// ----------------------------------------------------
// TOKEN → FULL USER PROFILE (ADMIN API)
// ----------------------------------------------------
func (c *Client) Me(accessToken string) (*AuthUser, error) {
	// ------------------------------------------------
	// 1) Resolve identity via OIDC UserInfo (CORRECT)
	// ------------------------------------------------
	ui, err := c.getUserInfo(accessToken)
	if err != nil {
		return nil, err
	}

	userID := ui.Sub
	if userID == "" {
		return nil, errors.New("userinfo missing sub")
	}

	// ------------------------------------------------
	// 2) Fetch REALM roles from Admin API
	// ------------------------------------------------
	realmRoles := []string{}
	isAdmin := false

	res, err := c.Get("users/" + userID + "/role-mappings/realm")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		var roles []struct {
			Name string `json:"name"`
		}

		if err := json.NewDecoder(res.Body).Decode(&roles); err != nil {
			return nil, err
		}

		for _, r := range roles {
			realmRoles = append(realmRoles, r.Name)
			if r.Name == "admin" {
				isAdmin = true
			}
		}
	}

	// ------------------------------------------------
	// 3) Fetch CLIENT roles from Admin API
	// ------------------------------------------------
	clientRoles := map[string][]string{}

	clients, err := c.ListClients()
	if err != nil {
		return nil, err
	}

	for _, client := range clients {
		res, err := c.Get(
			fmt.Sprintf(
				"users/%s/role-mappings/clients/%s",
				userID,
				client.ID,
			),
		)
		if err != nil || res.StatusCode != http.StatusOK {
			continue
		}

		var roles []struct {
			Name string `json:"name"`
		}

		if err := json.NewDecoder(res.Body).Decode(&roles); err != nil {
			res.Body.Close()
			continue
		}
		res.Body.Close()

		for _, r := range roles {
			clientRoles[client.ClientID] = append(clientRoles[client.ClientID], r.Name)
		}
	}

	// ------------------------------------------------
	// 4) Fetch FULL user profile from Admin API
	// ------------------------------------------------
	res, err = c.Get("users/" + userID)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf(
			"failed to fetch user profile [%d]: %s",
			res.StatusCode,
			string(body),
		)
	}

	var kcUser map[string]any
	if err := json.NewDecoder(res.Body).Decode(&kcUser); err != nil {
		return nil, err
	}

	getKCString := func(key string) string {
		if v, ok := kcUser[key].(string); ok {
			return v
		}
		return ""
	}

	firstName := getKCString("firstName")
	lastName := getKCString("lastName")

	fullName := strings.TrimSpace(firstName + " " + lastName)
	if fullName == "" {
		fullName = ui.Name
	}

	enabled := false
	if v, ok := kcUser["enabled"].(bool); ok {
		enabled = v
	}

	emailVerified := ui.EmailVerified

	var createdAt *time.Time
	if v, ok := kcUser["createdTimestamp"].(float64); ok {
		t := time.UnixMilli(int64(v))
		createdAt = &t
	}

	// ------------------------------------------------
	// 5) Return merged result
	// ------------------------------------------------
	return &AuthUser{
		ID:            userID,
		Username:      ui.PreferredUsername,
		Email:         ui.Email,
		FirstName:     firstName,
		LastName:      lastName,
		FullName:      fullName,
		IsAdmin:       isAdmin,
		RealmRoles:    realmRoles,
		ClientRoles:   clientRoles,
		Enabled:       enabled,
		EmailVerified: emailVerified,
		LastLoginAt:   createdAt,
	}, nil
}

func (c *Client) getUserInfo(accessToken string) (*KCUserInfo, error) {
	url := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/userinfo",
		c.BaseURL,
		c.Realm,
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf(
			"userinfo failed [%d]: %s",
			res.StatusCode,
			string(body),
		)
	}

	var ui KCUserInfo
	if err := json.NewDecoder(res.Body).Decode(&ui); err != nil {
		return nil, err
	}

	if ui.Sub == "" {
		return nil, errors.New("userinfo response missing sub")
	}

	return &ui, nil
}

// ----------------------------------------------------
// ADMIN API HELPERS
// ----------------------------------------------------
func (c *Client) doRequest(method, path string, body any) (*http.Response, error) {
	adminURL := fmt.Sprintf("%s/admin/realms/%s/%s", c.BaseURL, c.Realm, path)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, adminURL, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AdminToken)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c *Client) Get(path string) (*http.Response, error) {
	return c.doRequest(http.MethodGet, path, nil)
}
func (c *Client) Post(path string, body any) (*http.Response, error) {
	return c.doRequest(http.MethodPost, path, body)
}
func (c *Client) Put(path string, body any) (*http.Response, error) {
	return c.doRequest(http.MethodPut, path, body)
}
func (c *Client) Delete(path string) (*http.Response, error) {
	return c.doRequest(http.MethodDelete, path, nil)
}
