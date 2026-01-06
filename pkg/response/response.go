package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success: statusCode < 400,
		Data:    data,
	})
}

func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

func Error(w http.ResponseWriter, statusCode int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   err,
	})
}

func BadRequest(w http.ResponseWriter, err string) {
	Error(w, http.StatusBadRequest, err)
}

func Unauthorized(w http.ResponseWriter, err string) {
	Error(w, http.StatusUnauthorized, err)
}

func Forbidden(w http.ResponseWriter, err string) {
	Error(w, http.StatusForbidden, err)
}

func NotFound(w http.ResponseWriter, err string) {
	Error(w, http.StatusNotFound, err)
}

func InternalError(w http.ResponseWriter, err string) {
	Error(w, http.StatusInternalServerError, err)
}
