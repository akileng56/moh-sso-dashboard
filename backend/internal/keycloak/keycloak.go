package keycloak

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// NewClient creates a new Keycloak client
func NewClient(baseURL, realm, clientID, secret string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Realm:      realm,
		ClientID:   clientID,
		Secret:     secret,
		httpClient: &http.Client{},
	}
}

// Authenticate retrieves an access token using client credentials
func (c *Client) Authenticate() error {
	form := url.Values{}
	form.Add("grant_type", "client_credentials")
	form.Add("client_id", c.ClientID)
	form.Add("client_secret", c.Secret)

	res, err := http.PostForm(fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.BaseURL, c.Realm), form)
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

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(res.Body)
		res.Body.Close()
		return nil, fmt.Errorf("keycloak API error [%d]: %s", res.StatusCode, string(bodyBytes))
	}

	return res, nil
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
