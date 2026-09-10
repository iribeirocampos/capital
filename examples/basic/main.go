// Command basic demonstrates the most common operations exposed by the
// capital package: logging in, listing positions, opening one, and closing
// it again.
//
// It reads credentials from the environment so real credentials never end
// up in source control:
//
//	CAPITAL_IDENTIFIER  account email/username
//	CAPITAL_API_KEY     API key generated in the platform's API settings
//	CAPITAL_PASSWORD    account password
//	CAPITAL_DEMO        "true" to use the demo environment (recommended)
//
// Run it with: go run ./examples/basic
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/iribeirocampos/capital"
)

func main() {
	client := capital.NewClient(
		os.Getenv("CAPITAL_IDENTIFIER"),
		os.Getenv("CAPITAL_API_KEY"),
		os.Getenv("CAPITAL_PASSWORD"),
		capital.WithDemo(os.Getenv("CAPITAL_DEMO") == "true"),
	)

	ctx := context.Background()

	// Login is optional — it happens automatically on first use — but
	// calling it explicitly lets you fail fast on bad credentials.
	if err := client.Login(ctx); err != nil {
		log.Fatalf("login: %v", err)
	}

	positions, err := client.ListPositions(ctx)
	if err != nil {
		log.Fatalf("list positions: %v", err)
	}
	fmt.Printf("open positions: %d\n", len(positions))

	conf, err := client.CreatePosition(ctx, "GOLD", capital.Buy, 1,
		capital.WithStopLevel(1900),
		capital.WithProfitLevel(2000),
	)
	if err != nil {
		log.Fatalf("create position: %v", err)
	}
	fmt.Printf("opened position %s (%s)\n", conf.DealID, conf.Status)

	if _, err := client.ClosePosition(ctx, conf.DealID); err != nil {
		log.Fatalf("close position: %v", err)
	}
	fmt.Printf("closed position %s\n", conf.DealID)
}
