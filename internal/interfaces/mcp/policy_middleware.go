package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	domainpolicy "github.com/felixgeelhaar/acai/internal/domain/policy"
	policy "github.com/felixgeelhaar/acai/internal/infrastructure/policy"
)

// PolicyMiddleware wraps an MCP Server and enforces access control and redaction policies.
// It intercepts HandleToolJSON to check ACL before execution and applies redaction to responses.
type PolicyMiddleware struct {
	inner  *Server
	engine *policy.Engine
}

// NewPolicyMiddleware creates a middleware that wraps the given server with policy enforcement.
func NewPolicyMiddleware(inner *Server, engine *policy.Engine) *PolicyMiddleware {
	return &PolicyMiddleware{
		inner:  inner,
		engine: engine,
	}
}

// HandleToolJSON checks ACL, delegates to inner server, and applies redaction.
func (pm *PolicyMiddleware) HandleToolJSON(ctx context.Context, tool string, rawInput json.RawMessage) (json.RawMessage, error) {
	// Extract meeting context from input for ACL check
	meetingCtx := pm.extractMeetingContext(rawInput)

	// Check access control
	if err := pm.engine.CheckAccess(tool, meetingCtx); err != nil {
		return nil, fmt.Errorf("%s: %w", tool, err)
	}

	// Delegate to inner server
	result, err := pm.inner.HandleToolJSON(ctx, tool, rawInput)
	if err != nil {
		return nil, err
	}

	// Apply redaction if enabled
	if pm.engine.RedactionEnabled() {
		result = pm.redactJSON(result)
	}

	return result, nil
}

// Inner returns the wrapped server for direct access (e.g., transport setup).
func (pm *PolicyMiddleware) Inner() *Server {
	return pm.inner
}

// extractMeetingContext pulls meeting metadata from tool input for policy evaluation.
func (pm *PolicyMiddleware) extractMeetingContext(rawInput json.RawMessage) domainpolicy.MeetingContext {
	var input struct {
		MeetingID string   `json:"meeting_id"`
		ID        string   `json:"id"`
		Tags      []string `json:"tags"`
	}
	if err := json.Unmarshal(rawInput, &input); err != nil {
		log.Printf("policy: failed to extract meeting context: %v", err)
	}

	meetingID := input.MeetingID
	if meetingID == "" {
		meetingID = input.ID
	}

	return domainpolicy.MeetingContext{
		MeetingID: meetingID,
		Tags:      input.Tags,
	}
}

// redactJSON applies redaction rules to all string values in a JSON structure.
func (pm *PolicyMiddleware) redactJSON(data json.RawMessage) json.RawMessage {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return data
	}

	redacted := pm.redactValue(raw)
	result, err := json.Marshal(redacted)
	if err != nil {
		return data
	}
	return result
}

// redactValue recursively applies redaction to JSON values.
func (pm *PolicyMiddleware) redactValue(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return pm.engine.Redact(val)
	case map[string]interface{}:
		result := make(map[string]interface{}, len(val))
		for k, v := range val {
			if pm.isSpeakerField(k) {
				if s, ok := v.(string); ok {
					result[k] = pm.engine.RedactSpeaker(s)
					continue
				}
			}
			result[k] = pm.redactValue(v)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, v := range val {
			result[i] = pm.redactValue(v)
		}
		return result
	default:
		return v
	}
}

// isSpeakerField returns true if the JSON key represents a speaker/person field.
func (pm *PolicyMiddleware) isSpeakerField(key string) bool {
	lower := strings.ToLower(key)
	return lower == "speaker" || lower == "owner" || lower == "author" || lower == "name"
}
