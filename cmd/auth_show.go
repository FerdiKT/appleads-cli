package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	authCmd.AddCommand(authShowCmd)
}

var authShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show active profile configuration (token redacted)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, profile, err := loadProfile()
		if err != nil {
			return err
		}

		exp := ""
		if profile.TokenExpiresAt != nil {
			exp = profile.TokenExpiresAt.UTC().Format(time.RFC3339)
		}

		out := map[string]any{
			"profile":          opts.Profile,
			"active_profile":   cfg.EffectiveProfileName(opts.Profile),
			"client_id":        profile.ClientID,
			"team_id":          profile.TeamID,
			"key_id":           profile.KeyID,
			"org_id":           profile.OrgID,
			"private_key_path": profile.PrivateKeyPath,
			"api_version":      profile.EffectiveAPIVersion(),
			"api_base_url":     profile.EffectiveAPIBaseURL(),
			"auth_url":         profile.EffectiveAuthURL(),
			"token_cached":     profile.AccessToken != "",
			"token_expires_at": exp,
		}

		if opts.Output == "json" {
			return printJSON(out)
		}

		fmt.Printf("profile: %s\n", opts.Profile)
		fmt.Printf("active_profile: %s\n", cfg.EffectiveProfileName(opts.Profile))
		fmt.Printf("client_id: %s\n", profile.ClientID)
		fmt.Printf("team_id: %s\n", profile.TeamID)
		fmt.Printf("key_id: %s\n", profile.KeyID)
		fmt.Printf("org_id: %d\n", profile.OrgID)
		fmt.Printf("private_key_path: %s\n", profile.PrivateKeyPath)
		fmt.Printf("api: %s/api/%s\n", profile.EffectiveAPIBaseURL(), profile.EffectiveAPIVersion())
		fmt.Printf("auth_url: %s\n", profile.EffectiveAuthURL())
		fmt.Printf("token_cached: %t\n", profile.AccessToken != "")
		fmt.Printf("token_expires_at: %s\n", exp)
		return nil
	},
}
