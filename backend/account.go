package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

///////////
// TYPES

type Account struct {
	ID       int       `json:"id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	Picture  string    `json:"picture"`
	Bio      string    `json:"bio"`
	Created  time.Time `json:"created"`
}

type AccountHandler struct {
	db *pgxpool.Pool
}

func NewAccountHandler(db *pgxpool.Pool) *AccountHandler {
	return &AccountHandler{db: db}
}

////////////
// ROUTES

var (
	AccountRE       = regexp.MustCompile(`^\/accounts\/?$`)
	AccountREWithID = regexp.MustCompile(`^\/accounts\/(\d+)$`)
)

func (h *AccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	switch {
	// CREATE ACCOUNT
	case AccountRE.MatchString(url) && r.Method == http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "error parsing form", http.StatusInternalServerError)
			return
		}
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")
		_, err = h.CreateAccount(email, username, password)
		if err != nil {
			http.Error(w, fmt.Sprintf("error creating account: %v", err), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		return

	// GET ACCOUNT
	case AccountREWithID.MatchString(url) && r.Method == http.MethodGet:
		id, err := getAccountIDFromURL(url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		account, err := h.GetAccountByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		bytes, err := json.Marshal(account)
		if err != nil {
			http.Error(w, "error marshalling json", http.StatusInternalServerError)
		}
		w.Write(bytes)
		return
	}
}

/////////////
// HELPERS

func getAccountIDFromURL(url string) (int, error) {
	groups := AccountRE.FindStringSubmatch(url)
	if len(groups) != 2 {
		return -1, fmt.Errorf("invalid URL")
	}
	id, err := strconv.Atoi(groups[1])
	if err != nil {
		return -1, fmt.Errorf("error parsing id as int: %w", err)
	}
	return id, nil
}

////////////
// CREATE

func (h *AccountHandler) CreateAccount(email string, username string, password string) (userID int, err error) {
	// Check that email, username, and password are not blank
	if strings.TrimSpace(email) == "" ||
		strings.TrimSpace(username) == "" ||
		strings.TrimSpace(password) == "" {
		return -1, fmt.Errorf("empty email, username, and/or password")
	}
	// Hash password
	hashed, err := HashPassword(password)
	if err != nil {
		return -1, fmt.Errorf("error hashing password: %w", err)
	}
	// Validate email address
	_, err = mail.ParseAddress(email)
	if err != nil {
		return -1, fmt.Errorf("invalid email")
	}
	// Check that email is unique
	acc, err := h.GetAccountByEmail(email)
	if err != nil {
		return -1, fmt.Errorf("error checking if account is unique: %w", err)
	}
	if acc != nil {
		return -1, fmt.Errorf("account with given email already exists")
	}
	// Check that username is unique
	acc, err = h.GetAccountByUsername(email)
	if err != nil {
		return -1, fmt.Errorf("error checking if account is unique: %w", err)
	}
	if acc != nil {
		return -1, fmt.Errorf("account with given username already exists")
	}
	// Add account to database
	rows, err := h.db.Query(context.Background(),
		`INSERT INTO accounts (email, username, password)
		 VALUES($1, $2, $3)
		 RETURNING id`, email, username, hashed)
	if err != nil {
		return -1, fmt.Errorf("error inserting account into database: %w", err)
	}
	if !rows.Next() {
		return -1, fmt.Errorf("error inserting account into database: %w", err)
	}
	var accID int
	err = rows.Scan(&accID)
	if err != nil {
		return -1, fmt.Errorf("error scanning rows: %w", err)
	}
	return accID, nil
}

//////////
// READ

func (h *AccountHandler) GetAllAccounts() (*[]Account, error) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, email, username, picture, bio, created
		 FROM accounts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var accounts []Account
	for rows.Next() {
		var a Account
		err = rows.Scan(&a.ID, &a.Email, &a.Username, &a.Picture, &a.Bio, &a.Created)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}

	return &accounts, nil
}

func (h *AccountHandler) GetAccountByID(id int) (*Account, error) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, email, username, picture, bio, created
		 FROM accounts WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var a Account
	if !rows.Next() {
		return nil, nil
	}
	err = rows.Scan(&a.ID, &a.Email, &a.Username, &a.Picture, &a.Bio, &a.Created)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, fmt.Errorf("multiple rows returned")
	}
	return &a, nil
}

func (h *AccountHandler) GetAccountByUsername(username string) (*Account, error) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, email, username, picture, bio, created
		 FROM accounts WHERE username=$1`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var a Account
	// Account DNE
	if !rows.Next() {
		return nil, nil
	}
	err = rows.Scan(&a.ID, &a.Email, &a.Username, &a.Picture, &a.Bio, &a.Created)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, fmt.Errorf("multiple rows returned")
	}
	return &a, nil
}

