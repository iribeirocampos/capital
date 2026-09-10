// Package capital is an unofficial Go client for the Capital.com Public API
// (https://open-api.capital.com/). It is not affiliated with Capital.com;
// use at your own risk.
package capital

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

const (
	liveBaseURL = "https://api-capital.backend-capital.com/api/v1"
	demoBaseURL = "https://demo-api-capital.backend-capital.com/api/v1"
)

// Client is a Capital.com REST API client. Create one with NewClient.
//
// Client keeps a persistent, authenticated session: the first call that
// needs one triggers an automatic login, and the session is transparently
// renewed if the API reports it has expired. A Client is safe for
// concurrent use by multiple goroutines.
type Client struct {
	identifier string
	apiKey     string
	password   string
	baseURL    string
	httpClient *http.Client

	mu            sync.RWMutex
	cst           string
	securityToken string
	accountID     string
}

// Option configures a Client created by NewClient.
type Option func(*Client)

// WithDemo points the client at the Capital.com demo trading environment
// instead of the live one. Defaults to false (live).
func WithDemo(demo bool) Option {
	return func(c *Client) {
		if demo {
			c.baseURL = demoBaseURL
		} else {
			c.baseURL = liveBaseURL
		}
	}
}

// WithHTTPClient overrides the *http.Client used to send requests. Defaults
// to http.DefaultClient.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithBaseURL overrides the API base URL. Mainly useful for testing against
// a mock server.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// NewClient creates a new Capital.com API client.
//
// identifier is the account email/username, apiKey is an API key generated
// in the platform's API settings, and password is the account password. The
// password is never transmitted in plaintext: it is RSA-encrypted locally
// using a key fetched from the API as part of Login.
//
// By default the client targets the live trading environment; pass
// WithDemo(true) to use the demo environment instead.
func NewClient(identifier, apiKey, password string, opts ...Option) *Client {
	c := &Client{
		identifier: identifier,
		apiKey:     apiKey,
		password:   password,
		baseURL:    liveBaseURL,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// AccountID returns the currently active account ID, once a session has
// been established.
func (c *Client) AccountID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accountID
}

func (c *Client) setSession(cst, securityToken string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cst = cst
	c.securityToken = securityToken
}

func (c *Client) setAccountID(accountID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accountID = accountID
}

func (c *Client) clearSession() {
	c.setSession("", "")
}

func (c *Client) hasSession() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cst != "" && c.securityToken != ""
}

// ensureSession makes sure a session has been established, logging in
// automatically on first use.
func (c *Client) ensureSession(ctx context.Context) error {
	if c.hasSession() {
		return nil
	}
	return c.Login(ctx)
}

// rawRequest performs exactly one HTTP round trip and returns the raw
// response and body.
func (c *Client) rawRequest(ctx context.Context, method, path string, body any) (*http.Response, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("capital: encode request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("capital: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	c.mu.RLock()
	if c.apiKey != "" {
		req.Header.Set("X-CAP-API-KEY", c.apiKey)
	}
	if c.cst != "" {
		req.Header.Set("CST", c.cst)
	}
	if c.securityToken != "" {
		req.Header.Set("X-SECURITY-TOKEN", c.securityToken)
	}
	c.mu.RUnlock()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("capital: request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("capital: read response: %w", err)
	}
	return resp, respBody, nil
}

// doRequest performs an authenticated request and decodes a JSON response
// into out, if non-nil. It transparently re-authenticates and retries once
// if the API reports the session as invalid or expired.
func (c *Client) doRequest(ctx context.Context, method, path string, body, out any) error {
	_, err := c.doRequestHeaders(ctx, method, path, body, out, true)
	return err
}

// doRequestHeaders is doRequest's superset: it also returns the response
// headers (needed by Login to read CST/X-SECURITY-TOKEN) and lets the
// caller opt out of the auto re-login retry, which Login itself must do to
// avoid recursing into itself on a failed login.
func (c *Client) doRequestHeaders(ctx context.Context, method, path string, body, out any, retryOnAuthFailure bool) (http.Header, error) {
	resp, respBody, err := c.rawRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized && retryOnAuthFailure {
		c.clearSession()
		if loginErr := c.Login(ctx); loginErr == nil {
			resp, respBody, err = c.rawRequest(ctx, method, path, body)
			if err != nil {
				return nil, err
			}
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp.StatusCode, respBody)
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return nil, fmt.Errorf("capital: decode response: %w", err)
		}
	}
	return resp.Header, nil
}
