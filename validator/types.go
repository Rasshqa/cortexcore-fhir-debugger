package validator

// IssueSeverity indicates the criticality of the validation finding.
type IssueSeverity string

const (
	SeverityError   IssueSeverity = "FATAL_REJECT" // HTTP 400/422 from SATUSEHAT
	SeverityWarning IssueSeverity = "WARNING"      // Degraded data quality / deprecation
	SeverityInfo    IssueSeverity = "INFO"         // Advisory
)

// Issue represents a specific FHIR conformance violation matching SATUSEHAT rules.
type Issue struct {
	Resource   string        `json:"resource"`    // e.g. "Encounter", "Patient"
	Path       string        `json:"path"`        // e.g. "entry[0].resource.period.start"
	RuleID     string        `json:"rule_id"`     // e.g. "KEMKES-ENC-001"
	Severity   IssueSeverity `json:"severity"`    // FATAL_REJECT | WARNING | INFO
	Message    string        `json:"message"`     // Human message in Bahasa Indonesia
	Diagnostic string        `json:"diagnostic"`  // Technical spec failure
	Suggestion string        `json:"suggestion"`  // Actionable instruction for SIMRS engineer
	CanAutoFix bool          `json:"can_auto_fix"`// Whether the auto-sanitizer can resolve this
}

// ValidationReport is the complete result returned to the UI.
type ValidationReport struct {
	TotalResources int     `json:"total_resources"`
	Passed         bool    `json:"passed"`
	ErrorCount     int     `json:"error_count"`
	WarningCount   int     `json:"warning_count"`
	Issues         []Issue `json:"issues"`
	BundleType     string  `json:"bundle_type,omitempty"`
}
