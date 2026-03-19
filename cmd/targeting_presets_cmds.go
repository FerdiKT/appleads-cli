package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var targetingCountryCmd = &cobra.Command{
	Use:   "country",
	Short: "Country targeting presets",
}

var targetingDeviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Device targeting presets",
}

var targetingCountryAddFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Codes      string
	DryRun     bool
	Yes        bool
}

var targetingCountryRemoveFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Codes      string
	DryRun     bool
	Yes        bool
}

var targetingCountryOnlyFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Codes      string
	DryRun     bool
	Yes        bool
}

var targetingDeviceSetFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Classes    string
	DryRun     bool
	Yes        bool
}

var targetingDeviceOnlyIPhoneFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	DryRun     bool
	Yes        bool
}

var targetingDeviceOnlyIPadFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	DryRun     bool
	Yes        bool
}

var targetingDeviceOnlyBothFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	DryRun     bool
	Yes        bool
}

func init() {
	targetingCmd.AddCommand(targetingCountryCmd)
	targetingCmd.AddCommand(targetingDeviceCmd)

	targetingCountryCmd.AddCommand(targetingCountryAddCmd)
	targetingCountryCmd.AddCommand(targetingCountryRemoveCmd)
	targetingCountryCmd.AddCommand(targetingCountryOnlyCmd)

	addTargetingIDsFlags(targetingCountryAddCmd, &targetingCountryAddFlags.OrgID, &targetingCountryAddFlags.CampaignID, &targetingCountryAddFlags.AdGroupID)
	targetingCountryAddCmd.Flags().StringVar(&targetingCountryAddFlags.Codes, "codes", "", "Comma separated country codes (e.g. US,CA,DE)")
	targetingCountryAddCmd.Flags().BoolVar(&targetingCountryAddFlags.DryRun, "dry-run", false, "Print payload without sending update request")
	targetingCountryAddCmd.Flags().BoolVar(&targetingCountryAddFlags.Yes, "yes", false, "Skip interactive confirmation")
	_ = targetingCountryAddCmd.MarkFlagRequired("codes")

	addTargetingIDsFlags(targetingCountryRemoveCmd, &targetingCountryRemoveFlags.OrgID, &targetingCountryRemoveFlags.CampaignID, &targetingCountryRemoveFlags.AdGroupID)
	targetingCountryRemoveCmd.Flags().StringVar(&targetingCountryRemoveFlags.Codes, "codes", "", "Comma separated country codes to remove from included list")
	targetingCountryRemoveCmd.Flags().BoolVar(&targetingCountryRemoveFlags.DryRun, "dry-run", false, "Print payload without sending update request")
	targetingCountryRemoveCmd.Flags().BoolVar(&targetingCountryRemoveFlags.Yes, "yes", false, "Skip interactive confirmation")
	_ = targetingCountryRemoveCmd.MarkFlagRequired("codes")

	addTargetingIDsFlags(targetingCountryOnlyCmd, &targetingCountryOnlyFlags.OrgID, &targetingCountryOnlyFlags.CampaignID, &targetingCountryOnlyFlags.AdGroupID)
	targetingCountryOnlyCmd.Flags().StringVar(&targetingCountryOnlyFlags.Codes, "codes", "", "Comma separated country codes to keep as the only included countries")
	targetingCountryOnlyCmd.Flags().BoolVar(&targetingCountryOnlyFlags.DryRun, "dry-run", false, "Print payload without sending update request")
	targetingCountryOnlyCmd.Flags().BoolVar(&targetingCountryOnlyFlags.Yes, "yes", false, "Skip interactive confirmation")
	_ = targetingCountryOnlyCmd.MarkFlagRequired("codes")

	targetingDeviceCmd.AddCommand(targetingDeviceSetCmd)
	targetingDeviceCmd.AddCommand(targetingDeviceOnlyIPhoneCmd)
	targetingDeviceCmd.AddCommand(targetingDeviceOnlyIPadCmd)
	targetingDeviceCmd.AddCommand(targetingDeviceOnlyBothCmd)

	addTargetingIDsFlags(targetingDeviceSetCmd, &targetingDeviceSetFlags.OrgID, &targetingDeviceSetFlags.CampaignID, &targetingDeviceSetFlags.AdGroupID)
	targetingDeviceSetCmd.Flags().StringVar(&targetingDeviceSetFlags.Classes, "classes", "", "Comma separated device classes (IPHONE,IPAD)")
	targetingDeviceSetCmd.Flags().BoolVar(&targetingDeviceSetFlags.DryRun, "dry-run", false, "Print payload without sending update request")
	targetingDeviceSetCmd.Flags().BoolVar(&targetingDeviceSetFlags.Yes, "yes", false, "Skip interactive confirmation")
	_ = targetingDeviceSetCmd.MarkFlagRequired("classes")

	addTargetingIDsFlags(targetingDeviceOnlyIPhoneCmd, &targetingDeviceOnlyIPhoneFlags.OrgID, &targetingDeviceOnlyIPhoneFlags.CampaignID, &targetingDeviceOnlyIPhoneFlags.AdGroupID)
	targetingDeviceOnlyIPhoneCmd.Flags().BoolVar(&targetingDeviceOnlyIPhoneFlags.DryRun, "dry-run", false, "Print payload without sending update request")
	targetingDeviceOnlyIPhoneCmd.Flags().BoolVar(&targetingDeviceOnlyIPhoneFlags.Yes, "yes", false, "Skip interactive confirmation")

	addTargetingIDsFlags(targetingDeviceOnlyIPadCmd, &targetingDeviceOnlyIPadFlags.OrgID, &targetingDeviceOnlyIPadFlags.CampaignID, &targetingDeviceOnlyIPadFlags.AdGroupID)
	targetingDeviceOnlyIPadCmd.Flags().BoolVar(&targetingDeviceOnlyIPadFlags.DryRun, "dry-run", false, "Print payload without sending update request")
	targetingDeviceOnlyIPadCmd.Flags().BoolVar(&targetingDeviceOnlyIPadFlags.Yes, "yes", false, "Skip interactive confirmation")

	addTargetingIDsFlags(targetingDeviceOnlyBothCmd, &targetingDeviceOnlyBothFlags.OrgID, &targetingDeviceOnlyBothFlags.CampaignID, &targetingDeviceOnlyBothFlags.AdGroupID)
	targetingDeviceOnlyBothCmd.Flags().BoolVar(&targetingDeviceOnlyBothFlags.DryRun, "dry-run", false, "Print payload without sending update request")
	targetingDeviceOnlyBothCmd.Flags().BoolVar(&targetingDeviceOnlyBothFlags.Yes, "yes", false, "Skip interactive confirmation")
}

var targetingCountryAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add countries to the included list",
	RunE: func(cmd *cobra.Command, args []string) error {
		codes, err := parseCountryCodes(targetingCountryAddFlags.Codes)
		if err != nil {
			return err
		}
		return runTargetingPresetMutation(targetingCountryAddFlags.OrgID, targetingCountryAddFlags.CampaignID, targetingCountryAddFlags.AdGroupID, targetingCountryAddFlags.DryRun, targetingCountryAddFlags.Yes, "country add", func(next map[string]any) error {
			included, excluded := readIncludedExcludedStringSets(next, "country")
			for _, code := range codes {
				included[code] = struct{}{}
				delete(excluded, code)
			}
			writeIncludedExcludedDimension(next, "country", included, excluded)
			return nil
		})
	},
}

var targetingCountryRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove countries from the included list",
	RunE: func(cmd *cobra.Command, args []string) error {
		codes, err := parseCountryCodes(targetingCountryRemoveFlags.Codes)
		if err != nil {
			return err
		}
		return runTargetingPresetMutation(targetingCountryRemoveFlags.OrgID, targetingCountryRemoveFlags.CampaignID, targetingCountryRemoveFlags.AdGroupID, targetingCountryRemoveFlags.DryRun, targetingCountryRemoveFlags.Yes, "country remove", func(next map[string]any) error {
			included, excluded := readIncludedExcludedStringSets(next, "country")
			for _, code := range codes {
				delete(included, code)
			}
			writeIncludedExcludedDimension(next, "country", included, excluded)
			return nil
		})
	},
}

