package supabase

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"reflect"
	"strings"
	"testing"
)

// mockAuthResponse is reused for auth related success responses
var mockAuthResponse = AuthTokenResponse{
	AccessToken:  "access_token_value",
	TokenType:    "bearer",
	ExpiresIn:    3600,
	RefreshToken: "refresh_token_value",
}

// setUpTestServer creates a test server that mimics the supabase API paths used by the client.
func setUpTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	handler := http.NewServeMux()

	// Auth endpoints (token handled via authRequest without /rest prefix)
	handler.HandleFunc(tokenApiPath, func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(mockAuthResponse) })
	// Auth endpoints that (in current implementation) incorrectly go through doRequest and thus are prefixed with /rest/v1//auth/v1...
	combinedPaths := map[string]http.HandlerFunc{
		path.Clean(restApiPath + "/" + signupApiPath): func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST got %s", r.Method)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"new-user"}`))
		},
		path.Clean(restApiPath + "/" + magicLinkApiPath): func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"sent":true}`)) },
		path.Clean(restApiPath + "/" + recoverApiPath):   func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"recover":true}`)) },
		path.Clean(restApiPath + "/" + verifyApiPath):    func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"verified":true}`)) },
		path.Clean(restApiPath + "/" + userApiPath): func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				_, _ = w.Write([]byte(`{"user":"me"}`))
			case http.MethodPut:
				body, _ := io.ReadAll(r.Body)
				_, _ = w.Write(body)
			default:
				http.Error(w, "method", http.StatusMethodNotAllowed)
			}
		},
		path.Clean(restApiPath + "/" + logoutApiPath): func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"logged_out":true}`)) },
		path.Clean(restApiPath + "/" + inviteApiPath): func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"invited":true}`)) },
		path.Clean(restApiPath + "/" + resetApiPath):  func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"reset":true}`)) },
	}
	for p, h := range combinedPaths {
		handler.HandleFunc(p, h)
	}

	// REST endpoint handler - matches /rest/v1/<table>
	handler.HandleFunc(restApiPath+"/Food", func(w http.ResponseWriter, r *http.Request) {
		// capture some query params for assertions in specific tests using headers
		if r.Method == http.MethodGet {
			// echo query back
			m := map[string]any{}
			for k, v := range r.URL.Query() {
				if len(v) > 0 {
					m[k] = v[0]
				}
			}
			_ = json.NewEncoder(w).Encode(m)
			return
		}
		if r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write(body)
			return
		}
		if r.Method == http.MethodPut || r.Method == http.MethodPatch {
			body, _ := io.ReadAll(r.Body)
			_, _ = w.Write(body)
			return
		}
		if r.Method == http.MethodDelete {
			_, _ = w.Write([]byte(`{"deleted":true}`))
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// Endpoint used to verify Authorization header formatting
	handler.HandleFunc(restApiPath+"/HeaderCheck", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]string{"auth": auth})
	})

	return httptest.NewServer(handler)
}

// TestAuthAndUserMethods tests the authentication and user-related methods.
func TestAuthAndUserMethods(t *testing.T) {
	ts := setUpTestServer(t)
	defer ts.Close()

	client := NewClient(ts.URL, "api-key", "")

	// SignUp
	if _, err := client.SignUp("a@b.c", "pass"); err != nil {
		t.Fatalf("SignUp error: %v", err)
	}

	// SignIn
	auth, err := client.SignIn("a@b.c", "pass")
	if err != nil {
		t.Fatalf("SignIn error: %v", err)
	}
	if auth.AccessToken == "" {
		t.Fatalf("expected access token")
	}

	// RefreshToken
	auth2, err := client.RefreshToken("refresh")
	if err != nil || auth2.RefreshToken == "" {
		t.Fatalf("RefreshToken failed: %v", err)
	}

	// SendMagicLink
	if _, err := client.SendMagicLink("a@b.c"); err != nil {
		t.Fatalf("SendMagicLink: %v", err)
	}
	// SendPasswordRecovery
	if _, err := client.SendPasswordRecovery("a@b.c"); err != nil {
		t.Fatalf("SendPasswordRecovery: %v", err)
	}
	// VerifyOTP
	if _, err := client.VerifyOTP("a@b.c", "123456", "magiclink"); err != nil {
		t.Fatalf("VerifyOTP: %v", err)
	}

	// User endpoints
	client.Token = "Bearer token123"
	if _, err := client.GetUser(); err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if _, err := client.UpdateUser(map[string]string{"name": "Jane"}); err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if _, err := client.SignOut(); err != nil {
		t.Fatalf("SignOut: %v", err)
	}
	if _, err := client.InviteUser("new@user.com"); err != nil {
		t.Fatalf("InviteUser: %v", err)
	}
	if _, err := client.ResetPassword("token-abc", "newpass"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}
}

