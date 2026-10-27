package batch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fhir-debugger/validator"
)

// FileResult holds the validation outcome for a single JSON file.
type FileResult struct {
	FileName    string `json:"file_name"`
	Path        string `json:"path"`
	Passed      bool   `json:"passed"`
	ErrorCount  int    `json:"error_count"`
	TotalIssues int    `json:"total_issues"`
	TopIssue    string `json:"top_issue,omitempty"`
}

// BatchSummary aggregates stats across hundreds/thousands of JSON log files.
type BatchSummary struct {
	TotalFiles   int          `json:"total_files"`
	CleanFiles   int          `json:"clean_files"`
	RejectFiles  int          `json:"reject_files"`
	TotalErrors  int          `json:"total_errors"`
	Files        []FileResult `json:"files"`
	IssueCounts  map[string]int `json:"issue_counts"` // rule_id -> count
}

// ScanDirectory walks a directory, parses each .json file, and compiles an audit summary.
func ScanDirectory(dirPath string, maxFiles int) (*BatchSummary, error) {
	if maxFiles <= 0 {
		maxFiles = 500
	}

	summary := &BatchSummary{
		Files:       make([]FileResult, 0),
		IssueCounts: make(map[string]int),
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
			return nil
		}
		if len(summary.Files) >= maxFiles {
			return filepath.SkipDir
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		summary.TotalFiles++
		report, err := validator.ValidatePayload(raw)
		if err != nil {
			summary.RejectFiles++
			summary.TotalErrors++
			summary.Files = append(summary.Files, FileResult{
				FileName:    info.Name(),
				Path:        path,
				Passed:      false,
				ErrorCount:  1,
				TotalIssues: 1,
				TopIssue:    fmt.Sprintf("Invalid JSON: %v", err),
			})
			return nil
		}

		fr := FileResult{
			FileName:    info.Name(),
			Path:        path,
			Passed:      report.Passed,
			ErrorCount:  report.ErrorCount,
			TotalIssues: len(report.Issues),
		}

		if len(report.Issues) > 0 {
			fr.TopIssue = report.Issues[0].Message
			for _, iss := range report.Issues {
				summary.IssueCounts[iss.RuleID]++
			}
		}

		if report.Passed {
			summary.CleanFiles++
		} else {
			summary.RejectFiles++
			summary.TotalErrors += report.ErrorCount
		}

		summary.Files = append(summary.Files, fr)
		return nil
	})

	return summary, err
}

// GenerateAuditMarkdown creates a professional report for hospital management.
func GenerateAuditMarkdown(summary *BatchSummary, orgName string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# LAPORAN AUDIT INTEROPERABILITAS SATUSEHAT KEMENKES RI\n\n"))
	sb.WriteString(fmt.Sprintf("- **Fasilitas Kesehatan:** %s\n", orgName))
	sb.WriteString(fmt.Sprintf("- **Total Berkas Terverifikasi:** %d\n", summary.TotalFiles))
	complianceRate := 0.0
	if summary.TotalFiles > 0 {
		complianceRate = (float64(summary.CleanFiles) / float64(summary.TotalFiles)) * 100.0
	}
	sb.WriteString(fmt.Sprintf("- **Tingkat Kepatuhan (Compliance Rate):** %.2f%%\n", complianceRate))
	sb.WriteString(fmt.Sprintf("- **Berkas Lolos (Siap Kirim):** %d\n", summary.CleanFiles))
	sb.WriteString(fmt.Sprintf("- **Berkas Ditolak (Perlu Perbaikan):** %d\n\n", summary.RejectFiles))

	sb.WriteString("## Rekapitulasi Pelanggaran Standar Terbanyak\n\n")
	sb.WriteString("| Kode Aturan Kemenkes | Jumlah Temuan | Deskripsi Ringkas |\n")
	sb.WriteString("|---|---|---|\n")
	for rule, count := range summary.IssueCounts {
		sb.WriteString(fmt.Sprintf("| `%s` | %d | Pelanggaran profil Kemenkes |\n", rule, count))
	}

	sb.WriteString("\n---\n*Digenerate oleh CortexCore :: FHIR-Reject Debugger — AetherCortex Solution*\n")
	sb.WriteString("*Architected by Rashqa Andrean Fitrah Sulaeman*\n")
	return sb.String()
}
