package capital

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// testServer wires up an httptest.Server that fakes just enough of the
// Capital.com session flow (encryption key + login) for tests to exercise
// real Login()/doRequest() logic end to end, and lets each test register
// additional handlers for the endpoints it cares about.
type testServer struct {
	t          *testing.T
	mux        *http.ServeMux
	srv        *httptest.Server
	rsaKey     *rsa.PrivateKey
	loginCalls int
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	ts := &testServer{t: t, mux: http.NewServeMux(), rsaKey: key}

	ts.mux.HandleFunc("/session/encryptionKey", func(w http.ResponseWriter, r *http.Request) {
		derKey, err := x509.MarshalPKIXPublicKey(&ts.rsaKey.PublicKey)
		if err != nil {
			t.Fatalf("marshal public key: %v", err)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"encryptionKey": base64.StdEncoding.EncodeToString(derKey),
			"timeStamp":     1700000000000,
		})
	})

	ts.mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			ts.loginCalls++
			w.Header().Set("CST", "test-cst")
			w.Header().Set("X-SECURITY-TOKEN", "test-token")
			writeJSON(w, http.StatusOK, map[string]any{"accountId": "acc-1", "currency": "USD"})
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, map[string]any{})
		default:
			http.NotFound(w, r)
		}
	})

	ts.srv = httptest.NewServer(ts.mux)
	t.Cleanup(ts.srv.Close)
	return ts
}

func (ts *testServer) client() *Client {
	return NewClient("user@example.com", "api-key", "hunter2", WithBaseURL(ts.srv.URL))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func TestLoginEncryptsPasswordAndStoresSession(t *testing.T) {
	ts := newTestServer(t)
	c := ts.client()

	if err := c.Login(context.Background()); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if !c.hasSession() {
		t.Fatal("expected session to be established after Login")
	}
	if got := c.AccountID(); got != "acc-1" {
		t.Fatalf("AccountID() = %q, want %q", got, "acc-1")
	}
	if ts.loginCalls != 1 {
		t.Fatalf("expected exactly 1 login call, got %d", ts.loginCalls)
	}
}

func TestDoRequestMapsAPIError(t *testing.T) {
	ts := newTestServer(t)
	ts.mux.HandleFunc("/broken", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errorCode": "error.invalid.input"})
	})
	c := ts.client()

	err := c.doRequest(context.Background(), http.MethodGet, "/broken", nil, nil)
	if err == nil {
		t.Fatal("expected an error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusBadRequest)
	}
	if apiErr.Code != "error.invalid.input" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "error.invalid.input")
	}
}

func TestDoRequestReAuthenticatesOnceOnSessionExpiry(t *testing.T) {
	ts := newTestServer(t)
	c := ts.client()
	if err := c.Login(context.Background()); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	calls := 0
	ts.mux.HandleFunc("/expiring", func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"errorCode": "error.invalid.session.token"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.doRequest(context.Background(), http.MethodGet, "/expiring", nil, &out); err != nil {
		t.Fatalf("doRequest() error = %v", err)
	}
	if !out.OK {
		t.Error("expected ok=true after retry")
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls to /expiring (original + retry), got %d", calls)
	}
	if ts.loginCalls != 2 {
		t.Fatalf("expected 2 login calls (initial + re-login), got %d", ts.loginCalls)
	}
}

func TestCreateAndClosePosition(t *testing.T) {
	ts := newTestServer(t)
	c := ts.client()

	ts.mux.HandleFunc("/positions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		var body createPositionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Epic != "GOLD" || body.Direction != Buy || body.Size != 1 {
			t.Errorf("unexpected request body: %+v", body)
		}
		if body.StopLevel == nil || *body.StopLevel != 1900 {
			t.Errorf("expected stopLevel=1900, got %+v", body.StopLevel)
		}
		writeJSON(w, http.StatusOK, map[string]any{"dealReference": "ref-1"})
	})
	ts.mux.HandleFunc("/confirms/ref-1", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"dealReference": "ref-1",
			"dealId":        "deal-1",
			"status":        "OPEN",
			"epic":          "GOLD",
		})
	})
	ts.mux.HandleFunc("/positions/deal-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"dealReference": "ref-2"})
	})
	ts.mux.HandleFunc("/confirms/ref-2", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"dealReference": "ref-2", "dealId": "deal-1", "status": "CLOSED"})
	})

	conf, err := c.CreatePosition(context.Background(), "GOLD", Buy, 1, WithStopLevel(1900))
	if err != nil {
		t.Fatalf("CreatePosition() error = %v", err)
	}
	if conf.DealID != "deal-1" {
		t.Fatalf("DealID = %q, want %q", conf.DealID, "deal-1")
	}

	closeConf, err := c.ClosePosition(context.Background(), conf.DealID)
	if err != nil {
		t.Fatalf("ClosePosition() error = %v", err)
	}
	if closeConf.Status != "CLOSED" {
		t.Fatalf("Status = %q, want %q", closeConf.Status, "CLOSED")
	}
}

func TestSearchMarkets(t *testing.T) {
	ts := newTestServer(t)
	c := ts.client()

	ts.mux.HandleFunc("/markets", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("searchTerm"); got != "gold" {
			t.Errorf("searchTerm = %q, want %q", got, "gold")
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"markets": []map[string]any{{"epic": "GOLD", "instrumentName": "Gold"}},
		})
	})

	markets, err := c.SearchMarkets(context.Background(), "gold")
	if err != nil {
		t.Fatalf("SearchMarkets() error = %v", err)
	}
	if len(markets) != 1 || markets[0].Epic != "GOLD" {
		t.Fatalf("unexpected markets: %+v", markets)
	}
}

func TestParseAPIErrorFallsBackToRawBody(t *testing.T) {
	err := parseAPIError(http.StatusInternalServerError, []byte("not json"))
	if err.Code != "" {
		t.Errorf("expected empty Code for non-JSON body, got %q", err.Code)
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Error() = %q, expected it to mention status 500", err.Error())
	}
}
