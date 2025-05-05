package main

import (
	"context"
	"encoding/json"
	"errors"
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
	IdentityRouteRE  = regexp.MustCompile(`^\/me\/?$`)
)

func (h *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path

	switch {
	// IDENTITY ROUTE
	case IdentityRouteRE.MatchString(url) && r.Method == http.MethodGet:
		access := h.RefreshAccessIfNeeded(w, r)
		_, err := jwt.Parse(access.Value, func(t *jwt.Token) (interface{}, error) {
			return h.accessSecret, nil
		})
		if err != nil {
			http.Error(w, "error parsing access cookie", http.StatusUnauthorized)
		}
		claims, err := h.GetClaimsFromToken(access.Value)
		userID := claims.UserID
		data, err := json.Marshal(map[string]int{
			"userID": userID,
		})
		if err != nil {
			http.Error(w, "error marshalling json", http.StatusInternalServerError)
		}
		w.Write(data)
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
			return
		}
		// Create account
		userID, err := h.accountHandler.CreateAccount(email, username, password)
		if err != nil || userID < 0 {
			http.Error(w, fmt.Sprintf("error creating account: %v", err), http.StatusInternalServerError)
			return
		}
		// Login
		h.SetAuthCookies(w, r, userID)
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

		h.SetAuthCookies(w, r, userID)
		w.WriteHeader(http.StatusOK)
		return

	// RESTRICTED ROUTE
	case RestrictedPathRE.MatchString(url):
		h.RefreshAccessIfNeeded(w, r)
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

func (h *AuthMiddleware) SetAuthCookies(w http.ResponseWriter, r *http.Request, userID int) {
	accessCookie, errGenAccess := h.GenerateAccessCookie(userID)
	refreshCookie, errGenRefresh := h.GenerateRefreshCookie(userID)
	if errGenAccess != nil || errGenRefresh != nil {
		http.Error(w, "error generating tokens", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)
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
	if authDetails == nil {
		return -1, fmt.Errorf("error getting auth details")
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

func (h *AuthMiddleware) GenerateAccessCookie(userid int) (*http.Cookie, error) {
	accessClaims := &Claims{
		UserID: userid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExpiration)),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(h.accessSecret))
	if err != nil {
		return nil, err
	}
	accessCookie := &http.Cookie{
		Name:     "access",
		Value:    accessTokenString,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	return accessCookie, nil
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

func (h *AuthMiddleware) GenerateRefreshCookie(userid int) (*http.Cookie, error) {
	refreshClaims := &Claims{
		UserID: userid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenExpiration)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(h.refreshSecret))
	if err != nil {
		return nil, err
	}
	refreshCookie := &http.Cookie{
		Name:     "refresh",
		Value:    refreshTokenString,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	return refreshCookie, nil
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

func (h *AuthMiddleware) RefreshAccessIfNeeded(w http.ResponseWriter, r *http.Request) (accessCookie *http.Cookie) {
	// Check if access token is still valid
	currentAccessCookie, _ := r.Cookie("access")
	if currentAccessCookie != nil {
		access := currentAccessCookie.Value
		currentAccess, err := jwt.Parse(access, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.accessSecret), nil
		})
		switch {
		case currentAccess.Valid:
			// Valid token >> continue request
			return &http.Cookie{
				Name:     "access",
				Value:    currentAccess.Raw,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}
		case errors.Is(err, jwt.ErrTokenExpired):
			// Token expired >> continue to refresh
			break
		default:
			// Error other than token expired >> unauthorized
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return nil
		}
	}

	refreshCookie, _ := r.Cookie("refresh")
	if refreshCookie == nil {
		// Remove cookies, unauthorized
		h.DeleteAuthCookies(w, r)
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}

	claims := &Claims{}
	_, err := jwt.ParseWithClaims(refreshCookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.refreshSecret), nil
	})
	switch {
	case err == nil:
		break
	case errors.Is(err, jwt.ErrTokenExpired):
		h.DeleteAuthCookies(w, r)
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	default:
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}
	newAccessCookie, err := h.GenerateAccessCookie(claims.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	http.SetCookie(w, newAccessCookie)
	return newAccessCookie
}

func (h *AuthMiddleware) DeleteAuthCookies(w http.ResponseWriter, r *http.Request) {
	access := http.Cookie{
		Name:     "access",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	refresh := http.Cookie{
		Name:     "refresh",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

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
