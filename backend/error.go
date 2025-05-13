package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type ErrCode string

const (
	// Registration Codes
	AccountWithEmailExists    ErrCode = "ACCOUNT_EMAIL_EXISTS"
	AccountWithUsernameExists ErrCode = "ACCOUNT_USERNAME_EXISTS"
	BadRegistrationInfo       ErrCode = "BAD_REGISTRATION_INFO"
	BadEmail                  ErrCode = "BAD_EMAIL"

	// Access Codes
	NoAccessToken           ErrCode = "NO_ACCESS_TOKEN"
	NoRefreshToken          ErrCode = "NO_REFRESH_TOKEN"
	TokenExpired            ErrCode = "TOKEN_EXPIRED"
	TokenInvalid            ErrCode = "TOKEN_INVALID"
	RefreshTokenInvalidated ErrCode = "REFRESH_TOKEN_INVALIDATED"
	BadAuthHeader           ErrCode = "BAD_AUTH_HEADER"
	BadClaims               ErrCode = "BAD_CLAIMS"
	PasswordIncorrect       ErrCode = "PASSWORD_INCORRECT"

	// General Errors
	NotFound        ErrCode = "NOT_FOUND"
	IllegalArgument ErrCode = "ILLEGAL_ARGUMENT"
	InternalError   ErrCode = "INTERNAL_ERROR"
	DatabaseError   ErrCode = "DATABASE_ERROR"
)

/////////////////////
// INTERNAL ERRORS

type AppError struct {
	Code ErrCode
	Err  error
}

func NewAppError(code ErrCode, err any) *AppError {
	if err != nil {
		if e, ok := err.(error); ok {
			return &AppError{
				Code: code,
				Err:  e,
			}
		} else if e, ok := err.(string); ok {
			return &AppError{
				Code: code,
				Err:  errors.New(e),
			}
		} else {
			return &AppError{
				Code: code,
				Err:  fmt.Errorf("unknown"),
			}
		}
	} else {
		return &AppError{
			Code: code,
			Err:  fmt.Errorf("unspecified"),
		}
	}
}

func (ae *AppError) Error() string {
	return fmt.Sprintf("%s: %s", string(ae.Code), ae.Err.Error())
}

func (ae *AppError) Unwrap() error {
	return ae.Err
}

/////////////////////
// ERROR RESPONSES

type ErrorResponse struct {
	HttpStatus int     `json:"-"`
	ErrCode    ErrCode `json:"errcode,omitempty"`
	Err        error   `json:"-"`
}

// Creates a new ErrorResponse object
func NewErrorResponse(httpStatus int, errCode ErrCode, err error) *ErrorResponse {
	if err != nil {
		return &ErrorResponse{
			HttpStatus: httpStatus,
			ErrCode:    errCode,
			Err:        err,
		}
	} else {
		return &ErrorResponse{
			HttpStatus: httpStatus,
			ErrCode:    errCode,
			Err:        fmt.Errorf("unspecified"),
		}
	}
}

// Logs error and writes error response
func (e *ErrorResponse) LogAndWrite(w http.ResponseWriter, r *http.Request) {
	clientIP := r.Context().Value(clientIPKey).(string)
	log.Printf("%s, IP: %s, ERR: %s\n", e.ErrCode, clientIP, e.Err.Error())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.HttpStatus)
	errBytes, err := json.Marshal(e)
	if err != nil {
		log.Printf("error marshalling error: %v\n", err)
		return
	}
	w.Write(errBytes)
}
