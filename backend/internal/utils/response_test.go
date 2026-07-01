package utils

import (
	"encoding/json"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestSendSuccess(t *testing.T) {
	recorder := httptest.NewRecorder()
	data := map[string]string{"id": "123"}

	if err := SendSuccess(recorder, http.StatusCreated, "Created", data); err != nil {
		t.Fatalf("SendSuccess returned an error: %v", err)
	}

	assertResponse(t, recorder, http.StatusCreated, dto.Response{
		Status:  StatusSuccess,
		Message: "Created",
		Data:    map[string]any{"id": "123"},
		Errors:  nil,
	})
}

func TestSendError(t *testing.T) {
	recorder := httptest.NewRecorder()
	fieldErrors := map[string]string{"email": "is required"}

	if err := SendError(recorder, http.StatusBadRequest, "Validation failed", fieldErrors); err != nil {
		t.Fatalf("SendError returned an error: %v", err)
	}

	assertResponse(t, recorder, http.StatusBadRequest, dto.Response{
		Status:  StatusError,
		Message: "Validation failed",
		Data:    nil,
		Errors:  fieldErrors,
	})
}

func assertResponse(t *testing.T, recorder *httptest.ResponseRecorder, statusCode int, expected dto.Response) {
	t.Helper()

	if recorder.Code != statusCode {
		t.Fatalf("status code = %d, want %d", recorder.Code, statusCode)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var response dto.Response
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !reflect.DeepEqual(response, expected) {
		t.Fatalf("response = %#v, want %#v", response, expected)
	}
}
