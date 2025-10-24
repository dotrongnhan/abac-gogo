package evaluator

import (
	"strings"
)

// PathResolver defines the interface for resolving attribute paths in context
type PathResolver interface {
	Resolve(path string, context map[string]interface{}) (value interface{}, found bool)
}

// CompositePathResolver orchestrates multiple path resolution strategies
type CompositePathResolver struct {
	resolvers []PathResolver
}

// NewCompositePathResolver creates a new composite resolver with default strategies
func NewCompositePathResolver() *CompositePathResolver {
	return &CompositePathResolver{
		resolvers: []PathResolver{
			&DirectPathResolver{},
			&DotNotationResolver{},
			&ColonFallbackResolver{},
			&ShortcutResolver{},
		},
	}
}

// Resolve tries each resolver in order until one succeeds
func (cpr *CompositePathResolver) Resolve(path string, context map[string]interface{}) (interface{}, bool) {
	for _, resolver := range cpr.resolvers {
		if value, found := resolver.Resolve(path, context); found {
			return value, true
		}
	}
	return nil, false
}

// DirectPathResolver looks up the path directly in the context map
type DirectPathResolver struct{}

func (dpr *DirectPathResolver) Resolve(path string, context map[string]interface{}) (interface{}, bool) {
	value, exists := context[path]
	return value, exists
}

// DotNotationResolver handles nested path access using dot notation
type DotNotationResolver struct{}

func (dnr *DotNotationResolver) Resolve(path string, context map[string]interface{}) (interface{}, bool) {
	// Only process paths with dots
	if !strings.Contains(path, ".") {
		return nil, false
	}

	parts := strings.Split(path, ".")
	return navigateNestedMap(parts, context)
}

// ColonFallbackResolver converts dot notation to colon notation and tries direct access
// This supports legacy flat key formats like "user:department"
type ColonFallbackResolver struct{}

func (cfr *ColonFallbackResolver) Resolve(path string, context map[string]interface{}) (interface{}, bool) {
	// Only try if path contains dots
	if !strings.Contains(path, ".") {
		return nil, false
	}

	// Convert first dot to colon: "user.department" -> "user:department"
	flatPath := strings.Replace(path, ".", ":", 1)
	value, exists := context[flatPath]
	return value, exists
}

// ShortcutResolver handles special shortcuts for common patterns
// For example: "user.department" might map to "user.attributes.department"
type ShortcutResolver struct{}

func (sr *ShortcutResolver) Resolve(path string, context map[string]interface{}) (interface{}, bool) {
	// Only process paths with dots
	if !strings.Contains(path, ".") {
		return nil, false
	}

	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return nil, false
	}

	firstPart := parts[0]

	// Check if this is a shortcut-enabled prefix (user or resource)
	if firstPart != "user" && firstPart != "resource" {
		return nil, false
	}

	// Try to find the structured context
	structuredContext, exists := context[firstPart]
	if !exists {
		return nil, false
	}

	structMap, ok := structuredContext.(map[string]interface{})
	if !ok {
		return nil, false
	}

	// Check if it has an "attributes" sub-map
	attributes, exists := structMap["attributes"]
	if !exists {
		return nil, false
	}

	attrMap, ok := attributes.(map[string]interface{})
	if !ok {
		return nil, false
	}

	// Navigate the remaining path in the attributes map
	remainingParts := parts[1:]
	return navigateNestedMap(remainingParts, attrMap)
}

// navigateNestedMap is a unified function to navigate through nested maps
// It replaces the duplicated logic in getNestedValue and getNestedValueFromMap
func navigateNestedMap(parts []string, startMap map[string]interface{}) (interface{}, bool) {
	current := startMap

	for i, part := range parts {
		// Last part - return the value
		if i == len(parts)-1 {
			if value, exists := current[part]; exists {
				return value, true
			}
			return nil, false
		}

		// Navigate deeper
		nextLevel, exists := current[part]
		if !exists {
			return nil, false
		}

		// Type assert to map for next iteration
		nextMap, ok := nextLevel.(map[string]interface{})
		if !ok {
			// Can't navigate deeper if not a map
			return nil, false
		}

		current = nextMap
	}

	return nil, false
}
