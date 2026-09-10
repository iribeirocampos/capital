package capital

import (
	"context"
	"net/http"
)

// Position represents an open trading position.
type Position struct {
	Position struct {
		DealID         string    `json:"dealId"`
		DealReference  string    `json:"dealReference"`
		CreatedDate    string    `json:"createdDate"`
		Direction      Direction `json:"direction"`
		Size           float64   `json:"size"`
		Level          float64   `json:"level"`
		Currency       string    `json:"currency"`
		GuaranteedStop bool      `json:"guaranteedStop"`
		TrailingStop   bool      `json:"trailingStop"`
		StopLevel      float64   `json:"stopLevel"`
		ProfitLevel    float64   `json:"profitLevel"`
	} `json:"position"`
	Market struct {
		Epic           string  `json:"epic"`
		InstrumentName string  `json:"instrumentName"`
		MarketStatus   string  `json:"marketStatus"`
		Bid            float64 `json:"bid"`
		Offer          float64 `json:"offer"`
	} `json:"market"`
}

type positionsResponse struct {
	Positions []Position `json:"positions"`
}

// ListPositions returns all open positions on the active account.
func (c *Client) ListPositions(ctx context.Context) ([]Position, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var resp positionsResponse
	if err := c.doRequest(ctx, http.MethodGet, "/positions", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Positions, nil
}

// GetPosition returns a single open position by deal ID.
func (c *Client) GetPosition(ctx context.Context, dealID string) (*Position, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var pos Position
	if err := c.doRequest(ctx, http.MethodGet, "/positions/"+dealID, nil, &pos); err != nil {
		return nil, err
	}
	return &pos, nil
}

type createPositionRequest struct {
	Epic      string    `json:"epic"`
	Direction Direction `json:"direction"`
	Size      float64   `json:"size"`
	stopProfitParams
}

// CreatePosition opens a new trading position and returns its confirmation.
//
//	conf, err := client.CreatePosition(ctx, "GOLD", capital.Buy, 1,
//		capital.WithStopLevel(1900), capital.WithProfitLevel(2000))
func (c *Client) CreatePosition(ctx context.Context, epic string, direction Direction, size float64, opts ...TradeOption) (*DealConfirmation, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	req := createPositionRequest{Epic: epic, Direction: direction, Size: size}
	for _, opt := range opts {
		opt(&req.stopProfitParams)
	}
	var ref dealReferenceResponse
	if err := c.doRequest(ctx, http.MethodPost, "/positions", req, &ref); err != nil {
		return nil, err
	}
	return c.confirmDeal(ctx, ref.DealReference)
}

type updatePositionRequest struct {
	stopProfitParams
}

// UpdatePosition updates the stop-loss/take-profit settings of an open
// position and returns its confirmation.
func (c *Client) UpdatePosition(ctx context.Context, dealID string, opts ...TradeOption) (*DealConfirmation, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var req updatePositionRequest
	for _, opt := range opts {
		opt(&req.stopProfitParams)
	}
	var ref dealReferenceResponse
	if err := c.doRequest(ctx, http.MethodPut, "/positions/"+dealID, req, &ref); err != nil {
		return nil, err
	}
	return c.confirmDeal(ctx, ref.DealReference)
}

// ClosePosition closes an open position and returns its confirmation.
func (c *Client) ClosePosition(ctx context.Context, dealID string) (*DealConfirmation, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var ref dealReferenceResponse
	if err := c.doRequest(ctx, http.MethodDelete, "/positions/"+dealID, nil, &ref); err != nil {
		return nil, err
	}
	return c.confirmDeal(ctx, ref.DealReference)
}
