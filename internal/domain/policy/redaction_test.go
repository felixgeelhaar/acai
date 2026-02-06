package policy

import "testing"

func TestRedactionConfig_Defaults(t *testing.T) {
	rc := RedactionConfig{}
	if rc.Enabled {
		t.Error("should be disabled by default")
	}
	if len(rc.Rules) != 0 {
		t.Errorf("got %d rules, want 0", len(rc.Rules))
	}
}

func TestRedactionRule_EmailType(t *testing.T) {
	r := RedactionRule{
		Type:        RedactionEmails,
		Replacement: "[EMAIL]",
	}
	if r.Type != RedactionEmails {
		t.Errorf("got type %q", r.Type)
	}
	if r.Replacement != "[EMAIL]" {
		t.Errorf("got replacement %q", r.Replacement)
	}
}

func TestRedactionRule_SpeakersType(t *testing.T) {
	r := RedactionRule{
		Type:        RedactionSpeakers,
		Replacement: "Speaker {n}",
	}
	if r.Type != RedactionSpeakers {
		t.Errorf("got type %q", r.Type)
	}
}

func TestRedactionRule_KeywordsType(t *testing.T) {
	r := RedactionRule{
		Type:        RedactionKeywords,
		Replacement: "[REDACTED]",
		Keywords:    []string{"confidential", "salary"},
	}
	if len(r.Keywords) != 2 {
		t.Errorf("got %d keywords", len(r.Keywords))
	}
}

func TestRedactionRule_PatternsType(t *testing.T) {
	r := RedactionRule{
		Type:        RedactionPatterns,
		Replacement: "[MATCH]",
		Pattern:     `\d{3}-\d{2}-\d{4}`,
	}
	if r.Pattern == "" {
		t.Error("expected non-empty pattern")
	}
}

func TestRedactionConfig_WithRules(t *testing.T) {
	rc := RedactionConfig{
		Enabled: true,
		Rules: []RedactionRule{
			{Type: RedactionEmails, Replacement: "[EMAIL]"},
			{Type: RedactionKeywords, Replacement: "[REDACTED]", Keywords: []string{"secret"}},
		},
	}
	if !rc.Enabled {
		t.Error("should be enabled")
	}
	if len(rc.Rules) != 2 {
		t.Errorf("got %d rules", len(rc.Rules))
	}
}
