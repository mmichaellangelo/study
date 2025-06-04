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

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

///////////
// TYPES

type Account struct {
	ID       int         `json:"id"`
	Email    string      `json:"email"`
	Username string      `json:"username"`
	Picture  pgtype.Text `json:"picture"`
	Bio      pgtype.Text `json:"bio"`
	Created  time.Time   `json:"created"`
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
	AccountREWithID = regexp.MustCompile(`^\/accounts\/(\d+)\/?$`)
)

func (h *AccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(claimsKey).(*AccountClaims)
	url := r.URL.Path
	switch {

	// CREATE ACCOUNT ROUTE
	case AccountRE.MatchString(url) && r.Method == http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "error parsing form", http.StatusInternalServerError)
			return
		}
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")
		_, ae := h.CreateAccount(email, username, password)
		if ae != nil {
			er := NewErrorResponse(http.StatusInternalServerError, ae.Code, ae.Err)
			er.LogAndWrite(w, r)
			return
		}
		w.WriteHeader(http.StatusCreated)
		return

	// GET ACCOUNT ROUTE
	case AccountREWithID.MatchString(url) && r.Method == http.MethodGet:
		id, err := getAccountIDFromURL(url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Currently you may only access your own account
		// any other attempted access will return status unauthorized.
		if claims.UserID != id {
			er := NewErrorResponse(http.StatusUnauthorized, AccessNotAllowed, fmt.Errorf("account access unauthorized."))
			er.LogAndWrite(w, r)
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

	// DELETE ACCOUNT ROUTE
	case AccountREWithID.MatchString(url) && r.Method == http.MethodDelete:
		// Check that password authentication has been provided and is valid,
		// otherwise unauthorized
		if r.Context().Value(passwordAuthKey).(int) == -1 {
			er := NewErrorResponse(http.StatusUnauthorized, PasswordAuthRequired, nil)
			er.LogAndWrite(w, r)
			return
		}
		id, err := getAccountIDFromURL(url)
		if err != nil {
			er := NewErrorResponse(http.StatusBadRequest, IllegalArgument, err)
			er.LogAndWrite(w, r)
			return
		}
		if id != claims.UserID {
			er := NewErrorResponse(http.StatusUnauthorized, AccessNotAllowed, fmt.Errorf("tried to delete someone else's account"))
			er.LogAndWrite(w, r)
			return
		}
		err = h.DeleteAccount(id)
		if err != nil {
			er := NewErrorResponse(http.StatusInternalServerError, InternalError, err)
			er.LogAndWrite(w, r)
			return
		}
		DeleteAuthCookies(w, r)
		w.WriteHeader(http.StatusOK)
		return
	}
}

/////////////
// HELPERS

func getAccountIDFromURL(url string) (int, error) {
	groups := AccountREWithID.FindStringSubmatch(url)
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

/*
Creates a new account.

Params:

	email: email address
	username: username
	password: password unhashed

Returns:

	userID: id of newly created account or -1 if not successful
	err:
		BadRegistrationInfo if email, username, and/or password are empty or all whitespace
		AccountWithEmailExists if an account with the given email already exists
		AccountWithUsernameExists if an account with the given username already exists
		BadEmail if email address cannot be parsed
		InternalError error for an internal error
*/
func (h *AccountHandler) CreateAccount(email string, username string, password string) (account *Account, ae *AppError) {
	// Check that email, username, and password are not blank
	if strings.TrimSpace(email) == "" ||
		strings.TrimSpace(username) == "" ||
		strings.TrimSpace(password) == "" {
		return nil, NewAppError(BadRegistrationInfo, "blank username, email, or password")
	}
	// Hash password
	hashed, err := HashPassword(password)
	if err != nil {
		return nil, NewAppError(InternalError, err)
	}
	// Validate email address
	_, err = mail.ParseAddress(email)
	if err != nil {
		return nil, NewAppError(BadEmail, err)
	}
	// Check that email is unique
	acc, err := h.GetAccountByEmail(email)
	if err != nil {
		return nil, NewAppError(InternalError, err)
	}
	if acc != nil {
		return nil, NewAppError(AccountWithEmailExists, nil)
	}
	// Check that username is unique
	acc, err = h.GetAccountByUsername(username)
	if err != nil {
		return nil, NewAppError(InternalError, err)
	}
	if acc != nil {
		return nil, NewAppError(AccountWithUsernameExists, nil)
	}
	// Add account to database
	rows, err := h.db.Query(context.Background(),
		`INSERT INTO accounts (email, username, password)
		 VALUES($1, $2, $3)
		 RETURNING id, email, username, picture, bio created`, email, username, hashed)
	if err != nil {
		return nil, NewAppError(DatabaseError, err)
	}
	if !rows.Next() {
		return nil, NewAppError(DatabaseError, fmt.Errorf("unknown error occurred while inserting account into database"))
	}
	var a Account
	err = rows.Scan(&a.ID, &a.Email, &a.Username, &a.Picture, &a.Bio, &a.Created)
	if err != nil {
		return nil, NewAppError(DatabaseError, err)
	}
	return &a, nil
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

// Gets account corresponding to given id
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

// Gets account corresponding to given username
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

// Gets account corresponding to given email
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

// Updates an account's email
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

// Updates an account's username
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

// Updates an account's bio
func (h *AccountHandler) UpdateBio(id int, bio string) error {
	_, err := h.db.Exec(context.Background(),
		`UPDATE accounts
		 SET bio=$1 WHERE id=$2`, bio, id)
	if err != nil {
		return fmt.Errorf("error updating bio: %w", err)
	}
	return nil
}

// Updates an account's profile picture
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

// Deletes an account's profile picture
func (h *AccountHandler) DeletePicture(id int) error {
	_, err := h.db.Exec(context.Background(),
		`UPDATE accounts
		 SET picture=$1 WHERE id=$2`, nil, id)
	if err != nil {
		return fmt.Errorf("error deleting picture: %w", err)
	}
	return nil
}

// Deletes an account permanently
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
