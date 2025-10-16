package xhttp

import (
	"context"
	"net/http"
	"time"
)

const (
	RetryCountKey = "retry-count"
)

type XHTTP struct {
	MaxRetryCount uint8
	NowRetryCount uint8
	Cooldown      time.Duration
}

type RetryHook func(attempt *XHTTP, err error)

type Transport struct {
	RetryCount uint8
	Cooldown   time.Duration
	Base       http.RoundTripper
	OnRetry    RetryHook // Optional. Called on each retry.
}

// RoundTrip executes a single HTTP transaction, adding retry logic.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	xHTTP := &XHTTP{t.RetryCount, 0, t.Cooldown}

	req = req.WithContext(context.WithValue(req.Context(), RetryCountKey, xHTTP))

	resp, err := base.RoundTrip(req)
	if err == nil {
		return resp, nil
	}

	for i := uint8(1); i <= t.RetryCount; i++ {
		xHTTP.NowRetryCount = i
		req = req.WithContext(context.WithValue(req.Context(), RetryCountKey, xHTTP))

		if t.OnRetry != nil {
			t.OnRetry(xHTTP, err)
		}

		if t.Cooldown > 0 {
			time.Sleep(t.Cooldown)
		}

		resp, err = base.RoundTrip(req)
		if err == nil {
			break
		}
	}

	return resp, err
}

// NewClient creates a new http.Client with our custom retry transport.
func NewClient(retryCount uint8, cd time.Duration, hook RetryHook) *http.Client {
	return &http.Client{
		Transport: &Transport{
			RetryCount: retryCount,
			Cooldown:   cd,
			OnRetry:    hook,
		},
	}
}

// GetRetryCount extracts the xhttp from a request's context.
// It returns nil if the retry count key is not found.
func GetRetryCount(req *http.Request) *XHTTP {
	if req == nil {
		return nil
	}
	if xHTTP, ok := req.Context().Value(RetryCountKey).(*XHTTP); ok {
		return xHTTP
	}
	return nil
}
