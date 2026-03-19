package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ferdikt/appleads-cli/internal/keys"
	"github.com/spf13/cobra"
)

var authKeygenFlags struct {
	PrivateKeyOut string
	PublicKeyOut  string
	Force         bool
	SaveProfile   bool
	ShowPublicKey bool
}

func init() {
	authCmd.AddCommand(authKeygenCmd)

	authKeygenCmd.Flags().StringVar(&authKeygenFlags.PrivateKeyOut, "private-key-out", "", "Path to write generated private key (.p8)")
	authKeygenCmd.Flags().StringVar(&authKeygenFlags.PublicKeyOut, "public-key-out", "", "Path to write generated public key (.pem)")
	authKeygenCmd.Flags().BoolVar(&authKeygenFlags.Force, "force", false, "Overwrite output files if they exist")
	authKeygenCmd.Flags().BoolVar(&authKeygenFlags.SaveProfile, "save-profile", true, "Save private key path to selected profile")
	authKeygenCmd.Flags().BoolVar(&authKeygenFlags.ShowPublicKey, "show-public-key", true, "Print generated public key for Apple Ads upload")
}

var authKeygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate Apple Ads OAuth key pair (P-256)",
	Long:  "Generates a local ES256 key pair, stores private/public keys in files, and prints the public key to upload in Apple Ads.",
	RunE: func(cmd *cobra.Command, args []string) error {
		privateOut, publicOut, err := resolveKeyOutputPaths(authKeygenFlags.PrivateKeyOut, authKeygenFlags.PublicKeyOut, opts.Profile)
		if err != nil {
			return err
		}

		if err := ensureWritablePath(privateOut, authKeygenFlags.Force); err != nil {
			return err
		}
		if err := ensureWritablePath(publicOut, authKeygenFlags.Force); err != nil {
			return err
		}

		privatePEM, publicPEM, err := keys.GenerateP256KeyPair()
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(privateOut), 0o700); err != nil {
			return fmt.Errorf("create private key directory: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(publicOut), 0o755); err != nil {
			return fmt.Errorf("create public key directory: %w", err)
		}

		if err := os.WriteFile(privateOut, privatePEM, 0o600); err != nil {
			return fmt.Errorf("write private key: %w", err)
		}
		if err := os.WriteFile(publicOut, publicPEM, 0o644); err != nil {
			return fmt.Errorf("write public key: %w", err)
		}

		if authKeygenFlags.SaveProfile {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			p := cfg.EnsureProfile(opts.Profile)
			p.PrivateKeyPath = privateOut
			p.AccessToken = ""
			p.TokenExpiresAt = nil
			if err := cfg.Save(opts.ConfigPath); err != nil {
				return err
			}
		}

		if opts.Output == "json" {
			return printJSON(map[string]any{
				"profile":          opts.Profile,
				"private_key_path": privateOut,
				"public_key_path":  publicOut,
				"saved_profile":    authKeygenFlags.SaveProfile,
				"public_key":       string(publicPEM),
			})
		}

		fmt.Printf("private key saved: %s\n", privateOut)
		fmt.Printf("public key saved: %s\n", publicOut)
		if authKeygenFlags.SaveProfile {
			fmt.Printf("profile %q updated with private_key_path.\n", opts.Profile)
		}
		fmt.Println("Upload the PUBLIC KEY below in Apple Ads > Client Credentials:")
		if authKeygenFlags.ShowPublicKey {
			fmt.Println()
			fmt.Print(string(publicPEM))
		} else {
			fmt.Printf("(hidden; use --show-public-key)\n")
		}
		return nil
	},
}

func resolveKeyOutputPaths(privateOut, publicOut, profile string) (string, string, error) {
	if profile == "" {
		profile = "default"
	}
	if privateOut != "" && publicOut != "" {
		return privateOut, publicOut, nil
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", "", fmt.Errorf("resolve user config dir: %w", err)
	}
	keyDir := filepath.Join(base, "appleads", "keys")
	if privateOut == "" {
		privateOut = filepath.Join(keyDir, profile+".p8")
	}
	if publicOut == "" {
		publicOut = filepath.Join(keyDir, profile+".public.pem")
	}
	return privateOut, publicOut, nil
}

func ensureWritablePath(path string, force bool) error {
	if path == "" {
		return fmt.Errorf("empty output path")
	}
	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("file already exists: %s (use --force to overwrite)", path)
	}
	return nil
}
