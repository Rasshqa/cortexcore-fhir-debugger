package validator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	reNIK       = regexp.MustCompile(`^\d{16}$`)
	reISO8601   = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(T\d{2}:\d{2}:\d{2}(\.\d+)?([+-]\d{2}:\d{2}|Z)?)?$`)
	reTimezoned = regexp.MustCompile(`[+-]\d{2}:\d{2}$|Z$`)
	reUUID      = regexp.MustCompile(`^urn:uuid:[0-9a-fA-F\-]{36}$`)
)

// ValidatePayload parses raw JSON and validates against SATUSEHAT profiles.
func ValidatePayload(raw []byte) (*ValidationReport, error) {
	var doc map[string]interface{}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	report := &ValidationReport{Issues: []Issue{}}

	resourceType, _ := doc["resourceType"].(string)
	if resourceType == "" {
		report.Issues = append(report.Issues, Issue{
			Resource: "root", Path: "resourceType", RuleID: "FHIR-GEN-001",
			Severity: SeverityError, Message: "Properti 'resourceType' wajib ada di setiap resource FHIR",
			Diagnostic: "Missing mandatory field: resourceType",
			Suggestion: "Tambahkan field \"resourceType\": \"<NamaResource>\" di root JSON.",
		})
		report.ErrorCount = len(report.Issues)
		return report, nil
	}

	if resourceType == "Bundle" {
		report.BundleType, _ = doc["type"].(string)
		validateBundle(doc, report)
	} else {
		validateResource(doc, resourceType, "root", report)
	}

	for _, iss := range report.Issues {
		if iss.Severity == SeverityError {
			report.ErrorCount++
		} else if iss.Severity == SeverityWarning {
			report.WarningCount++
		}
	}
	report.Passed = report.ErrorCount == 0
	return report, nil
}

func validateBundle(doc map[string]interface{}, report *ValidationReport) {
	btype, _ := doc["type"].(string)
	if btype != "transaction" && btype != "collection" && btype != "batch" {
		report.Issues = append(report.Issues, Issue{
			Resource: "Bundle", Path: "type", RuleID: "KEMKES-BND-001",
			Severity: SeverityError, Message: "Bundle type harus 'transaction', 'batch', atau 'collection'",
			Diagnostic: fmt.Sprintf("Received type: '%s'", btype),
			Suggestion: "Setel Bundle.type = \"transaction\" untuk pengiriman standar ke SATUSEHAT.",
		})
	}

	entries, ok := doc["entry"].([]interface{})
	if !ok || len(entries) == 0 {
		report.Issues = append(report.Issues, Issue{
			Resource: "Bundle", Path: "entry", RuleID: "KEMKES-BND-002",
			Severity: SeverityError, Message: "Bundle harus memiliki minimal 1 entry",
			Diagnostic: "entry[] is empty or missing",
			Suggestion: "Tambahkan setidaknya satu entry berisi resource (misal Encounter).",
		})
		return
	}

	refs := map[string]bool{}
	for i, e := range entries {
		entry, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		resource, ok := entry["resource"].(map[string]interface{})
		if !ok {
			report.Issues = append(report.Issues, Issue{
				Resource: "Bundle", Path: fmt.Sprintf("entry[%d].resource", i), RuleID: "KEMKES-BND-003",
				Severity: SeverityError, Message: fmt.Sprintf("entry[%d] tidak memiliki 'resource'", i),
				Diagnostic: "Missing resource object inside entry",
				Suggestion: "Setiap entry Bundle harus memiliki properti 'resource' berisi resource FHIR.",
			})
			continue
		}

		rt, _ := resource["resourceType"].(string)
		report.TotalResources++
		entryPath := fmt.Sprintf("entry[%d].resource", i)
		validateResource(resource, rt, entryPath, report)

		// Track fullUrl for reference validation
		fullURL, _ := entry["fullUrl"].(string)
		if fullURL != "" {
			refs[fullURL] = true
		}
	}
}

func validateResource(res map[string]interface{}, rt, basePath string, report *ValidationReport) {
	switch rt {
	case "Patient":
		validatePatient(res, basePath, report)
	case "Encounter":
		validateEncounter(res, basePath, report)
	case "Condition":
		validateCondition(res, basePath, report)
	case "Observation":
		validateObservation(res, basePath, report)
	case "Procedure":
		validateProcedure(res, basePath, report)
	case "MedicationRequest":
		validateMedicationRequest(res, basePath, report)
	}
}

// ───── Patient ─────
func validatePatient(res map[string]interface{}, base string, r *ValidationReport) {
	identifiers := getArray(res, "identifier")
	foundNIK := false
	for j, id := range identifiers {
		idm, _ := id.(map[string]interface{})
		system, _ := idm["system"].(string)
		value, _ := idm["value"].(string)
		if system == "https://fhir.kemkes.go.id/id/nik" {
			foundNIK = true
			if !reNIK.MatchString(value) {
				r.Issues = append(r.Issues, Issue{
					Resource: "Patient", Path: fmt.Sprintf("%s.identifier[%d].value", base, j),
					RuleID: "KEMKES-PAT-001", Severity: SeverityError,
					Message:    fmt.Sprintf("NIK '%s' bukan 16 digit numerik yang valid", value),
					Diagnostic: "NIK must be exactly 16 numeric characters",
					Suggestion: "Pastikan NIK pasien diisi 16 digit angka tanpa spasi/karakter lain.",
					CanAutoFix: true,
				})
			}
		}
	}
	if !foundNIK {
		r.Issues = append(r.Issues, Issue{
			Resource: "Patient", Path: base + ".identifier", RuleID: "KEMKES-PAT-002",
			Severity: SeverityError,
			Message:    "Identifier NIK Kemenkes tidak ditemukan",
			Diagnostic: "Missing identifier with system 'https://fhir.kemkes.go.id/id/nik'",
			Suggestion: "Tambahkan identifier: {system: 'https://fhir.kemkes.go.id/id/nik', value: '<NIK16digit>'}.",
			CanAutoFix: true,
		})
	}

	requireField(res, "name", base, "Patient", "KEMKES-PAT-003", "Nama pasien wajib diisi", r)
	requireField(res, "gender", base, "Patient", "KEMKES-PAT-004", "Gender pasien wajib diisi (male/female/other/unknown)", r)
	requireField(res, "birthDate", base, "Patient", "KEMKES-PAT-005", "Tanggal lahir pasien wajib diisi (YYYY-MM-DD)", r)
}

// ───── Encounter ─────
func validateEncounter(res map[string]interface{}, base string, r *ValidationReport) {
	status, _ := res["status"].(string)
	validStatuses := map[string]bool{"planned": true, "arrived": true, "triaged": true, "in-progress": true, "onleave": true, "finished": true, "cancelled": true}
	if !validStatuses[status] {
		r.Issues = append(r.Issues, Issue{
			Resource: "Encounter", Path: base + ".status", RuleID: "KEMKES-ENC-001",
			Severity: SeverityError,
			Message:    fmt.Sprintf("Encounter status '%s' tidak valid", status),
			Diagnostic: "Expected: planned|arrived|triaged|in-progress|onleave|finished|cancelled",
			Suggestion: "Gunakan status sesuai lifecycle FHIR Encounter (misal 'finished' untuk kunjungan selesai).",
		})
	}

	classObj, _ := res["class"].(map[string]interface{})
	if classObj != nil {
		code, _ := classObj["code"].(string)
		validClasses := map[string]bool{"AMB": true, "IMP": true, "EMER": true, "OBSENC": true, "PRENC": true}
		if !validClasses[code] {
			r.Issues = append(r.Issues, Issue{
				Resource: "Encounter", Path: base + ".class.code", RuleID: "KEMKES-ENC-002",
				Severity: SeverityWarning,
				Message:    fmt.Sprintf("Class '%s' bukan standar V3 ActCode umum", code),
				Diagnostic: "Common values: AMB (rawat jalan), IMP (rawat inap), EMER (IGD)",
				Suggestion: "Periksa kode class Encounter sesuai ValueSet http://terminology.hl7.org/CodeSystem/v3-ActCode.",
			})
		}
	} else {
		r.Issues = append(r.Issues, Issue{
			Resource: "Encounter", Path: base + ".class", RuleID: "KEMKES-ENC-003",
			Severity: SeverityError,
			Message:    "Encounter.class wajib ada (AMB/IMP/EMER)",
			Diagnostic: "Missing mandatory field: class",
			Suggestion: "Tambahkan {\"class\": {\"system\": \"http://terminology.hl7.org/CodeSystem/v3-ActCode\", \"code\": \"AMB\"}}.",
		})
	}

	validateTimestampField(res, "period.start", base, "Encounter", "KEMKES-ENC-004", r)
	requireField(res, "subject", base, "Encounter", "KEMKES-ENC-005", "Encounter.subject (referensi Patient) wajib ada", r)
}

// ───── Condition ─────
func validateCondition(res map[string]interface{}, base string, r *ValidationReport) {
	requireField(res, "subject", base, "Condition", "KEMKES-CON-001", "Condition.subject (referensi Patient) wajib ada", r)

	code := getNestedMap(res, "code")
	if code == nil {
		r.Issues = append(r.Issues, Issue{
			Resource: "Condition", Path: base + ".code", RuleID: "KEMKES-CON-002",
			Severity: SeverityError, Message: "Condition.code (kode diagnosis ICD-10) wajib ada",
			Diagnostic: "Missing mandatory field: code",
			Suggestion: "Tambahkan code dengan system 'http://hl7.org/fhir/sid/icd-10' dan kode diagnosis.",
		})
	} else {
		validateCodingSystem(code, "http://hl7.org/fhir/sid/icd-10", base+".code", "Condition", "KEMKES-CON-003", "ICD-10", r)
	}

	clinicalStatus := getNestedMap(res, "clinicalStatus")
	if clinicalStatus != nil {
		validateCodingSystem(clinicalStatus, "http://terminology.hl7.org/CodeSystem/condition-clinical", base+".clinicalStatus", "Condition", "KEMKES-CON-004", "condition-clinical", r)
	}
}

// ───── Observation ─────
func validateObservation(res map[string]interface{}, base string, r *ValidationReport) {
	requireField(res, "status", base, "Observation", "KEMKES-OBS-001", "Observation.status wajib ada (registered/preliminary/final/amended)", r)
	requireField(res, "subject", base, "Observation", "KEMKES-OBS-002", "Observation.subject (referensi Patient) wajib ada", r)

	code := getNestedMap(res, "code")
	if code == nil {
		r.Issues = append(r.Issues, Issue{
			Resource: "Observation", Path: base + ".code", RuleID: "KEMKES-OBS-003",
			Severity: SeverityError, Message: "Observation.code (kode LOINC) wajib ada",
			Diagnostic: "Missing mandatory field: code",
			Suggestion: "Tambahkan code dengan system 'http://loinc.org' dan kode observasi.",
		})
	} else {
		validateCodingSystem(code, "http://loinc.org", base+".code", "Observation", "KEMKES-OBS-004", "LOINC", r)
	}
}

// ───── Procedure ─────
func validateProcedure(res map[string]interface{}, base string, r *ValidationReport) {
	requireField(res, "status", base, "Procedure", "KEMKES-PRC-001", "Procedure.status wajib ada (completed/in-progress/not-done)", r)
	requireField(res, "subject", base, "Procedure", "KEMKES-PRC-002", "Procedure.subject (referensi Patient) wajib ada", r)

	code := getNestedMap(res, "code")
	if code == nil {
		r.Issues = append(r.Issues, Issue{
			Resource: "Procedure", Path: base + ".code", RuleID: "KEMKES-PRC-003",
			Severity: SeverityError, Message: "Procedure.code (kode ICD-9-CM) wajib ada",
			Diagnostic: "Missing mandatory field: code",
			Suggestion: "Tambahkan code dengan system 'http://hl7.org/fhir/sid/icd-9-cm' dan kode tindakan.",
		})
	} else {
		validateCodingSystem(code, "http://hl7.org/fhir/sid/icd-9-cm", base+".code", "Procedure", "KEMKES-PRC-004", "ICD-9-CM", r)
	}
}

// ───── MedicationRequest ─────
func validateMedicationRequest(res map[string]interface{}, base string, r *ValidationReport) {
	requireField(res, "status", base, "MedicationRequest", "KEMKES-MED-001", "MedicationRequest.status wajib ada (active/completed/cancelled)", r)
	requireField(res, "subject", base, "MedicationRequest", "KEMKES-MED-002", "MedicationRequest.subject (referensi Patient) wajib ada", r)

	intent, _ := res["intent"].(string)
	if intent != "order" && intent != "plan" && intent != "proposal" {
		r.Issues = append(r.Issues, Issue{
			Resource: "MedicationRequest", Path: base + ".intent", RuleID: "KEMKES-MED-003",
			Severity: SeverityError,
			Message:    fmt.Sprintf("MedicationRequest.intent '%s' tidak valid", intent),
			Diagnostic: "Expected: order | plan | proposal",
			Suggestion: "Setel intent = \"order\" untuk resep obat standar ke SATUSEHAT.",
		})
	}

	// Check medication reference or codeable concept
	medRef := getNestedMap(res, "medicationReference")
	medCC := getNestedMap(res, "medicationCodeableConcept")
	if medRef == nil && medCC == nil {
		r.Issues = append(r.Issues, Issue{
			Resource: "MedicationRequest", Path: base + ".medication[x]", RuleID: "KEMKES-MED-004",
			Severity: SeverityError,
			Message:    "MedicationRequest harus memiliki medicationReference atau medicationCodeableConcept",
			Diagnostic: "Missing medication[x] — no drug identified",
			Suggestion: "Gunakan medicationCodeableConcept dengan system KFA atau medicationReference ke resource Medication.",
		})
	} else if medCC != nil {
		validateCodingSystem(medCC, "http://sys-ids.kemkes.go.id/kfa", base+".medicationCodeableConcept", "MedicationRequest", "KEMKES-MED-005", "KFA Kemenkes", r)
	}
}

// ───── Helpers ─────

func requireField(res map[string]interface{}, field, base, resource, ruleID, message string, r *ValidationReport) {
	if _, ok := res[field]; !ok {
		r.Issues = append(r.Issues, Issue{
			Resource: resource, Path: base + "." + field, RuleID: ruleID,
			Severity:   SeverityError,
			Message:    message,
			Diagnostic: fmt.Sprintf("Missing mandatory field: %s", field),
			Suggestion: fmt.Sprintf("Tambahkan field '%s' pada resource %s.", field, resource),
		})
	}
}

func validateTimestampField(res map[string]interface{}, dotPath, base, resource, ruleID string, r *ValidationReport) {
	parts := strings.SplitN(dotPath, ".", 2)
	var val string
	if len(parts) == 2 {
		nested := getNestedMap(res, parts[0])
		if nested != nil {
			val, _ = nested[parts[1]].(string)
		}
	} else {
		val, _ = res[dotPath].(string)
	}

	if val == "" {
		return // Treated as optional, may be caught by requireField
	}

	if !reISO8601.MatchString(val) {
		r.Issues = append(r.Issues, Issue{
			Resource: resource, Path: base + "." + dotPath, RuleID: ruleID,
			Severity: SeverityError,
			Message:    fmt.Sprintf("Timestamp '%s' tidak sesuai format ISO-8601", val),
			Diagnostic: "Expected format: YYYY-MM-DDThh:mm:ss+07:00",
			Suggestion: "Gunakan format ISO-8601 dengan timezone offset (misal +07:00 untuk WIB).",
			CanAutoFix: true,
		})
	} else if !reTimezoned.MatchString(val) && strings.Contains(val, "T") {
		r.Issues = append(r.Issues, Issue{
			Resource: resource, Path: base + "." + dotPath, RuleID: ruleID + "b",
			Severity: SeverityWarning,
			Message:    "Timestamp tidak memiliki timezone offset — SATUSEHAT mungkin menolak",
			Diagnostic: fmt.Sprintf("Value '%s' has no timezone offset", val),
			Suggestion: "Tambahkan offset timezone, misal +07:00 (WIB), +08:00 (WITA), +09:00 (WIT).",
			CanAutoFix: true,
		})
	}
}

func validateCodingSystem(cc map[string]interface{}, expectedSystem, path, resource, ruleID, systemName string, r *ValidationReport) {
	codings := getArray(cc, "coding")
	if len(codings) == 0 {
		r.Issues = append(r.Issues, Issue{
			Resource: resource, Path: path + ".coding", RuleID: ruleID,
			Severity: SeverityError,
			Message:    fmt.Sprintf("Coding array kosong — kode %s wajib diisi", systemName),
			Diagnostic: "coding[] is empty or missing",
			Suggestion: fmt.Sprintf("Tambahkan coding dengan system '%s' dan kode yang valid.", expectedSystem),
		})
		return
	}

	found := false
	for _, c := range codings {
		cm, _ := c.(map[string]interface{})
		sys, _ := cm["system"].(string)
		if sys == expectedSystem {
			found = true
			code, _ := cm["code"].(string)
			if code == "" {
				r.Issues = append(r.Issues, Issue{
					Resource: resource, Path: path + ".coding[].code", RuleID: ruleID + "-code",
					Severity: SeverityError, Message: "Kode terminologi kosong",
					Diagnostic: fmt.Sprintf("coding with system '%s' has empty code", expectedSystem),
					Suggestion: "Isi field 'code' dengan kode terminologi yang valid.",
				})
			}
			break
		}
	}

	if !found {
		r.Issues = append(r.Issues, Issue{
			Resource: resource, Path: path + ".coding[].system", RuleID: ruleID + "-sys",
			Severity: SeverityError,
			Message:    fmt.Sprintf("System URI '%s' tidak ditemukan di coding array", expectedSystem),
			Diagnostic: fmt.Sprintf("Expected coding system: %s", expectedSystem),
			Suggestion: fmt.Sprintf("Pastikan salah satu coding menggunakan system: '%s'.", expectedSystem),
			CanAutoFix: true,
		})
	}
}

func getArray(m map[string]interface{}, key string) []interface{} {
	arr, _ := m[key].([]interface{})
	return arr
}

func getNestedMap(m map[string]interface{}, key string) map[string]interface{} {
	v, _ := m[key].(map[string]interface{})
	return v
}
