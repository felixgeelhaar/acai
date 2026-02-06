// Package policy implements policy loading, evaluation, and content redaction.
// It maps domain policy value objects to infrastructure-level YAML configuration
// and provides the engine that the MCP middleware uses.
package policy

import (
	"fmt"
	"os"

	domainpolicy "github.com/felixgeelhaar/granola-mcp/internal/domain/policy"
	"gopkg.in/yaml.v3"
)

// yamlPolicy is the serialization format for policy YAML files.
type yamlPolicy struct {
	DefaultEffect string         `yaml:"default_effect"`
	Rules         []yamlRule     `yaml:"rules"`
	Redaction     yamlRedaction  `yaml:"redaction"`
}

type yamlRule struct {
	Name       string         `yaml:"name"`
	Effect     string         `yaml:"effect"`
	Tools      []string       `yaml:"tools"`
	Conditions yamlConditions `yaml:"conditions"`
}

type yamlConditions struct {
	MeetingTags []string `yaml:"meeting_tags"`
}

type yamlRedaction struct {
	Enabled bool               `yaml:"enabled"`
	Rules   []yamlRedactRule   `yaml:"rules"`
}

type yamlRedactRule struct {
	Type        string   `yaml:"type"`
	Replacement string   `yaml:"replacement"`
	Keywords    []string `yaml:"keywords"`
	Pattern     string   `yaml:"pattern"`
}

// LoadResult contains both the access policy and redaction config.
type LoadResult struct {
	Policy    domainpolicy.Policy
	Redaction domainpolicy.RedactionConfig
}

// LoadFromFile reads and parses a YAML policy file.
func LoadFromFile(path string) (*LoadResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read policy file: %w", err)
	}
	return LoadFromBytes(data)
}

// LoadFromBytes parses YAML policy data.
func LoadFromBytes(data []byte) (*LoadResult, error) {
	var yp yamlPolicy
	if err := yaml.Unmarshal(data, &yp); err != nil {
		return nil, fmt.Errorf("parse policy YAML: %w", err)
	}

	defaultEffect := domainpolicy.EffectAllow
	if yp.DefaultEffect == "deny" {
		defaultEffect = domainpolicy.EffectDeny
	}

	rules := make([]domainpolicy.Rule, len(yp.Rules))
	for i, yr := range yp.Rules {
		effect := domainpolicy.EffectAllow
		if yr.Effect == "deny" {
			effect = domainpolicy.EffectDeny
		}
		rules[i] = domainpolicy.Rule{
			Name:   yr.Name,
			Effect: effect,
			Tools:  yr.Tools,
			Conditions: domainpolicy.Conditions{
				MeetingTags: yr.Conditions.MeetingTags,
			},
		}
	}

	redactRules := make([]domainpolicy.RedactionRule, len(yp.Redaction.Rules))
	for i, rr := range yp.Redaction.Rules {
		redactRules[i] = domainpolicy.RedactionRule{
			Type:        domainpolicy.RedactionType(rr.Type),
			Replacement: rr.Replacement,
			Keywords:    rr.Keywords,
			Pattern:     rr.Pattern,
		}
	}

	return &LoadResult{
		Policy: domainpolicy.Policy{
			DefaultEffect: defaultEffect,
			Rules:         rules,
		},
		Redaction: domainpolicy.RedactionConfig{
			Enabled: yp.Redaction.Enabled,
			Rules:   redactRules,
		},
	}, nil
}
