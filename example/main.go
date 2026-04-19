package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jtclarkjr/router-go"
	"github.com/jtclarkjr/router-go/middleware"
	supabase "github.com/jtclarkjr/supabase-go-rest"
	"github.com/jtclarkjr/supabase-go-rest/example/utils"
)

var (
	supabaseUrl    = "https://your-project.supabase.co"
	supabaseKey    = "your-supabase-api-key"
	oauthRedirect  = "https://your-app.example.com/auth/callback"
)

type FoodCreate struct {
	UserID     uuid.UUID `json:"user_id"`
	Restaurant string    `json:"restaurant"`
	Rating     int64     `json:"rating"`
	FoodName   string    `json:"food_name"`
	Opinion    string    `json:"opinion"`
	Image      string    `json:"image"`
}

type FoodUpdate struct {
	Id         *int64    `json:"id,omitempty"`
	UserID     uuid.UUID `json:"user_id"`
	Restaurant string    `json:"restaurant"`
	Rating     int64     `json:"rating"`
	FoodName   string    `json:"food_name"`
	Opinion    string    `json:"opinion"`
	Image      string    `json:"image"`
}

// Handler for GET /food
// https://docs.postgrest.org/en/v12/references/api/tables_views.html#get
func getFoodHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Authorization token missing", http.StatusUnauthorized)
		return
	}

	client := supabase.NewClient(supabaseUrl, supabaseKey, token)

	qb := client.From("Food")

	// Apply known query params from the request URL.
	// Unrecognised params are treated as equality filters.
	q := r.URL.Query()
	if sel := q.Get("select"); sel != "" {
		qb = qb.Select(sel)
	}
	if order := q.Get("order"); order != "" {
		// order param expected as "<column>.asc" or "<column>.desc"
		asc := !strings.HasSuffix(order, ".desc")
		col := strings.TrimSuffix(strings.TrimSuffix(order, ".desc"), ".asc")
		qb = qb.Order(col, map[string]bool{"ascending": asc})
	}
	if limit := q.Get("limit"); limit != "" {
		n := 0
		fmt.Sscanf(limit, "%d", &n)
		if n > 0 {
			qb = qb.Limit(n)
		}
	}
	reserved := map[string]bool{"select": true, "order": true, "limit": true}
	for key := range q {
		if !reserved[key] {
			qb = qb.Eq(key, q.Get(key))
		}
	}

	body, err := qb.Execute()
	if err != nil {
		http.Error(w, "Error fetching data from Supabase", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(body); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handler for POST /food
// https://docs.postgrest.org/en/v12/references/api/tables_views.html#create
func createFoodHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization token missing", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	userID, err := utils.ExtractUserId(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	client := supabase.NewClient(supabaseUrl, supabaseKey, authHeader)

	var food FoodCreate
	if err := json.NewDecoder(r.Body).Decode(&food); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	food.UserID = userID

	body, err := client.From("Food").Insert(food).Execute()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create food data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(body); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handler for PATCH /food/{itemId}
// https://docs.postgrest.org/en/v12/references/api/tables_views.html#update
func patchFoodHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization token missing", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	userID, err := utils.ExtractUserId(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	itemId := router.URLParam(r, "itemId")
	if itemId == "" {
		http.Error(w, "Missing item ID", http.StatusBadRequest)
		return
	}

	client := supabase.NewClient(supabaseUrl, supabaseKey, authHeader)

	var food FoodUpdate
	if err := json.NewDecoder(r.Body).Decode(&food); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	food.UserID = userID

	body, err := client.From("Food").Update(food).Eq("id", itemId).Execute()
	if err != nil {
		log.Printf("Supabase PATCH request error: %v", err)
		http.Error(w, fmt.Sprintf("Failed to update food data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(body); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handler for DELETE /food/{itemId}
// https://docs.postgrest.org/en/v12/references/api/tables_views.html#delete
func deleteFoodHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Authorization token missing", http.StatusUnauthorized)
		return
	}

	itemId := router.URLParam(r, "itemId")
	if itemId == "" {
		http.Error(w, "Missing item ID", http.StatusBadRequest)
		return
	}

	client := supabase.NewClient(supabaseUrl, supabaseKey, token)
	body, err := client.From("Food").Delete().Eq("id", itemId).Execute()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete food data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(body); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handler for POST /auth/login (email + password)
func authLoginHandler(w http.ResponseWriter, r *http.Request) {
	var payload supabase.TokenRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if payload.Email == "" || payload.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	client := supabase.NewClient(supabaseUrl, supabaseKey, "")
	authResponse, err := client.SignIn(payload.Email, payload.Password)
	if err != nil {
		if errors.Is(err, supabase.ErrRequestFailed) {
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to authenticate: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authResponse); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handler for POST /auth/anonymous
// Creates an anonymous session. Enable "Allow anonymous sign-ins" in Supabase Auth settings.
func anonLoginHandler(w http.ResponseWriter, r *http.Request) {
	client := supabase.NewClient(supabaseUrl, supabaseKey, "")
	authResponse, err := client.SignInAnonymously()
	if err != nil {
		if errors.Is(err, supabase.ErrRequestFailed) {
			http.Error(w, "Anonymous sign-ins are not enabled", http.StatusForbidden)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to create anonymous session: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authResponse); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handler for GET /auth/oauth?provider=github
// Returns the OAuth authorization URL; the client redirects the user's browser to it.
func oauthLoginHandler(w http.ResponseWriter, r *http.Request) {
	providerStr := r.URL.Query().Get("provider")
	if providerStr == "" {
		http.Error(w, "provider query param required (e.g. github, google, discord)", http.StatusBadRequest)
		return
	}

	client := supabase.NewClient(supabaseUrl, supabaseKey, "")

	// Generate a PKCE pair. Store the Verifier in your session so it can be
	// retrieved in the callback handler.
	pkce, err := supabase.GeneratePKCEPair()
	if err != nil {
		http.Error(w, "Failed to generate PKCE pair", http.StatusInternalServerError)
		return
	}

	// TODO: persist pkce.Verifier in a server-side session keyed by a state param.
	// Here we log it as a placeholder.
	log.Printf("PKCE verifier (store in session): %s", pkce.Verifier)

	authURL, err := client.GetOAuthURL(
		supabase.OAuthProvider(providerStr),
		oauthRedirect,
		nil, // use project default scopes
		&pkce,
	)
	if err != nil {
		http.Error(w, "Failed to build OAuth URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"url": authURL}); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handler for GET /auth/callback?code=...
// Called by Supabase after the user authorises with the OAuth provider.
func oauthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "code query param missing", http.StatusBadRequest)
		return
	}

	// TODO: retrieve the PKCE verifier from your session store using the state param.
	// Here we use a placeholder; replace with your real session lookup.
	codeVerifier := r.URL.Query().Get("code_verifier") // placeholder — use session in production

	client := supabase.NewClient(supabaseUrl, supabaseKey, "")
	authResponse, err := client.ExchangeCodeForSession(code, codeVerifier)
	if err != nil {
		if errors.Is(err, supabase.ErrRequestFailed) {
			http.Error(w, "OAuth code exchange failed", http.StatusUnauthorized)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to exchange code: %v", err), http.StatusInternalServerError)
		return
	}

	// authResponse.ProviderToken contains the raw provider token (e.g. GitHub token)
	// if you need to call provider APIs directly.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authResponse); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handler for POST /auth/id-token
// Exchange a provider ID token (e.g. from Google Sign-In SDK) for a Supabase session.
func idTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Provider string `json:"provider"`
		IdToken  string `json:"id_token"`
		Nonce    string `json:"nonce,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if payload.Provider == "" || payload.IdToken == "" {
		http.Error(w, "provider and id_token are required", http.StatusBadRequest)
		return
	}

	client := supabase.NewClient(supabaseUrl, supabaseKey, "")
	authResponse, err := client.SignInWithIdToken(
		supabase.OAuthProvider(payload.Provider),
		payload.IdToken,
		payload.Nonce,
	)
	if err != nil {
		if errors.Is(err, supabase.ErrRequestFailed) {
			http.Error(w, "ID token authentication failed", http.StatusUnauthorized)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to authenticate: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authResponse); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func main() {
	r := router.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/v1", func(r *router.Router) {
		// Email / password auth
		r.Post("/auth/login", authLoginHandler)

		// Anonymous auth
		r.Post("/auth/anonymous", anonLoginHandler)

		// OAuth — redirect-based flow (PKCE)
		r.Get("/auth/oauth", oauthLoginHandler)    // step 1: get redirect URL
		r.Get("/auth/callback", oauthCallbackHandler) // step 2: exchange code for session

		// OAuth — ID token flow (Google Sign-In SDK, Apple, etc.)
		r.Post("/auth/id-token", idTokenHandler)

		// Food CRUD
		r.Get("/food", getFoodHandler)
		r.Post("/food", createFoodHandler)
		r.Patch("/food/{itemId}", patchFoodHandler)
		r.Delete("/food/{itemId}", deleteFoodHandler)
	})

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
