package mcp_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	domainpolicy "github.com/felixgeelhaar/granola-mcp/internal/domain/policy"
	policy "github.com/felixgeelhaar/granola-mcp/internal/infrastructure/policy"
	mcpiface "github.com/felixgeelhaar/granola-mcp/internal/interfaces/mcp"
)

func TestPolicyMiddleware_AllowedTool(t *testing.T) {
	repo := newMockRepo()
	repo.addMeeting(mustMeeting(t, "m-1", "Sprint Planning"))
	srv := newTestServer(repo)

	engine := policy.NewEngine(&policy.LoadResult{
		Policy: domainpolicy.Policy{DefaultEffect: domainpolicy.EffectAllow},
	})
	mw := mcpiface.NewPolicyMiddleware(srv, engine)

	raw, err := mw.HandleToolJSON(context.Background(), "list_meetings", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []mcpiface.MeetingResult
	if err := json.Unmarshal(raw, &results); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
}

func TestPolicyMiddleware_DeniedTool(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	engine := policy.NewEngine(&policy.LoadResult{
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
	mw := mcpiface.NewPolicyMiddleware(srv, engine)

	_, err := mw.HandleToolJSON(context.Background(), "get_transcript", json.RawMessage(`{"meeting_id":"m-1","tags":["confidential"]}`))
	if err == nil {
		t.Fatal("expected error for denied tool")
	}
	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("expected access denied error, got: %v", err)
	}
}

func TestPolicyMiddleware_DeniedByDefaultDeny(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	engine := policy.NewEngine(&policy.LoadResult{
		Policy: domainpolicy.Policy{DefaultEffect: domainpolicy.EffectDeny},
	})
	mw := mcpiface.NewPolicyMiddleware(srv, engine)

	_, err := mw.HandleToolJSON(context.Background(), "list_meetings", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error for default deny policy")
	}
}

func TestPolicyMiddleware_Redaction_EmailsInResponse(t *testing.T) {
	repo := newMockRepo()
	m := mustMeeting(t, "m-1", "alice@example.com discussed budget")
	repo.addMeeting(m)

	srv := newTestServer(repo)
	engine := policy.NewEngine(&policy.LoadResult{
		Policy: domainpolicy.Policy{DefaultEffect: domainpolicy.EffectAllow},
		Redaction: domainpolicy.RedactionConfig{
			Enabled: true,
			Rules: []domainpolicy.RedactionRule{
				{Type: domainpolicy.RedactionEmails, Replacement: "[EMAIL]"},
			},
		},
	})
	mw := mcpiface.NewPolicyMiddleware(srv, engine)

	raw, err := mw.HandleToolJSON(context.Background(), "list_meetings", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(raw)
	if strings.Contains(output, "alice@example.com") {
		t.Errorf("email should be redacted in output: %s", output)
	}
	if !strings.Contains(output, "[EMAIL]") {
		t.Errorf("expected [EMAIL] replacement in output: %s", output)
	}
}

func TestPolicyMiddleware_Redaction_KeywordsInResponse(t *testing.T) {
	repo := newMockRepo()
	m := mustMeeting(t, "m-1", "Discuss salary ranges")
	repo.addMeeting(m)

	srv := newTestServer(repo)
	engine := policy.NewEngine(&policy.LoadResult{
		Policy: domainpolicy.Policy{DefaultEffect: domainpolicy.EffectAllow},
		Redaction: domainpolicy.RedactionConfig{
			Enabled: true,
			Rules: []domainpolicy.RedactionRule{
				{Type: domainpolicy.RedactionKeywords, Replacement: "[REDACTED]", Keywords: []string{"salary"}},
			},
		},
	})
	mw := mcpiface.NewPolicyMiddleware(srv, engine)

	raw, err := mw.HandleToolJSON(context.Background(), "list_meetings", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(raw)
	if strings.Contains(output, "salary") {
		t.Errorf("keyword should be redacted in output: %s", output)
	}
}

func TestPolicyMiddleware_NoRedaction_WhenDisabled(t *testing.T) {
	repo := newMockRepo()
	m := mustMeeting(t, "m-1", "alice@example.com discussed salary")
	repo.addMeeting(m)

	srv := newTestServer(repo)
	engine := policy.NewEngine(&policy.LoadResult{
		Policy: domainpolicy.Policy{DefaultEffect: domainpolicy.EffectAllow},
		Redaction: domainpolicy.RedactionConfig{
			Enabled: false,
			Rules: []domainpolicy.RedactionRule{
				{Type: domainpolicy.RedactionEmails, Replacement: "[EMAIL]"},
			},
		},
	})
	mw := mcpiface.NewPolicyMiddleware(srv, engine)

	raw, err := mw.HandleToolJSON(context.Background(), "list_meetings", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(raw)
	if !strings.Contains(output, "alice@example.com") {
		t.Errorf("email should NOT be redacted when disabled: %s", output)
	}
}

func TestPolicyMiddleware_ConditionalDeny_MatchingTags(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	engine := policy.NewEngine(&policy.LoadResult{
		Policy: domainpolicy.Policy{
			DefaultEffect: domainpolicy.EffectAllow,
			Rules: []domainpolicy.Rule{
				{
					Name:       "block-confidential",
					Effect:     domainpolicy.EffectDeny,
					Tools:      []string{"get_transcript", "export_embeddings"},
					Conditions: domainpolicy.Conditions{MeetingTags: []string{"confidential"}},
				},
			},
		},
	})
	mw := mcpiface.NewPolicyMiddleware(srv, engine)

	// Non-matching tags should be allowed
	_, err := mw.HandleToolJSON(context.Background(), "get_transcript", json.RawMessage(`{"meeting_id":"m-1","tags":["public"]}`))
	// This will fail because the mock repo doesn't have m-1 transcript, but it should NOT be a policy error
	if err != nil && strings.Contains(err.Error(), "access denied") {
		t.Errorf("non-confidential meeting should not be denied: %v", err)
	}
}

func TestPolicyMiddleware_ConditionalDeny_NoTags(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	engine := policy.NewEngine(&policy.LoadResult{
		Policy: domainpolicy.Policy{
			DefaultEffect: domainpolicy.EffectAllow,
			Rules: []domainpolicy.Rule{
				{
					Name:       "block-confidential",
					Effect:     domainpolicy.EffectDeny,
					Tools:      []string{"get_transcript"},
					Conditions: domainpolicy.Conditions{MeetingTags: []string{"confidential"}},
				},
			},
		},
	})
	mw := mcpiface.NewPolicyMiddleware(srv, engine)

	// No tags at all — should not match the condition, so allowed
	_, err := mw.HandleToolJSON(context.Background(), "get_transcript", json.RawMessage(`{"meeting_id":"m-1"}`))
	if err != nil && strings.Contains(err.Error(), "access denied") {
		t.Errorf("request without tags should not be denied: %v", err)
	}
}

func TestPolicyMiddleware_Inner(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)
	engine := policy.NewEngine(&policy.LoadResult{
		Policy: domainpolicy.Policy{DefaultEffect: domainpolicy.EffectAllow},
	})
	mw := mcpiface.NewPolicyMiddleware(srv, engine)

	if mw.Inner() != srv {
		t.Error("Inner() should return the wrapped server")
	}
}

func TestPolicyMiddleware_InnerError_PassesThrough(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)
	engine := policy.NewEngine(&policy.LoadResult{
		Policy: domainpolicy.Policy{DefaultEffect: domainpolicy.EffectAllow},
	})
	mw := mcpiface.NewPolicyMiddleware(srv, engine)

	_, err := mw.HandleToolJSON(context.Background(), "unknown_tool", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error from inner server")
	}
	if strings.Contains(err.Error(), "access denied") {
		t.Error("error should be from inner server, not policy")
	}
}

func TestPolicyMiddleware_SpeakerRedaction(t *testing.T) {
	repo := newMockRepo()
	m := mustMeeting(t, "m-1", "Team Sync")
	repo.addMeeting(m)

	srv := newTestServer(repo)
	engine := policy.NewEngine(&policy.LoadResult{
		Policy: domainpolicy.Policy{DefaultEffect: domainpolicy.EffectAllow},
		Redaction: domainpolicy.RedactionConfig{
			Enabled: true,
			Rules: []domainpolicy.RedactionRule{
				{Type: domainpolicy.RedactionSpeakers, Replacement: "Speaker {n}"},
			},
		},
	})
	mw := mcpiface.NewPolicyMiddleware(srv, engine)

	raw, err := mw.HandleToolJSON(context.Background(), "get_meeting", json.RawMessage(`{"id":"m-1"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Just ensure redaction ran without error — the meeting has no participants with names to redact
	if len(raw) == 0 {
		t.Error("expected non-empty response")
	}
}

func TestPolicyMiddleware_FullIntegration(t *testing.T) {
	yamlData := `
default_effect: allow
rules:
  - name: block-confidential
    effect: deny
    tools: [get_transcript]
    conditions:
      meeting_tags: [confidential]
redaction:
  enabled: true
  rules:
    - type: emails
      replacement: "[EMAIL]"
    - type: keywords
      keywords: [salary]
      replacement: "[REDACTED]"
`
	result, err := policy.LoadFromBytes([]byte(yamlData))
	if err != nil {
		t.Fatalf("load policy: %v", err)
	}

	repo := newMockRepo()
	m := mustMeeting(t, "m-1", "Discuss alice@test.com salary")
	repo.addMeeting(m)

	srv := newTestServer(repo)
	engine := policy.NewEngine(result)
	mw := mcpiface.NewPolicyMiddleware(srv, engine)

	// Test 1: list_meetings should work but with redacted content
	raw, err := mw.HandleToolJSON(context.Background(), "list_meetings", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("list_meetings error: %v", err)
	}
	output := string(raw)
	if strings.Contains(output, "alice@test.com") {
		t.Errorf("email not redacted: %s", output)
	}
	if strings.Contains(output, "salary") {
		t.Errorf("keyword not redacted: %s", output)
	}

	// Test 2: get_transcript for confidential meeting should be denied
	_, err = mw.HandleToolJSON(context.Background(), "get_transcript", json.RawMessage(`{"meeting_id":"m-1","tags":["confidential"]}`))
	if err == nil {
		t.Fatal("expected access denied for confidential transcript")
	}
	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("expected access denied error, got: %v", err)
	}

	// Test 3: get_transcript for non-confidential should not be policy-denied
	_, err = mw.HandleToolJSON(context.Background(), "get_transcript", json.RawMessage(`{"meeting_id":"m-1","tags":["public"]}`))
	if err != nil && strings.Contains(err.Error(), "access denied") {
		t.Errorf("public meeting should not be denied: %v", err)
	}
}
