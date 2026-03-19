package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ferdikt/appleads-cli/internal/appleads"
	"github.com/ferdikt/appleads-cli/internal/config"
	"github.com/spf13/cobra"
)

var authOrgsFlags struct {
	Select bool
}

func init() {
	authCmd.AddCommand(authOrgsCmd)
	authOrgsCmd.Flags().BoolVar(&authOrgsFlags.Select, "select", true, "Interactively select and save org_id from the list")
}

var authOrgsCmd = &cobra.Command{
	Use:   "orgs",
	Short: "Fetch accessible organizations and optionally select one",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, profile, err := loadProfile()
		if err != nil {
			return err
		}

		token, err := ensureAccessToken(context.Background(), cfg, profile)
		if err != nil {
			return err
		}

		resp, apiVersionUsed, err := fetchUserACLsWithFallback(profile, token)
		if err != nil {
			return err
		}
		if len(resp.Data) == 0 {
			return fmt.Errorf("no organizations found for this token")
		}

		if opts.Output == "json" {
			return printJSON(map[string]any{
				"profile":          opts.Profile,
				"api_version_used": apiVersionUsed,
				"current_org_id":   profile.OrgID,
				"orgs":             resp.Data,
			})
		}

		w := tableWriter()
		fmt.Fprintln(w, "#\tORG_ID\tORG_NAME\tPARENT_ORG_ID\tCURRENCY\tTIMEZONE\tROLES")
		for i, org := range resp.Data {
			roles := strings.Join(org.RoleNames, ",")
			fmt.Fprintf(w, "%d\t%d\t%s\t%d\t%s\t%s\t%s\n", i+1, org.OrgID, org.OrgName, org.ParentOrgID, org.Currency, org.TimeZone, roles)
		}
		_ = w.Flush()
		fmt.Printf("API version used for ACL lookup: %s\n", apiVersionUsed)

		if !authOrgsFlags.Select || !stdinIsTTY() {
			if profile.OrgID <= 0 {
				fmt.Println("org_id is not set. Re-run with --select to save one.")
			}
			return nil
		}

		selectedOrgID, err := promptOrgSelection(resp.Data, profile.OrgID)
		if err != nil {
			return err
		}
		if selectedOrgID == 0 {
			fmt.Println("selection skipped.")
			return nil
		}

		profile.OrgID = selectedOrgID
		if err := cfg.Save(opts.ConfigPath); err != nil {
			return err
		}

		fmt.Printf("saved org_id=%d to profile %q\n", selectedOrgID, opts.Profile)
		return nil
	},
}

func fetchUserACLsWithFallback(profile *config.Profile, token string) (*appleads.UserACLListResponse, string, error) {
	versions := []string{strings.Trim(profile.EffectiveAPIVersion(), "/")}
	if versions[0] != "v4" {
		versions = append(versions, "v4")
	}

	var lastErr error
	for _, version := range versions {
		client := &appleads.Client{
			BaseURL: apiBaseURLForVersion(profile, version),
			Token:   token,
		}
		resp, err := client.ListUserACLs(context.Background())
		if err == nil {
			return resp, version, nil
		}
		lastErr = err
	}

	return nil, "", fmt.Errorf("fetch organizations failed: %w", lastErr)
}

func promptOrgSelection(orgs []appleads.UserACL, currentOrgID int64) (int64, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		if currentOrgID > 0 {
			fmt.Printf("Select org number to save (Enter to keep current org_id=%d): ", currentOrgID)
		} else {
			fmt.Print("Select org number to save (Enter to skip): ")
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			return 0, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			return 0, nil
		}

		index, err := strconv.Atoi(line)
		if err != nil || index < 1 || index > len(orgs) {
			fmt.Printf("invalid choice. Enter 1..%d\n", len(orgs))
			continue
		}

		return orgs[index-1].OrgID, nil
	}
}

func stdinIsTTY() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