var targetingCountryOnlyCmd = &cobra.Command{
	Use:   "only",
	Short: "Set included countries to exactly the provided list",
	RunE: func(cmd *cobra.Command, args []string) error {
		codes, err := parseCountryCodes(targetingCountryOnlyFlags.Codes)
		if err != nil {
			return err
		}
		return runTargetingPresetMutation(targetingCountryOnlyFlags.OrgID, targetingCountryOnlyFlags.CampaignID, targetingCountryOnlyFlags.AdGroupID, targetingCountryOnlyFlags.DryRun, targetingCountryOnlyFlags.Yes, "country only", func(next map[string]any) error {
			included := make(map[string]struct{}, len(codes))
			for _, code := range codes {
				included[code] = struct{}{}
			}
			writeIncludedExcludedDimension(next, "country", included, map[string]struct{}{})
			return nil
		})
	},
}

var targetingDeviceSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set included device classes",
	RunE: func(cmd *cobra.Command, args []string) error {
		classes, err := parseDeviceClasses(targetingDeviceSetFlags.Classes)
		if err != nil {
			return err
		}
		return runTargetingPresetMutation(targetingDeviceSetFlags.OrgID, targetingDeviceSetFlags.CampaignID, targetingDeviceSetFlags.AdGroupID, targetingDeviceSetFlags.DryRun, targetingDeviceSetFlags.Yes, "device set", func(next map[string]any) error {
			next["deviceClass"] = map[string]any{
				"included": toAnySlice(classes),
			}
			return nil
		})
	},
}

var targetingDeviceOnlyIPhoneCmd = &cobra.Command{
	Use:   "only-iphone",
	Short: "Shortcut preset: include only IPHONE",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTargetingPresetMutation(targetingDeviceOnlyIPhoneFlags.OrgID, targetingDeviceOnlyIPhoneFlags.CampaignID, targetingDeviceOnlyIPhoneFlags.AdGroupID, targetingDeviceOnlyIPhoneFlags.DryRun, targetingDeviceOnlyIPhoneFlags.Yes, "device only-iphone", func(next map[string]any) error {
			next["deviceClass"] = map[string]any{
				"included": []any{"IPHONE"},
			}
			return nil
		})
	},
}

var targetingDeviceOnlyIPadCmd = &cobra.Command{
	Use:   "only-ipad",
	Short: "Shortcut preset: include only IPAD",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTargetingPresetMutation(targetingDeviceOnlyIPadFlags.OrgID, targetingDeviceOnlyIPadFlags.CampaignID, targetingDeviceOnlyIPadFlags.AdGroupID, targetingDeviceOnlyIPadFlags.DryRun, targetingDeviceOnlyIPadFlags.Yes, "device only-ipad", func(next map[string]any) error {
			next["deviceClass"] = map[string]any{
				"included": []any{"IPAD"},
			}
			return nil
		})
	},
}

var targetingDeviceOnlyBothCmd = &cobra.Command{
	Use:   "only-both",
	Short: "Shortcut preset: include IPHONE and IPAD",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTargetingPresetMutation(targetingDeviceOnlyBothFlags.OrgID, targetingDeviceOnlyBothFlags.CampaignID, targetingDeviceOnlyBothFlags.AdGroupID, targetingDeviceOnlyBothFlags.DryRun, targetingDeviceOnlyBothFlags.Yes, "device only-both", func(next map[string]any) error {
			next["deviceClass"] = map[string]any{
				"included": []any{"IPHONE", "IPAD"},
			}
			return nil
		})
	},
}

