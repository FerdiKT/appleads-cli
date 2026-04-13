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

	"github.com/ferdikt/appleads-cli/internal/config"
	"github.com/ferdikt/appleads-cli/internal/keys"
	"github.com/spf13/cobra"
)

func init() {
	authCmd.AddCommand(authInitCmd)
}

var authInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive profile setup wizard",
	Long: strings.TrimSpace(`
Interactive setup for an Apple Ads profile.

The wizard follows the real Apple Ads OAuth order:
  1. Confirm which Apple Ads user will own the access.
  2. Generate or reuse a local P-256 private key.
  3. Upload the PUBLIC KEY in Apple Ads and collect clientId/teamId/keyId.
  4. Save credentials locally.
  5. Optionally mint a token now.

Use this when you are setting up a profile for the first time or rotating keys.
`),
	Example: strings.TrimSpace(`
  appleads auth init
  appleads -p agency auth init
`),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		p := cfg.EnsureProfile(opts.Profile)

		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Apple Ads auth setup (%s)\n", opts.Profile)
		fmt.Println("Press Enter to keep current values.")

		hasAppleAdsLogin, err := promptYesNo(reader, "Do you already have an Apple Ads login for Apple Ads Advanced?", true)
		if err != nil {
			return err
		}
		if !hasAppleAdsLogin {
			fmt.Println()
			fmt.Println("Before continuing:")
			fmt.Println("  - Create or activate the Apple ID that will be used with Apple Ads.")
			fmt.Println("  - Sign in to Apple Ads Advanced once in the browser.")
			fmt.Println("  - Make sure this Apple ID can access the correct Apple Ads account.")
			fmt.Println()
			if err := promptContinue(reader, "Press Enter after that Apple Ads login exists and can sign in"); err != nil {
				return err
			}
		}

		hasAPIAccess, err := promptYesNo(reader, "Is this the designated API user Apple ID for the target account, or another login with confirmed API access?", true)
		if err != nil {
			return err
		}
		if !hasAPIAccess {
			fmt.Println()
			fmt.Println("Before continuing:")
			fmt.Println("  - Sign in with the primary/admin Apple Ads account.")
			fmt.Println("  - In User Management, invite a separate Apple ID that will be used as the API user.")
			fmt.Println("  - Grant that Apple ID an API-capable role.")
			fmt.Println("  - In many setups this is an API Manager-style role or equivalent API access.")
			fmt.Println("  - Then sign back in as that invited API user to create client credentials.")
			fmt.Println("  - For RevenueCat-style third-party authorization, Account Admin or Campaign Group Manager is typically required.")
			fmt.Println()
			if err := promptContinue(reader, "Press Enter after the separate API user can sign in and access the target account"); err != nil {
				return err
			}
		}

		haveCreds := p.ClientID != "" && p.TeamID != "" && p.KeyID != ""
		haveCreds, err = promptYesNo(reader, "Do you already have clientId, teamId, and keyId from Apple Ads?", haveCreds)
		if err != nil {
			return err
		}

		privateKeyPath, publicPEM, err := resolveWizardPrivateKey(reader, opts.Profile, p.PrivateKeyPath, !haveCreds)
		if err != nil {
			return err
		}
		if !haveCreds {
			printPublicKeyUploadStep(publicPEM)
			if err := promptContinue(reader, "Press Enter after you upload the public key in Apple Ads and copy clientId, teamId, and keyId"); err != nil {
				return err
			}
		}

		clientID, teamID, keyID, parsedAllCreds, err := promptCredentialBlock(reader, p.ClientID, p.TeamID, p.KeyID)
		if err != nil {
			return err
		}
		if !parsedAllCreds {
			clientID, err = promptString(reader, "Client ID", clientID, true)
			if err != nil {
				return err
			}
			teamID, err = promptString(reader, "Team ID", teamID, true)
			if err != nil {
				return err
			}
			keyID, err = promptString(reader, "Key ID (kid)", keyID, true)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("Parsed clientId, teamId, and keyId from the pasted block.")
		}
		apiVersion, err := promptString(reader, "API version", p.EffectiveAPIVersion(), true)
		if err != nil {
			return err
		}
		if !config.IsValidAPIVersion(apiVersion) {
			return fmt.Errorf("invalid API version %q: expected format like v5", apiVersion)
		}
		apiBaseURL, err := promptString(reader, "API base URL", p.EffectiveAPIBaseURL(), true)
		if err != nil {
			return err
		}
		authURL, err := promptString(reader, "OAuth token URL", p.EffectiveAuthURL(), true)
		if err != nil {
			return err
		}
		testToken, err := promptYesNo(reader, "Create a test access token now?", true)
		if err != nil {
			return err
		}

		p.ClientID = clientID
		p.TeamID = teamID
		p.KeyID = keyID
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
			selectOrgNow, err := promptYesNo(reader, "Fetch accessible orgs now and save one as org_id?", p.OrgID <= 0)
			if err != nil {
				return err
			}
			if selectOrgNow {
				if err := selectOrgForProfile(context.Background(), cfg, p); err != nil {
					return fmt.Errorf("token created but org selection failed: %w", err)
				}
			} else if p.OrgID <= 0 {
				fmt.Println("next: run `appleads auth orgs --select` before campaign/report API calls.")
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
		fmt.Println("next: run `appleads auth token`, then `appleads auth orgs --select`.")
		fmt.Println("token creation skipped.")
		return nil
	},
}

