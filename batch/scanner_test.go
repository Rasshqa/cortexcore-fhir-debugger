package batch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fhir-batch-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Valid bundle
	validJSON := `{
		"resourceType": "Bundle",
		"type": "transaction",
		"entry": [{
			"resource": {
				"resourceType": "Encounter",
				"status": "finished",
				"class": { "code": "AMB" },
				"period": { "start": "2026-03-31T08:00:00+07:00" },
				"subject": { "reference": "Patient/1" }
			}
		}]
	}`
	// Invalid bundle (missing class)
	invalidJSON := `{
		"resourceType": "Bundle",
		"type": "transaction",
		"entry": [{
			"resource": {
				"resourceType": "Encounter",
				"status": "finished",
				"period": { "start": "2026-03-31T08:00:00+07:00" },
				"subject": { "reference": "Patient/1" }
			}
		}]
	}`

	os.WriteFile(filepath.Join(tmpDir, "tx_01.json"), []byte(validJSON), 0644)
	os.WriteFile(filepath.Join(tmpDir, "tx_02.json"), []byte(invalidJSON), 0644)

	summary, err := ScanDirectory(tmpDir, 10)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if summary.TotalFiles != 2 {
		t.Errorf("expected 2 files, got %d", summary.TotalFiles)
	}

	if summary.CleanFiles != 1 || summary.RejectFiles != 1 {
		t.Errorf("expected 1 clean, 1 reject; got clean=%d, reject=%d", summary.CleanFiles, summary.RejectFiles)
	}
}
