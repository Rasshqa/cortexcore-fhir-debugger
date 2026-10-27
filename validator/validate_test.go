package validator

import (
	"testing"
)

func TestValidateEncounterMissingClass(t *testing.T) {
	raw := []byte(`{
		"resourceType": "Encounter",
		"status": "finished",
		"period": { "start": "2026-03-31T08:00:00+07:00" },
		"subject": { "reference": "Patient/10000001" }
	}`)

	report, err := ValidatePayload(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Passed {
		t.Errorf("expected failure due to missing class, but passed")
	}

	foundMissingClass := false
	for _, iss := range report.Issues {
		if iss.RuleID == "KEMKES-ENC-003" {
			foundMissingClass = true
		}
	}
	if !foundMissingClass {
		t.Errorf("expected KEMKES-ENC-003 error, not found")
	}
}

func TestValidatePatientInvalidNIK(t *testing.T) {
	raw := []byte(`{
		"resourceType": "Patient",
		"identifier": [
			{ "system": "https://fhir.kemkes.go.id/id/nik", "value": "12345" }
		],
		"name": [{ "text": "Budi Santoso" }],
		"gender": "male",
		"birthDate": "1990-05-12"
	}`)

	report, err := ValidatePayload(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Passed {
		t.Errorf("expected validation to fail for 5-digit NIK, but passed")
	}
}

func TestValidateValidBundleTransaction(t *testing.T) {
	raw := []byte(`{
		"resourceType": "Bundle",
		"type": "transaction",
		"entry": [
			{
				"fullUrl": "urn:uuid:6112ff39-9d0a-42c2-8b4b-9e1bfd8e4f1a",
				"resource": {
					"resourceType": "Encounter",
					"status": "finished",
					"class": { "system": "http://terminology.hl7.org/CodeSystem/v3-ActCode", "code": "AMB" },
					"period": { "start": "2026-03-31T08:00:00+07:00" },
					"subject": { "reference": "urn:uuid:patient-1" }
				}
			}
		]
	}`)

	report, err := ValidatePayload(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !report.Passed {
		t.Errorf("expected valid bundle to pass, got %d errors: %v", report.ErrorCount, report.Issues)
	}
}
