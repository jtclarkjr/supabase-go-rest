package supabase

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestGeneratePKCEPair verifies PKCE pair generation correctness.
func TestGeneratePKCEPair(t *testing.T) {
	p1, err := GeneratePKCEPair()
	if err != nil {
		t.Fatalf("GeneratePKCEPair error: %v", err)
	}
	if p1.Verifier == "" {
		t.Fatal("expected non-empty Verifier")
	}
	if p1.Challenge == "" {
		t.Fatal("expected non-empty Challenge")
	}

	// Re-derive challenge from verifier and confirm it matches.
	h := sha256.Sum256([]byte(p1.Verifier))
	expected := base64.RawURLEncoding.EncodeToString(h[:])
	if p1.Challenge != expected {
		t.Fatalf("challenge mismatch: got %s want %s", p1.Challenge, expected)
	}

	// Two calls should produce different verifiers.
	p2, err := GeneratePKCEPair()
	if err != nil {
		t.Fatalf("second GeneratePKCEPair error: %v", err)
	}
	if p1.Verifier == p2.Verifier {
		t.Fatal("expected unique verifiers across calls")
	}
}

// TestGetOAuthURL verifies URL construction for various parameter combinations.
func TestGetOAuthURL(t *testing.T) {
	client := NewClient("https://project.supabase.co", "api-key", "")

	t.Run("basic no pkce no scopes", func(t *testing.T) {
		u, err := client.GetOAuthURL(ProviderGitHub, "https://app.example.com/callback", nil, nil)
		if err != nil {
			t.Fatalf("GetOAuthURL error: %v", err)
		}
		parsed, _ := url.Parse(u)
		q := parsed.Query()
		if q.Get("provider") != "github" {
			t.Fatalf("expected provider=github, got %s", q.Get("provider"))
		}
		if q.Get("redirect_to") != "https://app.example.com/callback" {
			t.Fatalf("unexpected redirect_to: %s", q.Get("redirect_to"))
		}
		if q.Get("code_challenge") != "" {
			t.Fatal("unexpected code_challenge without pkce")
		}
	})

	t.Run("with scopes", func(t *testing.T) {
		u, err := client.GetOAuthURL(ProviderGoogle, "https://app.example.com/callback", []string{"email", "profile"}, nil)
		if err != nil {
			t.Fatalf("GetOAuthURL error: %v", err)
		}
		parsed, _ := url.Parse(u)
		scopes := parsed.Query().Get("scopes")
		if !strings.Contains(scopes, "email") || !strings.Contains(scopes, "profile") {
			t.Fatalf("unexpected scopes: %s", scopes)
		}
	})

	t.Run("with pkce", func(t *testing.T) {
		pkce, _ := GeneratePKCEPair()
		u, err := client.GetOAuthURL(ProviderDiscord, "https://app.example.com/callback", nil, &pkce)
		if err != nil {
			t.Fatalf("GetOAuthURL error: %v", err)
		}
		parsed, _ := url.Parse(u)
		q := parsed.Query()
		if q.Get("code_challenge") != pkce.Challenge {
			t.Fatalf("code_challenge mismatch: got %s want %s", q.Get("code_challenge"), pkce.Challenge)
		}
		if q.Get("code_challenge_method") != "S256" {
			t.Fatalf("expected S256, got %s", q.Get("code_challenge_method"))
		}
	})

	t.Run("empty redirectTo omitted", func(t *testing.T) {
		u, err := client.GetOAuthURL(ProviderGitHub, "", nil, nil)
		if err != nil {
			t.Fatalf("GetOAuthURL error: %v", err)
		}
		parsed, _ := url.Parse(u)
		if _, ok := parsed.Query()["redirect_to"]; ok {
			t.Fatal("redirect_to should be absent when empty")
		}
	})
}

// setUpOAuthTestServer creates a test server for OAuth-related endpoints.
// It handles tokenApiPath and dispatches on grant_type query param.
func setUpOAuthTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	handler := http.NewServeMux()

	handler.HandleFunc(tokenApiPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		grantType := r.URL.Query().Get("grant_type")
		switch grantType {
		case "pkce":
			var body pkcePayload
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			if body.AuthCode == "" || body.CodeVerifier == "" {
				http.Error(w, "missing fields", http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(mockAuthResponse)
		case "id_token":
			var body idTokenPayload
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			if body.Provider == "" || body.IdToken == "" {
				http.Error(w, "missing fields", http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(mockAuthResponse)
		default:
			// Let existing password/refresh flows through for completeness.
			_ = json.NewEncoder(w).Encode(mockAuthResponse)
		}
	})

	return httptest.NewServer(handler)
}

// TestExchangeCodeForSession tests the PKCE code exchange flow.
func TestExchangeCodeForSession(t *testing.T) {
	ts := setUpOAuthTestServer(t)
	defer ts.Close()
	client := NewClient(ts.URL, "api-key", "")

	t.Run("happy path", func(t *testing.T) {
		auth, err := client.ExchangeCodeForSession("auth-code-123", "verifier-abc")
		if err != nil {
			t.Fatalf("ExchangeCodeForSession error: %v", err)
		}
		if auth.AccessToken == "" {
			t.Fatal("expected non-empty AccessToken")
		}
		if auth.AccessToken != mockAuthResponse.AccessToken {
			t.Fatalf("unexpected AccessToken: got %s want %s", auth.AccessToken, mockAuthResponse.AccessToken)
		}
	})

	t.Run("server error returns ErrRequestFailed", func(t *testing.T) {
		errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		}))
		defer errServer.Close()
		errClient := NewClient(errServer.URL, "api-key", "")

		_, err := errClient.ExchangeCodeForSession("bad-code", "bad-verifier")
		if !errors.Is(err, ErrRequestFailed) {
			t.Fatalf("expected ErrRequestFailed, got %v", err)
		}
	})
}

// TestSignInWithIdToken tests the ID token exchange flow.
func TestSignInWithIdToken(t *testing.T) {
	ts := setUpOAuthTestServer(t)
	defer ts.Close()
	client := NewClient(ts.URL, "api-key", "")

	t.Run("happy path without nonce", func(t *testing.T) {
		auth, err := client.SignInWithIdToken(ProviderGoogle, "id-token-xyz", "")
		if err != nil {
			t.Fatalf("SignInWithIdToken error: %v", err)
		}
		if auth.AccessToken == "" {
			t.Fatal("expected non-empty AccessToken")
		}
	})

	t.Run("happy path with nonce", func(t *testing.T) {
		auth, err := client.SignInWithIdToken(ProviderApple, "id-token-xyz", "nonce-123")
		if err != nil {
			t.Fatalf("SignInWithIdToken with nonce error: %v", err)
		}
		if auth.AccessToken != mockAuthResponse.AccessToken {
			t.Fatalf("unexpected AccessToken: got %s want %s", auth.AccessToken, mockAuthResponse.AccessToken)
		}
	})

	t.Run("server error returns ErrRequestFailed", func(t *testing.T) {
		errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "forbidden", http.StatusForbidden)
		}))
		defer errServer.Close()
		errClient := NewClient(errServer.URL, "api-key", "")

		_, err := errClient.SignInWithIdToken(ProviderGoogle, "bad-token", "")
		if !errors.Is(err, ErrRequestFailed) {
			t.Fatalf("expected ErrRequestFailed, got %v", err)
		}
	})
}
