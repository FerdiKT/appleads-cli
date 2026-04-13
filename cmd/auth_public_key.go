package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ferdikt/appleads-cli/internal/keys"
	"github.com/spf13/cobra"
)

var authPublicKeyFlags struct {
	PrivateKeyPath string
}

func init() {
	authCmd.AddCommand(authPublicKeyCmd)
	authPublicKeyCmd.Flags().StringVar(&authPublicKeyFlags.PrivateKeyPath, "private-key-path", "", "Path to private key (.p8). Defaults to profile private_key_path")
}

var authPublicKeyCmd = &cobra.Command{
	Use:   "public-key",
	Short: "Print uploadable public key from private key",
	Long: strings.TrimSpace(`
Print the PUBLIC KEY derived from your local private key.

Upload this PEM block in Apple Ads > Account Settings > Client Credentials (or API).
Do not upload the private key.
`),
	Example: strings.TrimSpace(`
  appleads auth public-key
  appleads auth public-key --private-key-path ~/.config/appleads/default.p8
`),
	RunE: func(cmd *cobra.Command, args []string) error {
		privateKeyPath := authPublicKeyFlags.PrivateKeyPath
		if privateKeyPath == "" {
			_, p, err := loadProfile()
			if err != nil {
				return err
			}
			privateKeyPath = p.PrivateKeyPath
			if privateKeyPath == "" {
				return fmt.Errorf("private key path not found; pass --private-key-path or run appleads auth set")
			}
		}

		privatePEM, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return fmt.Errorf("read private key file %q: %w", privateKeyPath, err)
		}

		publicPEM, err := keys.PublicKeyFromPrivateKeyPEM(privatePEM)
		if err != nil {
			return err
		}

		if opts.Output == "json" {
			return printJSON(map[string]any{
				"private_key_path": privateKeyPath,
				"public_key":       string(publicPEM),
			})
		}

		fmt.Printf("Public key (upload this to Apple Ads):\n\n")
		fmt.Print(string(publicPEM))
		return nil
	},
}
