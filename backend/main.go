package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Get environment
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	ACCESS_SECRET := os.Getenv("ACCESS_SECRET")
	REFRESH_SECRET := os.Getenv("REFRESH_SECRET")

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

	authMux := NewAuthMiddleware(mux, db, accountHandler, ACCESS_SECRET, REFRESH_SECRET)

	fmt.Println("Starting server on port 8080")
	err = http.ListenAndServe(":8080", authMux)
	if err != nil {
		log.Fatal(err)
	}
}
