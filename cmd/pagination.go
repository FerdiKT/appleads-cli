package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ferdikt/appleads-cli/internal/appleads"
)

func asInt(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	case string:
		n, err := strconv.Atoi(x)
		if err == nil {
			return n, true
		}
	}
	return 0, false
}

func cloneValues(v url.Values) url.Values {
	out := url.Values{}
	for k, vals := range v {
		for _, val := range vals {
			out.Add(k, val)
		}
	}
	return out
}

func fetchAllPages(ctx context.Context, client *appleads.Client, path string, query url.Values, offset, limit int) (map[string]any, error) {
	q := cloneValues(query)
	if q == nil {
		q = url.Values{}
	}
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 20
	}

	acc := []any{}
	var totalResults int
	var hasTotal bool
	current := offset

	for {
		q.Set("offset", strconv.Itoa(current))
		q.Set("limit", strconv.Itoa(limit))

		var page map[string]any
		if err := client.DoJSON(ctx, http.MethodGet, path, q, nil, &page); err != nil {
			return nil, err
		}

		pageData, _ := page["data"].([]any)
		acc = append(acc, pageData...)

		pagination, _ := page["pagination"].(map[string]any)
		if pagination == nil {
			break
		}

		total, ok := asInt(pagination["totalResults"])
		if ok {
			totalResults = total
			hasTotal = true
		}
		startIndex, _ := asInt(pagination["startIndex"])
		itemsPerPage, _ := asInt(pagination["itemsPerPage"])

		if itemsPerPage <= 0 {
			itemsPerPage = len(pageData)
		}
		if itemsPerPage <= 0 {
			break
		}

		next := startIndex + itemsPerPage
		if hasTotal && next >= totalResults {
			break
		}
		if next <= current {
			break
		}
		current = next
	}

	p := map[string]any{
		"startIndex":   offset,
		"itemsPerPage": len(acc),
	}
	if hasTotal {
		p["totalResults"] = totalResults
	} else {
		p["totalResults"] = len(acc)
	}

	return map[string]any{
		"data":       acc,
		"pagination": p,
		"error":      nil,
	}, nil
}

func fetchAllPagesWithFallback(ctx context.Context, client *appleads.Client, paths []string, query url.Values, offset, limit int) (map[string]any, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no paths provided")
	}
	var lastErr error
	for _, p := range paths {
		resp, err := fetchAllPages(ctx, client, p, query, offset, limit)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func callListEndpoint(ctx context.Context, client *appleads.Client, path string, query url.Values, offset, limit int, all bool) error {
	if !all {
		return callAPIAndPrint(ctx, client, http.MethodGet, path, query, nil)
	}
	resp, err := fetchAllPages(ctx, client, path, query, offset, limit)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func callListEndpointWithFallback(ctx context.Context, client *appleads.Client, paths []string, query url.Values, offset, limit int, all bool) error {
	if !all {
		return callAPIAndPrintWithFallback(ctx, client, http.MethodGet, paths, query, nil)
	}
	resp, err := fetchAllPagesWithFallback(ctx, client, paths, query, offset, limit)
	if err != nil {
		return err
	}
	return printJSON(resp)
}
