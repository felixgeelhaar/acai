package policy

import "testing"

func TestPolicy_DefaultAllow(t *testing.T) {
	p := &Policy{DefaultEffect: EffectAllow}
	effect := p.Evaluate("get_meeting", MeetingContext{MeetingID: "m-1"})
	if effect != EffectAllow {
		t.Errorf("got %q, want allow", effect)
	}
}

func TestPolicy_DefaultDeny(t *testing.T) {
	p := &Policy{DefaultEffect: EffectDeny}
	effect := p.Evaluate("get_meeting", MeetingContext{MeetingID: "m-1"})
	if effect != EffectDeny {
		t.Errorf("got %q, want deny", effect)
	}
}

func TestPolicy_FirstMatchWins_DenyBeforeAllow(t *testing.T) {
	p := &Policy{
		DefaultEffect: EffectAllow,
		Rules: []Rule{
			{Name: "deny-transcripts", Effect: EffectDeny, Tools: []string{"get_transcript"}},
			{Name: "allow-all", Effect: EffectAllow},
		},
	}
	effect := p.Evaluate("get_transcript", MeetingContext{})
	if effect != EffectDeny {
		t.Errorf("got %q, want deny", effect)
	}
}

func TestPolicy_FirstMatchWins_AllowBeforeDeny(t *testing.T) {
	p := &Policy{
		DefaultEffect: EffectDeny,
		Rules: []Rule{
			{Name: "allow-listing", Effect: EffectAllow, Tools: []string{"list_meetings"}},
			{Name: "deny-all", Effect: EffectDeny},
		},
	}
	effect := p.Evaluate("list_meetings", MeetingContext{})
	if effect != EffectAllow {
		t.Errorf("got %q, want allow", effect)
	}
}

func TestPolicy_ToolNotInRuleToolList(t *testing.T) {
	p := &Policy{
		DefaultEffect: EffectAllow,
		Rules: []Rule{
			{Name: "deny-transcript", Effect: EffectDeny, Tools: []string{"get_transcript"}},
		},
	}
	// get_meeting is NOT in the deny rule's tool list, so should get default effect
	effect := p.Evaluate("get_meeting", MeetingContext{})
	if effect != EffectAllow {
		t.Errorf("got %q, want allow (tool not in deny list)", effect)
	}
}

func TestPolicy_EmptyToolList_MatchesAll(t *testing.T) {
	p := &Policy{
		DefaultEffect: EffectAllow,
		Rules: []Rule{
			{Name: "deny-all", Effect: EffectDeny, Tools: []string{}},
		},
	}
	effect := p.Evaluate("anything", MeetingContext{})
	if effect != EffectDeny {
		t.Errorf("got %q, want deny (empty tools = all)", effect)
	}
}

func TestPolicy_ConditionMeetingTags_Match(t *testing.T) {
	p := &Policy{
		DefaultEffect: EffectAllow,
		Rules: []Rule{
			{
				Name:       "deny-confidential",
				Effect:     EffectDeny,
				Tools:      []string{"get_transcript"},
				Conditions: Conditions{MeetingTags: []string{"confidential"}},
			},
		},
	}

	effect := p.Evaluate("get_transcript", MeetingContext{
		MeetingID: "m-1",
		Tags:      []string{"confidential", "important"},
	})
	if effect != EffectDeny {
		t.Errorf("got %q, want deny (tag matches)", effect)
	}
}

func TestPolicy_ConditionMeetingTags_NoMatch(t *testing.T) {
	p := &Policy{
		DefaultEffect: EffectAllow,
		Rules: []Rule{
			{
				Name:       "deny-confidential",
				Effect:     EffectDeny,
				Tools:      []string{"get_transcript"},
				Conditions: Conditions{MeetingTags: []string{"confidential"}},
			},
		},
	}

	// Meeting doesn't have the "confidential" tag, so rule doesn't match
	effect := p.Evaluate("get_transcript", MeetingContext{
		MeetingID: "m-1",
		Tags:      []string{"public"},
	})
	if effect != EffectAllow {
		t.Errorf("got %q, want allow (tags don't match)", effect)
	}
}

func TestPolicy_ConditionMeetingTags_EmptyMeetingTags(t *testing.T) {
	p := &Policy{
		DefaultEffect: EffectAllow,
		Rules: []Rule{
			{
				Name:       "deny-confidential",
				Effect:     EffectDeny,
				Conditions: Conditions{MeetingTags: []string{"confidential"}},
			},
		},
	}

	effect := p.Evaluate("get_meeting", MeetingContext{
		MeetingID: "m-1",
		Tags:      nil,
	})
	if effect != EffectAllow {
		t.Errorf("got %q, want allow (empty meeting tags)", effect)
	}
}

func TestPolicy_MultipleRulesComplex(t *testing.T) {
	p := &Policy{
		DefaultEffect: EffectAllow,
		Rules: []Rule{
			{
				Name:       "block-confidential-transcripts",
				Effect:     EffectDeny,
				Tools:      []string{"get_transcript", "export_embeddings"},
				Conditions: Conditions{MeetingTags: []string{"confidential"}},
			},
			{
				Name:       "allow-public-everything",
				Effect:     EffectAllow,
				Conditions: Conditions{MeetingTags: []string{"public"}},
			},
		},
	}

	tests := []struct {
		name   string
		tool   string
		tags   []string
		expect Effect
	}{
		{"confidential transcript denied", "get_transcript", []string{"confidential"}, EffectDeny},
		{"confidential embeddings denied", "export_embeddings", []string{"confidential"}, EffectDeny},
		{"confidential listing allowed", "list_meetings", []string{"confidential"}, EffectAllow},
		{"public transcript allowed", "get_transcript", []string{"public"}, EffectAllow},
		{"no tags default allow", "get_transcript", nil, EffectAllow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effect := p.Evaluate(tt.tool, MeetingContext{Tags: tt.tags})
			if effect != tt.expect {
				t.Errorf("got %q, want %q", effect, tt.expect)
			}
		})
	}
}

func TestPolicy_NoRules(t *testing.T) {
	p := &Policy{
		DefaultEffect: EffectDeny,
		Rules:         nil,
	}
	effect := p.Evaluate("get_meeting", MeetingContext{})
	if effect != EffectDeny {
		t.Errorf("got %q, want deny (no rules, default deny)", effect)
	}
}
