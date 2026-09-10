package capital

import (
	"context"
	"net/http"
)

// OrderType is the type of a working (pending) order.
type OrderType string

const (
	OrderLimit OrderType = "LIMIT"
	OrderStop  OrderType = "STOP"
)

// WorkingOrder represents a pending limit or stop order.
type WorkingOrder struct {
	WorkingOrderData struct {
		DealID        string    `json:"dealId"`
		DealReference string    `json:"dealReference"`
		Direction     Direction `json:"direction"`
		Epic          string    `json:"epic"`
		OrderSize     float64   `json:"orderSize"`
		OrderLevel    float64   `json:"orderLevel"`
		OrderType     OrderType `json:"orderType"`
		CreatedDate   string    `json:"createdDate"`
		GoodTillDate  string    `json:"goodTillDate"`
	} `json:"workingOrderData"`
	MarketData struct {
		Epic           string  `json:"epic"`
		InstrumentName string  `json:"instrumentName"`
		Bid            float64 `json:"bid"`
		Offer          float64 `json:"offer"`
	} `json:"marketData"`
}

type workingOrdersResponse struct {
	WorkingOrders []WorkingOrder `json:"workingOrders"`
}

// ListWorkingOrders returns all pending working orders on the active
// account.
func (c *Client) ListWorkingOrders(ctx context.Context) ([]WorkingOrder, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var resp workingOrdersResponse
	if err := c.doRequest(ctx, http.MethodGet, "/workingorders", nil, &resp); err != nil {
		return nil, err
	}
	return resp.WorkingOrders, nil
}

type createWorkingOrderRequest struct {
	Epic      string    `json:"epic"`
	Direction Direction `json:"direction"`
	Size      float64   `json:"size"`
	Level     float64   `json:"level"`
	Type      OrderType `json:"type"`
	stopProfitParams
}

// CreateWorkingOrder creates a new limit or stop order and returns its
// confirmation.
func (c *Client) CreateWorkingOrder(ctx context.Context, epic string, direction Direction, size, level float64, orderType OrderType, opts ...TradeOption) (*DealConfirmation, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	req := createWorkingOrderRequest{Epic: epic, Direction: direction, Size: size, Level: level, Type: orderType}
	for _, opt := range opts {
		opt(&req.stopProfitParams)
	}
	var ref dealReferenceResponse
	if err := c.doRequest(ctx, http.MethodPost, "/workingorders", req, &ref); err != nil {
		return nil, err
	}
	return c.confirmDeal(ctx, ref.DealReference)
}

type updateWorkingOrderRequest struct {
	Level float64 `json:"level"`
	stopProfitParams
}

// UpdateWorkingOrder updates the level/stop/profit settings of a pending
// order and returns its confirmation.
func (c *Client) UpdateWorkingOrder(ctx context.Context, dealID string, level float64, opts ...TradeOption) (*DealConfirmation, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	req := updateWorkingOrderRequest{Level: level}
	for _, opt := range opts {
		opt(&req.stopProfitParams)
	}
	var ref dealReferenceResponse
	if err := c.doRequest(ctx, http.MethodPut, "/workingorders/"+dealID, req, &ref); err != nil {
		return nil, err
	}
	return c.confirmDeal(ctx, ref.DealReference)
}

// DeleteWorkingOrder cancels a pending order and returns its confirmation.
func (c *Client) DeleteWorkingOrder(ctx context.Context, dealID string) (*DealConfirmation, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var ref dealReferenceResponse
	if err := c.doRequest(ctx, http.MethodDelete, "/workingorders/"+dealID, nil, &ref); err != nil {
		return nil, err
	}
	return c.confirmDeal(ctx, ref.DealReference)
}
