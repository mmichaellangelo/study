package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Card struct {
	ID      int       `json:"id"`
	SetID   int       `json:"set_id"`
	Front   string    `json:"front"`
	Back    string    `json:"back"`
	Created time.Time `json:"created"`
}

type CardUpdate struct {
	ID    *int    `json:"id"`
	Front *string `json:"front"`
	Back  *string `json:"back"`
	Type  string  `json:"type"`
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

func (h *CardHandler) CreateCard(set_id int, front string, back string) (*Card, error) {
	rows, err := h.db.Query(context.Background(),
		`INSERT INTO cards 
		 (set_id, front, back)
		 VALUES($1, $2, $3)
		 RETURNING id, set_id, front, back, created`, set_id, front, back)
	if err != nil {
		return nil, fmt.Errorf("error creating card: %w", err)
	}
	defer rows.Close()
	rows.Next()
	var c Card
	err = rows.Scan(&c.ID, &c.SetID, &c.Front, &c.Back, &c.Created)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

//////////
// READ

func (h *CardHandler) GetCardsBySetID(set_id int) (*[]Card, error) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, set_id, front, back, created
		 FROM cards WHERE set_id=$1
		 ORDER BY id ASC `, set_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []Card
	for rows.Next() {
		var c Card
		err := rows.Scan(&c.ID, &c.SetID, &c.Front, &c.Back, &c.Created)
		if err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return &cards, nil
}

////////////
// UPDATE

func (h *CardHandler) UpdateCard(u CardUpdate) error {
	_, err := h.db.Exec(context.Background(),
		`UPDATE cards SET front=$1, back=$2
		 WHERE id=$3`, u.Front, u.Back, u.ID)
	if err != nil {
		return err
	}
	return nil
}

////////////
// DELETE

func (h *CardHandler) DeleteCard(card_id int) error {
	_, err := h.db.Exec(context.Background(),
		`DELETE FROM cards WHERE id=$1`, card_id)
	if err != nil {
		return err
	}
	return nil
}
