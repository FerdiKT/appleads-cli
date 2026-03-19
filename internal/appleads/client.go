package appleads

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	OrgID      int64
	Token      string
	HTTPClient *http.Client
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("api returned %d: %s", e.StatusCode, e.Body)
}

const (
	defaultMaxAttempts = 4
	baseRetryDelay     = 250 * time.Millisecond
	maxRetryDelay      = 3 * time.Second
	maxResponseBytes   = 10 << 20 // 10 MB
)

func shouldRetryStatus(status int) bool {
	switch status {
	case 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

func shouldRetryErr(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	// Covers temporary transport-level failures.
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "timeout")
}

func retryDelay(attempt int) time.Duration {
	// attempt starts at 1 for the first retry.
	delay := baseRetryDelay << (attempt - 1)
	if delay > maxRetryDelay {
		delay = maxRetryDelay
	}
	// Add jitter (up to 25% of delay) to avoid thundering herd.
	jitter := time.Duration(rand.Int63n(int64(delay) / 4))
	return delay + jitter
}

func (c *Client) doJSON(ctx context.Context, method, endpoint string, query url.Values, reqBody any, out any) error {
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	if c.BaseURL == "" {
		return fmt.Errorf("base url is empty")
	}
	if c.Token == "" {
		return fmt.Errorf("access token is empty")
	}

	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("parse base url: %w", err)
	}
	u.Path = path.Join(u.Path, endpoint)
	u.RawQuery = query.Encode()

	var bodyBytes []byte
	if reqBody != nil {
		raw, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("encode request payload: %w", err)
		}
		bodyBytes = raw
	}

	maxAttempts := defaultMaxAttempts
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var body io.Reader
		if reqBody != nil {
			body = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.Token)
		if c.OrgID > 0 {
			req.Header.Set("X-AP-Context", fmt.Sprintf("orgId=%d", c.OrgID))
		}
		if reqBody != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request api: %w", err)
			if attempt < maxAttempts && shouldRetryErr(err) {
				time.Sleep(retryDelay(attempt))
				continue
			}
			return lastErr
		}

		rawResp, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
		resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("read api response: %w", readErr)
			if attempt < maxAttempts {
				time.Sleep(retryDelay(attempt))
				continue
			}
			return lastErr
		}

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			httpErr := &HTTPError{
				StatusCode: resp.StatusCode,
				Body:       strings.TrimSpace(string(rawResp)),
			}
			lastErr = httpErr
			if attempt < maxAttempts && shouldRetryStatus(resp.StatusCode) {
				time.Sleep(retryDelay(attempt))
				continue
			}
			return httpErr
		}

		if out == nil {
			return nil
		}
		if err := json.Unmarshal(rawResp, out); err != nil {
			return fmt.Errorf("decode api response: %w", err)
		}
		return nil
	}
	return lastErr
}

func (c *Client) DoJSON(ctx context.Context, method, endpoint string, query url.Values, reqBody any, out any) error {
	return c.doJSON(ctx, method, endpoint, query, reqBody, out)
}
