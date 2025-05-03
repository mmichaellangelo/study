package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDBPool(ctx context.Context) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, "postgres://mike:password@localhost:5432/db")
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return pool, nil
}
