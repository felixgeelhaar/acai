package policy

import (
	"testing"

	domainpolicy "github.com/felixgeelhaar/granola-mcp/internal/domain/policy"
)

func TestRedactor_Disabled(t *testing.T) {
	r := NewRedactor(domainpolicy.RedactionConfig{Enabled: false})
	input := "alice@example.com said salary is high"
	if got := r.Redact(input); got != input {
		t.Errorf("disabled redactor should return input unchanged, got %q", got)
	}
}

func TestRedactor_EmailRedaction(t *testing.T) {
	r := NewRedactor(domainpolicy.RedactionConfig{
		Enabled: true,
		Rules: []domainpolicy.RedactionRule{
			{Type: domainpolicy.RedactionEmails, Replacement: "[EMAIL]"},
		},
	})

	input := "Contact alice@example.com or bob@test.org for details"
	got := r.Redact(input)
	expected := "Contact [EMAIL] or [EMAIL] for details"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestRedactor_EmailRedaction_NoEmails(t *testing.T) {
	r := NewRedactor(domainpolicy.RedactionConfig{
		Enabled: true,
		Rules: []domainpolicy.RedactionRule{
			{Type: domainpolicy.RedactionEmails, Replacement: "[EMAIL]"},
		},
	})

	input := "No emails here"
	if got := r.Redact(input); got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

func TestRedactor_KeywordRedaction(t *testing.T) {
	r := NewRedactor(domainpolicy.RedactionConfig{
		Enabled: true,
		Rules: []domainpolicy.RedactionRule{
			{Type: domainpolicy.RedactionKeywords, Replacement: "[REDACTED]", Keywords: []string{"salary", "confidential"}},
		},
	})

	input := "The salary discussion was confidential and very important"
	got := r.Redact(input)
	if got == input {
		t.Error("keywords should be redacted")
	}
	if !containsStr(got, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output: %q", got)
	}
}

func TestRedactor_KeywordRedaction_CaseInsensitive(t *testing.T) {
	r := NewRedactor(domainpolicy.RedactionConfig{
		Enabled: true,
		Rules: []domainpolicy.RedactionRule{
			{Type: domainpolicy.RedactionKeywords, Replacement: "[X]", Keywords: []string{"Secret"}},
		},
	})

	input := "This is SECRET information and a secret plan"
	got := r.Redact(input)
	if containsStr(got, "SECRET") || containsStr(got, "secret") {
		t.Errorf("case-insensitive keyword not redacted: %q", got)
	}
}

func TestRedactor_PatternRedaction(t *testing.T) {
	r := NewRedactor(domainpolicy.RedactionConfig{
		Enabled: true,
		Rules: []domainpolicy.RedactionRule{
			{Type: domainpolicy.RedactionPatterns, Replacement: "[SSN]", Pattern: `\d{3}-\d{2}-\d{4}`},
		},
	})

	input := "SSN is 123-45-6789 for the record"
	got := r.Redact(input)
	expected := "SSN is [SSN] for the record"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestRedactor_SpeakerRedaction(t *testing.T) {
	r := NewRedactor(domainpolicy.RedactionConfig{
		Enabled: true,
		Rules: []domainpolicy.RedactionRule{
			{Type: domainpolicy.RedactionSpeakers, Replacement: "Speaker {n}"},
		},
	})

	// Map speakers
	s1 := r.RedactSpeaker("Alice")
	s2 := r.RedactSpeaker("Bob")
	s3 := r.RedactSpeaker("Alice") // Should return same as first

	if s1 != "Speaker 1" {
		t.Errorf("Alice = %q, want Speaker 1", s1)
	}
	if s2 != "Speaker 2" {
		t.Errorf("Bob = %q, want Speaker 2", s2)
	}
	if s3 != "Speaker 1" {
		t.Errorf("Alice (again) = %q, want Speaker 1", s3)
	}
}

func TestRedactor_SpeakerRedaction_Disabled(t *testing.T) {
	r := NewRedactor(domainpolicy.RedactionConfig{Enabled: false})

	got := r.RedactSpeaker("Alice")
	if got != "Alice" {
		t.Errorf("got %q, want Alice (disabled)", got)
	}
}

func TestRedactor_MultipleRules(t *testing.T) {
	r := NewRedactor(domainpolicy.RedactionConfig{
		Enabled: true,
		Rules: []domainpolicy.RedactionRule{
			{Type: domainpolicy.RedactionEmails, Replacement: "[EMAIL]"},
			{Type: domainpolicy.RedactionKeywords, Replacement: "[REDACTED]", Keywords: []string{"secret"}},
		},
	})

	input := "Email alice@test.com about the secret project"
	got := r.Redact(input)
	if containsStr(got, "alice@test.com") {
		t.Errorf("email not redacted: %q", got)
	}
	if containsStr(got, "secret") {
		t.Errorf("keyword not redacted: %q", got)
	}
}

func TestRedactor_EmptyRules(t *testing.T) {
	r := NewRedactor(domainpolicy.RedactionConfig{
		Enabled: true,
		Rules:   nil,
	})

	input := "alice@test.com said something secret"
	if got := r.Redact(input); got != input {
		t.Errorf("no rules should return input unchanged: %q", got)
	}
}

func TestRedactor_InvalidPatternIgnored(t *testing.T) {
	r := NewRedactor(domainpolicy.RedactionConfig{
		Enabled: true,
		Rules: []domainpolicy.RedactionRule{
			{Type: domainpolicy.RedactionPatterns, Replacement: "[X]", Pattern: `[invalid`},
		},
	})

	input := "some text"
	if got := r.Redact(input); got != input {
		t.Errorf("invalid pattern should be skipped: %q", got)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
