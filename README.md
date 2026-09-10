# capital

An unofficial Go client for the [Capital.com Public API](https://open-api.capital.com/).

I am in no way affiliated with Capital.com, use at your own risk. Inspired by
my Python wrapper, [python-capital](https://github.com/iribeirocampos/python-capital).

If you want to invest in CFDs, go to [capital.com](https://capital.com/). If
you want to automate interactions with Capital.com from Go, stick around.

## Features

- Simple handling of authentication, with end-to-end password encryption
- A persistent session that logs in lazily and re-authenticates automatically
  on expiry, instead of opening a new session on every call
- `context.Context` on every network call
- Positions and working (limit/stop) order management
- Market navigation, market search, and historical prices
- Client sentiment, watchlists, account activity and transaction history
- Typed errors (`*capital.APIError`) carrying the API's status code and error code
- Zero third-party dependencies (standard library only)

## Install

```bash
go get github.com/iribeirocampos/capital
```

## Quick start

[Register an account with Capital.com](https://capital.com/) and
[generate an API key](https://capital.com/trading/platform/?popup=settings&tab=APISettings).
To use a demo account, pass `capital.WithDemo(true)` when creating the client.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/iribeirocampos/capital"
)

func main() {
	client := capital.NewClient("identifier", "api-key", "password", capital.WithDemo(true))
	ctx := context.Background()

	// Login happens automatically on first use, but you can call it
	// explicitly to fail fast on bad credentials.
	if err := client.Login(ctx); err != nil {
		log.Fatal(err)
	}

	// List all open positions.
	positions, err := client.ListPositions(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Open a position, with optional risk-management parameters.
	conf, err := client.CreatePosition(ctx, "GOLD", capital.Buy, 1,
		capital.WithStopLevel(1900),
		capital.WithProfitLevel(2000),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("opened", conf.DealID)

	// Close it again.
	if _, err := client.ClosePosition(ctx, conf.DealID); err != nil {
		log.Fatal(err)
	}

	_ = positions
}
```

See [`examples/basic`](examples/basic/main.go) for a runnable version that
reads credentials from environment variables.

## Error handling

Every method returns a `*capital.APIError` (via the standard `error`
interface) when the API responds with a non-2xx status:

```go
if _, err := client.ClosePosition(ctx, dealID); err != nil {
	var apiErr *capital.APIError
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.StatusCode, apiErr.Code)
	}
}
```

## Session management

Unlike a "login per call" design, `Client` keeps its session (`CST` /
`X-SECURITY-TOKEN`) cached across calls and only re-authenticates when the
API reports the session has expired, or on first use. A `Client` is safe for
concurrent use by multiple goroutines.

## License

MIT
