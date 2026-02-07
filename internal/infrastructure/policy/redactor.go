package policy

import (
	"fmt"
	"regexp"
	"strings"

	domainpolicy "github.com/felixgeelhaar/acai/internal/domain/policy"
)

var emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

// Redactor applies redaction rules to text content.
type Redactor struct {
	config          domainpolicy.RedactionConfig
	speakerMap      map[string]string
	speakerCounter  int
	patternRegexes  []*compiledPattern
}

type compiledPattern struct {
	regex       *regexp.Regexp
	replacement string
}

// NewRedactor creates a redactor from a redaction config.
func NewRedactor(config domainpolicy.RedactionConfig) *Redactor {
	r := &Redactor{
		config:     config,
		speakerMap: make(map[string]string),
	}

	// Pre-compile pattern regexes
	for _, rule := range config.Rules {
		if rule.Type == domainpolicy.RedactionPatterns && rule.Pattern != "" {
			if re, err := regexp.Compile(rule.Pattern); err == nil {
				r.patternRegexes = append(r.patternRegexes, &compiledPattern{
					regex:       re,
					replacement: rule.Replacement,
				})
			}
		}
	}

	return r
}

// Redact applies all configured redaction rules to the input text.
func (r *Redactor) Redact(text string) string {
	if !r.config.Enabled || len(r.config.Rules) == 0 {
		return text
	}

	result := text
	for _, rule := range r.config.Rules {
		switch rule.Type {
		case domainpolicy.RedactionEmails:
			result = emailRegex.ReplaceAllString(result, rule.Replacement)
		case domainpolicy.RedactionSpeakers:
			result = r.redactSpeakers(result, rule.Replacement)
		case domainpolicy.RedactionKeywords:
			result = r.redactKeywords(result, rule.Keywords, rule.Replacement)
		case domainpolicy.RedactionPatterns:
			// Handled via pre-compiled regexes
		}
	}

	// Apply pre-compiled pattern regexes
	for _, cp := range r.patternRegexes {
		result = cp.regex.ReplaceAllString(result, cp.replacement)
	}

	return result
}

// RedactSpeaker anonymizes a speaker name using a consistent mapping.
func (r *Redactor) RedactSpeaker(name string) string {
	if !r.config.Enabled {
		return name
	}

	for _, rule := range r.config.Rules {
		if rule.Type == domainpolicy.RedactionSpeakers {
			return r.mapSpeaker(name, rule.Replacement)
		}
	}
	return name
}

func (r *Redactor) redactSpeakers(text, replacement string) string {
	// For text content, we can only redact speakers that are in our map
	result := text
	for original, anonymized := range r.speakerMap {
		result = strings.ReplaceAll(result, original, anonymized)
	}
	return result
}

func (r *Redactor) mapSpeaker(name, template string) string {
	if anon, ok := r.speakerMap[name]; ok {
		return anon
	}
	r.speakerCounter++
	anon := strings.ReplaceAll(template, "{n}", fmt.Sprintf("%d", r.speakerCounter))
	r.speakerMap[name] = anon
	return anon
}

func (r *Redactor) redactKeywords(text string, keywords []string, replacement string) string {
	result := text
	for _, kw := range keywords {
		// Case-insensitive replacement
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(kw) + `\b`)
		result = re.ReplaceAllString(result, replacement)
	}
	return result
}
