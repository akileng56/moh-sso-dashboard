package keycloak

import (
	"bytes"
	"context"
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
	BaseURL           string
	Realm             string
	AdminClientID     string
	AdminClientSecret string
	WebClientID       string
	WebClientSecret   string
	AdminToken        string
	httpClient        *http.Client
	cache             Cache
}

type KCUserInfo struct {
	UserID            string `json:"sub"`
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
	IsUser        bool                `json:"isUser"`
	RealmRoles    []string            `json:"realmRoles"`
	ClientRoles   map[string][]string `json:"clientRoles"`
	Enabled       bool                `json:"enabled"`
	EmailVerified bool                `json:"emailVerified"`
	LastLoginAt   *time.Time          `json:"lastLoginAt"`
}

// Cache
// ----------------------------------------------------
type Cache interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Del(ctx context.Context, key string) error
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
	cache Cache,
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
		cache:             cache,
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
	ctx := context.Background()

	ui, err := c.getUserInfo(accessToken)
	if err != nil {
		return nil, err
	}

	userID := ui.UserID
	if userID == "" {
		return nil, errors.New("userinfo missing user_id")
	}

	cacheKey := "auth:user:" + userID

	var cached AuthUser
	if ok, err := c.cache.Get(ctx, cacheKey, &cached); err == nil && ok {
		return &cached, nil
	}

	realmRoles := []string{}
	isAdmin := false
	isUser := false

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

			if r.Name == "user" {
				isUser = true
			}
		}
	}

	clientRoles := map[string][]string{}
	clients, err := c.ListClients()
	if err != nil {
		return nil, err
	}

	for _, client := range clients {
		res, err := c.Get(
			fmt.Sprintf("users/%s/role-mappings/clients/%s", userID, client.ID),
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

	res, err = c.Get("users/" + userID)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to fetch user profile [%d]: %s", res.StatusCode, string(body))
	}

	var kcUser map[string]any
	if err := json.NewDecoder(res.Body).Decode(&kcUser); err != nil {
		return nil, err
	}

	get := func(k string) string {
		if v, ok := kcUser[k].(string); ok {
			return v
		}
		return ""
	}

	firstName := get("firstName")
	lastName := get("lastName")
	fullName := strings.TrimSpace(firstName + " " + lastName)
	if fullName == "" {
		fullName = ui.Name
	}

	enabled, _ := kcUser["enabled"].(bool)

	var createdAt *time.Time
	if v, ok := kcUser["createdTimestamp"].(float64); ok {
		t := time.UnixMilli(int64(v))
		createdAt = &t
	}

	user := &AuthUser{
		ID:            userID,
		Username:      ui.PreferredUsername,
		Email:         ui.Email,
		FirstName:     firstName,
		LastName:      lastName,
		FullName:      fullName,
		IsAdmin:       isAdmin,
		IsUser:        isUser,
		RealmRoles:    realmRoles,
		ClientRoles:   clientRoles,
		Enabled:       enabled,
		EmailVerified: ui.EmailVerified,
		LastLoginAt:   createdAt,
	}

	_ = c.cache.Set(ctx, cacheKey, user, 10*time.Minute)

	return user, nil
}

// ----------------------------------------------------
// UserInfo
// ----------------------------------------------------
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
		return nil, fmt.Errorf("userinfo failed [%d]: %s", res.StatusCode, string(body))
	}

	var ui KCUserInfo
	if err := json.NewDecoder(res.Body).Decode(&ui); err != nil {
		return nil, err
	}

	if ui.UserID == "" {
		return nil, errors.New("userinfo response missing user_id")
	}

	return &ui, nil
}

// ----------------------------------------------------
// ADMIN API HELPERS
// ----------------------------------------------------

func (c *Client) doRequest(
	method string,
	path string,
	body any,
	rawBody io.Reader,
) (*http.Response, error) {

	adminURL := fmt.Sprintf(
		"%s/admin/realms/%s/%s",
		c.BaseURL,
		c.Realm,
		path,
	)

	var reqBody io.Reader

	if rawBody != nil {
		reqBody = rawBody
	} else if body != nil {
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

func (c *Client) DeleteWithBody(
	path string,
	body io.Reader,
) (*http.Response, error) {

	return c.doRequest(
		http.MethodDelete,
		path,
		nil,
		body,
	)
}

func (c *Client) Get(path string) (*http.Response, error) {
	return c.doRequest(http.MethodGet, path, nil, nil)
}

func (c *Client) Post(path string, body any) (*http.Response, error) {
	return c.doRequest(http.MethodPost, path, body, nil)
}

func (c *Client) Put(path string, body any) (*http.Response, error) {
	return c.doRequest(http.MethodPut, path, body, nil)
}

func (c *Client) Delete(path string) (*http.Response, error) {
	return c.doRequest(http.MethodDelete, path, nil, nil)
}
