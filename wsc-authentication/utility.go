package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	nex "github.com/PretendoNetwork/nex-go"
)

// deriveWSCNEXPassword matches the HMAC calculation in provider-route.patch.
// Both sides use the UTF-8 secret as key and the base-10 PID as the message.
func deriveWSCNEXPassword(pid uint32, secret string) (string, error) {
	if pid == 0 {
		return "", fmt.Errorf("PID must be non-zero")
	}
	if len([]byte(secret)) < 32 {
		return "", fmt.Errorf("WSC_NEX_PASSWORD_SECRET must contain at least 32 bytes")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strconv.FormatUint(uint64(pid), 10)))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func passwordFromPID(pid uint32) (string, uint32) {
	if secret := strings.TrimSpace(os.Getenv("WSC_NEX_PASSWORD_SECRET")); secret != "" {
		password, err := deriveWSCNEXPassword(pid, secret)
		if err != nil {
			logger.Warning("WSC NEX password secret is invalid")
			return "", nex.Errors.RendezVous.InvalidUsername
		}
		return password, 0
	}

	// Local-only compatibility mode. Public Pretendo integration should set the
	// same WSC_NEX_PASSWORD_SECRET on the account route and authentication server.
	user := getNEXAccountByPID(pid)
	if user == nil {
		return "", nex.Errors.RendezVous.InvalidUsername
	}
	password, ok := user["password"].(string)
	if !ok || password == "" {
		return "", nex.Errors.RendezVous.InvalidUsername
	}
	return password, 0
}
