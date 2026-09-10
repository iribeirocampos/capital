package capital

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// Resolution is the candle/bar resolution for historical price requests.
type Resolution string

const (
	ResolutionMinute   Resolution = "MINUTE"
	ResolutionMinute5  Resolution = "MINUTE_5"
	ResolutionMinute15 Resolution = "MINUTE_15"
	ResolutionMinute30 Resolution = "MINUTE_30"
	ResolutionHour     Resolution = "HOUR"
	ResolutionHour4    Resolution = "HOUR_4"
	ResolutionDay      Resolution = "DAY"
	ResolutionWeek     Resolution = "WEEK"
)

// PriceBidOffer holds a bid/ask price pair.
type PriceBidOffer struct {
	Bid float64 `json:"bid"`
	Ask float64 `json:"ask"`
}

// PricePoint is a single historical price candle.
type PricePoint struct {
	SnapshotTime     string        `json:"snapshotTime"`
	OpenPrice        PriceBidOffer `json:"openPrice"`
	ClosePrice       PriceBidOffer `json:"closePrice"`
	HighPrice        PriceBidOffer `json:"highPrice"`
	LowPrice         PriceBidOffer `json:"lowPrice"`
	LastTradedVolume float64       `json:"lastTradedVolume"`
}

type pricesResponse struct {
	Prices         []PricePoint `json:"prices"`
	InstrumentType string       `json:"instrumentType"`
}

// PricesOption configures an optional GetPrices query parameter.
type PricesOption func(url.Values)

// WithPricesResolution sets the candle resolution (default MINUTE).
func WithPricesResolution(r Resolution) PricesOption {
	return func(v url.Values) { v.Set("resolution", string(r)) }
}

// WithPricesMax limits the number of returned price points (default 10,
// API max 1000).
func WithPricesMax(max int) PricesOption {
	return func(v url.Values) { v.Set("max", strconv.Itoa(max)) }
}

// WithPricesFrom sets the start of the date range, formatted as
// "2006-01-02T15:04:05".
func WithPricesFrom(from string) PricesOption {
	return func(v url.Values) { v.Set("from", from) }
}

// WithPricesTo sets the end of the date range, formatted as
// "2006-01-02T15:04:05".
func WithPricesTo(to string) PricesOption {
	return func(v url.Values) { v.Set("to", to) }
}

// GetPrices returns historical prices for the given epic.
func (c *Client) GetPrices(ctx context.Context, epic string, opts ...PricesOption) ([]PricePoint, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("resolution", string(ResolutionMinute))
	q.Set("max", "10")
	for _, opt := range opts {
		opt(q)
	}
	path := "/prices/" + url.PathEscape(epic) + "?" + q.Encode()
	var resp pricesResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Prices, nil
}
