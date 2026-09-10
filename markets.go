package capital

import (
	"context"
	"net/http"
	"net/url"
)

// MarketNode is a category node in the market navigation hierarchy.
type MarketNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Market describes a tradable instrument.
type Market struct {
	Epic           string  `json:"epic"`
	InstrumentName string  `json:"instrumentName"`
	InstrumentType string  `json:"instrumentType"`
	MarketStatus   string  `json:"marketStatus"`
	Bid            float64 `json:"bid"`
	Offer          float64 `json:"offer"`
	High           float64 `json:"high"`
	Low            float64 `json:"low"`
}

type marketNavigationResponse struct {
	Nodes   []MarketNode `json:"nodes"`
	Markets []Market     `json:"markets"`
}

// MarketNavigation returns the top-level market categories.
func (c *Client) MarketNavigation(ctx context.Context) ([]MarketNode, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var resp marketNavigationResponse
	if err := c.doRequest(ctx, http.MethodGet, "/marketnavigation", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Nodes, nil
}

// MarketNavigationNode returns the sub-nodes and markets under the given
// category node.
func (c *Client) MarketNavigationNode(ctx context.Context, nodeID string) ([]MarketNode, []Market, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, nil, err
	}
	var resp marketNavigationResponse
	path := "/marketnavigation/" + url.PathEscape(nodeID) + "?limit=500"
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, nil, err
	}
	return resp.Nodes, resp.Markets, nil
}

type marketsSearchResponse struct {
	Markets []Market `json:"markets"`
}

// SearchMarkets returns markets matching the given search term.
func (c *Client) SearchMarkets(ctx context.Context, searchTerm string) ([]Market, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var resp marketsSearchResponse
	path := "/markets?searchTerm=" + url.QueryEscape(searchTerm)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Markets, nil
}

// GetMarket returns the details of a single market by epic.
func (c *Client) GetMarket(ctx context.Context, epic string) (*Market, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var market Market
	if err := c.doRequest(ctx, http.MethodGet, "/markets/"+url.PathEscape(epic), nil, &market); err != nil {
		return nil, err
	}
	return &market, nil
}
