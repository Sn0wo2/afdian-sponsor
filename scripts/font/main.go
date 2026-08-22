package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed charset.txt
var charset string

const (
	defaultFamily = "Noto Sans SC:wght@100..900"
	maxFontBytes  = 5 * 1024 * 1024
)

var GoogleFontURLPattern = regexp.MustCompile(`url\((?:'|")?(https://fonts\.gstatic\.com/[^)'"]+)`)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	family := flag.String("family", envOr("AFDIAN_FONT_FAMILY", defaultFamily), "Google Fonts CSS2 family value")
	out := flag.String("out", "", "output font file path (default: <repo root>/internal/font/font.ttf)")
	ghaEnv := flag.String("gha-env", "", "GitHub Actions env file to append AFDIAN_FONT_FILE to; requires AFDIAN_FONT_FAMILY to be set, otherwise exits without fetching")
	flag.Parse()

	if *ghaEnv != "" && os.Getenv("AFDIAN_FONT_FAMILY") == "" {
		fmt.Println("AFDIAN_FONT_FAMILY is not set; the font embedded in the binary at build time will be used")
		return nil
	}

	if *out == "" {
		root, err := findRepoRoot()
		if err != nil {
			return err
		}

		*out = filepath.Join(root, "internal", "font", "font.ttf")
	}

	if len([]rune(charset)) > 700 {
		fmt.Fprintf(os.Stderr, "warning: charset has %d characters, Google Fonts may serve the full font instead of a subset\n", len([]rune(charset)))
	}

	query := url.Values{}
	query.Set("display", "swap")
	query.Set("family", *family)
	query.Set("text", charset)

	stylesheet, err := fetchResource("https://fonts.googleapis.com/css2?" + query.Encode())
	if err != nil {
		return fmt.Errorf("fetch stylesheet: %w", err)
	}

	match := GoogleFontURLPattern.FindSubmatch(stylesheet)
	if len(match) != 2 {
		return errors.New("stylesheet does not contain a font URL")
	}

	fontURL, err := url.Parse(string(match[1]))
	if err != nil || fontURL.Scheme != "https" || fontURL.Hostname() != "fonts.gstatic.com" {
		return errors.New("stylesheet contains an invalid font URL")
	}

	font, err := fetchResource(fontURL.String())
	if err != nil {
		return fmt.Errorf("fetch font: %w", err)
	}

	if len(font) > maxFontBytes {
		return fmt.Errorf("font is %d bytes, larger than the %d byte limit; the charset may exceed Google Fonts' subset limit", len(font), maxFontBytes)
	}

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	if err := os.WriteFile(*out, font, 0o644); err != nil {
		return fmt.Errorf("write font to %s: %w", *out, err)
	}

	fmt.Printf("Fetched %s (%d bytes) for family %q and wrote to %s\n", fontURL.Host, len(font), *family, *out)

	if *ghaEnv != "" {
		absolute, err := filepath.Abs(*out)
		if err != nil {
			return fmt.Errorf("resolve font path: %w", err)
		}

		envFile, err := os.OpenFile(*ghaEnv, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("open GitHub Actions env file: %w", err)
		}

		defer func() {
			_ = envFile.Close()
		}()

		if _, err := fmt.Fprintf(envFile, "AFDIAN_FONT_FILE=%s\n", absolute); err != nil {
			return fmt.Errorf("write GitHub Actions env file: %w", err)
		}
	}

	return nil
}

func fetchResource(resourceURL string) ([]byte, error) {
	resp, err := http.Get(resourceURL)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4*1024))

		return nil, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxFontBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return data, nil
}

func findRepoRoot() (string, error) {
	directory, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory, nil
		}

		parent := filepath.Dir(directory)
		if parent == directory {
			return "", errors.New("could not find go.mod in any parent directory")
		}

		directory = parent
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); strings.TrimSpace(value) != "" {
		return value
	}

	return fallback
}
