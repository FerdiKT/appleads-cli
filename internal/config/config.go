package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	defaultAPIVersion = "v5"
	defaultAPIBaseURL = "https://api.searchads.apple.com"
	defaultAuthURL    = "https://appleid.apple.com/auth/oauth2/token"
)

type Config struct {
	ActiveProfile string              `json:"active_profile,omitempty"`
	Profiles      map[string]*Profile `json:"profiles"`
}

type Profile struct {
	ClientID       string     `json:"client_id,omitempty"`
	TeamID         string     `json:"team_id,omitempty"`
	KeyID          string     `json:"key_id,omitempty"`
	OrgID          int64      `json:"org_id,omitempty"`
	PrivateKeyPath string     `json:"private_key_path,omitempty"`
	PrivateKeyPEM  string     `json:"private_key_pem,omitempty"`
	APIVersion     string     `json:"api_version,omitempty"`
	APIBaseURL     string     `json:"api_base_url,omitempty"`
	AuthURL        string     `json:"auth_url,omitempty"`
	AccessToken    string     `json:"access_token,omitempty"`
	TokenExpiresAt *time.Time `json:"token_expires_at,omitempty"`
}

func DefaultPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	newPath := filepath.Join(base, "appleads", "config.json")
	oldPath := filepath.Join(base, "appads", "config.json")

	if _, err := os.Stat(newPath); err == nil {
		return newPath, nil
	}
	if _, err := os.Stat(oldPath); err == nil {
		// One-time migration from legacy appads path.
		data, readErr := os.ReadFile(oldPath)
		if readErr == nil {
			_ = os.MkdirAll(filepath.Dir(newPath), 0o755)
			if writeErr := os.WriteFile(newPath, data, 0o600); writeErr == nil {
				_ = os.Remove(oldPath) // clean up legacy file
				return newPath, nil
			}
		}
		return oldPath, nil
	}
	return newPath, nil
}

func Load(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("config path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{Profiles: map[string]*Profile{}}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]*Profile{}
	}
	return &cfg, nil
}

func (c *Config) Save(path string) error {
	if c == nil {
		return errors.New("config is nil")
	}
	if path == "" {
		return errors.New("config path is empty")
	}
	if c.Profiles == nil {
		c.Profiles = map[string]*Profile{}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (c *Config) GetProfile(name string) (*Profile, error) {
	if c == nil {
		return nil, errors.New("config is nil")
	}
	p, ok := c.Profiles[name]
	if !ok || p == nil {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	return p, nil
}

func (c *Config) EnsureProfile(name string) *Profile {
	if c.Profiles == nil {
		c.Profiles = map[string]*Profile{}
	}
	if p, ok := c.Profiles[name]; ok && p != nil {
		return p
	}
	p := &Profile{}
	c.Profiles[name] = p
	if c.ActiveProfile == "" && len(c.Profiles) == 1 {
		c.ActiveProfile = name
	}
	return p
}

func (c *Config) EffectiveProfileName(defaultName string) string {
	if c == nil {
		return defaultName
	}
	if c.ActiveProfile != "" {
		return c.ActiveProfile
	}
	if defaultName != "" {
		return defaultName
	}
	return "default"
}

func (c *Config) ProfileNames() []string {
	if c == nil || len(c.Profiles) == 0 {
		return nil
	}
	out := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func (p *Profile) EffectiveAPIVersion() string {
	if p.APIVersion == "" {
		return defaultAPIVersion
	}
	return p.APIVersion
}

func (p *Profile) EffectiveAPIBaseURL() string {
	if p.APIBaseURL == "" {
		return defaultAPIBaseURL
	}
	return p.APIBaseURL
}

func (p *Profile) EffectiveAuthURL() string {
	if p.AuthURL == "" {
		return defaultAuthURL
	}
	return p.AuthURL
}

func (p *Profile) ResolvePrivateKeyPEM() ([]byte, error) {
	switch {
	case p.PrivateKeyPEM != "":
		return []byte(p.PrivateKeyPEM), nil
	case p.PrivateKeyPath != "":
		data, err := os.ReadFile(p.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("read private key file %q: %w", p.PrivateKeyPath, err)
		}
		return data, nil
	default:
		return nil, errors.New("private key is not configured (set private_key_path or private_key_pem)")
	}
}
