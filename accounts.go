package capital

import (
	"context"
	"net/http"
)

// Account represents a trading account belonging to the authenticated user.
type Account struct {
	AccountID   string `json:"accountId"`
	AccountName string `json:"accountName"`
	Currency    string `json:"currency"`
	AccountType string `json:"accountType"`
	Preferred   bool   `json:"preferred"`
	Balance     struct {
		Balance    float64 `json:"balance"`
		Deposit    float64 `json:"deposit"`
		ProfitLoss float64 `json:"profitLoss"`
		Available  float64 `json:"available"`
	} `json:"balance"`
}

type accountsResponse struct {
	Accounts []Account `json:"accounts"`
}

// ListAccounts returns all accounts available to the authenticated user.
func (c *Client) ListAccounts(ctx context.Context) ([]Account, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var resp accountsResponse
	if err := c.doRequest(ctx, http.MethodGet, "/accounts", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Accounts, nil
}

// AccountPreferences holds leverage and hedging-mode settings for the
// active account. Leverages maps an asset class (e.g. "SHARES", "INDICES",
// "CRYPTOCURRENCIES") to its leverage ratio.
type AccountPreferences struct {
	Leverages   map[string]float64 `json:"leverages"`
	HedgingMode bool               `json:"hedgingMode"`
}

// GetAccountPreferences returns the leverage and hedging-mode settings for
// the active account.
func (c *Client) GetAccountPreferences(ctx context.Context) (*AccountPreferences, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var prefs AccountPreferences
	if err := c.doRequest(ctx, http.MethodGet, "/accounts/preferences", nil, &prefs); err != nil {
		return nil, err
	}
	return &prefs, nil
}

// UpdateAccountPreferences updates leverage and hedging-mode settings for
// the active account.
func (c *Client) UpdateAccountPreferences(ctx context.Context, prefs AccountPreferences) error {
	if err := c.ensureSession(ctx); err != nil {
		return err
	}
	return c.doRequest(ctx, http.MethodPut, "/accounts/preferences", prefs, nil)
}

type topUpRequest struct {
	Amount float64 `json:"amount"`
}

type topUpResponse struct {
	Successful bool `json:"successful"`
}

// TopUpDemoAccount adjusts the balance of a demo account by amount (range
// -400000 to 400000). It only works against the demo environment (see
// WithDemo) and reports whether the top-up succeeded.
func (c *Client) TopUpDemoAccount(ctx context.Context, amount float64) (bool, error) {
	if err := c.ensureSession(ctx); err != nil {
		return false, err
	}
	var resp topUpResponse
	if err := c.doRequest(ctx, http.MethodPost, "/accounts/topUp", topUpRequest{Amount: amount}, &resp); err != nil {
		return false, err
	}
	return resp.Successful, nil
}