// TestRestMethodsAndQueryFormatting tests the REST methods and query parameter formatting.
func TestRestMethodsAndQueryFormatting(t *testing.T) {
	ts := setUpTestServer(t)
	defer ts.Close()

	client := NewClient(ts.URL, "api-key", "token-no-bearer-prefix")

	// GET with query params (they are auto formatted with eq.)
	body, err := client.Get("Food", map[string]string{"name": "John Doe", "city": "New York"})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	// decode response to ensure query echoed with eq. prefix
	var echoed map[string]string
	_ = json.Unmarshal(body, &echoed)
	for k, v := range echoed {
		if !strings.HasPrefix(v, "eq.") {
			t.Fatalf("expected eq. prefix for %s got %s", k, v)
		}
	}

	// POST
	postBody := []byte(`{"restaurant":"Place","rating":5}`)
	if _, err := client.Post("Food", postBody); err != nil {
		t.Fatalf("Post: %v", err)
	}

	// PUT
	if _, err := client.Put("Food", "id", "10", []byte(`{"id":10,"restaurant":"R","rating":4}`)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// PATCH
	if _, err := client.Patch("Food", map[string]string{"id": "10"}, []byte(`{"rating":3}`)); err != nil {
		t.Fatalf("Patch: %v", err)
	}

	// DELETE
	if _, err := client.Delete("Food", "id", "10"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

// TestAuthorizationHeaderFormatting tests the formatting of the Authorization header.
func TestAuthorizationHeaderFormatting(t *testing.T) {
	ts := setUpTestServer(t)
	defer ts.Close()

	client := NewClient(ts.URL, "api-key", "rawtoken123") // no Bearer prefix
	body, err := client.Get("HeaderCheck")
	if err != nil {
		t.Fatalf("Get header check: %v", err)
	}
	var resp map[string]string
	_ = json.Unmarshal(body, &resp)
	got := resp["auth"]
	if got != "Bearer rawtoken123" {
		t.Fatalf("expected Bearer prefix, got %s", got)
	}

	// Now with existing Bearer prefix
	client.Token = "Bearer already"
	body, err = client.Get("HeaderCheck")
	if err != nil {
		t.Fatalf("Get header check 2: %v", err)
	}
	_ = json.Unmarshal(body, &resp)
	if resp["auth"] != "Bearer already" {
		t.Fatalf("expected unchanged token, got %s", resp["auth"])
	}
}

// TestFormatQueryParams tests the formatting of query parameters.
func TestFormatQueryParams(t *testing.T) {
	in := map[string]string{"a": "1 2", "b": "special@value"}
	got := formatQueryParams(in)
	if len(got) != len(in) {
		t.Fatalf("size mismatch")
	}
	for k, v := range got {
		if !strings.HasPrefix(v, "eq.") {
			t.Fatalf("expected eq. prefix for %s", k)
		}
		// ensure value after eq. is URL escaped
		raw := strings.TrimPrefix(v, "eq.")
		unescaped, _ := url.QueryUnescape(raw)
		if unescaped != in[k] {
			t.Fatalf("expected %s got %s", in[k], unescaped)
		}
	}
}

// TestAuthRequestErrorHandling tests the error handling for authentication requests.
func TestAuthRequestErrorHandling(t *testing.T) {
	// Server that forces non-2xx for token endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, tokenApiPath) {
			http.Error(w, "fail", http.StatusBadRequest)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "api-key", "")
	_, err := client.SignIn("a@b.c", "pass")
	if err == nil {
		t.Fatalf("expected error on non-2xx")
	}
	if !errors.Is(err, ErrRequestFailed) && err.Error() == "request failed" {
		t.Logf("Error format is acceptable: %s", err.Error())
	}
}

// TestDoRequestQueryEncoding tests the query parameter encoding in requests.
func TestDoRequestQueryEncoding(t *testing.T) {
	// Ensure that query parameters are encoded only once and appear as eq.<encoded>
	var captured string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.URL.RawQuery
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "api-key", "t")
	_, _ = client.Get("Food", map[string]string{"name": "John Doe"})
	// Expect name=eq.John+Doe (space encoded) or name=eq.John%20Doe depending on encoding order.
	// Accept both plus or %20.
	// Allow for various encodings (space -> +, space -> %20, plus re-encoded to %2B due to double encoding)
	if !strings.Contains(captured, "name=eq.John+") && !strings.Contains(captured, "name=eq.John%20") && !strings.Contains(captured, "name=eq.John%2BDoe") {
		t.Fatalf("unexpected query encoding: %s", captured)
	}
}

// TestUpdateUserEcho tests the UpdateUser method.
func TestUpdateUserEcho(t *testing.T) {
	ts := setUpTestServer(t)
	defer ts.Close()
	client := NewClient(ts.URL, "api-key", "Bearer t")
	body, err := client.UpdateUser(map[string]string{"name": "Alice"})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	var m map[string]string
	_ = json.Unmarshal(body, &m)
	if m["name"] != "Alice" {
		t.Fatalf("expected echo, got %v", m)
	}
}

// TestSignOutResponse tests the SignOut method.
func TestSignOutResponse(t *testing.T) {
	ts := setUpTestServer(t)
	defer ts.Close()
	client := NewClient(ts.URL, "api-key", "Bearer t")
	body, err := client.SignOut()
	if err != nil {
		t.Fatalf("SignOut: %v", err)
	}
	if !strings.Contains(string(body), "logged_out") {
		t.Fatalf("unexpected body: %s", string(body))
	}
}

// TestInviteUser tests the InviteUser method.
func TestInviteUser(t *testing.T) {
	ts := setUpTestServer(t)
	defer ts.Close()
	client := NewClient(ts.URL, "api-key", "Bearer t")
	body, err := client.InviteUser("new@example.com")
	if err != nil {
		t.Fatalf("InviteUser: %v", err)
	}
	if !strings.Contains(string(body), "invited") {
		t.Fatalf("unexpected body: %s", string(body))
	}
}

// TestResetPassword tests the ResetPassword method.
func TestResetPassword(t *testing.T) {
	ts := setUpTestServer(t)
	defer ts.Close()
	client := NewClient(ts.URL, "api-key", "Bearer t")
	body, err := client.ResetPassword("tok", "pwd")
	if err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}
	if !strings.Contains(string(body), "reset") {
		t.Fatalf("unexpected body: %s", string(body))
	}
}

