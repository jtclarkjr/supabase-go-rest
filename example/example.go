package main

// import (
// "encoding/json"
// "fmt"
// "github.com/jtclarkjr/supabase-go-rest"
// "net/http"

// "github.com/go-chi/chi/v5"
// "github.com/go-chi/chi/v5/middleware"
// )

// var (
// 	supabaseBaseUrl = "https://your-project.supabase.co"
// 	supabaseKey     = "your-supabase-api-key"
// )

//  func getFoodHandler(w http.ResponseWriter, r *http.Request) {
//		token := r.Header.Get("Authorization")
//		if token == "" {
//			http.Error(w, "Authorization token missing", http.StatusUnauthorized)
//			return
//		}
//
//		client := supabase.Client(supabaseUrl, supabaseKey, token)
//
//		query := r.URL.Query()
//		queryParams := make(map[string]string)
//		for key := range query {
//			// Modify the query parameter format to be compatible with Supabase
//			queryParams[key] = fmt.Sprintf("eq.%s", url.QueryEscape(query.Get(key)))
//		}
//
//		body, err := client.Get("Food", queryParams)
//		if err != nil {
//			http.Error(w, "Error fetching data from Supabase", http.StatusInternalServerError)
//			return
//		}
//
//		w.Header().Set("Content-Type", "application/json")
//		_, err = w.Write(body)
//		if err != nil {
//			log.Printf("Error writing response: %v", err)
//		}
//  }

// func createFoodHandler(w http.ResponseWriter, r *http.Request) {
// 	token := r.Header.Get("Authorization")

// 	client := supabase.Client(supabaseBaseUrl, supabaseKey, token)

// 	var data map[string]interface{}
// 	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	jsonData, err := json.Marshal(data)
// 	if err != nil {
// 		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
// 		return
// 	}

// 	body, err := client.Post("food", jsonData)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Failed to create food data: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(body)
// }

// func updateFoodHandler(w http.ResponseWriter, r *http.Request) {
// 	token := r.Header.Get("Authorization")

// 	// Extract the item ID from the URL path
// 	itemId := chi.URLParam(r, "itemId")
// 	if itemId == "" {
// 		http.Error(w, "Missing item ID", http.StatusBadRequest)
// 		return
// 	}

// 	client := supabase.Client(supabaseBaseUrl, supabaseKey, token)

// 	var data map[string]interface{}
// 	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	jsonData, err := json.Marshal(data)
// 	if err != nil {
// 		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
// 		return
// 	}

// 	body, err := client.Put(fmt.Sprintf("food?id=eq.%s", itemId), jsonData)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Failed to update food data: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(body)
// }

// func deleteFoodHandler(w http.ResponseWriter, r *http.Request) {
// 	token := r.Header.Get("Authorization")

// 	// Extract the item ID from the URL path
// 	itemId := chi.URLParam(r, "itemId")
// 	if itemId == "" {
// 		http.Error(w, "Missing item ID", http.StatusBadRequest)
// 		return
// 	}

// 	client := supabase.Client(supabaseBaseUrl, supabaseKey, token)
// 	body, err := client.Delete(fmt.Sprintf("food?id=eq.%s", itemId))
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Failed to delete food data: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(body)
// }

func main() {
	// r := chi.NewRouter()
	// r.Use(middleware.Logger)

	// r.Route("/v1", func(r chi.Router) {
	// 	r.Get("/food", getFoodHandler)
	// 	r.Post("/food", createFoodHandler)
	// 	r.Put("/food/{itemId}", updateFoodHandler)
	// 	r.Delete("/food/{itemId}", deleteFoodHandler)
	// })

	// fmt.Println("Server is running on port 8080")
	// http.ListenAndServe(":8080", r)
}
