package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type RetryTransport struct {
	Base          http.RoundTripper
	MaxRetries429 int
	MaxRetries5xx int
	BaseDelay     time.Duration
	Sleep         func(ctx context.Context, d time.Duration) error
}

func NewRetryTransport(base http.RoundTripper) *RetryTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &RetryTransport{
		Base:          base,
		MaxRetries429: 3,
		MaxRetries5xx: 2,
		BaseDelay:     500 * time.Millisecond,
	}
}

func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := ensureReplayableBody(req); err != nil {
		return nil, err
	}
	var resp *http.Response
	var err error
	retries429 := 0
	retries5xx := 0
	for {
		if req.GetBody != nil {
			if req.Body != nil {
				_ = req.Body.Close()
			}
			body, getErr := req.GetBody()
			if getErr != nil {
				return nil, fmt.Errorf("reset request body: %w", getErr)
			}
			req.Body = body
		}
		resp, err = t.Base.RoundTrip(req)
		if err != nil {
			return nil, fmt.Errorf("round trip: %w", err)
		}
		if resp.StatusCode < 400 {
			return resp, nil
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			if retries429 >= t.MaxRetries429 {
				return resp, nil
			}
			delay := t.calculateBackoff(retries429, resp)
			drainAndClose(resp.Body)
			if err := t.sleep(req.Context(), delay); err != nil {
				return nil, err
			}
			retries429++
			continue
		}
		if resp.StatusCode >= 500 {
			if retries5xx >= t.MaxRetries5xx {
				return resp, nil
			}
			drainAndClose(resp.Body)
			if err := t.sleep(req.Context(), t.BaseDelay); err != nil {
				return nil, err
			}
			retries5xx++
			continue
		}
		return resp, nil
	}
}

func (t *RetryTransport) calculateBackoff(attempt int, resp *http.Response) time.Duration {
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			if seconds < 0 {
				return 0
			}
			return time.Duration(seconds) * time.Second
		}
		if tm, err := http.ParseTime(retryAfter); err == nil {
			d := time.Until(tm)
			if d < 0 {
				return 0
			}
			return d
		}
	}
	if t.BaseDelay <= 0 {
		return 0
	}
	bd := t.BaseDelay * time.Duration(1<<attempt)
	if bd <= 0 {
		return 0
	}
	jitterRange := bd / 2
	if jitterRange <= 0 {
		return bd
	}
	jitter := time.Duration(rand.Int63n(int64(jitterRange)))
	return bd + jitter
}

func (t *RetryTransport) sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	if t.Sleep != nil {
		return t.Sleep(ctx, d)
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func ensureReplayableBody(req *http.Request) error {
	if req == nil || req.Body == nil || req.GetBody != nil {
		return nil
	}
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("read request body: %w", err)
	}
	_ = req.Body.Close()
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(bodyBytes)), nil
	}
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return nil
}

func drainAndClose(body io.ReadCloser) {
	if body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(body, 1<<20))
	_ = body.Close()
}
