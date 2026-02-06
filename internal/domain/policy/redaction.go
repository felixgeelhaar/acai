package policy

// RedactionType categorizes what kind of data to redact.
type RedactionType string

const (
	RedactionEmails   RedactionType = "emails"
	RedactionSpeakers RedactionType = "speakers"
	RedactionKeywords RedactionType = "keywords"
	RedactionPatterns RedactionType = "patterns"
)

// RedactionRule defines a single redaction operation.
type RedactionRule struct {
	Type        RedactionType
	Replacement string   // What to replace matches with
	Keywords    []string // For keyword type: specific words to redact
	Pattern     string   // For patterns type: regex pattern
}

// RedactionConfig groups all redaction rules.
type RedactionConfig struct {
	Enabled bool
	Rules   []RedactionRule
}
