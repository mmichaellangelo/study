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
type Claims struct {
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

	AuthorizationHeaderRe = regexp.MustCompile(`^Bearer (.+)$`)
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

	// REFRESH ROUTE
	case RefreshRouteRE.MatchString(url) && r.Method == http.MethodPost:
		currentRefresh, ae := h.GetAuthTokenFromHeader(r)
		if ae != nil {
			re := NewErrorResponse(http.StatusBadRequest, ae.Code, ae.Err)
			re.LogAndWrite(w, r)
			return
		}
		ae = h.VerifyRefresh(currentRefresh)
		if ae != nil {
			if ae.Code == InternalError {
				er := NewErrorResponse(http.StatusUnauthorized, InternalError, ae.Err)
				er.LogAndWrite(w, r)
				return
			} else {
				er := NewErrorResponse(http.StatusUnauthorized, ae.Code, ae.Err)
				er.LogAndWrite(w, r)
				return
			}
		}
		accessTokens, ae := h.RefreshAccess(currentRefresh)
		if ae != nil {
			er := NewErrorResponse(http.StatusUnauthorized, ae.Code, ae.Err)
			er.LogAndWrite(w, r)
			return
		}
		bytes, err := json.Marshal(accessTokens)
		if err != nil {
			er := NewErrorResponse(http.StatusInternalServerError, InternalError, err)
			er.LogAndWrite(w, r)
			return
		}
		w.Write(bytes)
		return

	// IDENTITY ROUTE
	case IdentityRouteRE.MatchString(url) && r.Method == http.MethodGet:
		log.Printf("Handled identity route for %s\n", clientIP)
		access, ae := h.GetAuthTokenFromHeader(r)
		if ae != nil {
			er := NewErrorResponse(http.StatusBadRequest, ae.Code, ae.Err)
			er.LogAndWrite(w, r)
			return
		}
		claims, ae := h.GetClaimsFromAccess(access)
		if ae != nil {
			er := NewErrorResponse(http.StatusUnauthorized, ae.Code, ae.Err)
			er.LogAndWrite(w, r)
			return
		}
		if claims == nil {
			er := NewErrorResponse(http.StatusBadRequest, BadClaims, fmt.Errorf("nil claims"))
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
		userID, ae := h.accountHandler.CreateAccount(email, username, password)
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
		h.SetAuthCookies(w, r, userID, username)
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

		userID, username, err := h.Authenticate(emailOrUsername, password)
		if err != nil {
			http.Error(w, fmt.Sprintf("error authenticating: %v", err), http.StatusInternalServerError)
			return
		}
		if userID < 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.SetAuthCookies(w, r, userID, username)
		return

	// LOGOUT ROUTE
	case LogoutPathRE.MatchString(url) && r.Method == http.MethodPost:
		log.Printf("Handled logout route for %s\n", clientIP)
		h.DeleteAuthCookies(w, r)
		return

	// RESTRICTED ROUTE
	default:
		log.Printf("Handled restricted route for %s\n", clientIP)
		access, err := h.GetAuthTokenFromHeader(r)
		if err != nil {
			ae, ok := err.(*AppError)
			if ok {
				er := NewErrorResponse(http.StatusUnauthorized, BadAuthHeader, ae.Err)
				er.LogAndWrite(w, r)
				return
			}
		}
		claims, err := h.GetClaimsFromAccess(access)
		if err != nil {
			er := NewErrorResponse(http.StatusUnauthorized, BadClaims, err)
			er.LogAndWrite(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		r = r.WithContext(ctx)
		h.next.ServeHTTP(w, r)
		return
	}

}

func (h *AuthMiddleware) GetAuthTokenFromHeader(r *http.Request) (string, *AppError) {
	authHeader := r.Header.Get("Authorization")
	if !AuthorizationHeaderRe.MatchString(authHeader) {
		return "", NewAppError(BadAuthHeader, "auth header bad format, does not match regex format")
	}
	groups := AuthorizationHeaderRe.FindStringSubmatch(authHeader)
	if len(groups) != 2 {
		return "", NewAppError(BadAuthHeader, "auth header bad format, more than 2 regex groups")
	}
	return groups[1], nil
}

// Sets both refresh and access cookies
func (h *AuthMiddleware) SetAuthCookies(w http.ResponseWriter, r *http.Request, userID int, username string) {
	accessCookie, errGenAccess := h.GenerateAccessCookie(userID, username)
	refreshCookie, errGenRefresh := h.GenerateRefreshCookie(userID, username)
	if errGenAccess != nil || errGenRefresh != nil {
		http.Error(w, "error generating tokens", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)
}

// Validates login credentials.
func (h *AuthMiddleware) Authenticate(emailOrUsername string, password string) (userID int, username string, err error) {
	if strings.TrimSpace(emailOrUsername) == "" || strings.TrimSpace(password) == "" {
		return -1, "", fmt.Errorf("empty username or password")
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
		return -1, "", fmt.Errorf("error getting account: %w", errGetAccount)
	}
	if authDetails == nil {
		return -1, "", fmt.Errorf("error getting auth details")
	}
	// Authenticate
	if !VerifyPassword(password, authDetails.Password) {
		return -1, "", nil
	}
	return authDetails.UserID, authDetails.Username, nil
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
func (h *AuthMiddleware) GenerateAccessCookie(userid int, username string) (*http.Cookie, error) {
	accessClaims := &Claims{
		UserID:   userid,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExpiration)),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(h.accessSecret))
	if err != nil {
		return nil, err
	}
	accessCookie := http.Cookie{
		Name:     "access",
		Value:    accessTokenString,
		Path:     "/",
		Expires:  accessClaims.ExpiresAt.Time,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	}
	return &accessCookie, nil
}

// Generates refresh token in the form of a cookie and stores the token in the database
func (h *AuthMiddleware) GenerateRefreshCookie(userid int, username string) (*http.Cookie, error) {
	refreshClaims := &Claims{
		UserID:   userid,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenExpiration)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(h.refreshSecret))
	if err != nil {
		return nil, err
	}
	refreshCookie := http.Cookie{
		Name:     "refresh",
		Value:    refreshTokenString,
		Path:     "/",
		Expires:  refreshClaims.ExpiresAt.Time,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	}
	// Add refresh token to database
	_, err = h.db.Exec(context.Background(),
		`INSERT INTO refreshtokens (account_id, token, expires)
		 VALUES($1, $2, $3)`, refreshClaims.UserID, refreshTokenString, refreshClaims.ExpiresAt.Time)
	if err != nil {
		log.Printf("error inserting refresh into table: %v", err)
		return nil, err
	}
	return &refreshCookie, nil
}

// Checks if an access or refresh token is still valid
func isTokenValid(token *jwt.Token) error {
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.ExpiresAt.Time.Before(time.Now()) {
			return NewAppError(TokenExpired, nil)
		}
		return nil
	} else {
		return NewAppError(TokenInvalid, nil)
	}
}

