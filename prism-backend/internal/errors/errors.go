package apperrors

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

type AppError struct {
	Code       string       `json:"code"`
	Message    string       `json:"message"`
	StatusCode int          `json:"-"`
	Details    []FieldError `json:"details,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func Internal(msg string) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    msg,
		StatusCode: http.StatusInternalServerError,
	}
}

func NotFound(msg string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    msg,
		StatusCode: http.StatusNotFound,
	}
}

func Conflict(msg string) *AppError {
	return &AppError{
		Code:       "CONFLICT",
		Message:    msg,
		StatusCode: http.StatusConflict,
	}
}

func Validation(fields ...FieldError) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Input tidak valid",
		StatusCode: http.StatusBadRequest,
		Details:    fields,
	}
}

func BusinessRule(msg string) *AppError {
	return &AppError{
		Code:       "BUSINESS_RULE_ERROR",
		Message:    msg,
		StatusCode: http.StatusUnprocessableEntity,
	}
}

func Unauthorized(msg string) *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    msg,
		StatusCode: http.StatusUnauthorized,
	}
}

func Forbidden(msg string) *AppError {
	return &AppError{
		Code:       "FORBIDDEN",
		Message:    msg,
		StatusCode: http.StatusForbidden,
	}
}

func FromPgError(err error) *AppError {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return Internal("Terjadi kesalahan database")
	}

	switch pgErr.Code {
	case "23505":
		return &AppError{
			Code:       "CONFLICT",
			Message:    "Data sudah ada",
			StatusCode: http.StatusConflict,
		}
	case "23503":
		return &AppError{
			Code:       "VALIDATION_ERROR",
			Message:    "Referensi data tidak valid",
			StatusCode: http.StatusBadRequest,
		}
	case "23514":
		return &AppError{
			Code:       "VALIDATION_ERROR",
			Message:    "Data tidak memenuhi aturan",
			StatusCode: http.StatusBadRequest,
		}
	default:
		return Internal("Terjadi kesalahan database")
	}
}
