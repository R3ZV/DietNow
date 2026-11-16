// Command smoke runs the MVP smoke test against a running deployment.
//
//	go run ./cmd/smoke -base-url http://localhost:8080 -register
//
// Credentials come from SMOKE_EMAIL / SMOKE_PASSWORD (never put them on the
// command line in CI, where arguments are logged). Use a dedicated
// non-production account. Exit code 0 means every step passed.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dietnow/backend/internal/smoke"
)

func main() {
	baseURL := flag.String("base-url", envOr("SMOKE_BASE_URL", "http://localhost:8080"), "deployment to test")
	register := flag.Bool("register", false, "register the account if login fails")
	version := flag.String("expect-version", os.Getenv("SMOKE_EXPECT_VERSION"), "fail unless /api/health reports this version")
	timeout := flag.Duration("timeout", 2*time.Minute, "overall timeout")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	err := smoke.Run(ctx, smoke.Options{
		BaseURL:       *baseURL,
		Email:         os.Getenv("SMOKE_EMAIL"),
		Password:      os.Getenv("SMOKE_PASSWORD"),
		Register:      *register,
		ExpectVersion: *version,
		Logf: func(format string, args ...any) {
			fmt.Printf("  ✓ "+format+"\n", args...)
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "SMOKE TEST FAILED against %s: %v\n", *baseURL, err)
		os.Exit(1)
	}
	fmt.Printf("SMOKE TEST PASSED against %s\n", *baseURL)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
