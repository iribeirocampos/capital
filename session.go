package capital

import (
	"context"
	"fmt"
	"net/http"
)

type encryptionKeyResponse struct {
	EncryptionKey string `json:"encryptionKey"`
	TimeStamp     int64  `json:"timeStamp"`
}

// SessionDetails describes the currently active account/session.
type SessionDetails struct {
	AccountID     string `json:"accountId"`
	ClientID      string `json:"clientId"`
	Currency      string `json:"currency"`
	StreamingHost string `json:"streamingHost"`
}

type sessionLoginRequest struct {
	Identifier        string `json:"identifier"`
	Password          string `json:"password"`
	EncryptedPassword bool   `json:"encryptedPassword"`
}

// Login authenticates with Capital.com and stores the resulting session
// tokens on the Client. Other methods call Login automatically on first use
// and after session expiry, so most callers never need to call it directly
// — it's exposed for cases where you want to fail fast on bad credentials
// or explicitly control when authentication happens.
func (c *Client) Login(ctx context.Context) error {
	var keyResp encryptionKeyResponse
	if _, err := c.doRequestHeaders(ctx, http.MethodGet, "/session/encryptionKey", nil, &keyResp, false); err != nil {
		return fmt.Errorf("capital: fetch encryption key: %w", err)
	}

	encryptedPassword, err := encryptPassword(c.password, keyResp.EncryptionKey, keyResp.TimeStamp)
	if err != nil {
		return fmt.Errorf("capital: encrypt password: %w", err)
	}

	payload := sessionLoginRequest{
		Identifier:        c.identifier,
		Password:          encryptedPassword,
		EncryptedPassword: true,
	}

	var sess SessionDetails
	headers, err := c.doRequestHeaders(ctx, http.MethodPost, "/session", payload, &sess, false)
	if err != nil {
		return fmt.Errorf("capital: login: %w", err)
	}

	cst := headers.Get("CST")
	securityToken := headers.Get("X-SECURITY-TOKEN")
	if cst == "" || securityToken == "" {
		return fmt.Errorf("capital: login succeeded but session headers were missing")
	}
	c.setSession(cst, securityToken)
	c.setAccountID(sess.AccountID)
	return nil
}

// Logout invalidates the current session on the server and clears local
// session state. It is a no-op if no session is active.
func (c *Client) Logout(ctx context.Context) error {
	if !c.hasSession() {
		return nil
	}
	err := c.doRequest(ctx, http.MethodDelete, "/session", nil, nil)
	c.clearSession()
	return err
}

// CurrentSession returns details about the currently active session,
// establishing one first if necessary.
func (c *Client) CurrentSession(ctx context.Context) (*SessionDetails, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	var sess SessionDetails
	if err := c.doRequest(ctx, http.MethodGet, "/session", nil, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

type switchAccountRequest struct {
	AccountID string `json:"accountId"`
}

// SwitchAccount changes the active trading account for the current session.
func (c *Client) SwitchAccount(ctx context.Context, accountID string) error {
	if err := c.ensureSession(ctx); err != nil {
		return err
	}
	if err := c.doRequest(ctx, http.MethodPut, "/session", switchAccountRequest{AccountID: accountID}, nil); err != nil {
		return err
	}
	c.setAccountID(accountID)
	return nil
}
