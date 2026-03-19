package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

var apiFlags struct {
	Method   string
	Path     string
	OrgID    int64
	Query    []string
	Body     string
	BodyFile string
}

func init() {
	rootCmd.AddCommand(apiCmd)

	apiCmd.Flags().StringVar(&apiFlags.Method, "method", "GET", "HTTP method (GET|POST|PUT|DELETE)")
	apiCmd.Flags().StringVar(&apiFlags.Path, "path", "", "API path, e.g. /campaigns or /reports/campaigns")
	apiCmd.Flags().Int64Var(&apiFlags.OrgID, "org-id", 0, "Organization ID override")
	apiCmd.Flags().StringArrayVar(&apiFlags.Query, "query", nil, "Query parameter as key=value (repeatable)")
	apiCmd.Flags().StringVar(&apiFlags.Body, "body", "", "Inline JSON request body")
	apiCmd.Flags().StringVar(&apiFlags.BodyFile, "body-file", "", "Path to JSON request body file")
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Raw Apple Ads API call for any endpoint",
	Example: "  appleads api --method GET --path /campaigns --query limit=20 --query offset=0\n" +
		"  appleads api --method POST --path /campaigns/find --body-file ./payloads/campaign-find.json\n" +
		"  appleads api --method POST --path /reports/campaigns --body '{\"startTime\":\"2026-03-01\",\"endTime\":\"2026-03-17\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		method := strings.ToUpper(strings.TrimSpace(apiFlags.Method))
		switch method {
		case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete:
		default:
			return fmt.Errorf("unsupported --method %q", method)
		}

		path := strings.TrimSpace(apiFlags.Path)
		if path == "" {
			return fmt.Errorf("--path is required")
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		query, err := parseQueryParams(apiFlags.Query)
		if err != nil {
			return err
		}

		requireOrg := strings.HasPrefix(path, "/acls") == false && strings.HasPrefix(path, "/me") == false
		client, _, _, err := authedClient(context.Background(), apiFlags.OrgID, requireOrg)
		if err != nil {
			return err
		}

		allowEmpty := method == http.MethodGet || method == http.MethodDelete
		payload, err := readJSONPayload(apiFlags.Body, apiFlags.BodyFile, allowEmpty)
		if err != nil {
			return err
		}

		return callAPIAndPrint(context.Background(), client, method, path, query, payload)
	},
}
