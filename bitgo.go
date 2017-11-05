// Package bitgo is a client for BitGo API v1.
package bitgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	// Default URL for API endpoints is a production environment.
	// You can change a base URL using WithBaseURL.
	// More about environments https://bitgo.github.io/bitgo-docs/#bitgo-api-endpoints.
	defaultBaseURL = "https://www.bitgo.com"
)

// Client manages communication with the BitGo REST-ful API.
type Client struct {
	httpClient  *http.Client
	baseURL     string
	accessToken string

	Wallet *walletService
}

// New returns a new Client which can be configured with options.
// By default requests are sent to https://www.bitgo.com.
func New(options ...func(*Client)) *Client {
	c := Client{
		httpClient: http.DefaultClient,
		baseURL:    defaultBaseURL,
	}

	c.Wallet = &walletService{client: &c}

	for _, opt := range options {
		opt(&c)
	}
	return &c
}

// WithHTTPClient sets Client's underlying HTTP Client.
func WithHTTPClient(httpClient *http.Client) func(*Client) {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBaseURL configures Client to use BitGo API domain.
// Usually it's a URL where your BitGo Express REST-ful API service runs.
func WithBaseURL(baseURL string) func(*Client) {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithAccesToken sets access token to authenticate API requests.
func WithAccesToken(token string) func(*Client) {
	return func(c *Client) {
		c.accessToken = token
	}
}

// NewRequest creates Request to access BitGo API.
// API path must not start or end with slash. Query string params are optional.
// If specified, the value pointed to by body is JSON encoded and included
// as the request body.
func (c *Client) NewRequest(ctx context.Context, method, path string, params url.Values, body interface{}) (*http.Request, error) {
	var urlStr string
	if params != nil {
		urlStr = fmt.Sprintf("%s/api/v1/%s?%s", c.baseURL, path, params.Encode())
	} else {
		urlStr = fmt.Sprintf("%s/api/v1/%s", c.baseURL, path)
	}

	jsonBody := bytes.Buffer{}
	if body != nil {
		err := json.NewEncoder(&jsonBody).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, urlStr, &jsonBody)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.accessToken != "" {
		bearer := fmt.Sprintf("Bearer %s", c.accessToken)
		req.Header.Set("Authorization", bearer)
	}

	return req, nil
}

// Do uses Client's HTTP client to execute the Request and
// unmarshals the Response into v.
// It also handles unmarshaling errors returned by the API.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		e := Error{HTTPStatusCode: resp.StatusCode}
		if err = json.NewDecoder(resp.Body).Decode(&e); err != nil {
			e.Message = "server error"
		}
		return resp, e
	}

	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}