func runTargetingPresetMutation(orgID, campaignID, adGroupID int64, dryRun, yes bool, title string, mutate func(next map[string]any) error) error {
	if campaignID <= 0 {
		return fmt.Errorf("--campaign-id must be > 0")
	}
	if adGroupID <= 0 {
		return fmt.Errorf("--adgroup-id must be > 0")
	}
	client, _, _, err := authedClient(context.Background(), orgID, true)
	if err != nil {
		return err
	}
	current, err := fetchTargetingDimensions(context.Background(), client, campaignID, adGroupID)
	if err != nil {
		return err
	}
	next := cloneMap(current)
	if err := mutate(next); err != nil {
		return err
	}
	payload := map[string]any{"targetingDimensions": next}
	if dryRun {
		return printJSON(payload)
	}
	if err := confirmJSONPayload(title, payload, yes); err != nil {
		return err
	}
	path := fmt.Sprintf("/campaigns/%d/adgroups/%d", campaignID, adGroupID)
	return callAPIAndPrint(context.Background(), client, "PUT", path, nil, payload)
}

func parseCountryCodes(raw string) ([]string, error) {
	values := strings.Split(raw, ",")
	seen := map[string]struct{}{}
	codes := make([]string, 0, len(values))
	for _, item := range values {
		code := strings.ToUpper(strings.TrimSpace(item))
		if code == "" {
			continue
		}
		if len(code) < 2 || len(code) > 3 {
			return nil, fmt.Errorf("invalid country code %q", code)
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	if len(codes) == 0 {
		return nil, fmt.Errorf("--codes must include at least one country code")
	}
	return codes, nil
}

func parseDeviceClasses(raw string) ([]string, error) {
	values := strings.Split(raw, ",")
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, item := range values {
		class := strings.ToUpper(strings.TrimSpace(item))
		if class == "" {
			continue
		}
		switch class {
		case "IPHONE", "IPAD":
		default:
			return nil, fmt.Errorf("invalid device class %q (allowed: IPHONE, IPAD)", class)
		}
		if _, ok := seen[class]; ok {
			continue
		}
		seen[class] = struct{}{}
		out = append(out, class)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("--classes must include at least one of IPHONE, IPAD")
	}
	return out, nil
}

func readIncludedExcludedStringSets(targetingDimensions map[string]any, key string) (map[string]struct{}, map[string]struct{}) {
	included := map[string]struct{}{}
	excluded := map[string]struct{}{}
	dimValue, _ := targetingDimensions[key].(map[string]any)
	if dimValue == nil {
		return included, excluded
	}
	for _, item := range toStringSlice(dimValue["included"]) {
		included[strings.ToUpper(strings.TrimSpace(item))] = struct{}{}
	}
	for _, item := range toStringSlice(dimValue["excluded"]) {
		excluded[strings.ToUpper(strings.TrimSpace(item))] = struct{}{}
	}
	return included, excluded
}

func writeIncludedExcludedDimension(targetingDimensions map[string]any, key string, includedSet, excludedSet map[string]struct{}) {
	included := sortedSetKeys(includedSet)
	excluded := sortedSetKeys(excludedSet)
	if len(included) == 0 && len(excluded) == 0 {
		delete(targetingDimensions, key)
		return
	}
	value := map[string]any{}
	if len(included) > 0 {
		value["included"] = toAnySlice(included)
	}
	if len(excluded) > 0 {
		value["excluded"] = toAnySlice(excluded)
	}
	targetingDimensions[key] = value
}

func toStringSlice(v any) []string {
	switch x := v.(type) {
	case []string:
		return x
	case []any:
		out := make([]string, 0, len(x))
		for _, item := range x {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func sortedSetKeys(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for item := range set {
		if strings.TrimSpace(item) == "" {
			continue
		}
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func toAnySlice(items []string) []any {
	out := make([]any, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}
