package main

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var accessTokenExpiration = (time.Minute * 5)
var refreshTokenExpiration = (time.Hour * 24)

type AuthMiddleware struct {
	next           http.Handler
	db             *pgxpool.Pool
	accountHandler *AccountHandler
	accessSecret   string
	refreshSecret  string
}

type Claims struct {
	UserID int `json:"userid"`
	jwt.RegisteredClaims
}

type AuthDetails struct {
	UserID   int
	Password string
}

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
	RestrictedPathRE = regexp.MustCompile(`^\/accounts\/.*$`)
	LoginPathRE      = regexp.MustCompile(`^\/login\/?$`)
	RegisterPathRE   = regexp.MustCompile(`^\/register\/?$`)
)

func (h *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	switch {

	// REGISTER ROUTE
	case RegisterPathRE.MatchString(url) && r.Method == http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "error parsing form", http.StatusBadRequest)
			return
		}
		// Validate parameters
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")
		if strings.TrimSpace(email) == "" ||
			strings.TrimSpace(username) == "" ||
			strings.TrimSpace(password) == "" {
			http.Error(w, "invalid email/username/password", http.StatusBadRequest)
		}
		// Create account
		userID, err := h.accountHandler.CreateAccount(email, username, password)
		if err != nil || userID < 0 {
			http.Error(w, fmt.Sprintf("error creating account: %v", err), http.StatusInternalServerError)
		}
		// Login
		h.Login(w, r, userID)
		w.WriteHeader(http.StatusOK)
		return

	// LOGIN ROUTE
	case LoginPathRE.MatchString(url) && r.Method == http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "error parsing form", http.StatusBadRequest)
			return
		}
		emailOrUsername := r.FormValue("emailorusername")
		password := r.FormValue("password")

		userID, err := h.Authenticate(emailOrUsername, password)
		if err != nil {
			http.Error(w, fmt.Sprintf("error authenticating: %v", err), http.StatusInternalServerError)
			return
		}
		if userID < 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		h.Login(w, r, userID)
		w.WriteHeader(http.StatusOK)
		return

	// RESTRICTED ROUTE
	case RestrictedPathRE.MatchString(url):
		accessCookie, err := r.Cookie("access")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "missing access token", http.StatusUnauthorized)
				return
			}
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		accessToken := accessCookie.Value

		claims, err := h.GetClaimsFromToken(accessToken)
		if err != nil {
			http.Error(w, "error getting claims", http.StatusUnauthorized)
			return
		}
		// add claims to context
		ctx := context.WithValue(r.Context(), "claims", claims)
		r = r.WithContext(ctx)
		h.next.ServeHTTP(w, r)
		return

	// UNRESTRICTED ROUTE
	default:
		h.next.ServeHTTP(w, r)
	}

}

func (h *AuthMiddleware) Login(w http.ResponseWriter, r *http.Request, userID int) {
	accessToken, errGenAccess := h.GenerateAccessToken(userID)
	refreshToken, errGenRefresh := h.GenerateRefreshToken(userID)
	if errGenAccess != nil || errGenRefresh != nil {
		http.Error(w, "error generating tokens", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthMiddleware) Authenticate(emailOrUsername string, password string) (userID int, err error) {
	if strings.TrimSpace(emailOrUsername) == "" || strings.TrimSpace(password) == "" {
		return -1, fmt.Errorf("empty username or password")
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
		return -1, fmt.Errorf("error getting account: %w", errGetAccount)
	}
	// Authenticate
	if !VerifyPassword(password, authDetails.Password) {
		return -1, nil
	}
	return authDetails.UserID, nil
}

func (h *AuthMiddleware) GetAuthDetailsByUsername(username string) (*AuthDetails, error) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, password FROM accounts WHERE username=$1`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Account DNE
	if !rows.Next() {
		return nil, nil
	}
	var a AuthDetails
	err = rows.Scan(&a.UserID, &a.Password)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (h *AuthMiddleware) GetAuthDetailsByEmail(email string) (*AuthDetails, error) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, password FROM accounts WHERE email=$1`, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Account DNE
	if !rows.Next() {
		return nil, nil
	}
	var a AuthDetails
	err = rows.Scan(&a.UserID, &a.Password)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (h *AuthMiddleware) GenerateAccessToken(userid int) (string, error) {
	accessClaims := &Claims{
		UserID: userid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExpiration)),
		},
	}
	accesstoken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accesstokenstring, err := accesstoken.SignedString([]byte(h.accessSecret))
	if err != nil {
		return "", err
	}
	return accesstokenstring, nil
}

func isTokenValid(token *jwt.Token) error {
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.ExpiresAt.Time.Before(time.Now()) {
			return fmt.Errorf("token is expired")
		}
		return nil
	} else {
		return fmt.Errorf("invalid token")
	}
}

func (h *AuthMiddleware) GenerateRefreshToken(userid int) (string, error) {
	refreshClaims := &Claims{
		UserID: userid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenExpiration)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshtokenstring, err := refreshToken.SignedString([]byte(h.refreshSecret))
	if err != nil {
		return "", err
	}
	return refreshtokenstring, nil
}

func (h *AuthMiddleware) VerifyAccessToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.accessSecret), nil
	})
	if err != nil {
		return err
	}
	err = isTokenValid(token)
	if err != nil {
		return err
	}
	return nil
}

func (h *AuthMiddleware) VerifyRefreshToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.refreshSecret), nil
	})
	if err != nil {
		return err
	}
	err = isTokenValid(token)
	if err != nil {
		return err
	}
	return nil
}

func (h *AuthMiddleware) GetClaimsFromToken(tokenString string) (*Claims, error) {
	claims := Claims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.accessSecret), nil
	})
	if err != nil {
		return nil, err
	}
	err = isTokenValid(token)
	if err != nil {
		return nil, err
	}
	return &claims, nil
}

func (h *AuthMiddleware) RefreshAccess(refresh string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(refresh, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.refreshSecret), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	newAccess, err := h.GenerateAccessToken(claims.UserID)
	if err != nil {
		return "", err
	}
	return newAccess, nil
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
