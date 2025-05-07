package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Card struct {
	ID      int         `json:"id"`
	SetID   int         `json:"set_id"`
	Front   pgtype.Text `json:"front"`
	Back    pgtype.Text `json:"back"`
	Created time.Time   `json:"created"`
}

type CardHandler struct {
	db *pgxpool.Pool
}

func NewCardHandler(db *pgxpool.Pool) *CardHandler {
	return &CardHandler{db: db}
}

////////////
// ROUTES

var (
	CardRE       = regexp.MustCompile(`^\/cards\/?$`)
	CardREWithID = regexp.MustCompile(`^\/cards\/(\d+)\/?$`)
)

func (h *CardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

////////////
// CREATE

func (h *CardHandler) CreateCard(stack_id int, front string, back string) error {
	if strings.TrimSpace(front) == "" && strings.TrimSpace(back) == "" {
		return fmt.Errorf("front and back are both empty")
	}
	_, err := h.db.Exec(context.Background(),
		`INSERT INTO cards stack_id, front, back
		 VALUES($1, $2, $3)`, stack_id, front, back)
	if err != nil {
		return fmt.Errorf("error creating card: %w", err)
	}
	return nil
}
