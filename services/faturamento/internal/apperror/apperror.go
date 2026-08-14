// Package apperror define um tipo de erro de domínio único, usado em toda a
// aplicação para que os handlers HTTP consigam mapear erros para status codes
// e mensagens consistentes, sem precisar inspecionar strings de erro.
package apperror

import "net/http"

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NotFound(message string) *AppError {
	return &AppError{Code: "NOT_FOUND", Message: message, HTTPStatus: http.StatusNotFound}
}

func Validation(message string) *AppError {
	return &AppError{Code: "VALIDATION_ERROR", Message: message, HTTPStatus: http.StatusBadRequest}
}

func Conflict(code, message string) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: http.StatusConflict}
}

func Unavailable(message string) *AppError {
	return &AppError{Code: "SERVICE_UNAVAILABLE", Message: message, HTTPStatus: http.StatusServiceUnavailable}
}

func Internal(message string) *AppError {
	return &AppError{Code: "INTERNAL_ERROR", Message: message, HTTPStatus: http.StatusInternalServerError}
}