func (h *AuthMiddleware) VerifyAccess(tokenString string) error {
	var claims Claims
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.accessSecret), nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (h *AuthMiddleware) VerifyRefresh(tokenString string) *AppError {
	var claims Claims
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
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

func (h *AuthMiddleware) GetClaimsFromRefresh(tokenString string) (*Claims, *AppError) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
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
	err = isTokenValid(token)
	if err != nil {
		return nil, NewAppError(TokenInvalid, err)
	}
	return &claims, nil
}

func (h *AuthMiddleware) GetClaimsFromAccess(tokenString string) (*Claims, *AppError) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
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
	err = isTokenValid(token)
	if err != nil {
		return nil, NewAppError(TokenInvalid, err)
	}
	return &claims, nil
}

// Returns new access token
func (h *AuthMiddleware) RefreshAccess(currentRefresh string) (*AccessTokens, *AppError) {
	claims, err := h.GetClaimsFromRefresh(currentRefresh)
	switch {
	case err == nil:
		break
	case errors.Is(err, jwt.ErrTokenExpired):
		return nil, NewAppError(TokenExpired, err)
	default:
		return nil, NewAppError(InternalError, err)
	}

	err = h.VerifyRefresh(currentRefresh)
	if err != nil {
		return nil, err
	}
	newAccessCookie, err := h.GenerateAccessCookie(refreshClaims.UserID, refreshClaims.Username)
	if err != nil {
		log.Printf("error generating new access cookie for %s: %v\n", clientIP, err)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	newAccessClaims, err := h.GetClaimsFromAccess(newAccessCookie.Value)
	if err != nil {
		log.Printf("error getting claims from newly generated access token for %s: %v\n", clientIP, err)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	log.Printf("refreshed access for %s\n", clientIP)
	http.SetCookie(w, newAccessCookie)
	return newAccessClaims
}

func (h *AuthMiddleware) DeleteRefreshFromDB(token string) error {
	claims, err := h.GetClaimsFromRefresh(token)
	if err != nil {
		return &AppError{Code: DatabaseError, Err: err}
	}
	_, err = h.db.Exec(context.Background(),
		`DELETE FROM refreshtokens WHERE account_id=$1 AND token=$2`, claims.UserID, token)
	if err != nil {
		return &AppError{Code: DatabaseError, Err: err}
	}
	return nil
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
