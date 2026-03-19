package appleads

import (
	"strings"
	"testing"
	"time"

	"github.com/ferdikt/appleads-cli/internal/keys"
	"github.com/golang-jwt/jwt/v5"
)

func TestBuildClientSecret_Success(t *testing.T) {
	privPEM, _, err := keys.GenerateP256KeyPair()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	secret, err := BuildClientSecret("SEARCHADS.team", "SEARCHADS.client", "key-123", privPEM, now)
	if err != nil {
		t.Fatalf("BuildClientSecret: %v", err)
	}

	if secret == "" {
		t.Fatal("secret is empty")
	}

	// Should be a valid JWT with 3 parts.
	parts := strings.Split(secret, ".")
	if len(parts) != 3 {
		t.Fatalf("JWT parts = %d, want 3", len(parts))
	}

	// Parse without verification to check claims.
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(secret, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("parse JWT: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	if claims["iss"] != "SEARCHADS.team" {
		t.Fatalf("iss = %v, want SEARCHADS.team", claims["iss"])
	}
	if claims["sub"] != "SEARCHADS.client" {
		t.Fatalf("sub = %v, want SEARCHADS.client", claims["sub"])
	}
	if claims["aud"] != "https://appleid.apple.com" {
		t.Fatalf("aud = %v", claims["aud"])
	}

	// kid header.
	if token.Header["kid"] != "key-123" {
		t.Fatalf("kid = %v, want key-123", token.Header["kid"])
	}
	if token.Header["alg"] != "ES256" {
		t.Fatalf("alg = %v, want ES256", token.Header["alg"])
	}

	// Expiry should be 180 days from now.
	iat := int64(claims["iat"].(float64))
	exp := int64(claims["exp"].(float64))
	diff := exp - iat
	expected := int64(180 * 24 * 3600)
	if diff != expected {
		t.Fatalf("exp - iat = %d, want %d (180 days)", diff, expected)
	}
}

func TestBuildClientSecret_MissingFields(t *testing.T) {
	privPEM, _, _ := keys.GenerateP256KeyPair()
	now := time.Now()

	tests := []struct {
		name     string
		teamID   string
		clientID string
		keyID    string
		key      []byte
	}{
		{"missing team_id", "", "client", "key", privPEM},
		{"missing client_id", "team", "", "key", privPEM},
		{"missing key_id", "team", "client", "", privPEM},
		{"missing key", "team", "client", "key", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildClientSecret(tt.teamID, tt.clientID, tt.keyID, tt.key, now)
			if err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
		})
	}
}

func TestBuildClientSecret_InvalidKey(t *testing.T) {
	_, err := BuildClientSecret("team", "client", "key", []byte("not-a-pem"), time.Now())
	if err == nil {
		t.Fatal("expected error for invalid PEM key")
	}
}
