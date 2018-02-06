// Package bitgo is a client for BitGo API v1.
package bitgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	// Default URL for API endpoints is a production environment.
	// You can change a base URL using WithBaseURL.
	// More about environments https://bitgo.github.io/bitgo-docs/#bitgo-api-endpoints.
	defaultBaseURL = "https://www.bitgo.com"
)

// Config configures a Client. Config is set by the ConfigOption
// values passed to NewClient.
type Config struct {
	httpClient  *http.Client
	baseURL     string
	accessToken string
}

// ConfigOption configures how we set up the Client.
type ConfigOption func(*Config)

// WithHTTPClient sets Client's underlying HTTP Client.
func WithHTTPClient(httpClient *http.Client) ConfigOption {
	return func(c *Config) {
		c.httpClient = httpClient
	}
}

// WithBaseURL configures Client to use BitGo API domain.
// Usually it's a URL where your BitGo Express REST-ful API service runs.
func WithBaseURL(baseURL string) ConfigOption {
	return func(c *Config) {
		c.baseURL = baseURL
	}
}

// WithAccesToken sets access token to authenticate API requests.
func WithAccesToken(token string) ConfigOption {
	return func(c *Config) {
		c.accessToken = token
	}
}

// Client manages communication with the BitGo REST-ful API.
type Client struct {
	config Config
	Wallet *walletService
}

// NewClient returns a Client which can be configured with config options.
// By default requests are sent to https://www.bitgo.com.
func NewClient(options ...ConfigOption) *Client {
	c := Client{
		config: Config{
			httpClient: http.DefaultClient,
			baseURL:    defaultBaseURL,
		},
	}

	c.Wallet = &walletService{client: &c}

	for _, opt := range options {
		opt(&c.config)
	}
	return &c
}

// NewRequest creates Request to access BitGo API.
// API path must not start or end with slash. Query string params are optional.
// If specified, the value pointed to by body is JSON encoded and included
// as the request body.
func (c *Client) NewRequest(ctx context.Context, method, path string, params url.Values, body interface{}) (*http.Request, error) {
	var urlStr string
	if params != nil {
		urlStr = fmt.Sprintf("%s/api/v1/%s?%s", c.config.baseURL, path, params.Encode())
	} else {
		urlStr = fmt.Sprintf("%s/api/v1/%s", c.config.baseURL, path)
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
	if c.config.accessToken != "" {
		bearer := fmt.Sprintf("Bearer %s", c.config.accessToken)
		req.Header.Set("Authorization", bearer)
	}

	return req, nil
}

// Do uses Client's HTTP client to execute the Request and
// unmarshals the Response into v.
// It also handles unmarshaling errors returned by the API.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.config.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(body, v)
		return resp, err
	}

	e := Error{
		HTTPStatusCode: resp.StatusCode,
		Body:           string(body),
	}
	_ = json.Unmarshal(body, &e)

	switch resp.StatusCode {
	case http.StatusAccepted:
		e.Type = ErrorTypeRequiresApproval
	case http.StatusBadRequest:
		e.Type = ErrorTypeInvalidRequest
	case http.StatusUnauthorized, http.StatusForbidden:
		e.Type = ErrorTypeAuthentication
	case http.StatusNotFound:
		e.Type = ErrorTypeNotFound
	case http.StatusTooManyRequests:
		e.Type = ErrorTypeRateLimit
	default:
		e.Type = ErrorTypeAPI
	}
	return resp, e
}
