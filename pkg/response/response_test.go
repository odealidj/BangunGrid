package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bangungrid/pkg/response"
)

type DummyData struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
}

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("X-Request-ID", "req-123")
	r.Header.Set("X-Trace-ID", "trace-456")

	data := DummyData{Name: "Semen Gresik 40kg", Price: 65000}
	response.JSONSuccess(w, r, http.StatusOK, "Berhasil memuat produk", data)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}

	var resp response.Response[DummyData]
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
	if resp.Data.Name != data.Name {
		t.Errorf("expected name %s, got %s", data.Name, resp.Data.Name)
	}
	if resp.Meta.RequestID != "req-123" {
		t.Errorf("expected request_id req-123, got %s", resp.Meta.RequestID)
	}
	if resp.Meta.TraceID != "trace-456" {
		t.Errorf("expected trace_id trace-456, got %s", resp.Meta.TraceID)
	}
}

func TestJSONError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)

	response.JSONError(w, r, http.StatusBadRequest, "Data tidak valid", "VALIDATION_ERROR", []response.ErrorDetail{
		{Field: "email", Message: "Format email tidak valid"},
	})

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", res.StatusCode)
	}

	var errResp response.ErrorResponse
	if err := json.NewDecoder(res.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Success {
		t.Errorf("expected success false, got true")
	}
	if errResp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %s", errResp.Error.Code)
	}
	if len(errResp.Error.Details) != 1 {
		t.Fatalf("expected 1 error detail, got %d", len(errResp.Error.Details))
	}
}
