package policy

import (
	domainpolicy "github.com/felixgeelhaar/granola-mcp/internal/domain/policy"
)

// Engine wraps a domain Policy and a Redactor for the MCP middleware.
type Engine struct {
	policy   domainpolicy.Policy
	redactor *Redactor
}

// NewEngine creates a policy engine from a load result.
func NewEngine(result *LoadResult) *Engine {
	return &Engine{
		policy:   result.Policy,
		redactor: NewRedactor(result.Redaction),
	}
}

// CheckAccess evaluates whether a tool call is allowed.
func (e *Engine) CheckAccess(tool string, ctx domainpolicy.MeetingContext) error {
	effect := e.policy.Evaluate(tool, ctx)
	if effect == domainpolicy.EffectDeny {
		return domainpolicy.ErrAccessDenied
	}
	return nil
}

// Redact applies redaction rules to content.
func (e *Engine) Redact(content string) string {
	return e.redactor.Redact(content)
}

// RedactSpeaker anonymizes a speaker name.
func (e *Engine) RedactSpeaker(name string) string {
	return e.redactor.RedactSpeaker(name)
}

// RedactionEnabled returns whether redaction is active.
func (e *Engine) RedactionEnabled() bool {
	return e.redactor.config.Enabled
}
