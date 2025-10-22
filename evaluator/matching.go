package evaluator

import (
	"regexp"
	"strings"
)

// ActionMatcher handles action pattern matching
type ActionMatcher struct{}

// NewActionMatcher creates a new action matcher
func NewActionMatcher() *ActionMatcher {
	return &ActionMatcher{}
}

// Match checks if an action matches a pattern
// Pattern format: <service>:<resource-type>:<operation>
// Supports wildcards: *, prefix-*, *-suffix, *-middle-*
func (am *ActionMatcher) Match(pattern, action string) bool {
	if pattern == "*" {
		return true
	}

	patternParts := strings.Split(pattern, ":")
	actionParts := strings.Split(action, ":")

	if len(patternParts) != len(actionParts) {
		return false
	}

	for i := 0; i < len(patternParts); i++ {
		if !am.matchSegment(patternParts[i], actionParts[i]) {
			return false
		}
	}
	return true
}

// matchSegment matches a single segment with wildcard support
func (am *ActionMatcher) matchSegment(pattern, value string) bool {
	if pattern == "*" {
		return true
	}
	if !strings.Contains(pattern, "*") {
		return pattern == value
	}
	return am.matchWildcard(pattern, value)
}

// matchWildcard converts wildcard pattern to regex and matches
func (am *ActionMatcher) matchWildcard(pattern, value string) bool {
	// Convert wildcard pattern to regex
	regexPattern := strings.ReplaceAll(pattern, "*", ".*")
	regexPattern = "^" + regexPattern + "$"

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return false
	}

	return regex.MatchString(value)
}

// ResourceMatcher handles resource pattern matching
type ResourceMatcher struct{}

// NewResourceMatcher creates a new resource matcher
func NewResourceMatcher() *ResourceMatcher {
	return &ResourceMatcher{}
}

// Match checks if a resource matches a pattern
// Pattern format: <service>:<resource-type>:<resource-id>
// Hierarchical: <service>:<parent-type>:<parent-id>/<child-type>:<child-id>
// Supports wildcards and variable substitution
func (rm *ResourceMatcher) Match(pattern, resource string, context map[string]interface{}) bool {
	if pattern == "*" {
		return true
	}

	// Validate resource format before matching
	if !rm.validateResourceFormat(resource) {
		return false
	}

	// Substitute variables in pattern
	expandedPattern := rm.substituteVariables(pattern, context)

	// Validate expanded pattern format (after variable substitution)
	if !rm.validateResourceFormat(expandedPattern) && expandedPattern != "*" {
		return false
	}

	// Handle hierarchical resources
	if strings.Contains(expandedPattern, "/") || strings.Contains(resource, "/") {
		return rm.matchHierarchical(expandedPattern, resource)
	}

	// Simple resource matching
	return rm.matchSimple(expandedPattern, resource)
}

// matchSimple handles simple resource pattern matching
func (rm *ResourceMatcher) matchSimple(pattern, resource string) bool {
	patternParts := strings.Split(pattern, ":")
	resourceParts := strings.Split(resource, ":")

	if len(patternParts) != len(resourceParts) {
		return false
	}

	for i := 0; i < len(patternParts); i++ {
		if !rm.matchSegment(patternParts[i], resourceParts[i]) {
			return false
		}
	}
	return true
}

// matchHierarchical handles hierarchical resource pattern matching
func (rm *ResourceMatcher) matchHierarchical(pattern, resource string) bool {
	// Split by '/' first, then by ':'
	patternParts := rm.parseHierarchical(pattern)
	resourceParts := rm.parseHierarchical(resource)

	if len(patternParts) != len(resourceParts) {
		return false
	}

	for i := 0; i < len(patternParts); i++ {
		if !rm.matchSimple(patternParts[i], resourceParts[i]) {
			return false
		}
	}
	return true
}

// parseHierarchical parses hierarchical resource path
func (rm *ResourceMatcher) parseHierarchical(path string) []string {
	return strings.Split(path, "/")
}

// matchSegment matches a single segment with wildcard support
func (rm *ResourceMatcher) matchSegment(pattern, value string) bool {
	if pattern == "*" {
		return true
	}
	if !strings.Contains(pattern, "*") {
		return pattern == value
	}
	return rm.matchWildcard(pattern, value)
}

// matchWildcard converts wildcard pattern to regex and matches
func (rm *ResourceMatcher) matchWildcard(pattern, value string) bool {
	// Convert wildcard pattern to regex
	regexPattern := strings.ReplaceAll(pattern, "*", ".*")
	regexPattern = "^" + regexPattern + "$"

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return false
	}

	return regex.MatchString(value)
}

// hasVariables checks if a string contains variable substitutions
func (rm *ResourceMatcher) hasVariables(str string) bool {
	return strings.Contains(str, "${") && strings.Contains(str, "}")
}

// validateResourceFormat validates resource format according to specification
// Format: <service>:<resource-type>:<resource-id> or hierarchical with '/'
func (rm *ResourceMatcher) validateResourceFormat(resource string) bool {
	if resource == "*" {
		return true
	}

	// Skip validation if resource contains variables (validate after substitution)
	if rm.hasVariables(resource) {
		return true
	}

	// Handle hierarchical resources
	if strings.Contains(resource, "/") {
		parts := strings.Split(resource, "/")
		for _, part := range parts {
			if !rm.validateSimpleResourceFormat(part) {
				return false
			}
		}
		return true
	}

	return rm.validateSimpleResourceFormat(resource)
}

// validateSimpleResourceFormat validates simple resource format (no hierarchy)
func (rm *ResourceMatcher) validateSimpleResourceFormat(resource string) bool {
	parts := strings.Split(resource, ":")

	// Must have at least 3 parts: service:type:id
	if len(parts) < 3 {
		return false
	}

	// No empty segments (except wildcards)
	for _, part := range parts {
		if part == "" {
			return false
		}
		// Allow variables in format ${...}
		if strings.HasPrefix(part, "${") && strings.HasSuffix(part, "}") {
			continue
		}
	}

	return true
}

// substituteVariables replaces ${...} variables in pattern
func (rm *ResourceMatcher) substituteVariables(pattern string, context map[string]interface{}) string {
	result := pattern

	// Find all ${...} patterns
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(pattern, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			varName := match[1]
			if value, exists := context[varName]; exists {
				if strValue, ok := value.(string); ok {
					result = strings.ReplaceAll(result, match[0], strValue)
				}
			}
		}
	}

	return result
}
