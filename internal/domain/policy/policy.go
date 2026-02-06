// Package policy provides domain value objects for per-meeting agent policies.
// Policy evaluation is a presentation concern — the domain stays pure.
package policy

// Effect determines whether a rule allows or denies access.
type Effect string

const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

// Conditions specify when a rule applies.
type Conditions struct {
	MeetingTags []string // Rule applies if meeting has any of these tags
}

// Rule is a named policy rule with effect, target tools, and conditions.
type Rule struct {
	Name       string
	Effect     Effect
	Tools      []string   // Tool names this rule applies to (empty = all tools)
	Conditions Conditions
}

// MeetingContext provides meeting metadata for policy evaluation.
type MeetingContext struct {
	MeetingID string
	Tags      []string
}

// Policy is the top-level policy value object.
// It contains an ordered list of rules and a default effect.
// Evaluation is first-match-wins.
type Policy struct {
	DefaultEffect Effect
	Rules         []Rule
}

// Evaluate checks if a tool invocation is allowed given the meeting context.
// Uses first-match-wins semantics. If no rule matches, applies DefaultEffect.
func (p *Policy) Evaluate(tool string, ctx MeetingContext) Effect {
	for _, rule := range p.Rules {
		if matchesRule(rule, tool, ctx) {
			return rule.Effect
		}
	}
	return p.DefaultEffect
}

// matchesRule checks if a rule applies to the given tool and meeting context.
func matchesRule(rule Rule, tool string, ctx MeetingContext) bool {
	// Check tool match (empty tools list means all tools)
	if len(rule.Tools) > 0 && !containsString(rule.Tools, tool) {
		return false
	}

	// Check conditions
	if len(rule.Conditions.MeetingTags) > 0 {
		if !hasAnyTag(ctx.Tags, rule.Conditions.MeetingTags) {
			return false
		}
	}

	return true
}

func containsString(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func hasAnyTag(actual, required []string) bool {
	for _, r := range required {
		for _, a := range actual {
			if a == r {
				return true
			}
		}
	}
	return false
}
