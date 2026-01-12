package apierror

import "net/http"

type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}

/* -----------------------------
 * Helpers
 * ----------------------------- */

func New(status int, code, message string) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
	}
}

/* -----------------------------
 * Common errors
 * ----------------------------- */

var (
	ErrClientAlreadyExists = New(
		http.StatusConflict,
		"CLIENT_ALREADY_EXISTS",
		"Client with this ID already exists",
	)

	ErrInvalidClientID = New(
		http.StatusBadRequest,
		"INVALID_CLIENT_ID",
		"Client ID contains invalid characters",
	)

	ErrUnauthorized = New(
		http.StatusUnauthorized,
		"UNAUTHORIZED",
		"You are not authenticated",
	)

	ErrInternal = New(
		http.StatusInternalServerError,
		"INTERNAL_ERROR",
		"Something went wrong on our side",
	)
)
