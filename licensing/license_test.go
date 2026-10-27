package licensing

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func TestVerifyLicenseKey(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	payload := LicensePayload{
		Organization: "RSUD Dr. Moewardi",
		Tier:         "ENTERPRISE",
		IssuedAt:     time.Now(),
		ExpiresAt:    time.Now().Add(365 * 24 * time.Hour),
		MaxNodes:     10,
	}

	payloadBytes, _ := json.Marshal(payload)
	sig := ed25519.Sign(priv, payloadBytes)

	token := base64.RawURLEncoding.EncodeToString(payloadBytes) + "." + base64.RawURLEncoding.EncodeToString(sig)

	status, err := VerifyLicenseKey(token, pub)
	if err != nil {
		t.Fatalf("expected valid license, got error: %v", err)
	}

	if !status.IsLicensed || status.Organization != "RSUD Dr. Moewardi" {
		t.Errorf("unexpected license status: %+v", status)
	}
}

func TestVerifyTamperedLicenseKey(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	fakeToken := base64.RawURLEncoding.EncodeToString([]byte(`{"org":"Hacker RS"}`)) + "." + base64.RawURLEncoding.EncodeToString([]byte("invalidsignaturebytes12345678901234567890123456789012345678901234567890"))

	_, err := VerifyLicenseKey(fakeToken, pub)
	if err == nil {
		t.Errorf("expected tampered license to fail, but it passed")
	}
}