func resolveWizardPrivateKey(reader *bufio.Reader, profileName, currentPath string, needPublicKey bool) (string, []byte, error) {
	if currentPath != "" {
		reuse, err := promptYesNo(reader, fmt.Sprintf("Reuse the current private key at %q?", currentPath), true)
		if err != nil {
			return "", nil, err
		}
		if reuse {
			path := currentPath
			if !needPublicKey {
				return path, nil, nil
			}
			publicPEM, err := publicKeyFromPrivateKeyPath(path)
			if err != nil {
				fmt.Printf("cannot read current private key: %v\n", err)
				path, err = promptFilePath(reader, "Private key path (.p8)", currentPath, true)
				if err != nil {
					return "", nil, err
				}
				publicPEM, err = publicKeyFromPrivateKeyPath(path)
				if err != nil {
					return "", nil, err
				}
			}
			return path, publicPEM, nil
		}
	}

	generateNow, err := promptYesNo(reader, "Generate a new local P-256 key pair now?", true)
	if err != nil {
		return "", nil, err
	}
	if generateNow {
		privateOut, publicOut, publicPEM, err := generateWizardKeyPair(reader, profileName)
		if err != nil {
			return "", nil, err
		}
		fmt.Printf("private key saved: %s\n", privateOut)
		fmt.Printf("public key saved: %s\n", publicOut)
		return privateOut, publicPEM, nil
	}

	path, err := promptFilePath(reader, "Private key path (.p8)", currentPath, true)
	if err != nil {
		return "", nil, err
	}
	if !needPublicKey {
		return path, nil, nil
	}
	publicPEM, err := publicKeyFromPrivateKeyPath(path)
	if err != nil {
		return "", nil, err
	}
	return path, publicPEM, nil
}

func printPublicKeyUploadStep(publicPEM []byte) {
	fmt.Println()
	fmt.Println("Upload this PUBLIC KEY in Apple Ads:")
	fmt.Println("  - Open Apple Ads > Account Settings > Client Credentials (or API).")
	fmt.Println("  - Paste the full PEM block below.")
	fmt.Println("  - Save it, then Apple will show clientId, teamId, and keyId.")
	fmt.Println()
	fmt.Print(string(publicPEM))
	fmt.Println()
}

