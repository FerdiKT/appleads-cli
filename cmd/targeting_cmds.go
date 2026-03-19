package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ferdikt/appleads-cli/internal/appleads"
	"github.com/spf13/cobra"
)

var targetingShowFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
}

var targetingSetFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Dimension  string
	Include    string
	Exclude    string
	Value      string
	DryRun     bool
	Yes        bool
}

var targetingClearFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Dimension  string
	DryRun     bool
	Yes        bool
}

var targetingReplaceFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Body       string
	BodyFile   string
	DryRun     bool
	Yes        bool
}

var targetingDimensionAliases = map[string]string{
	"country":         "country",
	"countries":       "country",
	"admin-area":      "adminArea",
	"adminarea":       "adminArea",
	"locality":        "locality",
	"age":             "age",
	"gender":          "gender",
	"device":          "deviceClass",
	"device-class":    "deviceClass",
	"deviceclass":     "deviceClass",
	"daypart":         "daypart",
	"app-categories":  "appCategories",
	"appcategories":   "appCategories",
	"app-downloaders": "appDownloaders",
	"appdownloaders":  "appDownloaders",
}

func init() {
	targetingCmd.AddCommand(targetingShowCmd)
	targetingCmd.AddCommand(targetingSetCmd)
	targetingCmd.AddCommand(targetingClearCmd)
	targetingCmd.AddCommand(targetingReplaceCmd)

	addTargetingIDsFlags(targetingShowCmd, &targetingShowFlags.OrgID, &targetingShowFlags.CampaignID, &targetingShowFlags.AdGroupID)

	addTargetingIDsFlags(targetingSetCmd, &targetingSetFlags.OrgID, &targetingSetFlags.CampaignID, &targetingSetFlags.AdGroupID)
	targetingSetCmd.Flags().StringVar(&targetingSetFlags.Dimension, "dimension", "", "Dimension key (country, admin-area, locality, age, gender, device-class, daypart, app-categories, app-downloaders)")
	targetingSetCmd.Flags().StringVar(&targetingSetFlags.Include, "include", "", "Comma separated included values")
	targetingSetCmd.Flags().StringVar(&targetingSetFlags.Exclude, "exclude", "", "Comma separated excluded values")
	targetingSetCmd.Flags().StringVar(&targetingSetFlags.Value, "value", "", "Raw JSON value for the dimension (overrides include/exclude)")
	targetingSetCmd.Flags().BoolVar(&targetingSetFlags.DryRun, "dry-run", false, "Print generated payload without sending update request")
	targetingSetCmd.Flags().BoolVar(&targetingSetFlags.Yes, "yes", false, "Skip interactive confirmation")
	_ = targetingSetCmd.MarkFlagRequired("dimension")

	addTargetingIDsFlags(targetingClearCmd, &targetingClearFlags.OrgID, &targetingClearFlags.CampaignID, &targetingClearFlags.AdGroupID)
	targetingClearCmd.Flags().StringVar(&targetingClearFlags.Dimension, "dimension", "", "Dimension key to clear")
	targetingClearCmd.Flags().BoolVar(&targetingClearFlags.DryRun, "dry-run", false, "Print generated payload without sending update request")
	targetingClearCmd.Flags().BoolVar(&targetingClearFlags.Yes, "yes", false, "Skip interactive confirmation")
	_ = targetingClearCmd.MarkFlagRequired("dimension")

	addTargetingIDsFlags(targetingReplaceCmd, &targetingReplaceFlags.OrgID, &targetingReplaceFlags.CampaignID, &targetingReplaceFlags.AdGroupID)
	targetingReplaceCmd.Flags().StringVar(&targetingReplaceFlags.Body, "body", "", "Inline JSON targeting payload")
	targetingReplaceCmd.Flags().StringVar(&targetingReplaceFlags.BodyFile, "body-file", "", "Path to JSON targeting payload file")
	targetingReplaceCmd.Flags().BoolVar(&targetingReplaceFlags.DryRun, "dry-run", false, "Print payload without sending update request")
	targetingReplaceCmd.Flags().BoolVar(&targetingReplaceFlags.Yes, "yes", false, "Skip interactive confirmation")
}

func addTargetingIDsFlags(cmd *cobra.Command, orgID, campaignID, adGroupID *int64) {
	cmd.Flags().Int64Var(orgID, "org-id", 0, "Organization ID override")
	cmd.Flags().Int64Var(campaignID, "campaign-id", 0, "Campaign ID")
	cmd.Flags().Int64Var(adGroupID, "adgroup-id", 0, "Ad group ID")
	_ = cmd.MarkFlagRequired("campaign-id")
	_ = cmd.MarkFlagRequired("adgroup-id")
}

var targetingShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show targeting dimensions for an ad group",
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetingShowFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if targetingShowFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		client, _, _, err := authedClient(context.Background(), targetingShowFlags.OrgID, true)
		if err != nil {
			return err
		}
		td, err := fetchTargetingDimensions(context.Background(), client, targetingShowFlags.CampaignID, targetingShowFlags.AdGroupID)
		if err != nil {
			return err
		}
		return printJSON(map[string]any{
			"campaignId":          targetingShowFlags.CampaignID,
			"adGroupId":           targetingShowFlags.AdGroupID,
			"targetingDimensions": td,
		})
	},
}

var targetingSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set one targeting dimension by merging into current ad-group targeting",
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetingSetFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if targetingSetFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		dimension, err := normalizeTargetingDimension(targetingSetFlags.Dimension)
		if err != nil {
			return err
		}

		client, _, _, err := authedClient(context.Background(), targetingSetFlags.OrgID, true)
		if err != nil {
			return err
		}
		current, err := fetchTargetingDimensions(context.Background(), client, targetingSetFlags.CampaignID, targetingSetFlags.AdGroupID)
		if err != nil {
			return err
		}

		next := cloneMap(current)
		if strings.TrimSpace(targetingSetFlags.Value) != "" {
			var parsed any
			if err := json.Unmarshal([]byte(targetingSetFlags.Value), &parsed); err != nil {
				return fmt.Errorf("parse --value JSON: %w", err)
			}
			next[dimension] = parsed
		} else {
			included := parseCSVValues(targetingSetFlags.Include)
			excluded := parseCSVValues(targetingSetFlags.Exclude)
			if len(included) == 0 && len(excluded) == 0 {
				return fmt.Errorf("set requires --value JSON or at least one of --include / --exclude")
			}
			value := map[string]any{}
			if len(included) > 0 {
				value["included"] = included
			}
			if len(excluded) > 0 {
				value["excluded"] = excluded
			}
			next[dimension] = value
		}

		payload := map[string]any{"targetingDimensions": next}
		if targetingSetFlags.DryRun {
			return printJSON(payload)
		}
		if err := confirmJSONPayload("targeting set", payload, targetingSetFlags.Yes); err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d", targetingSetFlags.CampaignID, targetingSetFlags.AdGroupID)
		return callAPIAndPrint(context.Background(), client, http.MethodPut, path, nil, payload)
	},
}

var targetingClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear one targeting dimension from ad-group targeting",
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetingClearFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if targetingClearFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		dimension, err := normalizeTargetingDimension(targetingClearFlags.Dimension)
		if err != nil {
			return err
		}

		client, _, _, err := authedClient(context.Background(), targetingClearFlags.OrgID, true)
		if err != nil {
			return err
		}
		current, err := fetchTargetingDimensions(context.Background(), client, targetingClearFlags.CampaignID, targetingClearFlags.AdGroupID)
		if err != nil {
			return err
		}

		next := cloneMap(current)
		delete(next, dimension)
		payload := map[string]any{"targetingDimensions": next}
		if targetingClearFlags.DryRun {
			return printJSON(payload)
		}
		if err := confirmJSONPayload("targeting clear", payload, targetingClearFlags.Yes); err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d", targetingClearFlags.CampaignID, targetingClearFlags.AdGroupID)
		return callAPIAndPrint(context.Background(), client, http.MethodPut, path, nil, payload)
	},
}

var targetingReplaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "Replace targeting dimensions with JSON payload",
	Example: "  appleads targeting replace --campaign-id 123456 --adgroup-id 987654 --body-file ./payloads/targeting-replace.json\n" +
		"  appleads targeting replace --campaign-id 123456 --adgroup-id 987654 --body '{\"country\":{\"included\":[\"US\"]}}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetingReplaceFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if targetingReplaceFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		payload, err := readJSONPayload(targetingReplaceFlags.Body, targetingReplaceFlags.BodyFile, false)
		if err != nil {
			return err
		}
		payloadMap, ok := payload.(map[string]any)
		if !ok {
			return fmt.Errorf("payload must be a JSON object")
		}
		if _, hasTD := payloadMap["targetingDimensions"]; !hasTD {
			payloadMap = map[string]any{"targetingDimensions": payloadMap}
		}

		if targetingReplaceFlags.DryRun {
			return printJSON(payloadMap)
		}
		if err := confirmJSONPayload("targeting replace", payloadMap, targetingReplaceFlags.Yes); err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), targetingReplaceFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d", targetingReplaceFlags.CampaignID, targetingReplaceFlags.AdGroupID)
		return callAPIAndPrint(context.Background(), client, http.MethodPut, path, nil, payloadMap)
	},
}

func fetchTargetingDimensions(ctx context.Context, client *appleads.Client, campaignID, adGroupID int64) (map[string]any, error) {
	path := fmt.Sprintf("/campaigns/%d/adgroups/%d", campaignID, adGroupID)
	var resp map[string]any
	if err := client.DoJSON(ctx, http.MethodGet, path, nil, nil, &resp); err != nil {
		return nil, err
	}
	obj := unwrapResponseData(resp)
	td, _ := obj["targetingDimensions"].(map[string]any)
	if td == nil {
		td = map[string]any{}
	}
	return td, nil
}

func unwrapResponseData(resp map[string]any) map[string]any {
	if dataObj, ok := resp["data"].(map[string]any); ok {
		return dataObj
	}
	return resp
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func normalizeTargetingDimension(input string) (string, error) {
	key := strings.TrimSpace(strings.ToLower(input))
	key = strings.ReplaceAll(key, "_", "-")
	if key == "" {
		return "", fmt.Errorf("--dimension is required")
	}
	if normalized, ok := targetingDimensionAliases[key]; ok {
		return normalized, nil
	}
	return "", fmt.Errorf("unsupported --dimension %q", input)
}

func parseCSVValues(raw string) []any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	values := make([]any, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		if n, err := strconv.ParseInt(token, 10, 64); err == nil {
			values = append(values, n)
			continue
		}
		values = append(values, token)
	}
	return values
}
