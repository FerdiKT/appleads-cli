package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(orgsCmd)
	orgsCmd.AddCommand(orgsListCmd)
	orgsCmd.AddCommand(orgsUseCmd)
}

var orgsCmd = &cobra.Command{
	Use:   "orgs",
	Short: "List and switch accessible organizations",
}

var orgsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List organizations visible to the current token",
	Example: strings.TrimSpace(`
  appleads orgs list
  appleads orgs list --json
`),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, profile, resp, apiVersionUsed, err := loadOrganizationsForCurrentProfile()
		if err != nil {
			return err
		}
		return renderOrganizationList(profile, resp, apiVersionUsed)
	},
}

var orgsUseCmd = &cobra.Command{
	Use:   "use <org-id>",
	Short: "Save one accessible org_id to the active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		orgID, err := parsePositiveInt64("org-id", args[0])
		if err != nil {
			return err
		}

		cfg, profile, resp, _, err := loadOrganizationsForCurrentProfile()
		if err != nil {
			return err
		}

		var selectedName string
		found := false
		for _, org := range resp.Data {
			if org.OrgID == orgID {
				selectedName = org.OrgName
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("org_id=%d is not visible to the current token", orgID)
		}

		if err := saveSelectedOrgID(cfg, profile, orgID); err != nil {
			return err
		}

		if opts.Output == "json" {
			return printJSON(map[string]any{
				"profile":  opts.Profile,
				"org_id":   orgID,
				"org_name": selectedName,
				"saved":    true,
			})
		}

		fmt.Printf("saved org_id=%d (%s) to profile %q\n", orgID, selectedName, opts.Profile)
		return nil
	},
}
