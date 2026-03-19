package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ferdikt/appleads-cli/internal/keys"
	"github.com/spf13/cobra"
)

func init() {
	authCmd.AddCommand(authInitCmd)
}

var authInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive profile setup wizard",
	Long:  "Guides you through setting Apple Ads OAuth profile values and optionally validates by creating a token.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		p := cfg.EnsureProfile(opts.Profile)

		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Profile setup wizard (%s)\n", opts.Profile)
		fmt.Println("Press Enter to keep current values.")

		clientID, err := promptString(reader, "Client ID", p.ClientID, true)
		if err != nil {
			return err
		}
		teamID, err := promptString(reader, "Team ID", p.TeamID, true)
		if err != nil {
			return err
		}
		keyID, err := promptString(reader, "Key ID (kid)", p.KeyID, true)
		if err != nil {
			return err
		}
		orgID, err := promptInt64(reader, "Organization ID (org_id, optional for token)", p.OrgID, false)
		if err != nil {
			return err
		}
		if p.PrivateKeyPath == "" {
			generateNow, err := promptYesNo(reader, "No private key found. Generate a new key pair now?", true)
			if err != nil {
				return err
			}
			if generateNow {
				privateOut, publicOut, err := resolveKeyOutputPaths("", "", opts.Profile)
				if err != nil {
					return err
				}
				if err := ensureWritablePath(privateOut, false); err != nil {
					return err
				}
				if err := ensureWritablePath(publicOut, false); err != nil {
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

				p.PrivateKeyPath = privateOut

				fmt.Printf("private key saved: %s\n", privateOut)
				fmt.Printf("public key saved: %s\n", publicOut)
				fmt.Println("Upload the PUBLIC KEY below in Apple Ads > Client Credentials:")
				fmt.Println()
				fmt.Print(string(publicPEM))
				fmt.Println()
			}
		}
		privateKeyPath, err := promptFilePath(reader, "Private key path (.p8)", p.PrivateKeyPath, true)
		if err != nil {
			return err
		}
		apiVersion, err := promptString(reader, "API version", p.EffectiveAPIVersion(), true)
		if err != nil {
			return err
		}
		apiBaseURL, err := promptString(reader, "API base URL", p.EffectiveAPIBaseURL(), true)
		if err != nil {
			return err
		}
		authURL, err := promptString(reader, "OAuth token URL", p.EffectiveAuthURL(), true)
		if err != nil {
			return err
		}
		testToken, err := promptYesNo(reader, "Create test access token now?", true)
		if err != nil {
			return err
		}

		p.ClientID = clientID
		p.TeamID = teamID
		p.KeyID = keyID
		p.OrgID = orgID
		p.PrivateKeyPath = privateKeyPath
		p.APIVersion = apiVersion
		p.APIBaseURL = apiBaseURL
		p.AuthURL = authURL
		p.AccessToken = ""
		p.TokenExpiresAt = nil

		if err := cfg.Save(opts.ConfigPath); err != nil {
			return err
		}

		if testToken {
			token, err := ensureAccessToken(context.Background(), cfg, p)
			if err != nil {
				return fmt.Errorf("profile saved but token creation failed: %w", err)
			}
			if opts.Output == "json" {
				return printJSON(map[string]any{
					"profile":      opts.Profile,
					"config":       opts.ConfigPath,
					"token_cached": token != "",
					"expires_at":   p.TokenExpiresAt.UTC().Format(time.RFC3339),
				})
			}
			fmt.Printf("profile saved: %s\n", opts.ConfigPath)
			fmt.Printf("token created, expires at: %s\n", p.TokenExpiresAt.UTC().Format(time.RFC3339))
			if p.OrgID <= 0 {
				fmt.Println("note: org_id is empty. Set it later before campaign/report API calls.")
			}
			return nil
		}

		if opts.Output == "json" {
			return printJSON(map[string]any{
				"profile":    opts.Profile,
				"config":     opts.ConfigPath,
				"saved":      true,
				"token_test": false,
			})
		}

		fmt.Printf("profile saved: %s\n", opts.ConfigPath)
		if p.OrgID <= 0 {
			fmt.Println("note: org_id is empty. Set it later before campaign/report API calls.")
		}
		fmt.Println("token test skipped.")
		return nil
	},
}

func promptString(reader *bufio.Reader, label, current string, required bool) (string, error) {
	for {
		val, err := promptLine(reader, label, current)
		if err != nil {
			return "", err
		}
		if val == "" {
			if required {
				fmt.Printf("%s is required.\n", label)
				continue
			}
			return "", nil
		}
		return val, nil
	}
}

func promptInt64(reader *bufio.Reader, label string, current int64, required bool) (int64, error) {
	def := ""
	if current > 0 {
		def = strconv.FormatInt(current, 10)
	}
	for {
		text, err := promptLine(reader, label, def)
		if err != nil {
			return 0, err
		}
		if text == "" {
			if required {
				fmt.Printf("%s is required.\n", label)
				continue
			}
			return 0, nil
		}
		v, err := strconv.ParseInt(text, 10, 64)
		if err != nil || v <= 0 {
			fmt.Printf("%s must be a positive integer.\n", label)
			continue
		}
		return v, nil
	}
}

func promptFilePath(reader *bufio.Reader, label, current string, required bool) (string, error) {
	for {
		path, err := promptString(reader, label, current, required)
		if err != nil {
			return "", err
		}
		if path == "" {
			return "", nil
		}
		fi, err := os.Stat(path)
		if err != nil {
			fmt.Printf("cannot read file: %v\n", err)
			continue
		}
		if fi.IsDir() {
			fmt.Println("path is a directory, expected a .p8 file.")
			continue
		}
		return path, nil
	}
}

func promptYesNo(reader *bufio.Reader, label string, defaultYes bool) (bool, error) {
	def := "y"
	if !defaultYes {
		def = "n"
	}
	for {
		val, err := promptLine(reader, label+" (y/n)", def)
		if err != nil {
			return false, err
		}
		switch strings.ToLower(val) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Println("please answer y or n.")
		}
	}
}

func promptLine(reader *bufio.Reader, label, current string) (string, error) {
	if current != "" {
		fmt.Printf("%s [%s]: ", label, current)
	} else {
		fmt.Printf("%s: ", label)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			line = strings.TrimSpace(line)
			if line == "" {
				return "", fmt.Errorf("input aborted")
			}
			return line, nil
		}
		return "", err
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return strings.TrimSpace(current), nil
	}
	return line, nil
}
