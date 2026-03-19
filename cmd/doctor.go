package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ferdikt/appleads-cli/internal/appleads"
	"github.com/ferdikt/appleads-cli/internal/config"
	"github.com/spf13/cobra"
)

type doctorCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"` // pass|warn|fail|skip
	Detail string `json:"detail"`
}

var doctorFlags struct {
	OrgID           int64
	NoNetwork       bool
	NoCampaignCheck bool
	StrictWarnings  bool
	NoTokenRefresh  bool
}

func init() {
	rootCmd.AddCommand(doctorCmd)

	doctorCmd.Flags().Int64Var(&doctorFlags.OrgID, "org-id", 0, "Organization ID override for org-scoped checks")
	doctorCmd.Flags().BoolVar(&doctorFlags.NoNetwork, "no-network", false, "Skip network/API checks")
	doctorCmd.Flags().BoolVar(&doctorFlags.NoCampaignCheck, "no-campaign-check", false, "Skip /campaigns health check")
	doctorCmd.Flags().BoolVar(&doctorFlags.StrictWarnings, "strict", false, "Exit non-zero when warnings are present")
	doctorCmd.Flags().BoolVar(&doctorFlags.NoTokenRefresh, "no-token-refresh", false, "Do not request a new token; only validate cached token")
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run profile/auth/org/API health checks",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		checks := make([]doctorCheck, 0, 16)
		add := func(name, status, detail string) {
			checks = append(checks, doctorCheck{Name: name, Status: status, Detail: detail})
		}

		cfg, err := config.Load(opts.ConfigPath)
		if err != nil {
			add("config.load", "fail", err.Error())
			return renderDoctor(checks, doctorFlags.StrictWarnings)
		}
		add("config.path", "pass", opts.ConfigPath)

		profile, err := cfg.GetProfile(opts.Profile)
		if err != nil {
			add("profile.load", "fail", err.Error())
			return renderDoctor(checks, doctorFlags.StrictWarnings)
		}
		add("profile.active", "pass", fmt.Sprintf("profile=%s active_profile=%s", opts.Profile, cfg.EffectiveProfileName(opts.Profile)))

		missing := make([]string, 0, 4)
		if strings.TrimSpace(profile.ClientID) == "" {
			missing = append(missing, "client_id")
		}
		if strings.TrimSpace(profile.TeamID) == "" {
			missing = append(missing, "team_id")
		}
		if strings.TrimSpace(profile.KeyID) == "" {
			missing = append(missing, "key_id")
		}
		if len(missing) > 0 {
			add("auth.required_fields", "fail", "missing: "+strings.Join(missing, ", "))
		} else {
			add("auth.required_fields", "pass", "client_id/team_id/key_id present")
		}

		privateKeyPEM, keyErr := profile.ResolvePrivateKeyPEM()
		if keyErr != nil {
			add("auth.private_key", "fail", keyErr.Error())
		} else {
			add("auth.private_key", "pass", "private key is readable")
		}

		if keyErr == nil && len(missing) == 0 {
			_, err := appleads.BuildClientSecret(profile.TeamID, profile.ClientID, profile.KeyID, privateKeyPEM, time.Now().UTC())
			if err != nil {
				add("auth.client_secret", "fail", err.Error())
			} else {
				add("auth.client_secret", "pass", "client secret can be generated")
			}
		} else {
			add("auth.client_secret", "skip", "skipped due to missing credentials/private key")
		}

		token := profile.AccessToken
		if token == "" {
			add("token.cached", "warn", "no cached token")
		} else {
			exp := ""
			if profile.TokenExpiresAt != nil {
				exp = profile.TokenExpiresAt.UTC().Format(time.RFC3339)
			}
			add("token.cached", "pass", "token present; expires_at="+exp)
		}

		if doctorFlags.NoNetwork {
			add("token.refresh", "skip", "skipped due to --no-network")
		} else if doctorFlags.NoTokenRefresh {
			add("token.refresh", "skip", "skipped due to --no-token-refresh")
		} else {
			tok, err := ensureAccessToken(ctx, cfg, profile)
			if err != nil {
				add("token.refresh", "fail", err.Error())
			} else {
				token = tok
				exp := ""
				if profile.TokenExpiresAt != nil {
					exp = profile.TokenExpiresAt.UTC().Format(time.RFC3339)
				}
				add("token.refresh", "pass", "ok; expires_at="+exp)
			}
		}

		orgID, orgErr := resolveOrgID(doctorFlags.OrgID, profile.OrgID)
		if orgErr != nil {
			add("org.resolve", "warn", orgErr.Error())
		} else {
			add("org.resolve", "pass", fmt.Sprintf("org_id=%d", orgID))
		}

		if doctorFlags.NoNetwork {
			add("api.me", "skip", "skipped due to --no-network")
			add("api.acls", "skip", "skipped due to --no-network")
			add("api.campaigns", "skip", "skipped due to --no-network")
			return renderDoctor(checks, doctorFlags.StrictWarnings)
		}
		if strings.TrimSpace(token) == "" {
			add("api.me", "fail", "cannot call API without token")
			add("api.acls", "fail", "cannot call API without token")
			add("api.campaigns", "skip", "skipped because token unavailable")
			return renderDoctor(checks, doctorFlags.StrictWarnings)
		}

		meClient := &appleads.Client{
			BaseURL: apiBaseURL(profile),
			Token:   token,
		}
		meOK := false
		var meErr error
		for _, p := range []string{"/me", "/me/details"} {
			var out map[string]any
			if err := meClient.DoJSON(ctx, http.MethodGet, p, nil, nil, &out); err == nil {
				add("api.me", "pass", "reachable via "+p)
				meOK = true
				break
			} else {
				meErr = err
			}
		}
		if !meOK {
			add("api.me", "fail", meErr.Error())
		}

		if _, version, err := fetchUserACLsWithFallback(profile, token); err != nil {
			add("api.acls", "fail", err.Error())
		} else {
			add("api.acls", "pass", "reachable (version="+version+")")
		}

		if doctorFlags.NoCampaignCheck {
			add("api.campaigns", "skip", "skipped due to --no-campaign-check")
		} else if orgErr != nil {
			add("api.campaigns", "warn", "skipped because org_id is missing")
		} else {
			cClient := &appleads.Client{
				BaseURL: apiBaseURL(profile),
				OrgID:   orgID,
				Token:   token,
			}
			q := url.Values{}
			q.Set("offset", "0")
			q.Set("limit", "1")
			var out map[string]any
			err := cClient.DoJSON(ctx, http.MethodGet, "/campaigns", q, nil, &out)
			if err != nil {
				add("api.campaigns", "fail", err.Error())
			} else {
				add("api.campaigns", "pass", "reachable with org context")
			}
		}

		return renderDoctor(checks, doctorFlags.StrictWarnings)
	},
}

