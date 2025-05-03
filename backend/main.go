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
	setHandler := NewSetHandler(db, accountHandler)
	cardHandler := NewCardHandler(db)

	mux := http.NewServeMux()

	mux.Handle("/accounts", accountHandler)
	mux.Handle("/sets", setHandler)
	mux.Handle("/cards", cardHandler)

	authMux := NewAuthMiddleware(mux, db)

	http.ListenAndServe(":8080", authMux)
}
