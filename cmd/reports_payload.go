package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func buildReportPayload(cmd *cobra.Command, entity string, flags reportCallFlags) (any, error) {
	rawPayload, err := readJSONPayload(flags.Body, flags.BodyFile, true)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{}
	if rawPayload != nil {
		existing, ok := rawPayload.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("report payload must be a JSON object")
		}
		payload = cloneMap(existing)
	}

	startDate := strings.TrimSpace(flags.StartDate)
	endDate := strings.TrimSpace(flags.EndDate)
	if startDate != "" {
		if _, err := parseReportDate(startDate); err != nil {
			return nil, fmt.Errorf("invalid --start-date: %w", err)
		}
	}
	if endDate != "" {
		if _, err := parseReportDate(endDate); err != nil {
			return nil, fmt.Errorf("invalid --end-date: %w", err)
		}
	}
	if startDate != "" && endDate == "" {
		endDate = startDate
	}
	if endDate != "" && startDate == "" {
		startDate = endDate
	}

	dateFlagsChanged := cmd.Flags().Changed("start-date") || cmd.Flags().Changed("end-date")
	timeZoneChanged := cmd.Flags().Changed("time-zone") || cmd.Flags().Changed("timezone")

	if dateFlagsChanged && startDate == "" && endDate == "" {
		return nil, fmt.Errorf("both --start-date and --end-date are empty")
	}
	if dateFlagsChanged || startDate != "" {
		payload["startTime"] = startDate
		payload["endTime"] = endDate
	}

	if cmd.Flags().Changed("granularity") {
		granularity := strings.ToUpper(strings.TrimSpace(flags.Granularity))
		switch granularity {
		case "HOURLY", "DAILY", "WEEKLY":
			payload["granularity"] = granularity
		case "":
			delete(payload, "granularity")
		default:
			return nil, fmt.Errorf("invalid --granularity %q", flags.Granularity)
		}
	}

	if cmd.Flags().Changed("row-totals") {
		payload["returnRowTotals"] = flags.ReturnRowTotals
	}

	if timeZoneChanged {
		timeZone := strings.TrimSpace(flags.TimeZone)
		if timeZone == "" {
			delete(payload, "timeZone")
		} else {
			payload["timeZone"] = timeZone
		}
	}

	if rawPayload == nil && startDate == "" && endDate == "" {
		return nil, fmt.Errorf("report request is required (pass --start-date/--end-date or --body/--body-file)")
	}

	selector := map[string]any{}
	if existingSelector, ok := payload["selector"].(map[string]any); ok {
		selector = cloneMap(existingSelector)
	}
	if _, ok := selector["pagination"].(map[string]any); !ok {
		selector["pagination"] = map[string]any{
			"offset": 0,
			"limit":  1000,
		}
	} else {
		pagination := cloneMap(selector["pagination"].(map[string]any))
		if _, ok := pagination["offset"]; !ok {
			pagination["offset"] = 0
		}
		if _, ok := pagination["limit"]; !ok {
			pagination["limit"] = 1000
		}
		selector["pagination"] = pagination
	}
	if _, ok := selector["orderBy"]; !ok {
		field, err := defaultReportOrderField(entity)
		if err != nil {
			return nil, err
		}
		selector["orderBy"] = []any{
			map[string]any{
				"field":     field,
				"sortOrder": "ASCENDING",
			},
		}
	}
	payload["selector"] = selector

	if _, hasTimeZone := payload["timeZone"]; !hasTimeZone {
		payload["timeZone"] = defaultReportTimeZone(entity)
	}
	if entity == "searchterms" && strings.TrimSpace(fmt.Sprint(payload["timeZone"])) != "ORTZ" {
		return nil, fmt.Errorf("searchterms reports require --time-zone ORTZ")
	}
	if _, hasGranularity := payload["granularity"]; !hasGranularity {
		if _, hasRowTotals := payload["returnRowTotals"]; !hasRowTotals {
			payload["returnRowTotals"] = true
		}
	}

	return payload, nil
}

func defaultReportOrderField(entity string) (string, error) {
	switch entity {
	case "campaigns":
		return "campaignId", nil
	case "adgroups":
		return "adGroupId", nil
	case "keywords":
		return "keywordId", nil
	case "searchterms":
		return "localSpend", nil
	case "ads":
		return "adId", nil
	case "impressionshare":
		return "impressions", nil
	default:
		return "", fmt.Errorf("unsupported report entity %q", entity)
	}
}

func defaultReportTimeZone(entity string) string {
	if entity == "searchterms" {
		return "ORTZ"
	}
	return "UTC"
}

func parseReportDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(value))
}
