package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ferdikt/appleads-cli/internal/appleads"
	"github.com/ferdikt/appleads-cli/internal/config"
)

func loadProfile() (*config.Config, *config.Profile, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return nil, nil, err
	}
	profile, err := cfg.GetProfile(opts.Profile)
	if err != nil {
		return nil, nil, err
	}
	return cfg, profile, nil
}

func ensureAccessToken(ctx context.Context, cfg *config.Config, profile *config.Profile) (string, error) {
	if profile.AccessToken != "" && profile.TokenExpiresAt != nil {
		if profile.TokenExpiresAt.After(time.Now().Add(60 * time.Second)) {
			return profile.AccessToken, nil
		}
	}

	privateKeyPEM, err := profile.ResolvePrivateKeyPEM()
	if err != nil {
		return "", err
	}

	clientSecret, err := appleads.BuildClientSecret(profile.TeamID, profile.ClientID, profile.KeyID, privateKeyPEM, time.Now().UTC())
	if err != nil {
		return "", err
	}

	tokenResp, err := appleads.RequestAccessToken(ctx, nil, profile.EffectiveAuthURL(), profile.ClientID, clientSecret)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	profile.AccessToken = tokenResp.AccessToken
	profile.TokenExpiresAt = &expiresAt
	if err := cfg.Save(opts.ConfigPath); err != nil {
		return "", err
	}
	return tokenResp.AccessToken, nil
}

func apiBaseURL(profile *config.Profile) string {
	return apiBaseURLForVersion(profile, profile.EffectiveAPIVersion())
}

func resolveOrgID(overrideOrgID, profileOrgID int64) (int64, error) {
	if overrideOrgID > 0 {
		return overrideOrgID, nil
	}
	if profileOrgID > 0 {
		return profileOrgID, nil
	}
	return 0, fmt.Errorf("org_id is not set. Pass --org-id, or run `appleads auth orgs --select`, or `appleads auth set --org-id ...`")
}

func apiBaseURLForVersion(profile *config.Profile, version string) string {
	base := strings.TrimRight(profile.EffectiveAPIBaseURL(), "/")
	version = strings.Trim(version, "/")
	if version == "" {
		version = profile.EffectiveAPIVersion()
	}
	return fmt.Sprintf("%s/api/%s", base, version)
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func tableWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
}
