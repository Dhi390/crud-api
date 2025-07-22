package main

import (
	"log"
	"net/http"
	"time"

	"crud-api/config"
	"crud-api/handlers"

	"github.com/gorilla/mux"
)

func main() {
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer db.Close()

	//r := mux.NewRouter()
	r := mux.NewRouter().StrictSlash(true)

	// Routes
	r.HandleFunc("/users", handlers.CreateUser(db)).Methods("POST")
	r.HandleFunc("/users", handlers.GetAllUsers(db)).Methods("GET")
	r.HandleFunc("/users/{id}", handlers.GetUser(db)).Methods("GET")
	r.HandleFunc("/users/{id}", handlers.UpdateUser(db)).Methods("PUT")
	r.HandleFunc("/users/{id}", handlers.PatchUser(db)).Methods("PATCH")
	r.HandleFunc("/users/{id}", handlers.DeleteUser(db)).Methods("DELETE")

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("404 → %s %s", r.Method, r.URL.Path)
		http.Error(w, "Route not found", http.StatusNotFound)
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println(" Server running at http://localhost:8080")
	log.Fatal(srv.ListenAndServe())
}
