package capital

import (
	"context"
	"net/http"
	"net/url"
)

// Sentiment holds the proportion of clients long/short on a market.
type Sentiment struct {
	MarketID string  `json:"marketId"`
	LongPct  float64 `json:"longPositionPercentage"`
	ShortPct float64 `json:"shortPositionPercentage"`
}

// ClientSentiment returns the client sentiment (long/short percentages) for
// the given market ID.
func (c *Client) ClientSentiment(ctx context.Context, marketID string) (*Sentiment, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var sent Sentiment
	path := "/clientsentiment/" + url.PathEscape(marketID)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &sent); err != nil {
		return nil, err
	}
	return &sent, nil
}
