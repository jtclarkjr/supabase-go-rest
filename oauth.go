package supabase

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// OAuthProvider is a typed string for OAuth provider names accepted by Supabase.
// Callers can also cast arbitrary strings: OAuthProvider("custom_provider").
type OAuthProvider string

const (
	ProviderGitHub   OAuthProvider = "github"
	ProviderGoogle   OAuthProvider = "google"
	ProviderDiscord  OAuthProvider = "discord"
	ProviderApple    OAuthProvider = "apple"
	ProviderFacebook OAuthProvider = "facebook"
	ProviderTwitter  OAuthProvider = "twitter"
	ProviderSlack    OAuthProvider = "slack"
	ProviderSpotify  OAuthProvider = "spotify"
	ProviderTwitch   OAuthProvider = "twitch"
	ProviderLinkedIn OAuthProvider = "linkedin_oidc"
	ProviderNotion   OAuthProvider = "notion"
	ProviderZoom     OAuthProvider = "zoom"
)

// PKCEPair holds a PKCE code verifier and its derived S256 code challenge.
// Generate one before calling GetOAuthURL, persist Verifier server-side
// (e.g., in a session), then pass it to ExchangeCodeForSession.
type PKCEPair struct {
	Verifier  string
	Challenge string
}

// pkcePayload is the request body for a PKCE token exchange.
type pkcePayload struct {
	AuthCode     string `json:"auth_code"`
	CodeVerifier string `json:"code_verifier"`
}

// idTokenPayload is the request body for an ID-token exchange.
type idTokenPayload struct {
	Provider string `json:"provider"`
	IdToken  string `json:"id_token"`
	Nonce    string `json:"nonce,omitempty"`
}

// generatePKCEVerifier generates a cryptographically random base64url-encoded verifier.
func generatePKCEVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// derivePKCEChallenge derives the S256 code challenge from a verifier.
func derivePKCEChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// GeneratePKCEPair generates a new PKCE verifier+challenge pair.
// The Verifier must be stored by the caller and passed to ExchangeCodeForSession.
func GeneratePKCEPair() (PKCEPair, error) {
	verifier, err := generatePKCEVerifier()
	if err != nil {
		return PKCEPair{}, fmt.Errorf("pkce: failed to generate verifier: %w", err)
	}
	return PKCEPair{
		Verifier:  verifier,
		Challenge: derivePKCEChallenge(verifier),
	}, nil
}

// GetOAuthURL returns the Supabase /auth/v1/authorize URL for the given provider.
// The caller is responsible for redirecting the end-user's browser to this URL.
// If pkce is non-nil, code_challenge and code_challenge_method=S256 are appended;
// the caller must persist pkce.Verifier for use with ExchangeCodeForSession.
// scopes may be nil or empty to use the project's default scopes.
func (c *Client) GetOAuthURL(provider OAuthProvider, redirectTo string, scopes []string, pkce *PKCEPair) (string, error) {
	base := fmt.Sprintf("%s%s", c.BaseUrl, authorizeApiPath)
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("GetOAuthURL: failed to parse base URL: %w", err)
	}
	q := u.Query()
	q.Set("provider", string(provider))
	if redirectTo != "" {
		q.Set("redirect_to", redirectTo)
	}
	if len(scopes) > 0 {
		q.Set("scopes", strings.Join(scopes, " "))
	}
	if pkce != nil {
		q.Set("code_challenge", pkce.Challenge)
		q.Set("code_challenge_method", "S256")
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// ExchangeCodeForSession exchanges an OAuth authorization code for a Supabase session
// using the PKCE flow. codeVerifier must be the Verifier from the PKCEPair used when
// building the authorization URL with GetOAuthURL.
func (c *Client) ExchangeCodeForSession(code string, codeVerifier string) (*AuthTokenResponse, error) {
	payload := pkcePayload{
		AuthCode:     code,
		CodeVerifier: codeVerifier,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ExchangeCodeForSession: failed to marshal payload: %w", err)
	}

	urlStr := fmt.Sprintf("%s%s?grant_type=pkce", c.BaseUrl, tokenApiPath)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("ExchangeCodeForSession: failed to create request: %w", err)
	}
	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ExchangeCodeForSession: request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("ExchangeCodeForSession: error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("ExchangeCodeForSession: request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		return nil, ErrRequestFailed
	}

	var authResponse AuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return nil, fmt.Errorf("ExchangeCodeForSession: failed to decode response: %w", err)
	}
	return &authResponse, nil
}

// SignInWithIdToken exchanges a third-party ID token for a Supabase session.
// provider must be one of the providers enabled in your Supabase project that
// support the id_token grant (e.g. google, apple, azure, facebook, kakao).
// Pass an empty string for nonce if the provider does not require one.
func (c *Client) SignInWithIdToken(provider OAuthProvider, idToken string, nonce string) (*AuthTokenResponse, error) {
	payload := idTokenPayload{
		Provider: string(provider),
		IdToken:  idToken,
		Nonce:    nonce,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("SignInWithIdToken: failed to marshal payload: %w", err)
	}

	urlStr := fmt.Sprintf("%s%s?grant_type=id_token", c.BaseUrl, tokenApiPath)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("SignInWithIdToken: failed to create request: %w", err)
	}
	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SignInWithIdToken: request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("SignInWithIdToken: error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("SignInWithIdToken: request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		return nil, ErrRequestFailed
	}

	var authResponse AuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return nil, fmt.Errorf("SignInWithIdToken: failed to decode response: %w", err)
	}
	return &authResponse, nil
}
