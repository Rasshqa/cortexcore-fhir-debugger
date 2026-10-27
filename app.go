package main

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"os"

	"fhir-debugger/autofix"
	"fhir-debugger/batch"
	"fhir-debugger/licensing"
	"fhir-debugger/validator"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// AetherCortexPublicKey is the Ed25519 public key for license verification.
const AetherCortexPublicKey = "O3XbT7xJ7n0tHw2kY8zP1qW3eR5tY7uI9oP0aS2dF4g="

type App struct {
	ctx context.Context
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// ValidateJSON validates raw JSON/Bundle payload against SATUSEHAT profiles.
func (a *App) ValidateJSON(payload string) *validator.ValidationReport {
	report, err := validator.ValidatePayload([]byte(payload))
	if err != nil {
		return &validator.ValidationReport{
			Issues: []validator.Issue{{
				Resource: "root", RuleID: "PARSE-001",
				Severity: validator.SeverityError,
				Message:  "JSON tidak valid: " + err.Error(),
			}},
			ErrorCount: 1,
		}
	}
	return report
}

// AutoFixJSON applies heuristic sanitization and returns the patched JSON + patch log.
func (a *App) AutoFixJSON(payload string) map[string]interface{} {
	fixed, patches, err := autofix.SanitizePayload([]byte(payload))
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	return map[string]interface{}{
		"fixed":   string(fixed),
		"patches": patches,
	}
}

// OpenJSONFile shows a file picker dialog and returns file content.
func (a *App) OpenJSONFile() map[string]interface{} {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Pilih Berkas FHIR Bundle JSON",
		Filters: []runtime.FileFilter{{DisplayName: "JSON Files", Pattern: "*.json"}},
	})
	if err != nil || path == "" {
		return map[string]interface{}{"error": "file tidak dipilih"}
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return map[string]interface{}{"error": "gagal membaca file: " + err.Error()}
	}
	return map[string]interface{}{
		"path":    path,
		"content": string(raw),
	}
}

// SaveFixedJSON saves a patched JSON to a .fixed.json file.
func (a *App) SaveFixedJSON(content string) map[string]interface{} {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Simpan JSON Terkoreksi",
		DefaultFilename: "payload.fixed.json",
		Filters:         []runtime.FileFilter{{DisplayName: "JSON Files", Pattern: "*.json"}},
	})
	if err != nil || path == "" {
		return map[string]interface{}{"error": "simpan dibatalkan"}
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return map[string]interface{}{"error": "gagal menyimpan: " + err.Error()}
	}
	return map[string]interface{}{"saved": path}
}

// BatchScanDirectory scans a folder of SIMRS JSON log files.
func (a *App) BatchScanDirectory(maxFiles int) map[string]interface{} {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Pilih Folder Log SIMRS untuk Batch Scan",
	})
	if err != nil || dir == "" {
		return map[string]interface{}{"error": "folder tidak dipilih"}
	}

	summary, err := batch.ScanDirectory(dir, maxFiles)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	return map[string]interface{}{"summary": summary, "dir": dir}
}

// VerifyLicense validates an Enterprise license key offline.
func (a *App) VerifyLicense(licenseKey string) *licensing.LicenseStatus {
	pubKeyBytes, err := base64.StdEncoding.DecodeString(AetherCortexPublicKey)
	if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
		st := licensing.DefaultCommunityStatus()
		return &st
	}

	status, err := licensing.VerifyLicenseKey(licenseKey, ed25519.PublicKey(pubKeyBytes))
	if err != nil {
		st := licensing.DefaultCommunityStatus()
		return &st
	}
	return status
}
