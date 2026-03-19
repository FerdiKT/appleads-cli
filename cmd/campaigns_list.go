package cmd

import (
	"context"
	"fmt"

	"github.com/ferdikt/appleads-cli/internal/appleads"
	"github.com/spf13/cobra"
)

var campaignsListFlags struct {
	Offset int
	Limit  int
	OrgID  int64
	All    bool
}

func init() {
	campaignsCmd.AddCommand(campaignsListCmd)
	campaignsListCmd.Flags().IntVar(&campaignsListFlags.Offset, "offset", 0, "Pagination offset")
	campaignsListCmd.Flags().IntVar(&campaignsListFlags.Limit, "limit", 20, "Pagination limit")
	campaignsListCmd.Flags().Int64Var(&campaignsListFlags.OrgID, "org-id", 0, "Organization ID override (uses profile org_id when omitted)")
	campaignsListCmd.Flags().BoolVar(&campaignsListFlags.All, "all", false, "Fetch all pages")
}

var campaignsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List campaigns",
	RunE: func(cmd *cobra.Command, args []string) error {
		if campaignsListFlags.Limit <= 0 {
			return fmt.Errorf("--limit must be > 0")
		}
		if campaignsListFlags.Offset < 0 {
			return fmt.Errorf("--offset must be >= 0")
		}

		cfg, profile, err := loadProfile()
		if err != nil {
			return err
		}
		orgID, err := resolveOrgID(campaignsListFlags.OrgID, profile.OrgID)
		if err != nil {
			return err
		}

		token, err := ensureAccessToken(context.Background(), cfg, profile)
		if err != nil {
			return err
		}

		client := &appleads.Client{
			BaseURL: apiBaseURL(profile),
			OrgID:   orgID,
			Token:   token,
		}

		var resp *appleads.CampaignListResponse
		if campaignsListFlags.All {
			allData := []appleads.Campaign{}
			current := campaignsListFlags.Offset
			for {
				page, err := client.ListCampaigns(context.Background(), current, campaignsListFlags.Limit)
				if err != nil {
					return err
				}
				allData = append(allData, page.Data...)

				p := page.Pagination
				start, _ := asInt(p["startIndex"])
				items, _ := asInt(p["itemsPerPage"])
				total, ok := asInt(p["totalResults"])
				if items <= 0 {
					items = len(page.Data)
				}
				if items <= 0 {
					resp = &appleads.CampaignListResponse{Data: allData, Pagination: map[string]any{
						"startIndex":   campaignsListFlags.Offset,
						"itemsPerPage": len(allData),
						"totalResults": len(allData),
					}}
					break
				}
				next := start + items
				if ok && next >= total {
					resp = &appleads.CampaignListResponse{Data: allData, Pagination: map[string]any{
						"startIndex":   campaignsListFlags.Offset,
						"itemsPerPage": len(allData),
						"totalResults": total,
					}}
					break
				}
				if next <= current {
					resp = &appleads.CampaignListResponse{Data: allData, Pagination: map[string]any{
						"startIndex":   campaignsListFlags.Offset,
						"itemsPerPage": len(allData),
						"totalResults": len(allData),
					}}
					break
				}
				current = next
			}
		} else {
			page, err := client.ListCampaigns(context.Background(), campaignsListFlags.Offset, campaignsListFlags.Limit)
			if err != nil {
				return err
			}
			resp = page
		}

		if opts.Output == "json" {
			return printJSON(resp)
		}

		w := tableWriter()
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tDISPLAY\tSERVING\tDAILY_BUDGET")
		for _, c := range resp.Data {
			daily := ""
			if c.DailyBudget != nil {
				daily = c.DailyBudget.Amount + " " + c.DailyBudget.Currency
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n", c.ID, c.Name, c.Status, c.DisplayStatus, c.ServingStatus, daily)
		}
		return w.Flush()
	},
}
