package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ferdikt/appleads-cli/internal/appleads"
	"github.com/spf13/cobra"
)

type campaignHealthRow struct {
	CampaignID     int64    `json:"campaign_id"`
	CampaignName   string   `json:"campaign_name"`
	CampaignStatus string   `json:"campaign_status"`
	DisplayStatus  string   `json:"display_status"`
	ServingStatus  string   `json:"serving_status"`
	TodaySpend     string   `json:"today_spend"`
	Reasons        []string `json:"reasons,omitempty"`
	Action         string   `json:"action"`
}

var campaignsHealthFlags struct {
	OrgID    int64
	Date     string
	TimeZone string
}

func init() {
	campaignsCmd.AddCommand(campaignsHealthCmd)

	campaignsHealthCmd.Flags().Int64Var(&campaignsHealthFlags.OrgID, "org-id", 0, "Organization ID override")
	campaignsHealthCmd.Flags().StringVar(&campaignsHealthFlags.Date, "date", "", "Report date (YYYY-MM-DD, default: today UTC)")
	campaignsHealthCmd.Flags().StringVar(&campaignsHealthFlags.TimeZone, "time-zone", "UTC", "Report timezone")
	campaignsHealthCmd.Flags().StringVar(&campaignsHealthFlags.TimeZone, "timezone", "UTC", "Alias for --time-zone")
}

var campaignsHealthCmd = &cobra.Command{
	Use:   "health",
	Short: "Summarize campaign serving health, reasons, spend, and next action",
	Example: "  appleads campaigns health\n" +
		"  appleads campaigns health --date 2026-04-16 --output json",
	RunE: func(cmd *cobra.Command, args []string) error {
		reportDate := strings.TrimSpace(campaignsHealthFlags.Date)
		if reportDate == "" {
			reportDate = time.Now().UTC().Format("2006-01-02")
		}
		if _, err := parseReportDate(reportDate); err != nil {
			return fmt.Errorf("invalid --date: %w", err)
		}

		client, _, _, err := authedClient(context.Background(), campaignsHealthFlags.OrgID, true)
		if err != nil {
			return err
		}

		rows, err := fetchCampaignHealthRows(context.Background(), client, reportDate, strings.TrimSpace(campaignsHealthFlags.TimeZone))
		if err != nil {
			return err
		}

		if opts.Output == "json" {
			return printJSON(map[string]any{
				"date":      reportDate,
				"time_zone": campaignsHealthFlags.TimeZone,
				"rows":      rows,
			})
		}

		w := tableWriter()
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tDISPLAY\tSERVING\tTODAY_SPEND\tREASONS\tACTION")
		for _, row := range rows {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				row.CampaignID,
				row.CampaignName,
				row.CampaignStatus,
				row.DisplayStatus,
				row.ServingStatus,
				row.TodaySpend,
				strings.Join(row.Reasons, ","),
				row.Action,
			)
		}
		return w.Flush()
	},
}

func fetchCampaignHealthRows(ctx context.Context, client *appleads.Client, reportDate, timeZone string) ([]campaignHealthRow, error) {
	basePayload := map[string]any{
		"startTime": reportDate,
		"endTime":   reportDate,
		"timeZone":  timeZone,
		"selector": map[string]any{
			"orderBy": []any{
				map[string]any{
					"field":     "campaignId",
					"sortOrder": "ASCENDING",
				},
			},
			"pagination": map[string]any{
				"offset": 0,
				"limit":  1000,
			},
		},
		"returnRowTotals": true,
	}

	offset := 0
	rows := make([]campaignHealthRow, 0, 64)
	for {
		payload := cloneMap(basePayload)
		selector := cloneMap(basePayload["selector"].(map[string]any))
		pagination := cloneMap(selector["pagination"].(map[string]any))
		pagination["offset"] = offset
		selector["pagination"] = pagination
		payload["selector"] = selector

		var resp map[string]any
		if err := client.DoJSON(ctx, http.MethodPost, "/reports/campaigns", nil, payload, &resp); err != nil {
			return nil, err
		}

		pageRows := parseCampaignHealthPage(resp)
		rows = append(rows, pageRows...)

		paginationMap, _ := resp["pagination"].(map[string]any)
		if paginationMap == nil {
			break
		}
		startIndex, _ := asInt(paginationMap["startIndex"])
		itemsPerPage, _ := asInt(paginationMap["itemsPerPage"])
		totalResults, ok := asInt(paginationMap["totalResults"])
		if itemsPerPage <= 0 || !ok {
			break
		}
		next := startIndex + itemsPerPage
		if next >= totalResults {
			break
		}
		offset = next
	}

	return rows, nil
}

