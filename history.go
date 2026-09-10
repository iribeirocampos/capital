package capital

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// Activity is a single entry in the account's activity history.
type Activity struct {
	Date    string `json:"date"`
	DateUTC string `json:"dateUTC"`
	Epic    string `json:"epic"`
	DealID  string `json:"dealId"`
	Source  string `json:"source"`
	Type    string `json:"type"`
	Status  string `json:"status"`
}

type activityResponse struct {
	Activities []Activity `json:"activities"`
}

// HistoryOption configures optional query parameters for AccountActivity
// and TransactionHistory.
type HistoryOption func(url.Values)

// WithHistoryFrom sets the start of the date range, formatted as
// "2006-01-02T15:04:05".
func WithHistoryFrom(from string) HistoryOption {
	return func(v url.Values) { v.Set("from", from) }
}

// WithHistoryTo sets the end of the date range, formatted as
// "2006-01-02T15:04:05".
func WithHistoryTo(to string) HistoryOption {
	return func(v url.Values) { v.Set("to", to) }
}

// WithHistoryLastPeriod limits results to the last N seconds.
func WithHistoryLastPeriod(seconds int) HistoryOption {
	return func(v url.Values) { v.Set("lastPeriod", strconv.Itoa(seconds)) }
}

// AccountActivity returns account activity history. The API limits a
// single request to a 1-day date range.
func (c *Client) AccountActivity(ctx context.Context, opts ...HistoryOption) ([]Activity, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	q := url.Values{}
	for _, opt := range opts {
		opt(q)
	}
	path := "/history/activity"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	var resp activityResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Activities, nil
}

// Transaction is a single entry in the account's transaction history.
type Transaction struct {
	Date            string `json:"date"`
	DateUTC         string `json:"dateUtc"`
	InstrumentName  string `json:"instrumentName"`
	TransactionType string `json:"transactionType"`
	Reference       string `json:"reference"`
	Size            string `json:"size"`
	Currency        string `json:"currency"`
	Status          string `json:"status"`
}

type transactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
}

// TransactionHistory returns the account's transaction history.
func (c *Client) TransactionHistory(ctx context.Context, opts ...HistoryOption) ([]Transaction, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	q := url.Values{}
	for _, opt := range opts {
		opt(q)
	}
	path := "/history/transactions"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	var resp transactionsResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Transactions, nil
}
