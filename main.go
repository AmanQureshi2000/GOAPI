package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Item represents the data structure for our DB table
type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var db *sql.DB

func main() {
	// 1. Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 2. Build Connection String
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
	)

	// 3. Connect to Database
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	port := os.Getenv("PORT")
    if port == "" {
        port = "8080" // Default for local dev
    }

	// 4. Define Routes
	http.HandleFunc("/items", handleItems)

	fmt.Printf("Server starting on port %s\n", port)
    // IMPORTANT: Listen on "0.0.0.0" so Render can route traffic to it
    log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

func handleItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	switch r.Method {
	case "GET":
		if id != 0 {
			var item Item
			err := db.QueryRow("SELECT id, name, description FROM items WHERE id = $1", id).Scan(&item.ID, &item.Name, &item.Description)
			if err != nil {
				http.Error(w, "Item not found", http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(item)
		} else {
			rows, _ := db.Query("SELECT id, name, description FROM items")
			var items []Item
			for rows.Next() {
				var item Item
				rows.Scan(&item.ID, &item.Name, &item.Description)
				items = append(items, item)
			}
			json.NewEncoder(w).Encode(items)
		}

	case "POST":
		var item Item
		json.NewDecoder(r.Body).Decode(&item)
		if item.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		_, err := db.Exec("INSERT INTO items (name, description) VALUES ($1, $2)", item.Name, item.Description)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Item created successfully"})

	case "PUT":
		if id == 0 {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}
		var item Item
		json.NewDecoder(r.Body).Decode(&item)
		_, err := db.Exec("UPDATE items SET name = $1, description = $2 WHERE id = $3", item.Name, item.Description, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Item updated successfully"})

	case "DELETE":
		if id == 0 {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}
		db.Exec("DELETE FROM items WHERE id = $1", id)
		json.NewEncoder(w).Encode(map[string]string{"message": "Item deleted successfully"})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}