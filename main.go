package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // PostgreSQL driver
	"log"
	"net/http"
)

const (
	host     = "host.docker.internal"
	port     = 5432
	user     = "postgres"
	password = "972011"
	dbname   = "keyvaluedb"
)

var db *sql.DB

// Initialize DB connection and create table if it doesn't exist
func initDB() {
	// Format the connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Open a connection to the database
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Error opening database connection: ", err)
	}

	// Ping the database to check if the connection is successful
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database: ", err)
	}
	fmt.Println("Successfully connected to PostgreSQL!")

	// Create table if it doesn't exist
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS key_value_store (
		key VARCHAR(255) PRIMARY KEY,
		value TEXT
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Error creating table: %v\n", err)
	}
	fmt.Println("Table 'key_value_store' created successfully (if it didn't exist).")
}

// API Endpoint: GET /getValue/{key}
func getValue(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	var value string
	err := db.QueryRow("SELECT value FROM key_value_store WHERE key = $1", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			value = "null" // Return "null" if the key doesn't exist
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"key": "%s", "value": "%s"}`, key, value)
}

// API Endpoint: POST /createValue
func createValue(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert key-value into the database
	_, err := db.Exec("INSERT INTO key_value_store (key, value) VALUES ($1, $2)", request.Key, request.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Key-Value pair %s:%s created successfully", request.Key, request.Value)
}

func main() {
	// Initialize DB connection and create table
	initDB()
	defer db.Close()

	// Set up API routes
	r := mux.NewRouter()
	r.HandleFunc("/getValue/{key}", getValue).Methods("GET")
	r.HandleFunc("/createValue", createValue).Methods("POST")

	// Start the HTTP server
	log.Fatal(http.ListenAndServe(":8080", r))
}
