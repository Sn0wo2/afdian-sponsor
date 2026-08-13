package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Sn0wo2/afdian-sponsor/internal/xhttp"
	"github.com/Sn0wo2/afdian-sponsor/version"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	client := xhttp.NewClient(3, 2*time.Second, func(attempt xhttp.RetryAttempt, err error) {
		fmt.Printf("HTTP request failed, retrying... (attempt: %d, max: %d, cd: %s, error: %v)\n", attempt.Number, attempt.Limit, attempt.Cooldown, err)
	})

	fmt.Printf("afdian-sponsor %s\n", version.String())

	config, err := LoadConfig(os.LookupEnv)
	if err != nil {
		return err
	}

	sponsors, err := QuerySponsors(client, config.UserID, config.APIToken, config.TotalSponsor)
	if err != nil {
		return err
	}

	sponsors.SortBy(config.Sort)

	svg, err := NewSVGRenderer(client, config).Render(sponsors)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(config.Output), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	if err := os.WriteFile(config.Output, []byte(svg), 0o644); err != nil { //nolint:gosec // Generated SVGs are intended to be publicly readable.
		return fmt.Errorf("write SVG to %s: %w", config.Output, err)
	}

	fmt.Printf("SVG file saved to %s\n", config.Output)

	fmt.Println("❤ Thank you for using afdian-sponsor(https://github.com/Sn0wo2/afdian-sponsor)~")

	return nil
}
