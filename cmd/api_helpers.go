package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ferdikt/appleads-cli/internal/appleads"
	"github.com/ferdikt/appleads-cli/internal/config"
)

func authedClient(ctx context.Context, overrideOrgID int64, requireOrg bool) (*appleads.Client, *config.Config, *config.Profile, error) {
	cfg, profile, err := loadProfile()
	if err != nil {
		return nil, nil, nil, err
	}

	token, err := ensureAccessToken(ctx, cfg, profile)
	if err != nil {
		return nil, nil, nil, err
	}

	orgID := int64(0)
	if requireOrg || overrideOrgID > 0 || profile.OrgID > 0 {
		orgID, err = resolveOrgID(overrideOrgID, profile.OrgID)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	client := &appleads.Client{
		BaseURL: apiBaseURL(profile),
		OrgID:   orgID,
		Token:   token,
	}
	return client, cfg, profile, nil
}

func parseQueryParams(items []string) (url.Values, error) {
	q := url.Values{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
			return nil, fmt.Errorf("invalid --query value %q (expected key=value)", item)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		q.Add(key, value)
	}
	return q, nil
}

func readJSONPayload(body, bodyFile string, allowEmpty bool) (any, error) {
	raw := ""
	switch {
	case strings.TrimSpace(body) != "":
		raw = body
	case strings.TrimSpace(bodyFile) != "":
		data, err := os.ReadFile(bodyFile)
		if err != nil {
			return nil, fmt.Errorf("read body file: %w", err)
		}
		raw = string(data)
	default:
		if info, err := os.Stdin.Stat(); err == nil && (info.Mode()&os.ModeCharDevice) == 0 {
			stdin, err := io.ReadAll(os.Stdin)
			if err == nil && len(strings.TrimSpace(string(stdin))) > 0 {
				raw = string(stdin)
			}
		}
	}

	if strings.TrimSpace(raw) == "" {
		if allowEmpty {
			return nil, nil
		}
		return nil, fmt.Errorf("request body is required (pass --body, --body-file, or pipe JSON)")
	}

	var payload any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("parse JSON body: %w", err)
	}
	return payload, nil
}

func callAPIAndPrint(ctx context.Context, client *appleads.Client, method, path string, query url.Values, payload any) error {
	var resp any
	if err := client.DoJSON(ctx, method, path, query, payload, &resp); err != nil {
		return err
	}
	return printJSON(resp)
}

func withNotFoundHint(err error, hint string) error {
	if err == nil {
		return nil
	}
	var httpErr *appleads.HTTPError
	if errors.As(err, &httpErr) && httpErr.StatusCode == 404 {
		return fmt.Errorf("%s: %w", hint, err)
	}
	return err
}

func callAPIAndPrintWithFallback(ctx context.Context, client *appleads.Client, method string, paths []string, query url.Values, payload any) error {
	var lastErr error
	for i, p := range paths {
		var resp any
		err := client.DoJSON(ctx, method, p, query, payload, &resp)
		if err == nil {
			return printJSON(resp)
		}
		lastErr = err
		if i < len(paths)-1 {
			time.Sleep(50 * time.Millisecond)
		}
	}
	return fmt.Errorf("request failed for all candidate paths %v: %w", paths, lastErr)
}
