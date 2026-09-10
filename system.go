package capital

import (
	"context"
	"net/http"
)

// ServerTime returns the Capital.com server time as a Unix millisecond
// timestamp. It does not require an active session.
func (c *Client) ServerTime(ctx context.Context) (int64, error) {
	var resp struct {
		ServerTime int64 `json:"serverTime"`
	}
	if err := c.doRequest(ctx, http.MethodGet, "/time", nil, &resp); err != nil {
		return 0, err
	}
	return resp.ServerTime, nil
}

// Ping checks that the API is reachable, returning nil if the service
// reports itself healthy. It does not require an active session.
func (c *Client) Ping(ctx context.Context) error {
	var resp struct {
		Status string `json:"status"`
	}
	return c.doRequest(ctx, http.MethodGet, "/ping", nil, &resp)
}