func renderDoctor(checks []doctorCheck, strictWarnings bool) error {
	passCount := 0
	warnCount := 0
	failCount := 0
	skipCount := 0
	for _, c := range checks {
		switch c.Status {
		case "pass":
			passCount++
		case "warn":
			warnCount++
		case "fail":
			failCount++
		case "skip":
			skipCount++
		}
	}

	if opts.Output == "json" {
		_ = printJSON(map[string]any{
			"summary": map[string]any{
				"pass": passCount,
				"warn": warnCount,
				"fail": failCount,
				"skip": skipCount,
			},
			"checks": checks,
		})
	} else {
		w := tableWriter()
		fmt.Fprintln(w, "CHECK\tSTATUS\tDETAIL")
		for _, c := range checks {
			fmt.Fprintf(w, "%s\t%s\t%s\n", c.Name, strings.ToUpper(c.Status), c.Detail)
		}
		_ = w.Flush()
		fmt.Printf("summary: pass=%d warn=%d fail=%d skip=%d\n", passCount, warnCount, failCount, skipCount)
	}

	if failCount > 0 {
		return fmt.Errorf("doctor failed (%d check(s))", failCount)
	}
	if strictWarnings && warnCount > 0 {
		return fmt.Errorf("doctor has warnings (%d) and --strict is set", warnCount)
	}
	return nil
}
