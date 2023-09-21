package response

import (
	"encoding/json"
	"net/http"
	"time"
)

// Meta memuat metadata audit, pelacakan terdistribusi, dan timestamp
type Meta struct {
	RequestID string    `json:"request_id"`
	TraceID   string    `json:"trace_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Pagination memuat metadata halaman data array
type Pagination struct {
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
	HasNextPage bool  `json:"has_next_page"`
	HasPrevPage bool  `json:"has_prev_page"`
}

// ErrorDetail memuat rincian kesalahan per field
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// ErrorPayload memuat kode error mesin dan daftar rincian error
type ErrorPayload struct {
	Code    string        `json:"code"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// Response adalah amplop response objek tunggal standar BangunGrid
type Response[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Meta    Meta   `json:"meta"`
}

// PaginatedResponse adalah amplop response array berhalaman standar BangunGrid
type PaginatedResponse[T any] struct {
	Success    bool       `json:"success"`
	Message    string     `json:"message"`
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
	Meta       Meta       `json:"meta"`
}

// ErrorResponse adalah amplop response error standar BangunGrid
type ErrorResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Error   ErrorPayload `json:"error"`
	Meta    Meta         `json:"meta"`
}

// ExtractMeta mengambil informasi request_id dan trace_id dari HTTP header
func ExtractMeta(r *http.Request) Meta {
	requestID := r.Header.Get("X-Request-ID")
	traceID := r.Header.Get("X-Trace-ID")

	return Meta{
		RequestID: requestID,
		TraceID:   traceID,
		Timestamp: time.Now().UTC(),
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// JSONSuccess menulis response sukses objek tunggal
func JSONSuccess[T any](w http.ResponseWriter, r *http.Request, httpStatus int, message string, data T) {
	writeJSON(w, httpStatus, Response[T]{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    ExtractMeta(r),
	})
}

// JSONPaginated menulis response sukses array dengan pagination
func JSONPaginated[T any](w http.ResponseWriter, r *http.Request, httpStatus int, message string, data []T, pagination Pagination) {
	writeJSON(w, httpStatus, PaginatedResponse[T]{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
		Meta:       ExtractMeta(r),
	})
}

// JSONError menulis response error terstandarisasi
func JSONError(w http.ResponseWriter, r *http.Request, httpStatus int, message string, code string, details []ErrorDetail) {
	writeJSON(w, httpStatus, ErrorResponse{
		Success: false,
		Message: message,
		Error: ErrorPayload{
			Code:    code,
			Details: details,
		},
		Meta: ExtractMeta(r),
	})
}
