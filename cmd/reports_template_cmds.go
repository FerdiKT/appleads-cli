package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var reportsTemplateFlags struct {
	OrgID       int64
	CampaignID  int64
	Preset      string
	Granularity string
	TimeZone    string
	Run         bool
}

func init() {
	reportsCmd.AddCommand(reportsTemplateCmd)

	reportsTemplateCmd.Flags().Int64Var(&reportsTemplateFlags.OrgID, "org-id", 0, "Organization ID override")
	reportsTemplateCmd.Flags().Int64Var(&reportsTemplateFlags.CampaignID, "campaign-id", 0, "Campaign ID (required for adgroups/keywords/searchterms/ads/impressionshare)")
	reportsTemplateCmd.Flags().StringVar(&reportsTemplateFlags.Preset, "preset", "last-7d", "Date preset (today|yesterday|last-7d|last-30d)")
	reportsTemplateCmd.Flags().StringVar(&reportsTemplateFlags.Granularity, "granularity", "DAILY", "Report granularity (DAILY|HOURLY|WEEKLY)")
	reportsTemplateCmd.Flags().StringVar(&reportsTemplateFlags.TimeZone, "time-zone", "UTC", "IANA timezone or UTC")
	reportsTemplateCmd.Flags().BoolVar(&reportsTemplateFlags.Run, "run", false, "Call API immediately with generated payload")
}

var reportsTemplateCmd = &cobra.Command{
	Use:   "template <entity>",
	Short: "Generate a ready report payload (and optionally run it)",
	Long:  "Entity: campaigns|adgroups|keywords|searchterms|ads|impressionshare",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entity := strings.ToLower(strings.TrimSpace(args[0]))
		path, needsCampaign, err := reportEntityPath(entity, reportsTemplateFlags.CampaignID)
		if err != nil {
			return err
		}
		if needsCampaign && reportsTemplateFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0 for entity %q", entity)
		}

		payload, err := buildReportTemplatePayload(reportsTemplateFlags.Preset, reportsTemplateFlags.Granularity, reportsTemplateFlags.TimeZone)
		if err != nil {
			return err
		}

		if !reportsTemplateFlags.Run {
			return printJSON(map[string]any{
				"entity":  entity,
				"path":    path,
				"payload": payload,
			})
		}

		client, _, _, err := authedClient(context.Background(), reportsTemplateFlags.OrgID, true)
		if err != nil {
			return err
		}
		if entity == "impressionshare" {
			paths := []string{
				path,
				strings.ReplaceAll(path, "impressionshare", "impressionShare"),
				strings.ReplaceAll(path, "impressionshare", "impression-share"),
				strings.ReplaceAll(path, "impressionshare", "impression_share"),
			}
			return callAPIAndPrintWithFallback(context.Background(), client, http.MethodPost, paths, nil, payload)
		}
		return callAPIAndPrint(context.Background(), client, http.MethodPost, path, nil, payload)
	},
}

func reportEntityPath(entity string, campaignID int64) (path string, needsCampaign bool, err error) {
	switch entity {
	case "campaigns":
		return "/reports/campaigns", false, nil
	case "adgroups":
		return fmt.Sprintf("/reports/campaigns/%d/adgroups", campaignID), true, nil
	case "keywords":
		return fmt.Sprintf("/reports/campaigns/%d/keywords", campaignID), true, nil
	case "searchterms":
		return fmt.Sprintf("/reports/campaigns/%d/searchterms", campaignID), true, nil
	case "ads":
		return fmt.Sprintf("/reports/campaigns/%d/ads", campaignID), true, nil
	case "impressionshare":
		return fmt.Sprintf("/reports/campaigns/%d/impressionshare", campaignID), true, nil
	default:
		return "", false, fmt.Errorf("unsupported entity %q", entity)
	}
}

func buildReportTemplatePayload(preset, granularity, timeZone string) (map[string]any, error) {
	today := time.Now().UTC()
	start, end, err := resolveDatePreset(strings.ToLower(strings.TrimSpace(preset)), today)
	if err != nil {
		return nil, err
	}
	granularity = strings.ToUpper(strings.TrimSpace(granularity))
	if granularity == "" {
		granularity = "DAILY"
	}
	switch granularity {
	case "HOURLY", "DAILY", "WEEKLY":
	default:
		return nil, fmt.Errorf("invalid --granularity %q", granularity)
	}
	if strings.TrimSpace(timeZone) == "" {
		timeZone = "UTC"
	}

	return map[string]any{
		"startTime":   start.Format("2006-01-02"),
		"endTime":     end.Format("2006-01-02"),
		"timeZone":    timeZone,
		"granularity": granularity,
		"selector": map[string]any{
			"pagination": map[string]any{
				"offset": 0,
				"limit":  1000,
			},
		},
	}, nil
}

func resolveDatePreset(preset string, now time.Time) (start, end time.Time, err error) {
	base := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	switch preset {
	case "today":
		return base, base, nil
	case "yesterday":
		y := base.AddDate(0, 0, -1)
		return y, y, nil
	case "last-7d":
		return base.AddDate(0, 0, -7), base.AddDate(0, 0, -1), nil
	case "last-30d":
		return base.AddDate(0, 0, -30), base.AddDate(0, 0, -1), nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid --preset %q", preset)
	}
}
