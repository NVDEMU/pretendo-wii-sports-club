package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const internalAPIPort = "8080"

type internalNEXTokenRequest struct {
	PID uint32 `json:"pid"`
}

type internalNEXTokenResponse struct {
	XMLName     xml.Name `xml:"nex_token"`
	Host        string   `xml:"host"`
	NexPassword string   `xml:"nex_password"`
	PID         uint32   `xml:"pid"`
	Port        uint16   `xml:"port"`
	Token       string   `xml:"token"`
}

func validPublicHost(host string) bool {
	host = strings.TrimSpace(host)
	if host == "" || len(host) > 253 || strings.ContainsAny(host, "/\\\\ \t\r\n<>") {
		return false
	}
	if net.ParseIP(host) != nil {
		return true
	}
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, ch := range label {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
				(ch >= '0' && ch <= '9') || ch == '-') {
				return false
			}
		}
	}
	return true
}

func makeSessionPassword() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate NEX session password: %w", err)
	}
	return hex.EncodeToString(randomBytes), nil
}

func makeSessionToken() (string, error) {
	randomBytes := make([]byte, 36)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate NEX session token: %w", err)
	}
	return base64.StdEncoding.EncodeToString(randomBytes), nil
}

func newInternalNEXTokenHandler(apiToken, publicHost string, publicPort uint16) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/internal/wsc/nex-token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if len(apiToken) < 32 || publicPort == 0 || !validPublicHost(publicHost) {
			http.Error(w, "internal NEX token service is not configured", http.StatusServiceUnavailable)
			return
		}

		provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if len(provided) != len(apiToken) ||
			subtle.ConstantTimeCompare([]byte(provided), []byte(apiToken)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		defer r.Body.Close()
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var request internalNEXTokenRequest
		if err := decoder.Decode(&request); err != nil {
			http.Error(w, "invalid JSON request", http.StatusBadRequest)
			return
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			http.Error(w, "request must contain exactly one JSON object", http.StatusBadRequest)
			return
		}
		if request.PID == 0 {
			http.Error(w, "pid must be an existing non-zero PID", http.StatusBadRequest)
			return
		}
		if mongoCollection == nil {
			http.Error(w, "NEX account database is unavailable", http.StatusServiceUnavailable)
			return
		}

		passwordCandidate, err := makeSessionPassword()
		if err != nil {
			logger.Warning(err.Error())
			http.Error(w, "could not issue NEX credentials", http.StatusInternalServerError)
			return
		}
		token, err := makeSessionToken()
		if err != nil {
			logger.Warning(err.Error())
			http.Error(w, "could not issue NEX credentials", http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		// NEX account passwords are stable per PID; only the separate token is
		// rotated per request. $setOnInsert prevents one request from invalidating
		// another request's login by changing an established account password.
		var account bson.M
		err = mongoCollection.FindOneAndUpdate(
			ctx,
			bson.D{{Key: "pid", Value: request.PID}},
			bson.D{{Key: "$setOnInsert", Value: bson.D{
				{Key: "pid", Value: request.PID},
				{Key: "password", Value: passwordCandidate},
			}}},
			options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
		).Decode(&account)
		if err != nil {
			logger.Warning(fmt.Sprintf("Could not provision NEX credentials for PID %d: %v", request.PID, err))
			http.Error(w, "could not store NEX credentials", http.StatusInternalServerError)
			return
		}
		password, ok := account["password"].(string)
		if !ok || password == "" {
			logger.Warning(fmt.Sprintf("NEX account password is missing or invalid for PID %d", request.PID))
			http.Error(w, "NEX account credentials are invalid", http.StatusInternalServerError)
			return
		}

		response := internalNEXTokenResponse{
			Host:        publicHost,
			NexPassword: password,
			PID:         request.PID,
			Port:        publicPort,
			Token:       token,
		}
		body, err := xml.Marshal(response)
		if err != nil {
			logger.Warning("Could not encode NEX token response: " + err.Error())
			http.Error(w, "could not encode NEX token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(append([]byte(xml.Header), body...))
		fmt.Printf("Issued WSC NEX session credentials for PID %d\n", request.PID)
	})
	return mux
}

func startInternalNEXTokenAPI() {
	apiToken := strings.TrimSpace(os.Getenv("WSC_INTERNAL_API_TOKEN"))
	publicHost := strings.TrimSpace(os.Getenv("NEX_AUTH_PUBLIC_HOST"))
	if apiToken == "" && publicHost == "" {
		fmt.Println("WSC internal NEX token API disabled (optional configuration not set)")
		return
	}
	if len(apiToken) < 32 || !validPublicHost(publicHost) {
		fmt.Fprintln(os.Stderr, "WSC internal NEX token API configuration invalid: set a 32+ character WSC_INTERNAL_API_TOKEN and a valid NEX_AUTH_PUBLIC_HOST together")
		os.Exit(2)
	}

	portText := strings.TrimSpace(os.Getenv("NEX_AUTH_PUBLIC_PORT"))
	if portText == "" {
		portText = "60014"
	}
	parsedPort, err := strconv.ParseUint(portText, 10, 16)
	if err != nil || parsedPort == 0 {
		fmt.Fprintln(os.Stderr, "NEX_AUTH_PUBLIC_PORT must be an integer from 1 to 65535")
		os.Exit(2)
	}

	handler := newInternalNEXTokenHandler(apiToken, publicHost, uint16(parsedPort))
	server := &http.Server{
		Addr:              ":" + internalAPIPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	go func() {
		fmt.Println("WSC internal NEX token API listening on the private container network port " + internalAPIPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Warning("WSC internal NEX token API stopped: " + err.Error())
		}
	}()
}