func parseCampaignHealthPage(resp map[string]any) []campaignHealthRow {
	dataMap, _ := resp["data"].(map[string]any)
	reportingResponse, _ := dataMap["reportingDataResponse"].(map[string]any)
	rowItems, _ := reportingResponse["row"].([]any)
	out := make([]campaignHealthRow, 0, len(rowItems))
	for _, item := range rowItems {
		rowMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		metadata, _ := rowMap["metadata"].(map[string]any)
		total, _ := rowMap["total"].(map[string]any)
		campaignID, _ := anyToInt64(metadata["campaignId"])
		reasons := anyToStringSlice(metadata["servingStateReasons"])
		out = append(out, campaignHealthRow{
			CampaignID:     campaignID,
			CampaignName:   strings.TrimSpace(fmt.Sprint(metadata["campaignName"])),
			CampaignStatus: strings.TrimSpace(fmt.Sprint(metadata["campaignStatus"])),
			DisplayStatus:  strings.TrimSpace(fmt.Sprint(metadata["displayStatus"])),
			ServingStatus:  strings.TrimSpace(fmt.Sprint(metadata["servingStatus"])),
			TodaySpend:     formatMoneyValue(total["localSpend"]),
			Reasons:        reasons,
			Action:         recommendCampaignAction(reasons, metadata),
		})
	}
	return out
}

func recommendCampaignAction(reasons []string, metadata map[string]any) string {
	if len(reasons) == 0 && strings.EqualFold(fmt.Sprint(metadata["servingStatus"]), "RUNNING") {
		return "No action needed"
	}
	for _, reason := range reasons {
		switch reason {
		case "NO_AVAILABLE_AD_GROUPS", "AD_GROUP_MISSING":
			return "Create or enable at least one ad group"
		case "PAUSED_BY_USER":
			return "Enable the campaign if it should serve"
		case "CAMPAIGN_END_DATE_REACHED":
			return "Extend the campaign end date or duplicate into a new flight"
		case "CAMPAIGN_START_DATE_IN_FUTURE":
			return "Wait for the start date or move it earlier"
		case "DAILY_CAP_EXHAUSTED":
			return "Raise the daily budget or wait for the next reset"
		case "TOTAL_BUDGET_EXHAUSTED":
			return "Increase the total budget"
		case "BO_EXHAUSTED", "BO_END_DATE_REACHED":
			return "Fund or renew the budget order"
		case "NO_ELIGIBLE_COUNTRIES":
			return "Check storefront targeting and App Store availability"
		case "CREDIT_CARD_DECLINED", "NO_PAYMENT_METHOD_ON_FILE", "ORG_CHARGE_BACK_DISPUTED", "ORG_PAYMENT_TYPE_CHANGED":
			return "Fix billing on the Apple Ads account"
		case "APP_NOT_PUBLISHED_YET", "APP_NOT_ELIGIBLE", "APP_NOT_ELIGIBLE_SEARCHADS":
			return "Verify app eligibility and storefront availability"
		}
	}
	if strings.EqualFold(fmt.Sprint(metadata["displayStatus"]), "ON_HOLD") {
		return "Review serving reasons and ad group readiness"
	}
	return "Inspect campaign settings and serving reasons"
}

func anyToStringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		text := strings.TrimSpace(fmt.Sprint(item))
		if text == "" || text == "<nil>" {
			continue
		}
		out = append(out, text)
	}
	return out
}

func anyToInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err == nil {
			return n, true
		}
	}
	return 0, false
}

func formatMoneyValue(value any) string {
	money, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	amount := strings.TrimSpace(fmt.Sprint(money["amount"]))
	currency := strings.TrimSpace(fmt.Sprint(money["currency"]))
	if amount == "" || amount == "<nil>" {
		amount = "0"
	}
	if currency == "" || currency == "<nil>" {
		return amount
	}
	return amount + " " + currency
}
