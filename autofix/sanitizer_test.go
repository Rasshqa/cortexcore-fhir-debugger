package autofix

import (
	"strings"
	"testing"
)

func TestSanitizeTimestampAndTimezone(t *testing.T) {
	raw := []byte(`{
		"resourceType": "Encounter",
		"period": { "start": "2026-03-31 08:30:00" }
	}`)

	fixed, patches, err := SanitizePayload(raw)
	if err != nil {
		t.Fatalf("sanitizer failed: %v", err)
	}

	if len(patches) == 0 {
		t.Fatalf("expected patches, got 0")
	}

	if !strings.Contains(string(fixed), "2026-03-31T08:30:00+07:00") {
		t.Errorf("expected timestamp with +07:00 offset, got: %s", string(fixed))
	}
}

func TestSanitizeNIK(t *testing.T) {
	raw := []byte(`{
		"resourceType": "Patient",
		"identifier": [
			{ "system": "https://fhir.kemkes.go.id/id/nik", "value": "3201-1234-5678-0001" }
		]
	}`)

	fixed, patches, err := SanitizePayload(raw)
	if err != nil {
		t.Fatalf("sanitizer failed: %v", err)
	}

	if len(patches) == 0 {
		t.Fatalf("expected patches, got 0")
	}

	if !strings.Contains(string(fixed), "3201123456780001") {
		t.Errorf("expected cleaned NIK, got: %s", string(fixed))
	}
}
