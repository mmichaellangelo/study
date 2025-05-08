package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Set struct {
	ID          int         `json:"id"`
	AccountID   int         `json:"account_id"`
	Name        pgtype.Text `json:"name"`
	Description pgtype.Text `json:"description"`
	Created     time.Time   `json:"created"`
	Cards       *[]Card     `json:"cards"`
}

type SetHandler struct {
	db             *pgxpool.Pool
	accountHandler *AccountHandler
	cardHandler    *CardHandler
}

type CardData struct {
	Front pgtype.Text `json:"front"`
	Back  pgtype.Text `json:"back"`
}

func NewSetHandler(db *pgxpool.Pool, accountHandler *AccountHandler, cardHandler *CardHandler) *SetHandler {
	return &SetHandler{db: db, accountHandler: accountHandler, cardHandler: cardHandler}
}

var (
	SetRE           = regexp.MustCompile((`^\/sets\/?$`))
	SetREWithID     = regexp.MustCompile(`^\/sets\/(\d+)\/?$`)
	CardREWithSetID = regexp.MustCompile(`^\/sets\/(\d+)\/?$`)
)

func (h *SetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	claims := r.Context().Value("claims").(*Claims)
	switch {

	// CREATE SET ROUTE
	case SetRE.MatchString(url) && r.Method == http.MethodPost:
		setID, err := h.CreateSet(claims.UserID)
		if err != nil {
			http.Error(w, "error creating set", http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(map[string]int{
			"id": setID,
		})
		if err != nil {
			http.Error(w, "error creating set", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return

	// GET SET BY ID ROUTE
	case SetREWithID.MatchString(url) && r.Method == http.MethodGet:
		groups := SetREWithID.FindStringSubmatch(url)
		if len(groups) != 2 {
			http.Error(w, "invalid url", http.StatusBadRequest)
			return
		}
		setID, err := strconv.Atoi(groups[1])
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		set, err := h.GetSetByIDWithCards(setID)
		if err != nil {
			http.Error(w, "error getting set", http.StatusNotFound)
			log.Printf("error getting set: %v\n", err)
			return
		}
		data, err := json.Marshal(set)
		if err != nil {
			http.Error(w, "error marshalling json", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return

	// GET ACCOUNT SETS ROUTE
	case SetRE.MatchString(url) && r.Method == http.MethodGet:
		sets, err := h.GetSetsByAccountID(claims.UserID)
		if err != nil {
			http.Error(w, "error getting sets", http.StatusBadRequest)
			return
		}
		data, err := json.Marshal(sets)
		if err != nil {
			http.Error(w, "error marshalling json", http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return

	// INSERT CARD ROUTE
	case CardREWithSetID.MatchString(url) && r.Method == http.MethodPost:
		groups := CardREWithSetID.FindStringSubmatch(url)
		if len(groups) != 2 {
			log.Println("invalid URL")
			http.Error(w, "invalid URL", http.StatusBadRequest)
			return
		}
		set_id, err := strconv.Atoi(groups[1])
		if err != nil {
			log.Printf("error parsing id from url: %v\n", err)
			http.Error(w, "invalid ID", http.StatusBadRequest)
		}
		var cardData CardData
		defer r.Body.Close()
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("error reading body: %v\n", err)
			http.Error(w, "error reading body", http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(bytes, &cardData)
		if err != nil {
			log.Printf("error unmarshalling json: %v\n", err)
			http.Error(w, "error unmarshalling json", http.StatusBadRequest)
			return
		}
		card, err := h.cardHandler.CreateCard(set_id, cardData.Front, cardData.Back)
		if err != nil {
			log.Printf("error creating card: %v\n", err)
			http.Error(w, "error creating card", http.StatusInternalServerError)
			return
		}
		responseBytes, err := json.Marshal(card)
		if err != nil {
			log.Printf("error marshalling json for response: %v\n", err)
			http.Error(w, "error marshalling json for response", http.StatusInternalServerError)
			return
		}
		w.Write(responseBytes)
		return

	default:
		return
	}
}

////////////
// CREATE

func (h *SetHandler) CreateSet(account_id int) (id int, err error) {
	// Check that account exists
	// TODO perhaps make a SQL function that does this instead!
	rows, err := h.db.Query(context.Background(),
		`SELECT id FROM accounts WHERE id=$1`, account_id)
	if err != nil {
		return -1, fmt.Errorf("error querying account: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return -1, fmt.Errorf("account does not exist")
	}
	rows.Close()
	// Create set
	rows, err = h.db.Query(context.Background(),
		`INSERT INTO sets (account_id)
		 VALUES($1) RETURNING id`, account_id)
	if err != nil {
		return -1, fmt.Errorf("error inserting into database: %w", err)
	}
	defer rows.Close()
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
		`SELECT id, account_id, name, description, created
		 FROM sets WHERE id=$1`, set_id)
	if err != nil {
		return nil, fmt.Errorf("error getting set: %w", err)
	}
	defer rows.Close()
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

func (h *SetHandler) GetSetByIDWithCards(set_id int) (*Set, error) {
	set, err := h.GetSetByID(set_id)
	if err != nil {
		return nil, err
	}
	cards, err := h.cardHandler.GetCardsBySetID(set_id)
	if err != nil {
		return nil, err
	}
	set.Cards = cards
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
	if err != nil {
		return nil, fmt.Errorf("error scanning sets: %w", err)
	}
	defer rows.Close()
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
