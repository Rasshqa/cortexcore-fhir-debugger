package licensing

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// AetherCortexMasterPublicKey is the embedded Ed25519 public key.
// In production this verifies licenses issued by the founder/architecture key.
var AetherCortexMasterPublicKey, _ = base64.StdEncoding.DecodeString("O3XbT7xJ7n0tHw2kY8zP1qW3eR5tY7uI9oP0aS2dF4g=")

// LicensePayload is the signed claim inside a license key.
type LicensePayload struct {
	Organization string    `json:"org"`        // e.g. "RSUD Dr. Moewardi Surakarta"
	Tier         string    `json:"tier"`       // "ENTERPRISE" | "COMMUNITY"
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"` // Zero time = lifetime
	MaxNodes     int       `json:"max_nodes"`
}

// LicenseStatus represents the active license state.
type LicenseStatus struct {
	IsLicensed   bool      `json:"is_licensed"`
	Organization string    `json:"organization"`
	Tier         string    `json:"tier"`
	ExpiresAt    string    `json:"expires_at"`
	Features     []string  `json:"features"`
}

// DefaultCommunityStatus is returned when no valid license is installed.
func DefaultCommunityStatus() LicenseStatus {
	return LicenseStatus{
		IsLicensed:   false,
		Organization: "Community Edition (Free)",
		Tier:         "COMMUNITY",
		ExpiresAt:    "Unlimited",
		Features: []string{
			"Single Payload Interactive Validation",
			"SATUSEHAT Profile Error Diagnosis",
			"One-Click Auto-Fix Patch Generator",
			"Export Fixed JSON",
		},
	}
}

// EnterpriseFeatures returns unlocked enterprise perks.
func EnterpriseFeatures() []string {
	return []string{
		"Single Payload Interactive Validation",
		"SATUSEHAT Profile Error Diagnosis",
		"One-Click Auto-Fix Patch Generator",
		"Export Fixed JSON",
		"Batch Folder Scanner (Multi-File Logs)",
		"Directorate Audit Report Generator (PDF/HTML)",
		"Custom SIMRS Rule Overrides",
		"Priority Local Verification Engine",
	}
}

// VerifyLicenseKey validates a base64 encoded license token.
// Format: <base64(payload)>.<base64(signature)>
func VerifyLicenseKey(licenseStr string, pubKey ed25519.PublicKey) (*LicenseStatus, error) {
	if strings.TrimSpace(licenseStr) == "" {
		st := DefaultCommunityStatus()
		return &st, nil
	}

	parts := strings.Split(strings.TrimSpace(licenseStr), ".")
	if len(parts) != 2 {
		return nil, errors.New("format lisensi tidak valid (harus PAYLOAD.SIGNATURE)")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}

	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}

	// Verify cryptographic signature
	if !ed25519.Verify(pubKey, payloadBytes, sigBytes) {
		return nil, errors.New("tanda tangan digital lisensi tidak sah / dipalsukan")
	}

	var payload LicensePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("parse payload: %w", err)
	}

	if !payload.ExpiresAt.IsZero() && time.Now().After(payload.ExpiresAt) {
		return nil, fmt.Errorf("lisensi kadaluarsa pada %s", payload.ExpiresAt.Format("02 Jan 2006"))
	}

	return &LicenseStatus{
		IsLicensed:   true,
		Organization: payload.Organization,
		Tier:         payload.Tier,
		ExpiresAt:    payload.ExpiresAt.Format("02 Jan 2006"),
		Features:     EnterpriseFeatures(),
	}, nil
}
