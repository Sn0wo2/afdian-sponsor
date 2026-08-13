package xhttp

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type RetryAttempt struct {
	Number   uint8
	Limit    uint8
	Cooldown time.Duration
}

type RetryHook func(attempt RetryAttempt, err error)

type Transport struct {
	RetryCount uint8
	Cooldown   time.Duration
	Base       http.RoundTripper
	OnRetry    RetryHook
}

func (transport *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	base := transport.Base
	if base == nil {
		base = http.DefaultTransport
	}

	response, err := base.RoundTrip(request)
	if err == nil {
		return response, nil
	}

	discardResponse := func(response *http.Response) {
		if response == nil || response.Body == nil {
			return
		}

		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
	}

	discardResponse(response)

	if contextErr := request.Context().Err(); contextErr != nil {
		return nil, contextErr
	}

	if transport.RetryCount == 0 {
		return nil, err
	}

	if request.Body != nil && request.GetBody == nil {
		return nil, fmt.Errorf("request failed and its body cannot be replayed: %w", err)
	}

	for number := 1; number <= int(transport.RetryCount); number++ {
		if transport.OnRetry != nil {
			transport.OnRetry(RetryAttempt{
				Number:   uint8(number),
				Limit:    transport.RetryCount,
				Cooldown: transport.Cooldown,
			}, err)
		}

		if transport.Cooldown > 0 {
			timer := time.NewTimer(transport.Cooldown)
			select {
			case <-request.Context().Done():
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}

				return nil, request.Context().Err()
			case <-timer.C:
			}
		} else if contextErr := request.Context().Err(); contextErr != nil {
			return nil, contextErr
		}

		retryRequest := request.Clone(request.Context())
		if request.Body != nil {
			retryRequest.Body, err = request.GetBody()
			if err != nil {
				return nil, fmt.Errorf("recreate request body: %w", err)
			}
		}

		response, err = base.RoundTrip(retryRequest)
		if err == nil {
			return response, nil
		}

		discardResponse(response)
	}

	return nil, err
}

func NewClient(retryCount uint8, cooldown time.Duration, hook RetryHook) *http.Client {
	return &http.Client{
		Transport: &Transport{
			RetryCount: retryCount,
			Cooldown:   cooldown,
			OnRetry:    hook,
		},
	}
}