func promptCredentialBlock(reader *bufio.Reader, currentClientID, currentTeamID, currentKeyID string) (string, string, string, bool, error) {
	fmt.Println()
	fmt.Println("Paste the Apple credential block now if you want.")
	fmt.Println("Expected lines:")
	fmt.Println("  clientId SEARCHADS....")
	fmt.Println("  teamId SEARCHADS....")
	fmt.Println("  keyId  xxxxxxxx-....")
	fmt.Println("You can paste either:")
	fmt.Println("  - three single-line pairs, or")
	fmt.Println("  - alternating key/value lines copied from the UI.")
	fmt.Println("The wizard stops automatically after it detects clientId, teamId, and keyId.")
	fmt.Println("Press Enter on an empty line to skip and enter them one by one.")
	fmt.Print("Paste first line (or press Enter to skip): ")

	firstLine, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", "", "", false, err
	}
	firstLine = strings.TrimSpace(firstLine)
	if firstLine == "" {
		return currentClientID, currentTeamID, currentKeyID, false, nil
	}

	clientID, teamID, keyID := currentClientID, currentTeamID, currentKeyID
	explicit := map[string]bool{}
	pendingKey := ""
	clientID, teamID, keyID, pendingKey = applyCredentialLine(firstLine, clientID, teamID, keyID, pendingKey, explicit)

	for !(explicit["clientid"] && explicit["teamid"] && explicit["keyid"]) || pendingKey != "" {
		line, readErr := reader.ReadString('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return "", "", "", false, readErr
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		clientID, teamID, keyID, pendingKey = applyCredentialLine(line, clientID, teamID, keyID, pendingKey, explicit)
		if errors.Is(readErr, io.EOF) {
			break
		}
	}

	return clientID, teamID, keyID, explicit["clientid"] && explicit["teamid"] && explicit["keyid"], nil
}

func applyCredentialLine(line, clientID, teamID, keyID, pendingKey string, explicit map[string]bool) (string, string, string, string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return clientID, teamID, keyID, pendingKey
	}

	if pendingKey != "" {
		if isCredentialKey(strings.ToLower(line)) {
			pendingKey = strings.ToLower(line)
			return clientID, teamID, keyID, pendingKey
		}
		switch pendingKey {
		case "clientid":
			clientID = line
		case "teamid":
			teamID = line
		case "keyid":
			keyID = line
		}
		explicit[pendingKey] = true
		return clientID, teamID, keyID, ""
	}

	if isCredentialKey(strings.ToLower(line)) {
		return clientID, teamID, keyID, strings.ToLower(line)
	}

	key, value, ok := parseCredentialLine(line)
	if !ok {
		return clientID, teamID, keyID, pendingKey
	}
	switch key {
	case "clientid":
		clientID = value
	case "teamid":
		teamID = value
	case "keyid":
		keyID = value
	}
	explicit[key] = true
	return clientID, teamID, keyID, pendingKey
}

func parseCredentialLine(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", false
	}
	fields := strings.Fields(line)
	if len(fields) >= 2 {
		key := strings.ToLower(strings.TrimSuffix(strings.TrimSuffix(fields[0], ":"), "="))
		value := strings.TrimSpace(strings.Join(fields[1:], " "))
		if isCredentialKey(key) && value != "" {
			return key, value, true
		}
	}

	for _, sep := range []string{":", "="} {
		if left, right, ok := strings.Cut(line, sep); ok {
			key := strings.ToLower(strings.TrimSpace(left))
			value := strings.TrimSpace(right)
			if isCredentialKey(key) && value != "" {
				return key, value, true
			}
		}
	}

	return "", "", false
}

func isCredentialKey(key string) bool {
	switch key {
	case "clientid", "teamid", "keyid":
		return true
	default:
		return false
	}
}

