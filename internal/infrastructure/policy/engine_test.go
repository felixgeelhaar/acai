package policy

import (
	"testing"

	domainpolicy "github.com/felixgeelhaar/acai/internal/domain/policy"
)

func TestEngine_CheckAccess_Allowed(t *testing.T) {
	engine := NewEngine(&LoadResult{
		Policy: domainpolicy.Policy{DefaultEffect: domainpolicy.EffectAllow},
	})

	err := engine.CheckAccess("get_meeting", domainpolicy.MeetingContext{})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestEngine_CheckAccess_Denied(t *testing.T) {
	engine := NewEngine(&LoadResult{
		Policy: domainpolicy.Policy{
			DefaultEffect: domainpolicy.EffectAllow,
			Rules: []domainpolicy.Rule{
				{
					Name:       "block-transcripts",
					Effect:     domainpolicy.EffectDeny,
					Tools:      []string{"get_transcript"},
					Conditions: domainpolicy.Conditions{MeetingTags: []string{"confidential"}},
				},
			},
		},
	})

	err := engine.CheckAccess("get_transcript", domainpolicy.MeetingContext{
		Tags: []string{"confidential"},
	})
	if err != domainpolicy.ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestEngine_CheckAccess_DeniedToolNotMatched(t *testing.T) {
	engine := NewEngine(&LoadResult{
		Policy: domainpolicy.Policy{
			DefaultEffect: domainpolicy.EffectAllow,
			Rules: []domainpolicy.Rule{
				{Name: "block-transcripts", Effect: domainpolicy.EffectDeny, Tools: []string{"get_transcript"}},
			},
		},
	})

	err := engine.CheckAccess("list_meetings", domainpolicy.MeetingContext{})
	if err != nil {
		t.Errorf("expected nil for unrelated tool, got %v", err)
	}
}

func TestEngine_Redact(t *testing.T) {
	engine := NewEngine(&LoadResult{
		Redaction: domainpolicy.RedactionConfig{
			Enabled: true,
			Rules: []domainpolicy.RedactionRule{
				{Type: domainpolicy.RedactionEmails, Replacement: "[EMAIL]"},
			},
		},
	})

	got := engine.Redact("Contact alice@example.com")
	if got != "Contact [EMAIL]" {
		t.Errorf("got %q", got)
	}
}

func TestEngine_RedactSpeaker(t *testing.T) {
	engine := NewEngine(&LoadResult{
		Redaction: domainpolicy.RedactionConfig{
			Enabled: true,
			Rules: []domainpolicy.RedactionRule{
				{Type: domainpolicy.RedactionSpeakers, Replacement: "Speaker {n}"},
			},
		},
	})

	got := engine.RedactSpeaker("Alice")
	if got != "Speaker 1" {
		t.Errorf("got %q", got)
	}
}

func TestEngine_RedactionEnabled(t *testing.T) {
	engine := NewEngine(&LoadResult{
		Redaction: domainpolicy.RedactionConfig{Enabled: true},
	})
	if !engine.RedactionEnabled() {
		t.Error("expected redaction enabled")
	}

	engine2 := NewEngine(&LoadResult{
		Redaction: domainpolicy.RedactionConfig{Enabled: false},
	})
	if engine2.RedactionEnabled() {
		t.Error("expected redaction disabled")
	}
}

func TestEngine_FullIntegration(t *testing.T) {
	yaml := `
default_effect: allow
rules:
  - name: block-confidential-transcripts
    effect: deny
    tools: [get_transcript, export_embeddings]
    conditions:
      meeting_tags: [confidential]
redaction:
  enabled: true
  rules:
    - type: emails
      replacement: "[EMAIL]"
    - type: speakers
      replacement: "Speaker {n}"
    - type: keywords
      keywords: [salary]
      replacement: "[REDACTED]"
`
	result, err := LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	engine := NewEngine(result)

	// ACL check: confidential transcript should be denied
	err = engine.CheckAccess("get_transcript", domainpolicy.MeetingContext{
		Tags: []string{"confidential"},
	})
	if err != domainpolicy.ErrAccessDenied {
		t.Errorf("expected denied, got %v", err)
	}

	// ACL check: non-confidential should be allowed
	err = engine.CheckAccess("get_transcript", domainpolicy.MeetingContext{
		Tags: []string{"public"},
	})
	if err != nil {
		t.Errorf("expected allowed, got %v", err)
	}

	// Redaction: email
	got := engine.Redact("alice@test.com discussed salary")
	if findSubstring(got, "alice@test.com") {
		t.Errorf("email not redacted: %q", got)
	}
	if findSubstring(got, "salary") {
		t.Errorf("keyword not redacted: %q", got)
	}

	// Speaker anonymization
	s1 := engine.RedactSpeaker("Alice")
	s2 := engine.RedactSpeaker("Bob")
	if s1 == s2 {
		t.Error("different speakers should get different anonymized names")
	}
}
