package errs

import (
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
	ErrUserName = errors.New("username is required")
	ErrNoStars  = errors.New("stars <= rented books")
)

type ValidationErrorResponse struct {
	Message string `json:"message"`
	Errors  struct {
		AdditionalProperties string `json:"additionalProperties"`
	} `json:"errors"`
}