func generateWizardKeyPair(reader *bufio.Reader, profileName string) (string, string, []byte, error) {
	privateOut, publicOut, err := resolveKeyOutputPaths("", "", profileName)
	if err != nil {
		return "", "", nil, err
	}

	forceOverwrite := false
	if pathExists(privateOut) || pathExists(publicOut) {
		overwrite, err := promptYesNo(reader, fmt.Sprintf("Key files already exist for profile %q. Overwrite them?", profileName), false)
		if err != nil {
			return "", "", nil, err
		}
		if overwrite {
			forceOverwrite = true
		} else {
			privateOut, publicOut = uniqueKeyOutputPaths(privateOut, publicOut)
			fmt.Printf("Using new key file names:\n  - %s\n  - %s\n", privateOut, publicOut)
		}
	}
	if err := ensureWritablePath(privateOut, forceOverwrite); err != nil {
		return "", "", nil, err
	}
	if err := ensureWritablePath(publicOut, forceOverwrite); err != nil {
		return "", "", nil, err
	}

	privatePEM, publicPEM, err := keys.GenerateP256KeyPair()
	if err != nil {
		return "", "", nil, err
	}
	if err := os.MkdirAll(filepath.Dir(privateOut), 0o700); err != nil {
		return "", "", nil, fmt.Errorf("create private key directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(publicOut), 0o755); err != nil {
		return "", "", nil, fmt.Errorf("create public key directory: %w", err)
	}
	if err := os.WriteFile(privateOut, privatePEM, 0o600); err != nil {
		return "", "", nil, fmt.Errorf("write private key: %w", err)
	}
	if err := os.WriteFile(publicOut, publicPEM, 0o644); err != nil {
		return "", "", nil, fmt.Errorf("write public key: %w", err)
	}
	return privateOut, publicOut, publicPEM, nil
}

func uniqueKeyOutputPaths(privateOut, publicOut string) (string, string) {
	suffix := time.Now().UTC().Format("20060102-150405")
	return appendSuffixBeforeExt(privateOut, "-"+suffix), appendSuffixBeforeExt(publicOut, "-"+suffix)
}

func appendSuffixBeforeExt(path, suffix string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	return base + suffix + ext
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func publicKeyFromPrivateKeyPath(path string) ([]byte, error) {
	privatePEM, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key file %q: %w", path, err)
	}
	publicPEM, err := keys.PublicKeyFromPrivateKeyPEM(privatePEM)
	if err != nil {
		return nil, err
	}
	return publicPEM, nil
}

func selectOrgForProfile(ctx context.Context, cfg *config.Config, p *config.Profile) error {
	token, err := ensureAccessToken(ctx, cfg, p)
	if err != nil {
		return err
	}

	resp, apiVersionUsed, err := fetchUserACLsWithFallback(p, token)
	if err != nil {
		return err
	}
	if len(resp.Data) == 0 {
		return fmt.Errorf("no organizations found for this token")
	}

	w := tableWriter()
	fmt.Fprintln(w, "#\tORG_ID\tORG_NAME\tPARENT_ORG_ID\tCURRENCY\tTIMEZONE\tROLES")
	for i, org := range resp.Data {
		roles := strings.Join(org.RoleNames, ",")
		fmt.Fprintf(w, "%d\t%d\t%s\t%d\t%s\t%s\t%s\n", i+1, org.OrgID, org.OrgName, org.ParentOrgID, org.Currency, org.TimeZone, roles)
	}
	_ = w.Flush()
	fmt.Printf("API version used for ACL lookup: %s\n", apiVersionUsed)

	selectedOrgID, err := promptOrgSelection(resp.Data, p.OrgID)
	if err != nil {
		return err
	}
	if selectedOrgID == 0 {
		fmt.Println("selection skipped.")
		return nil
	}

	p.OrgID = selectedOrgID
	if err := cfg.Save(opts.ConfigPath); err != nil {
		return err
	}
	fmt.Printf("saved org_id=%d to profile %q\n", selectedOrgID, opts.Profile)
	return nil
}

func promptContinue(reader *bufio.Reader, label string) error {
	_, err := promptLine(reader, label, "")
	return err
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
