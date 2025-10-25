package path

import (
	"strings"
)

// ShortcutConfig defines a path shortcut mapping
type ShortcutConfig struct {
	// Prefix is the first part of the path (e.g., "user", "resource", "environment")
	Prefix string
	// TargetPath is the intermediate path to insert (e.g., ["attributes"], ["metadata"])
	TargetPath []string
}

// PathResolver defines the interface for resolving attribute paths in context
type PathResolver interface {
	Resolve(path string, context map[string]interface{}) (value interface{}, found bool)
}

// CompositePathResolver orchestrates multiple path resolution strategies
type CompositePathResolver struct {
	resolvers []PathResolver
}

// DefaultShortcuts returns the default shortcut configurations
func DefaultShortcuts() []ShortcutConfig {
	return []ShortcutConfig{
		{Prefix: "user", TargetPath: []string{"attributes"}},
		{Prefix: "resource", TargetPath: []string{"attributes"}},
	}
}

// NewCompositePathResolver creates a new composite resolver with default strategies
func NewCompositePathResolver() *CompositePathResolver {
	return NewCompositePathResolverWithShortcuts(DefaultShortcuts())
}

// NewCompositePathResolverWithShortcuts creates a new composite resolver with custom shortcut configs
func NewCompositePathResolverWithShortcuts(shortcuts []ShortcutConfig) *CompositePathResolver {
	return &CompositePathResolver{
		resolvers: []PathResolver{
			&DirectPathResolver{},
			NewArrayAccessResolver(), // High priority for array access
			&DotNotationResolver{},
			&ColonFallbackResolver{},
			NewShortcutResolver(shortcuts),
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

// NewDotNotationResolver creates a new dot notation resolver
func NewDotNotationResolver() *DotNotationResolver {
	return &DotNotationResolver{}
}

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

// ShortcutResolver handles configurable shortcuts for common patterns
// For example: "user.department" might map to "user.attributes.department"
type ShortcutResolver struct {
	shortcuts map[string][]string // prefix -> target path
}

// NewShortcutResolver creates a new shortcut resolver with given configurations
func NewShortcutResolver(configs []ShortcutConfig) *ShortcutResolver {
	shortcuts := make(map[string][]string)
	for _, config := range configs {
		shortcuts[config.Prefix] = config.TargetPath
	}
	return &ShortcutResolver{
		shortcuts: shortcuts,
	}
}

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

	// Check if this prefix has a configured shortcut
	targetPath, hasShortcut := sr.shortcuts[firstPart]
	if !hasShortcut {
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

	// Navigate to the target path (e.g., ["attributes"])
	current := structMap
	for _, pathPart := range targetPath {
		nextLevel, exists := current[pathPart]
		if !exists {
			return nil, false
		}

		nextMap, ok := nextLevel.(map[string]interface{})
		if !ok {
			return nil, false
		}

		current = nextMap
	}

	// Navigate the remaining path in the target map
	remainingParts := parts[1:]
	return navigateNestedMap(remainingParts, current)
}

// ArrayAccessResolver handles array index access in paths
// Supports: user.roles[0], user.roles.0
type ArrayAccessResolver struct {
	normalizer *PathNormalizer
}

// NewArrayAccessResolver creates a new array access resolver
func NewArrayAccessResolver() *ArrayAccessResolver {
	return &ArrayAccessResolver{
		normalizer: NewPathNormalizer(),
	}
}

func (aar *ArrayAccessResolver) Resolve(path string, context map[string]interface{}) (interface{}, bool) {
	// Only process paths with array notation
	if !strings.Contains(path, "[") && !aar.hasNumericPart(path) {
		return nil, false
	}

	// Parse path to get array indices
	pathInfo, err := aar.normalizer.NormalizePath(path)
	if err != nil || !pathInfo.HasArrayAccess {
		return nil, false
	}

	// Navigate with array access support
	return navigateWithArrayAccess(pathInfo.Parts, pathInfo.ArrayIndices, context)
}

// hasNumericPart checks if path has numeric parts like "roles.0"
func (aar *ArrayAccessResolver) hasNumericPart(path string) bool {
	parts := strings.Split(path, ".")
	for i, part := range parts {
		// Skip first part (can't be just a number)
		if i == 0 {
			continue
		}
		// Check if part is purely numeric
		if len(part) > 0 && part[0] >= '0' && part[0] <= '9' {
			return true
		}
	}
	return false
}

// navigateWithArrayAccess navigates through nested maps and arrays
func navigateWithArrayAccess(parts []string, arrayIndices map[int]int, startMap map[string]interface{}) (interface{}, bool) {
	var current interface{} = startMap

	for i, part := range parts {
		// Check if we need to access an array at this position
		if arrayIndex, hasArrayAccess := arrayIndices[i]; hasArrayAccess {
			// First, navigate to the field (which should be an array)
			currentMap, ok := current.(map[string]interface{})
			if !ok {
				return nil, false
			}

			arrayField, exists := currentMap[part]
			if !exists {
				return nil, false
			}

			// Now access the array
			currentArray, ok := arrayField.([]interface{})
			if !ok {
				return nil, false
			}

			// Check bounds
			if arrayIndex < 0 || arrayIndex >= len(currentArray) {
				return nil, false
			}

			// Access array element
			current = currentArray[arrayIndex]

			// If this is the last part, return the value
			if i == len(parts)-1 {
				return current, true
			}

			// Otherwise, current should be a map for next iteration
			// (continue will handle the type check in next iteration)
			continue
		}

		// Regular map navigation
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}

		// Last part - return the value
		if i == len(parts)-1 {
			if value, exists := currentMap[part]; exists {
				return value, true
			}
			return nil, false
		}

		// Navigate deeper
		nextLevel, exists := currentMap[part]
		if !exists {
			return nil, false
		}

		current = nextLevel
	}

	return nil, false
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
