package keycloak

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	BaseURL   string
	Realm     string
	ClientID  string
	Secret    string
	Token     string
}

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

	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return err
	}
	c.Token = body.AccessToken
	return nil
}
