package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load("/tmp/appleads-test-does-not-exist.json")
	if err != nil {
		t.Fatalf("Load non-existent: %v", err)
	}
	if cfg.Profiles == nil {
		t.Fatal("Profiles map should be initialized")
	}
	if len(cfg.Profiles) != 0 {
		t.Fatalf("expected 0 profiles, got %d", len(cfg.Profiles))
	}
}

func TestLoadEmptyPath(t *testing.T) {
	_, err := Load("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{
		ActiveProfile: "test",
		Profiles: map[string]*Profile{
			"test": {
				ClientID: "SEARCHADS.abc",
				TeamID:   "SEARCHADS.xyz",
				KeyID:    "key-123",
				OrgID:    12345,
			},
		},
	}

	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify file permissions.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Fatalf("file permissions = %o, want 600", perm)
	}

	// Load back and verify.
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ActiveProfile != "test" {
		t.Fatalf("ActiveProfile = %q, want test", loaded.ActiveProfile)
	}
	p, err := loaded.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if p.ClientID != "SEARCHADS.abc" {
		t.Fatalf("ClientID = %q, want SEARCHADS.abc", p.ClientID)
	}
	if p.OrgID != 12345 {
		t.Fatalf("OrgID = %d, want 12345", p.OrgID)
	}
}

func TestGetProfileNotFound(t *testing.T) {
	cfg := &Config{Profiles: map[string]*Profile{}}
	_, err := cfg.GetProfile("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
}

func TestEnsureProfile(t *testing.T) {
	cfg := &Config{}

	p := cfg.EnsureProfile("new")
	if p == nil {
		t.Fatal("EnsureProfile returned nil")
	}
	if cfg.ActiveProfile != "new" {
		t.Fatalf("ActiveProfile = %q, want new (auto-set for first profile)", cfg.ActiveProfile)
	}

	// Second profile should not change active.
	cfg.EnsureProfile("second")
	if cfg.ActiveProfile != "new" {
		t.Fatalf("ActiveProfile = %q, want new (should not change)", cfg.ActiveProfile)
	}
}

func TestEffectiveProfileName(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *Config
		defaultName string
		want        string
	}{
		{"nil config", nil, "fallback", "fallback"},
		{"active set", &Config{ActiveProfile: "prod"}, "dev", "prod"},
		{"no active, has default", &Config{}, "dev", "dev"},
		{"no active, no default", &Config{}, "", "default"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.EffectiveProfileName(tt.defaultName)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProfileNames(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]*Profile{
			"beta":    {},
			"alpha":   {},
			"charlie": {},
		},
	}
	names := cfg.ProfileNames()
	if len(names) != 3 {
		t.Fatalf("len = %d, want 3", len(names))
	}
	// Should be sorted.
	if names[0] != "alpha" || names[1] != "beta" || names[2] != "charlie" {
		t.Fatalf("names = %v, want [alpha beta charlie]", names)
	}
}

func TestProfileDefaults(t *testing.T) {
	p := &Profile{}
	if v := p.EffectiveAPIVersion(); v != "v5" {
		t.Fatalf("EffectiveAPIVersion = %q, want v5", v)
	}
	if u := p.EffectiveAPIBaseURL(); u != "https://api.searchads.apple.com" {
		t.Fatalf("EffectiveAPIBaseURL = %q", u)
	}
	if a := p.EffectiveAuthURL(); a != "https://appleid.apple.com/auth/oauth2/token" {
		t.Fatalf("EffectiveAuthURL = %q", a)
	}
}

func TestProfileOverrides(t *testing.T) {
	p := &Profile{
		APIVersion: "v4",
		APIBaseURL: "https://custom.api.com",
		AuthURL:    "https://custom.auth.com",
	}
	if v := p.EffectiveAPIVersion(); v != "v4" {
		t.Fatalf("EffectiveAPIVersion = %q, want v4", v)
	}
	if u := p.EffectiveAPIBaseURL(); u != "https://custom.api.com" {
		t.Fatalf("EffectiveAPIBaseURL = %q", u)
	}
	if a := p.EffectiveAuthURL(); a != "https://custom.auth.com" {
		t.Fatalf("EffectiveAuthURL = %q", a)
	}
}

func TestResolvePrivateKeyPEM_Inline(t *testing.T) {
	p := &Profile{PrivateKeyPEM: "-----BEGIN PRIVATE KEY-----\nfake\n-----END PRIVATE KEY-----"}
	data, err := p.ResolvePrivateKeyPEM()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if string(data) != p.PrivateKeyPEM {
		t.Fatalf("data mismatch")
	}
}

func TestResolvePrivateKeyPEM_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.pem")
	content := []byte("-----BEGIN PRIVATE KEY-----\nfilecontent\n-----END PRIVATE KEY-----")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	p := &Profile{PrivateKeyPath: path}
	data, err := p.ResolvePrivateKeyPEM()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if string(data) != string(content) {
		t.Fatalf("data mismatch")
	}
}

func TestResolvePrivateKeyPEM_NotConfigured(t *testing.T) {
	p := &Profile{}
	_, err := p.ResolvePrivateKeyPEM()
	if err == nil {
		t.Fatal("expected error when no key configured")
	}
}

func TestSaveNilConfig(t *testing.T) {
	var cfg *Config
	err := cfg.Save("/tmp/test.json")
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestTokenExpiresAtRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	exp := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	cfg := &Config{
		ActiveProfile: "test",
		Profiles: map[string]*Profile{
			"test": {
				ClientID:       "id",
				AccessToken:    "tok",
				TokenExpiresAt: &exp,
			},
		},
	}
	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	p := loaded.Profiles["test"]
	if p.TokenExpiresAt == nil {
		t.Fatal("TokenExpiresAt is nil after round-trip")
	}
	if !p.TokenExpiresAt.Equal(exp) {
		t.Fatalf("TokenExpiresAt = %v, want %v", p.TokenExpiresAt, exp)
	}
}
