package main

import (
	"testing"

	nex "github.com/PretendoNetwork/nex-go"
)

func TestDeriveWSCNEXPasswordIsStableAndPIDSpecific(t *testing.T) {
	secret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	a, err := deriveWSCNEXPassword(123456789, secret)
	if err != nil {
		t.Fatalf("deriveWSCNEXPassword: %v", err)
	}
	b, err := deriveWSCNEXPassword(123456789, secret)
	if err != nil {
		t.Fatalf("deriveWSCNEXPassword: %v", err)
	}
	c, err := deriveWSCNEXPassword(123456790, secret)
	if err != nil {
		t.Fatalf("deriveWSCNEXPassword: %v", err)
	}
	if a != b {
		t.Fatal("same secret and PID must yield the same WSC NEX password")
	}
	if a == c {
		t.Fatal("different PIDs must yield different WSC NEX passwords")
	}
	if len(a) != 64 {
		t.Fatalf("expected a 64-character SHA-256 hex password; got %d chars", len(a))
	}
	if a == secret {
		t.Fatal("derived password must not be the shared secret")
	}
}

func TestPasswordFromPIDUsesSharedSecretBeforeDatabase(t *testing.T) {
	secret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	t.Setenv("WSC_NEX_PASSWORD_SECRET", secret)
	want, err := deriveWSCNEXPassword(123456789, secret)
	if err != nil {
		t.Fatal(err)
	}
	got, code := passwordFromPID(123456789)
	if code != 0 {
		t.Fatalf("passwordFromPID returned error code %d", code)
	}
	if got != want {
		t.Fatal("passwordFromPID did not use shared HMAC-derived value")
	}
}

func TestDeriveWSCNEXPasswordRejectsShortSecretOrZeroPID(t *testing.T) {
	if _, err := deriveWSCNEXPassword(1, "too-short"); err == nil {
		t.Fatal("expected short shared secret to be rejected")
	}
	if _, err := deriveWSCNEXPassword(0, "0123456789abcdef0123456789abcdef"); err == nil {
		t.Fatal("expected zero PID to be rejected")
	}
}

func TestPasswordFromPIDShortSecretFailsClosed(t *testing.T) {
	t.Setenv("WSC_NEX_PASSWORD_SECRET", "short-secret")
	got, code := passwordFromPID(123)
	if got != "" || code != nex.Errors.RendezVous.InvalidUsername {
		t.Fatalf("expected invalid account for short configured secret; got %q / %d", got, code)
	}
}
