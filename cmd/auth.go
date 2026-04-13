package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(authCmd)
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Set up Apple Ads OAuth: keygen, upload public key, token, orgs",
	Long: strings.TrimSpace(`
Set up and manage Apple Ads OAuth access for the selected profile.

Recommended first-run flow:
  1. Make sure you have the Apple ID that will act as the API user for the target account.
  2. Run 'appleads auth init' to generate a local key pair and save profile values.
  3. Upload the generated PUBLIC KEY in Apple Ads > Account Settings > Client Credentials (or API).
  4. Copy the one-time values Apple shows: clientId, teamId, keyId.
  5. Run 'appleads auth token' to mint an access token.
  6. Run 'appleads auth orgs --select' to save the org_id used by campaign/report commands.

Notes:
  - OAuth does not remove Apple Ads account permissions; the user must still have the right access.
  - Apple's documented Campaign Management API flow uses a designated API user, usually a separate Apple ID invited via User Management.
  - For third-party OAuth authorizations (for example RevenueCat), Apple/RevenueCat require a user with sufficient Apple Ads permissions such as Account Admin or Campaign Group Manager.
  - org_id is not required to create a token, but most data commands need it.
`),
	Example: strings.TrimSpace(`
  appleads auth init
  appleads auth keygen
  appleads auth public-key
  appleads auth set --client-id "SEARCHADS..." --team-id "SEARCHADS..." --key-id "..."
  appleads auth token
  appleads auth orgs --select
`),
}
