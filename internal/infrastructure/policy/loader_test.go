package policy

import (
	"os"
	"path/filepath"
	"testing"

	domainpolicy "github.com/felixgeelhaar/granola-mcp/internal/domain/policy"
)

func TestLoadFromBytes_ValidYAML(t *testing.T) {
	yaml := `
default_effect: allow
rules:
  - name: block-confidential
    effect: deny
    tools:
      - get_transcript
      - export_embeddings
    conditions:
      meeting_tags:
        - confidential
redaction:
  enabled: true
  rules:
    - type: emails
      replacement: "[EMAIL]"
    - type: keywords
      keywords:
        - salary
        - confidential
      replacement: "[REDACTED]"
`

	result, err := LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Policy.DefaultEffect != domainpolicy.EffectAllow {
		t.Errorf("default effect = %q, want allow", result.Policy.DefaultEffect)
	}
	if len(result.Policy.Rules) != 1 {
		t.Fatalf("got %d rules, want 1", len(result.Policy.Rules))
	}
	rule := result.Policy.Rules[0]
	if rule.Name != "block-confidential" {
		t.Errorf("rule name = %q", rule.Name)
	}
	if rule.Effect != domainpolicy.EffectDeny {
		t.Errorf("rule effect = %q", rule.Effect)
	}
	if len(rule.Tools) != 2 {
		t.Errorf("got %d tools", len(rule.Tools))
	}
	if len(rule.Conditions.MeetingTags) != 1 || rule.Conditions.MeetingTags[0] != "confidential" {
		t.Errorf("conditions = %v", rule.Conditions)
	}

	if !result.Redaction.Enabled {
		t.Error("redaction should be enabled")
	}
	if len(result.Redaction.Rules) != 2 {
		t.Errorf("got %d redaction rules", len(result.Redaction.Rules))
	}
}

func TestLoadFromBytes_DefaultDeny(t *testing.T) {
	yaml := `default_effect: deny`
	result, err := LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Policy.DefaultEffect != domainpolicy.EffectDeny {
		t.Errorf("default effect = %q, want deny", result.Policy.DefaultEffect)
	}
}

func TestLoadFromBytes_EmptyYAML(t *testing.T) {
	result, err := LoadFromBytes([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty YAML should default to allow
	if result.Policy.DefaultEffect != domainpolicy.EffectAllow {
		t.Errorf("default effect = %q, want allow", result.Policy.DefaultEffect)
	}
}

func TestLoadFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadFromBytes([]byte(`{invalid: [yaml`))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadFromFile_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	data := `
default_effect: allow
rules:
  - name: test
    effect: deny
    tools: [get_transcript]
`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Policy.Rules) != 1 {
		t.Errorf("got %d rules", len(result.Policy.Rules))
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/policy.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadFromBytes_PatternRedaction(t *testing.T) {
	yaml := `
redaction:
  enabled: true
  rules:
    - type: patterns
      replacement: "[SSN]"
      pattern: '\d{3}-\d{2}-\d{4}'
`
	result, err := LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Redaction.Rules) != 1 {
		t.Fatalf("got %d rules", len(result.Redaction.Rules))
	}
	if result.Redaction.Rules[0].Pattern == "" {
		t.Error("expected non-empty pattern")
	}
}

func TestLoadFromBytes_SpeakerRedaction(t *testing.T) {
	yaml := `
redaction:
  enabled: true
  rules:
    - type: speakers
      replacement: "Speaker {n}"
`
	result, err := LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Redaction.Rules[0].Type != domainpolicy.RedactionSpeakers {
		t.Errorf("type = %q", result.Redaction.Rules[0].Type)
	}
}
