package main

import (
	"context"
	"log"
	"net/http"
)

func main() {
	// Init db connection
	db, err := InitDBPool(context.Background())
	if err != nil {
		log.Fatal("Error initializing DB connection: %w", err)
	}

	// Init handlers
	accountHandler := NewAccountHandler(db)

	mux := http.NewServeMux()

	mux.Handle("/accounts", accountHandler)

	http.ListenAndServe(":8080", mux)
}