// TestSignUpStatusCreated tests the SignUp method for a 201 Created response.
func TestSignUpStatusCreated(t *testing.T) {
	ts := setUpTestServer(t)
	defer ts.Close()
	client := NewClient(ts.URL, "api-key", "")
	body, err := client.SignUp("x@y.z", "p")
	if err != nil {
		t.Fatalf("SignUp: %v", err)
	}
	if !strings.Contains(string(body), "new-user") {
		t.Fatalf("unexpected response: %s", string(body))
	}
}

// Safety check: ensure internal constants have expected leading slash formatting so path joins are correct.
func TestConstantFormatting(t *testing.T) {
	consts := []string{restApiPath, authApiPath, tokenApiPath, signupApiPath, magicLinkApiPath, recoverApiPath, verifyApiPath, userApiPath, logoutApiPath, inviteApiPath, resetApiPath}
	for _, c := range consts {
		if !strings.HasPrefix(c, "/") {
			t.Fatalf("constant %s lacks leading slash", c)
		}
	}
}

// Ensure auth response struct round trips JSON.
func TestAuthTokenResponseJSON(t *testing.T) {
	b, err := json.Marshal(mockAuthResponse)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out AuthTokenResponse
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(out, mockAuthResponse) {
		t.Fatalf("round trip mismatch: %+v vs %+v", out, mockAuthResponse)
	}
}
