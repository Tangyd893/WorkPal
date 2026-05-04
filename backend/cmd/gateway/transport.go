package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"
)

type retryTransport struct {
	base        http.RoundTripper
	maxAttempts int
	backoff     time.Duration
	sleep       func(time.Duration)
}

func newRetryTransport(base http.RoundTripper, maxAttempts int, backoff time.Duration) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &retryTransport{
		base:        base,
		maxAttempts: maxInt(1, maxAttempts),
		backoff:     backoff,
		sleep:       time.Sleep,
	}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !isRetryableMethod(req.Method) || t.maxAttempts <= 1 {
		resp, err := t.base.RoundTrip(req)
		if resp != nil {
			resp.Header.Set("X-Gateway-Attempts", "1")
		}
		return resp, err
	}

	var lastErr error
	var lastResp *http.Response
	for attempt := 1; attempt <= t.maxAttempts; attempt++ {
		clonedReq, err := cloneRetryRequest(req)
		if err != nil {
			return nil, err
		}

		resp, err := t.base.RoundTrip(clonedReq)
		if err == nil && !shouldRetryStatus(resp.StatusCode) {
			resp.Header.Set("X-Gateway-Attempts", strconv.Itoa(attempt))
			return resp, nil
		}
		if resp != nil {
			resp.Header.Set("X-Gateway-Attempts", strconv.Itoa(attempt))
		}

		lastErr = err
		lastResp = resp
		if attempt == t.maxAttempts {
			break
		}
		if err == nil && resp != nil {
			_ = resp.Body.Close()
		}
		if t.backoff > 0 {
			t.sleep(t.backoff)
		}
	}

	return lastResp, lastErr
}

func cloneRetryRequest(req *http.Request) (*http.Request, error) {
	clonedReq := req.Clone(req.Context())
	if req.Body == nil {
		return clonedReq, nil
	}
	if req.GetBody == nil {
		return nil, errors.New("gateway retry requires GetBody when the request contains a body")
	}
	body, err := req.GetBody()
	if err != nil {
		return nil, err
	}
	clonedReq.Body = body
	return clonedReq, nil
}

func isRetryableMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func shouldRetryStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}
