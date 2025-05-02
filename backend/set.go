package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Set struct {
	ID          int       `json:"id"`
	AccountID   int       `json:"account_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
	Cards       *[]Card   `json:"cards"`
}

type SetHandler struct {
	db             *pgxpool.Pool
	accountHandler *AccountHandler
}

func NewSetHandler(db *pgxpool.Pool, accountHandler *AccountHandler) *SetHandler {
	return &SetHandler{db: db, accountHandler: accountHandler}
}

func (h *SetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

////////////
// CREATE

func (h *SetHandler) CreateSet(account_id int) (id int, err error) {
	// Check that account exists
	// TODO perhaps make a SQL function that does this instead!
	rows, err := h.db.Query(context.Background(),
		`SELECT id FROM accounts WHERE id=$1`, account_id)
	defer rows.Close()
	if err != nil {
		return -1, fmt.Errorf("error querying account: %w", err)
	}
	if !rows.Next() {
		return -1, fmt.Errorf("account does not exist")
	}
	rows.Close()
	// Create set
	rows, err = h.db.Query(context.Background(),
		`INSERT INTO sets account_id
		 VALUES($1) RETURNING id`, account_id)
	defer rows.Close()
	if err != nil {
		return -1, fmt.Errorf("error inserting into database: %w", err)
	}
	if !rows.Next() {
		return -1, fmt.Errorf("error inserting into database")
	}
	set_id := -1
	err = rows.Scan(&set_id)
	if err != nil {
		return -1, fmt.Errorf("error scanning row: %w", err)
	}
	if set_id == -1 {
		return -1, fmt.Errorf("error getting set id")
	}
	return set_id, nil
}

//////////
// READ

func (h *SetHandler) GetSetByID(set_id int) (*Set, error) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, acocunt_id, name, description, created
		 FROM sets WHERE id=$1`, set_id)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("error getting set: %w", err)
	}
	if !rows.Next() {
		return nil, nil
	}
	var s Set
	err = rows.Scan(&s.ID, &s.AccountID, &s.Name, &s.Description, &s.Created)
	if err != nil {
		return nil, fmt.Errorf("error scanning row: %w", err)
	}
	return &s, nil
}

func (h *SetHandler) GetSetWithCards(set_id int) (*Set, error) {
	// Get set info
	set, err := h.GetSetByID(set_id)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, nil
	}
	// Get cards in set
	var cards []Card
	rows, err := h.db.Query(context.Background(),
		`SELECT id, set_id, front, back, created
		 FROM cards WHERE set_id=$1`, set_id)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("error getting cards from db: %w", err)
	}
	for rows.Next() {
		var c Card
		err := rows.Scan(&c.ID, &c.SetID, &c.Front, &c.Back, &c.Created)
		if err != nil {
			return nil, fmt.Errorf("error scanning card: %w", err)
		}
		cards = append(cards, c)
	}
	if len(cards) == 0 {
		set.Cards = nil
	} else {
		set.Cards = &cards
	}
	return set, nil
}

func (h *SetHandler) GetSetsByAccountID(account_id int) (*[]Set, error) {
	// Check that account exists
	acc, err := h.accountHandler.GetAccountByID(account_id)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("account does not exist")
	}
	// Get sets
	rows, err := h.db.Query(context.Background(),
		`SELECT id, account_id, name, description, created
		 FROM sets WHERE account_id=$1`, account_id)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("error scanning sets: %w", err)
	}
	var sets []Set
	for rows.Next() {
		var s Set
		err := rows.Scan(&s.ID, &s.AccountID, &s.Name, &s.Description, &s.Created)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		sets = append(sets, s)
	}
	if len(sets) == 0 {
		return nil, nil
	}
	return &sets, nil
}

////////////
// UPDATE

func (h *SetHandler) UpdateName(set_id int, name string) error {
	_, err := h.db.Exec(context.Background(),
		`UPDATE sets SET name=$1 WHERE id=$2`, name, set_id)
	if err != nil {
		return fmt.Errorf("error updating name: %w", err)
	}
	return nil
}

func (h *SetHandler) UpdateDescription(set_id int, description string) error {
	_, err := h.db.Exec(context.Background(),
		`UPDATE sets SET description=$1 WHERE id=$2`, description, set_id)
	if err != nil {
		return fmt.Errorf("error updating description: %w", err)
	}
	return nil
}

////////////
// DELETE

func (h *SetHandler) DeleteSet(set_id int) error {
	// Check exists
	set, err := h.GetSetByID(set_id)
	if err != nil {
		return fmt.Errorf("error querying set: %w", err)
	}
	if set == nil {
		return fmt.Errorf("set does not exist")
	}
	_, err = h.db.Exec(context.Background(),
		`DELETE FROM sets WHERE id=$1`, set_id)
	if err != nil {
		return fmt.Errorf("error deleting set: %w", err)
	}
	return nil
}
