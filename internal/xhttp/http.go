package xhttp

import (
	"errors"
	"net/http"
	"time"
)

// Transport is a custom http.RoundTripper that adds retry logic.
type Transport struct {
	Base       http.RoundTripper
	RetryCount uint8
	Cooldown   time.Duration
}

// RoundTrip executes a single HTTP transaction, adding retry logic.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	for i := uint8(0); i < t.RetryCount; i++ {
		if t.Cooldown > 0 {
			time.Sleep(t.Cooldown)
		}
		if resp, err := base.RoundTrip(req); err == nil {
			return resp, nil
		}
	}

	return nil, errors.New("http transport retry limit exceeded")
}

func NewClient(retryCount uint8, cd time.Duration) *http.Client {
	return &http.Client{
		Transport: &Transport{
			RetryCount: retryCount,
			Cooldown:   cd,
		},
	}
}
