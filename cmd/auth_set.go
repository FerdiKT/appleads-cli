package cmd

import (
	"fmt"
	"strings"

	"github.com/ferdikt/appleads-cli/internal/config"
	"github.com/spf13/cobra"
)

var authSetFlags struct {
	ClientID       string
	TeamID         string
	KeyID          string
	OrgID          int64
	PrivateKeyPath string
	APIVersion     string
	APIBaseURL     string
	AuthURL        string
}

func init() {
	authCmd.AddCommand(authSetCmd)

	authSetCmd.Flags().StringVar(&authSetFlags.ClientID, "client-id", "", "OAuth client ID")
	authSetCmd.Flags().StringVar(&authSetFlags.TeamID, "team-id", "", "Apple team ID")
	authSetCmd.Flags().StringVar(&authSetFlags.KeyID, "key-id", "", "Private key ID (kid)")
	authSetCmd.Flags().Int64Var(&authSetFlags.OrgID, "org-id", 0, "Apple Ads organization ID")
	authSetCmd.Flags().StringVar(&authSetFlags.PrivateKeyPath, "private-key-path", "", "Path to .p8 private key")
	authSetCmd.Flags().StringVar(&authSetFlags.APIVersion, "api-version", "", "API version (default: v5)")
	authSetCmd.Flags().StringVar(&authSetFlags.APIBaseURL, "api-base-url", "", "API base URL (default: https://api.searchads.apple.com)")
	authSetCmd.Flags().StringVar(&authSetFlags.AuthURL, "auth-url", "", "OAuth token URL")
}

var authSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set profile credentials and API settings",
	Long: strings.TrimSpace(`
Save Apple Ads OAuth values in the selected profile.

The main values are:
  - client_id: shown by Apple after you upload a public key
  - team_id: shown by Apple with the client credentials
  - key_id: shown by Apple and used as the JWT kid
  - private_key_path: your local .p8 file that matches the uploaded public key
  - org_id: optional here, but required by most campaign/report commands
`),
	Example: strings.TrimSpace(`
  appleads auth set --client-id "SEARCHADS.abc" --team-id "SEARCHADS.def" --key-id "1234" --private-key-path ~/.config/appleads/default.p8
  appleads auth set --org-id 12345678
`),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		p := cfg.EnsureProfile(opts.Profile)

		if cmd.Flags().Changed("client-id") {
			p.ClientID = authSetFlags.ClientID
		}
		if cmd.Flags().Changed("team-id") {
			p.TeamID = authSetFlags.TeamID
		}
		if cmd.Flags().Changed("key-id") {
			p.KeyID = authSetFlags.KeyID
		}
		if cmd.Flags().Changed("org-id") {
			p.OrgID = authSetFlags.OrgID
		}
		if cmd.Flags().Changed("private-key-path") {
			p.PrivateKeyPath = authSetFlags.PrivateKeyPath
		}
		if cmd.Flags().Changed("api-version") {
			if authSetFlags.APIVersion != "" && !config.IsValidAPIVersion(authSetFlags.APIVersion) {
				return fmt.Errorf("invalid api version %q: expected format like v5", authSetFlags.APIVersion)
			}
			p.APIVersion = authSetFlags.APIVersion
		}
		if cmd.Flags().Changed("api-base-url") {
			p.APIBaseURL = authSetFlags.APIBaseURL
		}
		if cmd.Flags().Changed("auth-url") {
			p.AuthURL = authSetFlags.AuthURL
		}

		if err := cfg.Save(opts.ConfigPath); err != nil {
			return err
		}

		if opts.Output == "json" {
			return printJSON(map[string]any{
				"profile": opts.Profile,
				"config":  opts.ConfigPath,
				"saved":   true,
			})
		}

		fmt.Printf("profile %q saved to %s\n", opts.Profile, opts.ConfigPath)
		return nil
	},
}

func loadConfig() (*config.Config, error) {
	return config.Load(opts.ConfigPath)
}