func (h *AccountHandler) GetAccountByEmail(email string) (*Account, error) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, email, username, picture, bio, created
		 FROM accounts WHERE email=$1`, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var a Account
	if !rows.Next() {
		return nil, nil
	}
	err = rows.Scan(&a.ID, &a.Email, &a.Username, &a.Picture, &a.Bio, &a.Created)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, fmt.Errorf("multiple rows returned")
	}
	return &a, nil
}

////////////
// UPDATE

func (h *AccountHandler) UpdateEmail(id int, email string) error {
	// Validate email
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email")
	}
	// Check that user doesn't already exist with new email
	acc, err := h.GetAccountByEmail(email)
	if err != nil {
		return fmt.Errorf("error querying db for account: %w", err)
	}
	if acc != nil {
		return fmt.Errorf("account with email already exists")
	}
	// Update email
	_, err = h.db.Exec(context.Background(),
		`UPDATE accounts
		 SET email=$1 WHERE id=$2`, email, id)
	if err != nil {
		return fmt.Errorf("error updating email: %w", err)
	}
	return nil
}

func (h *AccountHandler) UpdateUsername(id int, username string) error {
	// Check that user doesn't already exist with new username
	acc, err := h.GetAccountByUsername(username)
	if err != nil {
		return fmt.Errorf("error querying db for account: %w", err)
	}
	if acc != nil {
		return fmt.Errorf("account with username already exists")
	}
	// Update username
	_, err = h.db.Exec(context.Background(),
		`UPDATE accounts
		 SET username=$1 WHERE id=$2`, username, id)
	if err != nil {
		return fmt.Errorf("error updating username: %w", err)
	}
	return nil
}

func (h *AccountHandler) UpdateBio(id int, bio string) error {
	_, err := h.db.Exec(context.Background(),
		`UPDATE accounts
		 SET bio=$1 WHERE id=$2`, bio, id)
	if err != nil {
		return fmt.Errorf("error updating bio: %w", err)
	}
	return nil
}

func (h *AccountHandler) UpdatePicture(id int, picture string) error {
	_, err := h.db.Exec(context.Background(),
		`UPDATE accounts
		 SET picture=$1 WHERE id=$2`, picture, id)
	if err != nil {
		return fmt.Errorf("error updating picture: %w", err)
	}
	return nil
}

////////////
// DELETE

func (h *AccountHandler) DeletePicture(id int) error {
	_, err := h.db.Exec(context.Background(),
		`UPDATE accounts
		 SET picture=$1 WHERE id=$2`, nil, id)
	if err != nil {
		return fmt.Errorf("error deleting picture: %w", err)
	}
	return nil
}

func (h *AccountHandler) DeleteAccount(id int) error {
	// Check that account exists
	acc, err := h.GetAccountByID(id)
	if err != nil {
		return fmt.Errorf("error querying db for account: %w", err)
	}
	if acc == nil {
		return fmt.Errorf("account does not exist")
	}
	// Delete account
	_, err = h.db.Exec(context.Background(),
		`DELETE FROM accounts
		 WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("error deleting account: %w", err)
	}
	return nil
}
