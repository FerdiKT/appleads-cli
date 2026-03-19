package appleads

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
)

func TestDoJSON_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Accept = %q", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	client := &Client{BaseURL: srv.URL, Token: "test-token"}
	var out map[string]string
	err := client.DoJSON(context.Background(), http.MethodGet, "/test", nil, nil, &out)
	if err != nil {
		t.Fatalf("DoJSON: %v", err)
	}
	if out["status"] != "ok" {
		t.Fatalf("status = %q, want ok", out["status"])
	}
}

func TestDoJSON_OrgIDHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Header.Get("X-AP-Context")
		if ctx != "orgId=42" {
			t.Errorf("X-AP-Context = %q, want orgId=42", ctx)
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := &Client{BaseURL: srv.URL, Token: "tok", OrgID: 42}
	err := client.DoJSON(context.Background(), http.MethodGet, "/test", nil, nil, nil)
	if err != nil {
		t.Fatalf("DoJSON: %v", err)
	}
}

func TestDoJSON_QueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("limit = %q", r.URL.Query().Get("limit"))
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := &Client{BaseURL: srv.URL, Token: "tok"}
	q := url.Values{}
	q.Set("limit", "10")
	err := client.DoJSON(context.Background(), http.MethodGet, "/test", q, nil, nil)
	if err != nil {
		t.Fatalf("DoJSON: %v", err)
	}
}

func TestDoJSON_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte(`{"error":"forbidden"}`))
	}))
	defer srv.Close()

	client := &Client{BaseURL: srv.URL, Token: "tok"}
	err := client.DoJSON(context.Background(), http.MethodGet, "/test", nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for 403")
	}
	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("error type = %T, want *HTTPError", err)
	}
	if httpErr.StatusCode != 403 {
		t.Fatalf("StatusCode = %d, want 403", httpErr.StatusCode)
	}
}

func TestDoJSON_Retry(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			w.WriteHeader(500)
			w.Write([]byte(`{"error":"internal"}`))
			return
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	client := &Client{BaseURL: srv.URL, Token: "tok"}
	var out map[string]any
	err := client.DoJSON(context.Background(), http.MethodGet, "/test", nil, nil, &out)
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if atomic.LoadInt32(&attempts) < 3 {
		t.Fatalf("expected at least 3 attempts, got %d", attempts)
	}
}

func TestDoJSON_NoRetryOn4xx(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(400)
		w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer srv.Close()

	client := &Client{BaseURL: srv.URL, Token: "tok"}
	err := client.DoJSON(context.Background(), http.MethodGet, "/test", nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for 400")
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Fatalf("should not retry 400, attempts = %d", attempts)
	}
}

func TestDoJSON_EmptyToken(t *testing.T) {
	client := &Client{BaseURL: "http://localhost", Token: ""}
	err := client.DoJSON(context.Background(), http.MethodGet, "/test", nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestDoJSON_EmptyBaseURL(t *testing.T) {
	client := &Client{BaseURL: "", Token: "tok"}
	err := client.DoJSON(context.Background(), http.MethodGet, "/test", nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for empty base URL")
	}
}

func TestDoJSON_PostWithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		w.Write([]byte(`{"created":true}`))
	}))
	defer srv.Close()

	client := &Client{BaseURL: srv.URL, Token: "tok"}
	body := map[string]any{"name": "test-campaign"}
	var out map[string]any
	err := client.DoJSON(context.Background(), http.MethodPost, "/campaigns", nil, body, &out)
	if err != nil {
		t.Fatalf("DoJSON POST: %v", err)
	}
}

func TestShouldRetryStatus(t *testing.T) {
	retryable := []int{429, 500, 502, 503, 504}
	for _, code := range retryable {
		if !shouldRetryStatus(code) {
			t.Errorf("shouldRetryStatus(%d) = false, want true", code)
		}
	}
	nonRetryable := []int{200, 201, 400, 401, 403, 404}
	for _, code := range nonRetryable {
		if shouldRetryStatus(code) {
			t.Errorf("shouldRetryStatus(%d) = true, want false", code)
		}
	}
}

func TestRetryDelay(t *testing.T) {
	d1 := retryDelay(1)
	d2 := retryDelay(2)

	// Each delay should increase (with jitter, at minimum equal to base without jitter).
	if d1 < baseRetryDelay {
		t.Fatalf("d1 = %v, want >= %v", d1, baseRetryDelay)
	}
	// d2 base is 500ms, d3 base is 1000ms
	if d2 < 2*baseRetryDelay {
		t.Fatalf("d2 = %v, want >= %v", d2, 2*baseRetryDelay)
	}

	// Should never exceed maxRetryDelay + 25% jitter.
	d10 := retryDelay(10)
	maxWithJitter := maxRetryDelay + maxRetryDelay/4
	if d10 > maxWithJitter {
		t.Fatalf("d10 = %v, exceeds max+jitter %v", d10, maxWithJitter)
	}
}
