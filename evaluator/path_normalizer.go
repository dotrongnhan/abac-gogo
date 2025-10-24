package evaluator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PathInfo contains parsed and normalized path information
type PathInfo struct {
	// Original raw path
	Raw string
	// Normalized path parts (without array indices)
	Parts []string
	// Map of part index to array index (e.g., parts[2] accesses array[5])
	ArrayIndices map[int]int
	// Whether this path contains array access
	HasArrayAccess bool
}

// PathNormalizer validates and normalizes attribute paths
type PathNormalizer struct {
	// Regex for array index notation: field[0] or field.0
	arrayIndexPattern *regexp.Regexp
}

// NewPathNormalizer creates a new path normalizer
func NewPathNormalizer() *PathNormalizer {
	return &PathNormalizer{
		arrayIndexPattern: regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\[(\d+)\]$|^([a-zA-Z_][a-zA-Z0-9_]*)\.(\d+)$`),
	}
}

// NormalizePath validates and normalizes a path string
func (pn *PathNormalizer) NormalizePath(path string) (*PathInfo, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	// Remove leading/trailing whitespace
	path = strings.TrimSpace(path)

	if path == "" {
		return nil, fmt.Errorf("path cannot be empty after trimming")
	}

	// Split by dots
	rawParts := strings.Split(path, ".")

	info := &PathInfo{
		Raw:            path,
		Parts:          make([]string, 0, len(rawParts)),
		ArrayIndices:   make(map[int]int),
		HasArrayAccess: false,
	}

	for _, part := range rawParts {
		// Skip empty parts (e.g., from "user..department")
		if part == "" {
			continue
		}

		// Check if this part has array index notation
		if strings.Contains(part, "[") || pn.isNumericPart(part, len(info.Parts) > 0) {
			fieldName, arrayIndex, err := pn.parseArrayAccess(part, len(info.Parts) > 0)
			if err != nil {
				return nil, fmt.Errorf("invalid array access in '%s': %w", part, err)
			}

			if fieldName != "" {
				// Add field name if it exists (e.g., "roles" from "roles[0]")
				info.Parts = append(info.Parts, fieldName)
			}

			if arrayIndex >= 0 {
				// Record array index for this position
				info.ArrayIndices[len(info.Parts)-1] = arrayIndex
				info.HasArrayAccess = true
			}
		} else {
			// Regular field name
			if !pn.isValidIdentifier(part) {
				return nil, fmt.Errorf("invalid identifier: '%s'", part)
			}
			info.Parts = append(info.Parts, part)
		}
	}

	if len(info.Parts) == 0 {
		return nil, fmt.Errorf("path must contain at least one valid part")
	}

	return info, nil
}

// parseArrayAccess parses array access notation
// Supports: field[0], field.0 (when isFollowing is true)
// Returns: fieldName, arrayIndex, error
func (pn *PathNormalizer) parseArrayAccess(part string, isFollowing bool) (string, int, error) {
	// Try bracket notation: field[0]
	if idx := strings.Index(part, "["); idx >= 0 {
		if !strings.HasSuffix(part, "]") {
			return "", -1, fmt.Errorf("unclosed bracket in '%s'", part)
		}

		fieldName := part[:idx]
		indexStr := part[idx+1 : len(part)-1]

		// Validate field name if it exists
		if fieldName != "" && !pn.isValidIdentifier(fieldName) {
			return "", -1, fmt.Errorf("invalid field name '%s'", fieldName)
		}

		// Parse index
		arrayIndex, err := strconv.Atoi(indexStr)
		if err != nil || arrayIndex < 0 {
			return "", -1, fmt.Errorf("invalid array index '%s'", indexStr)
		}

		return fieldName, arrayIndex, nil
	}

	// Try dot notation for numeric index: .0, .1, etc.
	// Only when following another part (e.g., "roles.0" not "0")
	if isFollowing && pn.isNumericPart(part, true) {
		arrayIndex, err := strconv.Atoi(part)
		if err != nil || arrayIndex < 0 {
			return "", -1, fmt.Errorf("invalid array index '%s'", part)
		}
		return "", arrayIndex, nil
	}

	return part, -1, nil
}

// isNumericPart checks if a part is purely numeric
func (pn *PathNormalizer) isNumericPart(part string, allowAsIndex bool) bool {
	if !allowAsIndex {
		return false
	}
	_, err := strconv.Atoi(part)
	return err == nil
}

// isValidIdentifier checks if a string is a valid identifier
func (pn *PathNormalizer) isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// Must start with letter or underscore
	first := rune(s[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Rest can be letters, digits, or underscores
	for _, r := range s[1:] {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	return true
}

// ParsePath is a convenience function to parse a path
func ParsePath(path string) (*PathInfo, error) {
	normalizer := NewPathNormalizer()
	return normalizer.NormalizePath(path)
}
