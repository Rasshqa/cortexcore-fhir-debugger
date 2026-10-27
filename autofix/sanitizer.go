package autofix

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	reTimestampNoTZ = regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?)([^+\-Z\d]|$)`)
	reTimestampSpac = regexp.MustCompile(`(\d{4}-\d{2}-\d{2}) (\d{2}:\d{2}:\d{2})`)
	reNIKDirty      = regexp.MustCompile(`[^\d]`)
)

// SanitizePayload applies all heuristic auto-fixes to raw JSON and returns
// a corrected version plus a list of applied patches.
func SanitizePayload(raw []byte) ([]byte, []string, error) {
	var doc interface{}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, nil, err
	}

	var patches []string
	walkAndFix(doc, &patches)

	out, err := json.MarshalIndent(doc, "", "  ")
	return out, patches, err
}

func walkAndFix(node interface{}, patches *[]string) {
	switch n := node.(type) {
	case map[string]interface{}:
		fixTimestamps(n, patches)
		fixNIKIdentifiers(n, patches)
		injectMissingSystemURIs(n, patches)
		for _, v := range n {
			walkAndFix(v, patches)
		}
	case []interface{}:
		for _, item := range n {
			walkAndFix(item, patches)
		}
	}
}

// fixTimestamps converts "2026-01-15 08:30:00" -> "2026-01-15T08:30:00+07:00"
// and "2026-01-15T08:30:00" (no TZ) -> "2026-01-15T08:30:00+07:00"
func fixTimestamps(m map[string]interface{}, patches *[]string) {
	tsFields := []string{"start", "end", "performedDateTime", "authoredOn", "effectiveDateTime", "issued"}
	for _, field := range tsFields {
		if val, ok := m[field].(string); ok {
			fixed := fixTimestamp(val)
			if fixed != val {
				m[field] = fixed
				*patches = append(*patches, "Timestamp '"+field+"': '"+val+"' → '"+fixed+"' (timezone offset WIB ditambahkan)")
			}
		}
	}
}

func fixTimestamp(s string) string {
	// Replace space separator with T: use ${1}T${2} so $1T isn't treated as a single capture group name
	s = reTimestampSpac.ReplaceAllString(s, "${1}T${2}")
	// Add +07:00 if timestamp has T but no timezone offset after T
	if idx := strings.Index(s, "T"); idx != -1 {
		afterT := s[idx:]
		hasOffset := strings.Contains(afterT, "+") ||
			strings.Contains(afterT, "-") || // e.g. -05:00
			strings.HasSuffix(afterT, "Z")
		if !hasOffset {
			s = s + "+07:00"
		}
	}
	return s
}

// fixNIKIdentifiers strips non-numeric chars from NIK values in identifier arrays.
func fixNIKIdentifiers(m map[string]interface{}, patches *[]string) {
	identifiers, ok := m["identifier"].([]interface{})
	if !ok {
		return
	}
	for i, id := range identifiers {
		idm, ok := id.(map[string]interface{})
		if !ok {
			continue
		}
		system, _ := idm["system"].(string)
		if !strings.Contains(system, "nik") {
			continue
		}
		value, _ := idm["value"].(string)
		clean := reNIKDirty.ReplaceAllString(value, "")
		if clean != value {
			idm["value"] = clean
			identifiers[i] = idm
			*patches = append(*patches, "NIK identifier: '"+value+"' → '"+clean+"' (karakter non-numerik dihapus)")
		}
	}
}

// injectMissingSystemURIs ensures known SATUSEHAT system URIs are present in coding arrays.
func injectMissingSystemURIs(m map[string]interface{}, patches *[]string) {
	rt, _ := m["resourceType"].(string)
	type uriHint struct {
		field  string
		system string
		label  string
	}
	hints := map[string]uriHint{
		"Condition":         {field: "code", system: "http://hl7.org/fhir/sid/icd-10", label: "ICD-10"},
		"Procedure":         {field: "code", system: "http://hl7.org/fhir/sid/icd-9-cm", label: "ICD-9-CM"},
		"Observation":       {field: "code", system: "http://loinc.org", label: "LOINC"},
		"MedicationRequest": {field: "medicationCodeableConcept", system: "http://sys-ids.kemkes.go.id/kfa", label: "KFA"},
	}

	hint, ok := hints[rt]
	if !ok {
		return
	}

	cc, ok := m[hint.field].(map[string]interface{})
	if !ok {
		return
	}
	codings, _ := cc["coding"].([]interface{})
	for _, c := range codings {
		cm, _ := c.(map[string]interface{})
		if sys, _ := cm["system"].(string); sys == hint.system {
			return // Already present
		}
		// If coding exists with correct code but wrong/missing system, inject
		if sys, _ := cm["system"].(string); sys == "" {
			cm["system"] = hint.system
			*patches = append(*patches, "System URI disuntikkan di "+rt+"."+hint.field+".coding[].system: '"+hint.system+"' ("+hint.label+")")
			return
		}
	}
}
