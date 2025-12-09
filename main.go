package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Visitor struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host,omitempty"`
	CheckedIn time.Time `json:"checked_in_at"`
}

var db *sql.DB

func main() {
	var err error
	port := os.Getenv("PORT")
	dbURL := os.Getenv("DATABASE_URL")
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))

	http.HandleFunc("/api/checkin", handleCheckin)
	http.HandleFunc("/api/visitors", handleVisitors)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleCheckin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var v Visitor
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := db.QueryRow(
		"INSERT INTO visitors (name, host) VALUES ($1, $2) RETURNING id, checked_in_at",
		v.Name, v.Host,
	).Scan(&v.ID, &v.CheckedIn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func handleVisitors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT id, name, host, checked_in_at FROM visitors ORDER BY checked_in_at DESC")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	visitors := []Visitor{}
	for rows.Next() {
		var v Visitor
		if err := rows.Scan(&v.ID, &v.Name, &v.Host, &v.CheckedIn); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		visitors = append(visitors, v)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(visitors)
}
