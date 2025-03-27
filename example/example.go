package example

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	supabase "github.com/jtclarkjr/supabase-go-rest"
	"github.com/jtclarkjr/supabase-go-rest/example/utils"
)

var (
	supabaseUrl = "https://your-project.supabase.co"
	supabaseKey = "your-supabase-api-key"
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
	Id         *int64    `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Restaurant string    `json:"restaurant"`
	Rating     int64     `json:"rating"`
	FoodName   string    `json:"food_name"`
	Opinion    string    `json:"opinion"`
	Image      string    `json:"image"`
}

// AuthTokenResponse represents the response from the /token endpoint
type AuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// Handler for GET
// https://docs.postgrest.org/en/v12/references/api/tables_views.html#get
func getFoodHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Authorization token missing", http.StatusUnauthorized)
		return
	}

	client := supabase.NewClient(supabaseUrl, supabaseKey, token)

	query := r.URL.Query()
	queryParams := make(map[string]string)
	for key := range query {
		queryParams[key] = query.Get(key)
	}
	// In query can do many things native to postgrest
	// Can sort using a column using like ?order=created_at.desc
	// Can filter using a column name ?food_name="pizza"

	body, err := client.Get("Food", queryParams)
	if err != nil {
		http.Error(w, "Error fetching data from Supabase", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(body)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handler for POST
// https://docs.postgrest.org/en/v12/references/api/tables_views.html#create
func createFoodHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization token missing", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// userId for RLS set for auth.id action only
	// ExtractUserId not included in example but this pull id from token
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

	jsonData, err := json.Marshal(food)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	log.Printf("Request payload: %s", string(jsonData))

	body, err := client.Post("Food", jsonData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create food data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(body)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handler for PUT
// https://docs.postgrest.org/en/v12/references/api/tables_views.html#put
func putFoodHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization token missing", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// userId for RLS set for auth.id action only
	// ExtractUserId not included in example but this pull id from token
	userID, err := utils.ExtractUserId(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	itemId := chi.URLParam(r, "itemId")
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
	foodId, err := strconv.ParseInt(itemId, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}
	food.Id = &foodId

	jsonData, err := json.Marshal(food)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	// Ensure the primary key is included in the request body
	primaryKey := "id"
	body, err := client.Put("Food", primaryKey, itemId, jsonData)
	if err != nil {
		log.Printf("Supabase PUT request error: %v", err)
		http.Error(w, fmt.Sprintf("Failed to update food data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(body)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handler for PATCH
// https://docs.postgrest.org/en/v12/references/api/tables_views.html#update
func patchFoodHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization token missing", http.StatusUnauthorized)
		return
	}

	client := supabase.NewClient(supabaseUrl, supabaseKey, authHeader)

	query := r.URL.Query()
	queryParams := make(map[string]string)
	for key := range query {
		queryParams[key] = query.Get(key)
	}

	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	jsonData, err := json.Marshal(updateData)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	body, err := client.Patch("Food", queryParams, jsonData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update food data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(body)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handler for DELETE
// https://docs.postgrest.org/en/v12/references/api/tables_views.html#delete
func deleteFoodHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")

	// Extract the item ID from the URL path
	itemId := chi.URLParam(r, "itemId")
	if itemId == "" {
		http.Error(w, "Missing item ID", http.StatusBadRequest)
		return
	}

	client := supabase.NewClient(supabaseUrl, supabaseKey, token)
	primaryKey := "id"
	body, err := client.Delete("Food", primaryKey, itemId)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete food data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(body)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handler for POST /auth/token (Login via Email and Password)
func authTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload supabase.TokenRequestPayload

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure email and password are provided
	if payload.Email == "" || payload.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	client := supabase.NewClient(supabaseUrl, supabaseKey, "") // Empty token initially

	// Perform SignIn
	authResponse, err := client.SignIn(payload.Email, payload.Password)
	if err != nil {
		// Handle specific errors from the supabase package
		if errors.Is(err, supabase.ErrRequestFailed) {
			http.Error(w, fmt.Sprintf("Supabase request failed: %v", err), http.StatusBadGateway)
			return
		}
		if errors.Is(err, supabase.ErrInvalidResponse) {
			http.Error(w, fmt.Sprintf("Invalid response from Supabase: %v", err), http.StatusInternalServerError)
			return
		}

		// Generic error fallback
		http.Error(w, fmt.Sprintf("Failed to authenticate: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the token response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authResponse); err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/v1", func(r chi.Router) {
		r.Post("/auth/token", authTokenHandler)

		r.Get("/food", getFoodHandler)
		r.Post("/food", createFoodHandler)
		r.Put("/food/{itemId}", putFoodHandler)
		r.Patch("/food", patchFoodHandler)
		r.Delete("/food/{itemId}", deleteFoodHandler)
	})

	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", r)
}
