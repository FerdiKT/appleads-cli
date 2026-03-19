package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ferdikt/appleads-cli/internal/appleads"
	"github.com/spf13/cobra"
)

var authTokenFlags struct {
	Show   bool
	NoSave bool
}

func init() {
	authCmd.AddCommand(authTokenCmd)

	authTokenCmd.Flags().BoolVar(&authTokenFlags.Show, "show", false, "Print raw access token")
	authTokenCmd.Flags().BoolVar(&authTokenFlags.NoSave, "no-save", false, "Do not save token in profile")
}

var authTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Create an OAuth access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, profile, err := loadProfile()
		if err != nil {
			return err
		}

		privateKeyPEM, err := profile.ResolvePrivateKeyPEM()
		if err != nil {
			return err
		}

		clientSecret, err := appleads.BuildClientSecret(profile.TeamID, profile.ClientID, profile.KeyID, privateKeyPEM, time.Now().UTC())
		if err != nil {
			return err
		}

		tokenResp, err := appleads.RequestAccessToken(context.Background(), nil, profile.EffectiveAuthURL(), profile.ClientID, clientSecret)
		if err != nil {
			return err
		}

		expiresAt := time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		if !authTokenFlags.NoSave {
			profile.AccessToken = tokenResp.AccessToken
			profile.TokenExpiresAt = &expiresAt
			if err := cfg.Save(opts.ConfigPath); err != nil {
				return err
			}
		}

		if authTokenFlags.Show {
			fmt.Println(tokenResp.AccessToken)
			return nil
		}

		if opts.Output == "json" {
			return printJSON(map[string]any{
				"profile":    opts.Profile,
				"expires_at": expiresAt.Format(time.RFC3339),
				"saved":      !authTokenFlags.NoSave,
			})
		}

		if authTokenFlags.NoSave {
			fmt.Printf("token created (expires at %s)\n", expiresAt.Format(time.RFC3339))
		} else {
			fmt.Printf("token created and saved to profile %q (expires at %s)\n", opts.Profile, expiresAt.Format(time.RFC3339))
		}
		return nil
	},
}
