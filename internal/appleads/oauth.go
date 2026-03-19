package appleads

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const oauthAudience = "https://appleid.apple.com"

type OAuthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
}

func BuildClientSecret(teamID, clientID, keyID string, privateKeyPEM []byte, now time.Time) (string, error) {
	if teamID == "" {
		return "", fmt.Errorf("team_id is required")
	}
	if clientID == "" {
		return "", fmt.Errorf("client_id is required")
	}
	if keyID == "" {
		return "", fmt.Errorf("key_id is required")
	}
	if len(privateKeyPEM) == 0 {
		return "", fmt.Errorf("private key is required")
	}

	key, err := jwt.ParseECPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}

	claims := jwt.MapClaims{
		"iss": teamID,
		"sub": clientID,
		"aud": oauthAudience,
		"iat": now.Unix(),
		"exp": now.Add(180 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = keyID

	signed, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("sign client secret: %w", err)
	}
	return signed, nil
}

func RequestAccessToken(ctx context.Context, httpClient *http.Client, authURL, clientID, clientSecret string) (*OAuthToken, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	if authURL == "" {
		return nil, fmt.Errorf("auth url is required")
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("scope", "searchadsorg")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var token OAuthToken
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	if token.AccessToken == "" {
		return nil, fmt.Errorf("token endpoint returned empty access_token")
	}
	return &token, nil
}
