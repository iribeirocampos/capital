package capital

import (
	"context"
	"net/http"
)

// Watchlist is a named list of market epics.
type Watchlist struct {
	WatchlistID string   `json:"watchlistId,omitempty"`
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Markets     []Market `json:"markets,omitempty"`
}

type watchlistsResponse struct {
	Watchlists []Watchlist `json:"watchlists"`
}

// ListWatchlists returns all watchlists for the authenticated user.
func (c *Client) ListWatchlists(ctx context.Context) ([]Watchlist, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var resp watchlistsResponse
	if err := c.doRequest(ctx, http.MethodGet, "/watchlists", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Watchlists, nil
}

type createWatchlistRequest struct {
	Name  string   `json:"name"`
	Epics []string `json:"epics,omitempty"`
}

type createWatchlistResponse struct {
	WatchlistID string `json:"watchlistId"`
	Status      string `json:"status"`
}

// CreateWatchlist creates a new watchlist, optionally pre-populated with
// the given epics, and returns its ID.
func (c *Client) CreateWatchlist(ctx context.Context, name string, epics ...string) (string, error) {
	if err := c.ensureSession(ctx); err != nil {
		return "", err
	}
	req := createWatchlistRequest{Name: name, Epics: epics}
	var resp createWatchlistResponse
	if err := c.doRequest(ctx, http.MethodPost, "/watchlists", req, &resp); err != nil {
		return "", err
	}
	return resp.WatchlistID, nil
}

// GetWatchlist returns a watchlist's details, including its markets.
func (c *Client) GetWatchlist(ctx context.Context, watchlistID string) (*Watchlist, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var wl Watchlist
	if err := c.doRequest(ctx, http.MethodGet, "/watchlists/"+watchlistID, nil, &wl); err != nil {
		return nil, err
	}
	return &wl, nil
}

type addMarketRequest struct {
	Epic string `json:"epic"`
}

// AddMarketToWatchlist adds a market (by epic) to the given watchlist.
func (c *Client) AddMarketToWatchlist(ctx context.Context, watchlistID, epic string) error {
	if err := c.ensureSession(ctx); err != nil {
		return err
	}
	return c.doRequest(ctx, http.MethodPut, "/watchlists/"+watchlistID, addMarketRequest{Epic: epic}, nil)
}

// RemoveMarketFromWatchlist removes a market (by epic) from the given
// watchlist.
func (c *Client) RemoveMarketFromWatchlist(ctx context.Context, watchlistID, epic string) error {
	if err := c.ensureSession(ctx); err != nil {
		return err
	}
	return c.doRequest(ctx, http.MethodDelete, "/watchlists/"+watchlistID+"/"+epic, nil, nil)
}

// DeleteWatchlist deletes an entire watchlist.
func (c *Client) DeleteWatchlist(ctx context.Context, watchlistID string) error {
	if err := c.ensureSession(ctx); err != nil {
		return err
	}
	return c.doRequest(ctx, http.MethodDelete, "/watchlists/"+watchlistID, nil, nil)
}
