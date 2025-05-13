package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var accessTokenExpiration = (time.Minute * 5)
var refreshTokenExpiration = (time.Hour * 24)

// Middleware to handle user auth
type AuthMiddleware struct {
	next           http.Handler
	db             *pgxpool.Pool
	accountHandler *AccountHandler
	accessSecret   string
	refreshSecret  string
}

// Claims to be included in restricted route context
type AccountClaims struct {
	UserID   int    `json:"userid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Details needed for user auth
type AuthDetails struct {
	UserID   int
	Username string
	Password string
}

type AccessTokens struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

// Creates a new Auth Middleware
func NewAuthMiddleware(handlerToWrap http.Handler,
	db *pgxpool.Pool, accountHandler *AccountHandler,
	accessSecret string, refreshSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		next:           handlerToWrap,
		db:             db,
		accountHandler: accountHandler,
		accessSecret:   accessSecret,
		refreshSecret:  refreshSecret,
	}
}

////////////
// ROUTES

var (
	LoginPathRE     = regexp.MustCompile(`^\/auth\/login\/?$`)
	LogoutPathRE    = regexp.MustCompile(`\/auth\/logout\/?$`)
	RegisterPathRE  = regexp.MustCompile(`^\/auth\/register\/?$`)
	IdentityRouteRE = regexp.MustCompile(`^\/auth\/me\/?$`)
	RefreshRouteRE  = regexp.MustCompile(`^\/auth\/refresh\/?$`)
)

type contextKey string

const clientIPKey contextKey = "clientIP"
const claimsKey contextKey = "claims"

// HTTP Routes
func (h *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path

	ORIGIN := os.Getenv("ORIGIN")

	w.Header().Set("Access-Control-Allow-Origin", ORIGIN)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}

	log.Printf("%s %s %s\n", clientIP, r.Method, url)
	ctx := context.WithValue(r.Context(), clientIPKey, clientIP)
	r = r.WithContext(ctx)

	// Handle OPTIONS preflight requests
	if r.Method == http.MethodOptions {
		fmt.Println("Handled OPTIONS request")
		w.WriteHeader(http.StatusOK) // Just need to return OK status with CORS headers
		return
	}

	switch {
	// IDENTITY ROUTE
	case IdentityRouteRE.MatchString(url) && r.Method == http.MethodGet:
		log.Printf("Handled identity route for %s\n", clientIP)
		claims, ae := h.RefreshAccess(w, r)
		if ae != nil {
			er := NewErrorResponse(http.StatusUnauthorized, ae.Code, ae.Err)
			er.LogAndWrite(w, r)
			return
		}
		data, err := json.Marshal(claims)
		if err != nil {
			er := NewErrorResponse(http.StatusInternalServerError, InternalError, err)
			er.LogAndWrite(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return

	// REGISTER ROUTE
	case RegisterPathRE.MatchString(url) && r.Method == http.MethodPost:
		log.Printf("Handled register route for %s\n", clientIP)
		err := r.ParseMultipartForm(0)
		if err != nil {
			er := NewErrorResponse(http.StatusBadRequest, InternalError, err)
			er.LogAndWrite(w, r)
			return
		}
		// Validate parameters
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")
		if strings.TrimSpace(email) == "" ||
			strings.TrimSpace(username) == "" ||
			strings.TrimSpace(password) == "" {
			er := NewErrorResponse(http.StatusBadRequest, BadRegistrationInfo, errors.New("bad email, username, and/or password"))
			er.LogAndWrite(w, r)
			return
		}
		// Create account
		_, ae := h.accountHandler.CreateAccount(email, username, password)
		if ae != nil {
			switch {
			case ae.Code == BadRegistrationInfo:
				er := NewErrorResponse(http.StatusBadRequest, ae.Code, ae.Err)
				er.LogAndWrite(w, r)
			case ae.Code == AccountWithEmailExists || ae.Code == AccountWithUsernameExists:
				er := NewErrorResponse(http.StatusForbidden, ae.Code, ae.Err)
				er.LogAndWrite(w, r)
				return
			}
		}
		// Login
		ae = h.Login(email, password, w)
		if ae != nil {
			er := NewErrorResponse(http.StatusUnauthorized, ae.Code, ae.Err)
			er.LogAndWrite(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		return

	// LOGIN ROUTE
	case LoginPathRE.MatchString(url) && r.Method == http.MethodPost:
		log.Printf("Handled login route for %s\n", clientIP)
		err := r.ParseMultipartForm(0)
		if err != nil {
			http.Error(w, "error parsing form", http.StatusBadRequest)
			return
		}
		emailOrUsername := r.FormValue("emailorusername")
		password := r.FormValue("password")

		ae := h.Login(emailOrUsername, password, w)
		if ae != nil {
			er := NewErrorResponse(http.StatusUnauthorized, ae.Code, ae.Err)
			er.LogAndWrite(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		return

	// LOGOUT ROUTE
	case LogoutPathRE.MatchString(url) && r.Method == http.MethodPost:
		log.Printf("Handled logout route for %s\n", clientIP)
		refresh, err := r.Cookie("refresh")
		if err != nil {
			er := NewErrorResponse(http.StatusBadRequest, NoRefreshToken, err)
			er.LogAndWrite(w, r)
			return
		}
		if refresh == nil {
			er := NewErrorResponse(http.StatusBadRequest, NoRefreshToken, err)
			er.LogAndWrite(w, r)
			return
		}
		ae := h.Logout(refresh.Value, w, r)
		if ae != nil {
			er := NewErrorResponse(http.StatusBadRequest, ae.Code, ae.Err)
			er.LogAndWrite(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		return

	// RESTRICTED ROUTE
	default:
		log.Printf("Handled restricted route for %s\n", clientIP)
		claims, ae := h.RefreshAccess(w, r)
		if ae != nil {
			er := NewErrorResponse(http.StatusUnauthorized, ae.Code, ae.Err)
			er.LogAndWrite(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		r = r.WithContext(ctx)
		h.next.ServeHTTP(w, r)
		return
	}

}

// Validates login credentials, stores refresh and sets cookies
func (h *AuthMiddleware) Login(emailOrUsername string, password string, w http.ResponseWriter) *AppError {
	if strings.TrimSpace(emailOrUsername) == "" || strings.TrimSpace(password) == "" {
		return NewAppError(IllegalArgument, "email or username is blank")
	}
	// Distinguish username/email
	_, errParseAddress := mail.ParseAddress(emailOrUsername)
	var errGetAccount error
	var authDetails *AuthDetails
	// Get password hash from db
	if errParseAddress != nil {
		// Username
		authDetails, errGetAccount = h.GetAuthDetailsByUsername(emailOrUsername)
	} else {
		// Email
		authDetails, errGetAccount = h.GetAuthDetailsByEmail(emailOrUsername)
	}
	if errGetAccount != nil {
		return NewAppError(NotFound, errGetAccount)
	}
	if authDetails == nil {
		return NewAppError(NotFound, "account not found")
	}
	// Authenticate
	if !VerifyPassword(password, authDetails.Password) {
		return NewAppError(PasswordIncorrect, nil)
	}
	refresh, ae := h.GenerateRefreshCookieAndStore(authDetails.UserID, authDetails.Username)
	if ae != nil {
		return ae
	}
	access, ae := h.GenerateAccessCookie(authDetails.UserID, authDetails.Username)
	if ae != nil {
		return ae
	}
	http.SetCookie(w, refresh)
	http.SetCookie(w, access)
	return nil
}

// Deletes refresh from db, sends cookie deletion request
func (h *AuthMiddleware) Logout(refresh string, w http.ResponseWriter, r *http.Request) *AppError {
	ae := h.DeleteRefreshFromDB(refresh)
	if ae != nil {
		return ae
	}
	h.DeleteAuthCookies(w, r)
	return nil
}

// Given username, returns auth details (userID and password)
func (h *AuthMiddleware) GetAuthDetailsByUsername(username string) (*AuthDetails, error) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, username, password FROM accounts WHERE username=$1`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Account DNE
	if !rows.Next() {
		return nil, nil
	}
	var a AuthDetails
	err = rows.Scan(&a.UserID, &a.Username, &a.Password)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Given email, returns auth details (userID and password)
func (h *AuthMiddleware) GetAuthDetailsByEmail(email string) (*AuthDetails, error) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, username, password FROM accounts WHERE email=$1`, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Account DNE
	if !rows.Next() {
		return nil, nil
	}
	var a AuthDetails
	err = rows.Scan(&a.UserID, &a.Username, &a.Password)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Generates access token in the form of a cookie
func (h *AuthMiddleware) GenerateAccessCookie(userid int, username string) (*http.Cookie, *AppError) {
	accessClaims := &AccountClaims{
		UserID:   userid,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExpiration)),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(h.accessSecret))
	if err != nil {
		return nil, NewAppError(InternalError, err)
	}
	accessCookie := http.Cookie{
		Name:     "access",
		Value:    accessTokenString,
		Path:     "/",
		Expires:  accessClaims.ExpiresAt.Time,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	}

	return &accessCookie, nil
}

// Generates refresh token in the form of a cookie and stores the token in the database
func (h *AuthMiddleware) GenerateRefreshCookieAndStore(userid int, username string) (*http.Cookie, *AppError) {
	refreshClaims := &AccountClaims{
		UserID:   userid,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenExpiration)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(h.refreshSecret))
	if err != nil {
		return nil, NewAppError(InternalError, err)
	}
	refreshCookie := http.Cookie{
		Name:     "refresh",
		Value:    refreshTokenString,
		Path:     "/",
		Expires:  refreshClaims.ExpiresAt.Time,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	}
	// Add refresh token to database
	_, err = h.db.Exec(context.Background(),
		`INSERT INTO refreshtokens (account_id, token, expires)
		 VALUES($1, $2, $3)`, refreshClaims.UserID, refreshTokenString, refreshClaims.ExpiresAt.Time)
	if err != nil {
		log.Printf("error inserting refresh into table: %v", err)
		return nil, NewAppError(DatabaseError, err)
	}
	return &refreshCookie, nil
}

func (h *AuthMiddleware) VerifyAccess(tokenString string) *AppError {
	var claims AccountClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(h.accessSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return NewAppError(TokenExpired, err)
		} else {
			return NewAppError(TokenInvalid, err)
		}
	}
	return nil
}

func (h *AuthMiddleware) VerifyRefresh(tokenString string) *AppError {
	var claims AccountClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(h.refreshSecret), nil
	})
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return NewAppError(TokenExpired, err)
		case errors.Is(err, jwt.ErrTokenInvalidClaims) || errors.Is(err, jwt.ErrInvalidType):
			return NewAppError(BadClaims, err)
		default:
			return NewAppError(InternalError, err)
		}
	}
	// Check that refresh token exists in database
	rows, err := h.db.Query(context.Background(),
		`SELECT token FROM refreshtokens
		 WHERE account_id=$1`, claims.UserID)
	if err != nil {
		return NewAppError(DatabaseError, err)
	}
	defer rows.Close()
	found := false
	for rows.Next() {
		var t string
		err := rows.Scan(&t)
		if err != nil {
			return NewAppError(DatabaseError, err)
		}
		if t == tokenString {
			found = true
			break
		}
	}
	if !found {
		return NewAppError(RefreshTokenInvalidated, "given refresh token has been invalidated")
	}
	return nil
}

func (h *AuthMiddleware) GetClaimsFromRefresh(tokenString string) (*AccountClaims, *AppError) {
	var claims AccountClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		return []byte(h.refreshSecret), nil
	})
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, NewAppError(TokenExpired, err)
		default:
			return nil, NewAppError(InternalError, err)
		}
	}
	return &claims, nil
}

func (h *AuthMiddleware) GetClaimsFromAccess(tokenString string) (*AccountClaims, *AppError) {
	var claims AccountClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		return []byte(h.accessSecret), nil
	})
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, &AppError{Code: TokenExpired, Err: err}
		default:
			return nil, &AppError{Code: InternalError, Err: err}
		}
	}
	return &claims, nil
}

/*
Refreshes access if needed.
Sets new cookies if refreshed and deletes old refresh from db.
Returns current access claims if still valid, otherwise returns new access claims.
*/
func (h *AuthMiddleware) RefreshAccess(w http.ResponseWriter, r *http.Request) (*AccountClaims, *AppError) {
	// Check if access is still valid
	currentAccess, err := r.Cookie("access")
	if err == nil && currentAccess != nil {
		ae := h.VerifyAccess(currentAccess.Value)
		switch {
		case ae == nil:
			// Access still valid >> send back access claims
			accessClaims, err := h.GetClaimsFromAccess(currentAccess.Value)
			if err != nil {
				return nil, NewAppError(BadClaims, err)
			}
			return accessClaims, nil
		case ae.Code == TokenExpired:
			// Access expired >> refresh
			break
		default:
			// Other error >> exit and return
			return nil, ae
		}
	}
	currentRefresh, err := r.Cookie("refresh")
	if err != nil {
		return nil, NewAppError(NoRefreshToken, err)
	}
	ae := h.VerifyRefresh(currentRefresh.Value)
	if ae != nil {
		return nil, ae
	}
	claims, ae := h.GetClaimsFromRefresh(currentRefresh.Value)
	if ae != nil {
		return nil, ae
	}
	ae = h.DeleteRefreshFromDB(currentRefresh.Value)
	if ae != nil {
		return nil, ae
	}
	newRefresh, ae := h.GenerateRefreshCookieAndStore(claims.UserID, claims.Username)
	if ae != nil {
		return nil, ae
	}
	newAccess, ae := h.GenerateAccessCookie(claims.UserID, claims.Username)
	if ae != nil {
		return nil, ae
	}
	newAccessClaims, ae := h.GetClaimsFromAccess(newAccess.Value)
	if ae != nil {
		return nil, ae
	}
	http.SetCookie(w, newAccess)
	http.SetCookie(w, newRefresh)
	return newAccessClaims, nil
}

func (h *AuthMiddleware) DeleteRefreshFromDB(token string) *AppError {
	claims, ae := h.GetClaimsFromRefresh(token)
	if ae != nil {
		return ae
	}
	_, err := h.db.Exec(context.Background(),
		`DELETE FROM refreshtokens WHERE account_id=$1 AND token=$2`, claims.UserID, token)
	if err != nil {
		return NewAppError(DatabaseError, err)
	}
	return nil
}

func (h *AuthMiddleware) DeleteAuthCookies(w http.ResponseWriter, r *http.Request) {
	clientIP := r.Context().Value(clientIPKey).(string)
	log.Printf("sending auth cookie delete request to %s\n", clientIP)
	access := http.Cookie{
		Name:     "access",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	}

	refresh := http.Cookie{
		Name:     "refresh",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	}

	// Set new cookies, trigger browser delete
	http.SetCookie(w, &access)
	http.SetCookie(w, &refresh)
}

// HashPassword generates a bcrypt hash for the given password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// VerifyPassword verifies if the given password matches the stored hash.
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
