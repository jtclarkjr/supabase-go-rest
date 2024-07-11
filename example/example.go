// Example pulled from working project

package example

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jtclarkjr/supabase-go-rest"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

func getFoodHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Authorization token missing", http.StatusUnauthorized)
		return
	}

	client := supabase.Client(supabaseUrl, supabaseKey, token)

	query := r.URL.Query()
	queryParams := make(map[string]string)
	for key := range query {
		// Modify the query parameter format to be compatible with Supabase
		queryParams[key] = fmt.Sprintf("eq.%s", url.QueryEscape(query.Get(key)))
	}

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

	client := supabase.Client(supabaseUrl, supabaseKey, authHeader)

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

// primary key need to be in request body for PUT
// https://docs.postgrest.org/en/v12/references/api/tables_views.html#put

func updateFoodHandler(w http.ResponseWriter, r *http.Request) {
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

	client := supabase.Client(supabaseUrl, supabaseKey, authHeader)

	var food FoodUpdate
	if err := json.NewDecoder(r.Body).Decode(&food); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	food.UserID = userID

	// Ensure the primary key is included in the request body
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

	url := fmt.Sprintf("Food?id=eq.%s", itemId)

	body, err := client.Put(url, jsonData)
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

func deleteFoodHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")

	// Extract the item ID from the URL path
	itemId := chi.URLParam(r, "itemId")
	if itemId == "" {
		http.Error(w, "Missing item ID", http.StatusBadRequest)
		return
	}

	client := supabase.Client(supabaseUrl, supabaseKey, token)
	body, err := client.Delete(fmt.Sprintf("Food?id=eq.%s", itemId))
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

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/food", getFoodHandler)
		r.Post("/food", createFoodHandler)
		r.Put("/food/{itemId}", updateFoodHandler)
		r.Delete("/food/{itemId}", deleteFoodHandler)
	})

	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", r)
}
