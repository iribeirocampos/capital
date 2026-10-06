package capital

import (
	"context"
	"net/http"
	"time"
)

// Direction indicates whether a position or order buys or sells.
type Direction string

const (
	Buy  Direction = "BUY"
	Sell Direction = "SELL"
)

// stopProfitParams holds the optional risk-management fields shared by
// position and working-order create/update requests. It is embedded
// (anonymously, so its fields marshal at the top level of the JSON body)
// into each request type.
type stopProfitParams struct {
	GuaranteedStop bool     `json:"guaranteedStop"`
	TrailingStop   bool     `json:"trailingStop"`
	StopLevel      *float64 `json:"stopLevel,omitempty"`
	StopDistance   *float64 `json:"stopDistance,omitempty"`
	StopAmount     *float64 `json:"stopAmount,omitempty"`
	ProfitLevel    *float64 `json:"profitLevel,omitempty"`
	ProfitDistance *float64 `json:"profitDistance,omitempty"`
	ProfitAmount   *float64 `json:"profitAmount,omitempty"`

	// GoodTillDate only applies to working orders.
	GoodTillDate string `json:"goodTillDate,omitempty"`
}

// TradeOption configures optional risk-management parameters on
// CreatePosition, UpdatePosition, CreateWorkingOrder and UpdateWorkingOrder
// — the Go equivalent of the optional keyword arguments python-capital
// exposes for the same calls.
type TradeOption func(*stopProfitParams)

// WithGuaranteedStop marks the stop as a guaranteed stop.
func WithGuaranteedStop(v bool) TradeOption {
	return func(p *stopProfitParams) { p.GuaranteedStop = v }
}

// WithTrailingStop marks the stop as a trailing stop.
func WithTrailingStop(v bool) TradeOption {
	return func(p *stopProfitParams) { p.TrailingStop = v }
}

// WithStopLevel sets an absolute stop-loss price level.
func WithStopLevel(v float64) TradeOption {
	return func(p *stopProfitParams) { p.StopLevel = &v }
}

// WithStopDistance sets a stop-loss distance from the current price.
func WithStopDistance(v float64) TradeOption {
	return func(p *stopProfitParams) { p.StopDistance = &v }
}

// WithStopAmount sets a stop-loss as a monetary amount.
func WithStopAmount(v float64) TradeOption {
	return func(p *stopProfitParams) { p.StopAmount = &v }
}

// WithProfitLevel sets an absolute take-profit price level.
func WithProfitLevel(v float64) TradeOption {
	return func(p *stopProfitParams) { p.ProfitLevel = &v }
}

// WithProfitDistance sets a take-profit distance from the current price.
func WithProfitDistance(v float64) TradeOption {
	return func(p *stopProfitParams) { p.ProfitDistance = &v }
}

// WithGoodTillDate sets when a working order expires (sent as UTC).
func WithGoodTillDate(t time.Time) TradeOption {
	return func(p *stopProfitParams) { p.GoodTillDate = t.UTC().Format("2006-01-02T15:04:05") }
}

// WithProfitAmount sets a take-profit as a monetary amount.
func WithProfitAmount(v float64) TradeOption {
	return func(p *stopProfitParams) { p.ProfitAmount = &v }
}

// DealConfirmation describes the outcome of a create/update/close position
// or working-order action, obtained via the /confirms endpoint.
type DealConfirmation struct {
	DealReference string    `json:"dealReference"`
	DealID        string    `json:"dealId"`
	DealStatus    string    `json:"dealStatus"`
	Status        string    `json:"status"`
	Reason        string    `json:"reason"`
	Direction     Direction `json:"direction"`
	Level         float64   `json:"level"`
	Size          float64   `json:"size"`
	Epic          string    `json:"epic"`

	// Raw is the verbatim JSON body of the /confirms response. The typed
	// fields above don't cover everything the API can send (e.g. the
	// affectedDeals array on multi-leg confirmations, or reject-specific
	// fields); callers that need to diagnose a REJECTED confirmation with
	// an empty Reason can fall back to inspecting this.
	Raw string `json:"-"`
}

type dealReferenceResponse struct {
	DealReference string `json:"dealReference"`
}

func (c *Client) confirmDeal(ctx context.Context, dealReference string) (*DealConfirmation, error) {
	var conf DealConfirmation
	raw, err := c.doRequestRawBody(ctx, http.MethodGet, "/confirms/"+dealReference, nil, &conf)
	if err != nil {
		return nil, err
	}
	conf.Raw = string(raw)
	return &conf, nil
}
