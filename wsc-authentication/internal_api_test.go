package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)


func TestSessionCredentialsUseSeparateFormats(t *testing.T) {
	password, err := makeSessionPassword()
	if err != nil {
		t.Fatalf("makeSessionPassword returned an error: %v", err)
	}
	passwordBytes, err := hex.DecodeString(password)
	if err != nil {
		t.Fatalf("password is not hex-encoded: %v", err)
	}
	if len(passwordBytes) != 16 {
		t.Fatalf("expected 16 random password bytes; got %d", len(passwordBytes))
	}

	token, err := makeSessionToken()
	if err != nil {
		t.Fatalf("makeSessionToken returned an error: %v", err)
	}
	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("token is not base64-encoded: %v", err)
	}
	if len(tokenBytes) != 36 {
		t.Fatalf("expected 36 random token bytes; got %d", len(tokenBytes))
	}
	if token == password {
		t.Fatal("NEX token and session password must be separate credentials")
	}
}

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
