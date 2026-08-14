// Package apperror define o tipo de erro de domínio usado pelos handlers HTTP
// para responder de forma consistente, sem inspecionar strings de erro.
package apperror

import "net/http"

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string { return e.Message }

func Validation(message string) *AppError {
	return &AppError{Code: "VALIDATION_ERROR", Message: message, HTTPStatus: http.StatusBadRequest}
}

func Internal(message string) *AppError {
	return &AppError{Code: "INTERNAL_ERROR", Message: message, HTTPStatus: http.StatusInternalServerError}
}
