package cmd

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var searchAppsFlags struct {
	OrgID  int64
	Query  string
	Limit  int
	Offset int
	All    bool
}

var searchGeoFlags struct {
	OrgID  int64
	Entity string
	Limit  int
	Offset int
	All    bool
}

func init() {
	searchCmd.AddCommand(searchAppsCmd)
	searchCmd.AddCommand(searchGeoCmd)

	searchAppsCmd.Flags().Int64Var(&searchAppsFlags.OrgID, "org-id", 0, "Organization ID override")
	searchAppsCmd.Flags().StringVar(&searchAppsFlags.Query, "query", "", "Search query text")
	searchAppsCmd.Flags().IntVar(&searchAppsFlags.Limit, "limit", 20, "Pagination limit")
	searchAppsCmd.Flags().IntVar(&searchAppsFlags.Offset, "offset", 0, "Pagination offset")
	searchAppsCmd.Flags().BoolVar(&searchAppsFlags.All, "all", false, "Fetch all pages")
	_ = searchAppsCmd.MarkFlagRequired("query")

	searchGeoCmd.Flags().Int64Var(&searchGeoFlags.OrgID, "org-id", 0, "Organization ID override")
	searchGeoCmd.Flags().StringVar(&searchGeoFlags.Entity, "entity", "Locality", "Geo entity (Country, AdminArea, Locality, District)")
	searchGeoCmd.Flags().IntVar(&searchGeoFlags.Limit, "limit", 20, "Pagination limit")
	searchGeoCmd.Flags().IntVar(&searchGeoFlags.Offset, "offset", 0, "Pagination offset")
	searchGeoCmd.Flags().BoolVar(&searchGeoFlags.All, "all", false, "Fetch all pages")
}

var searchAppsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Search App Store apps for Apple Ads",
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(searchAppsFlags.Query) == "" {
			return fmt.Errorf("--query is required")
		}
		if searchAppsFlags.Limit <= 0 {
			return fmt.Errorf("--limit must be > 0")
		}
		if searchAppsFlags.Offset < 0 {
			return fmt.Errorf("--offset must be >= 0")
		}

		client, _, _, err := authedClient(context.Background(), searchAppsFlags.OrgID, true)
		if err != nil {
			return err
		}
		q := url.Values{}
		q.Set("query", strings.TrimSpace(searchAppsFlags.Query))
		q.Set("limit", strconv.Itoa(searchAppsFlags.Limit))
		q.Set("offset", strconv.Itoa(searchAppsFlags.Offset))
		return callListEndpoint(context.Background(), client, "/search/apps", q, searchAppsFlags.Offset, searchAppsFlags.Limit, searchAppsFlags.All)
	},
}

var searchGeoCmd = &cobra.Command{
	Use:   "geo",
	Short: "Search geolocations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if searchGeoFlags.Limit <= 0 {
			return fmt.Errorf("--limit must be > 0")
		}
		if searchGeoFlags.Offset < 0 {
			return fmt.Errorf("--offset must be >= 0")
		}

		client, _, _, err := authedClient(context.Background(), searchGeoFlags.OrgID, true)
		if err != nil {
			return err
		}

		q := url.Values{}
		q.Set("entity", searchGeoFlags.Entity)
		q.Set("limit", strconv.Itoa(searchGeoFlags.Limit))
		q.Set("offset", strconv.Itoa(searchGeoFlags.Offset))

		return callListEndpointWithFallback(context.Background(), client, []string{
			"/search/geo",
			"/search/geolocations",
		}, q, searchGeoFlags.Offset, searchGeoFlags.Limit, searchGeoFlags.All)
	},
}
