package capital

import (
	"context"
	"fmt"
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

// ClosePositionPartial closes part of an open position by opening an
// opposite-direction trade on the same market for the given size.
//
// Capital.com's public API has no native partial-close endpoint — DELETE
// /positions/{dealId} always closes a position in full, with no size
// parameter. On accounts in netting mode (the default), opening a trade in
// the opposite direction on the same epic nets against the existing
// position instead of opening a second one — this is how partial closes
// work on Capital.com's own web/app UI. ClosePositionPartial checks the
// account's hedging-mode setting first and refuses to proceed if hedging
// mode is enabled, since in that mode the opposing trade would open a
// separate position rather than reduce this one.
func (c *Client) ClosePositionPartial(ctx context.Context, dealID string, size float64) (*DealConfirmation, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}

	prefs, err := c.GetAccountPreferences(ctx)
	if err != nil {
		return nil, err
	}
	if prefs.HedgingMode {
		return nil, fmt.Errorf("capital: partial close is not supported with hedging mode enabled")
	}

	pos, err := c.GetPosition(ctx, dealID)
	if err != nil {
		return nil, err
	}
	if size <= 0 || size >= pos.Position.Size {
		return nil, fmt.Errorf("capital: partial close size must be greater than 0 and less than the position size (%v)", pos.Position.Size)
	}

	opposite := Sell
	if pos.Position.Direction == Sell {
		opposite = Buy
	}
	return c.CreatePosition(ctx, pos.Market.Epic, opposite, size)
}
