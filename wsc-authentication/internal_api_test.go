package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidPublicHost(t *testing.T) {
	for _, host := range []string{"nex.example.net", "203.0.113.10", "2001:db8::1"} {
		if !validPublicHost(host) {
			t.Errorf("expected %q to be a valid host", host)
		}
	}
	for _, host := range []string{"", "https://example.net", "example.net/path", "bad host", "-starts-with-hyphen.example", "bad..example"} {
		if validPublicHost(host) {
			t.Errorf("expected %q to be invalid", host)
		}
	}
}

func TestInternalNEXTokenAPIRequiresBearerToken(t *testing.T) {
	handler := newInternalNEXTokenHandler("test-only-secret-with-at-least-32-characters", "nex.example.net", 60014)
	request := httptest.NewRequest(http.MethodPost, "/internal/wsc/nex-token", bytes.NewBufferString(`{"pid":123}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing bearer token; got %d", response.Code)
	}
}

func TestInternalNEXTokenAPIRejectsWrongMethod(t *testing.T) {
	handler := newInternalNEXTokenHandler("test-only-secret-with-at-least-32-characters", "nex.example.net", 60014)
	request := httptest.NewRequest(http.MethodGet, "/internal/wsc/nex-token", nil)
	request.Header.Set("Authorization", "Bearer test-only-secret-with-at-least-32-characters")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405; got %d", response.Code)
	}
	if !strings.Contains(response.Header().Get("Allow"), http.MethodPost) {
		t.Fatalf("expected POST in Allow header, got %q", response.Header().Get("Allow"))
	}
}

func TestInternalNEXTokenAPIRejectsBadPIDWithoutDatabase(t *testing.T) {
	handler := newInternalNEXTokenHandler("test-only-secret-with-at-least-32-characters", "nex.example.net", 60014)
	request := httptest.NewRequest(http.MethodPost, "/internal/wsc/nex-token", bytes.NewBufferString(`{"pid":0}`))
	request.Header.Set("Authorization", "Bearer test-only-secret-with-at-least-32-characters")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for PID 0; got %d body=%s", response.Code, response.Body.String())
	}
}

func TestInternalNEXTokenAPIRejectsUnknownFields(t *testing.T) {
	handler := newInternalNEXTokenHandler("test-only-secret-with-at-least-32-characters", "nex.example.net", 60014)
	request := httptest.NewRequest(http.MethodPost, "/internal/wsc/nex-token", bytes.NewBufferString(`{"pid":123,"password":"caller-controlled"}`))
	request.Header.Set("Authorization", "Bearer test-only-secret-with-at-least-32-characters")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown field; got %d body=%s", response.Code, response.Body.String())
	}
}
