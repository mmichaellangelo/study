package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDBPool(ctx context.Context) (*pgxpool.Pool, error) {
	DB := os.Getenv("DB")
	pool, err := pgxpool.New(ctx, fmt.Sprintf("postgres://mike:password@%s/db", DB))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return pool, nil
}
