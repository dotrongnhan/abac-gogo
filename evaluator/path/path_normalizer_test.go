package path

import (
	"testing"
)

func TestPathNormalizer_BasicPaths(t *testing.T) {
	normalizer := NewPathNormalizer()

	tests := []struct {
		name          string
		path          string
		expectedParts []string
		shouldError   bool
	}{
		{
			name:          "Simple path",
			path:          "user",
			expectedParts: []string{"user"},
			shouldError:   false,
		},
		{
			name:          "Nested path",
			path:          "user.department",
			expectedParts: []string{"user", "department"},
			shouldError:   false,
		},
		{
			name:          "Deep nested path",
			path:          "user.profile.work.department",
			expectedParts: []string{"user", "profile", "work", "department"},
			shouldError:   false,
		},
		{
			name:          "Path with underscores",
			path:          "user_profile.work_location",
			expectedParts: []string{"user_profile", "work_location"},
			shouldError:   false,
		},
		{
			name:        "Empty path",
			path:        "",
			shouldError: true,
		},
		{
			name:        "Only whitespace",
			path:        "   ",
			shouldError: true,
		},
		{
			name:          "Path with extra spaces (trimmed)",
			path:          "  user.department  ",
			expectedParts: []string{"user", "department"},
			shouldError:   false,
		},
		{
			name:          "Path with double dots (filtered)",
			path:          "user..department",
			expectedParts: []string{"user", "department"},
			shouldError:   false,
		},
		{
			name:        "Invalid identifier - starts with number",
			path:        "9user",
			shouldError: true,
		},
		{
			name:        "Invalid identifier - special characters",
			path:        "user-name",
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			info, err := normalizer.NormalizePath(test.path)

			if test.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(info.Parts) != len(test.expectedParts) {
				t.Errorf("Expected %d parts, got %d", len(test.expectedParts), len(info.Parts))
				return
			}

			for i, expected := range test.expectedParts {
				if info.Parts[i] != expected {
					t.Errorf("Part %d: expected '%s', got '%s'", i, expected, info.Parts[i])
				}
			}

			if info.HasArrayAccess {
				t.Errorf("Expected no array access, but HasArrayAccess is true")
			}
		})
	}
}

func TestPathNormalizer_ArrayAccess(t *testing.T) {
	normalizer := NewPathNormalizer()

	tests := []struct {
		name           string
		path           string
		expectedParts  []string
		expectedArrays map[int]int
		shouldError    bool
	}{
		{
			name:           "Array access with brackets",
			path:           "user.roles[0]",
			expectedParts:  []string{"user", "roles"},
			expectedArrays: map[int]int{1: 0},
			shouldError:    false,
		},
		{
			name:           "Array access with dot notation",
			path:           "user.roles.0",
			expectedParts:  []string{"user", "roles"},
			expectedArrays: map[int]int{1: 0},
			shouldError:    false,
		},
		{
			name:           "Multiple array accesses",
			path:           "data.items[0].tags[2]",
			expectedParts:  []string{"data", "items", "tags"},
			expectedArrays: map[int]int{1: 0, 2: 2},
			shouldError:    false,
		},
		{
			name:           "Array access then property",
			path:           "user.roles[0].name",
			expectedParts:  []string{"user", "roles", "name"},
			expectedArrays: map[int]int{1: 0},
			shouldError:    false,
		},
		{
			name:           "Nested array access mixed notation",
			path:           "data.items[0].tags.1.value",
			expectedParts:  []string{"data", "items", "tags", "value"},
			expectedArrays: map[int]int{1: 0, 2: 1},
			shouldError:    false,
		},
		{
			name:        "Invalid bracket - unclosed",
			path:        "user.roles[0",
			shouldError: true,
		},
		{
			name:        "Invalid bracket - not at end",
			path:        "user.roles[0]extra",
			shouldError: true,
		},
		{
			name:        "Invalid array index - negative",
			path:        "user.roles[-1]",
			shouldError: true,
		},
		{
			name:        "Invalid array index - non-numeric",
			path:        "user.roles[abc]",
			shouldError: true,
		},
		{
			name:           "Large array index",
			path:           "data.items[999]",
			expectedParts:  []string{"data", "items"},
			expectedArrays: map[int]int{1: 999},
			shouldError:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			info, err := normalizer.NormalizePath(test.path)

			if test.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(info.Parts) != len(test.expectedParts) {
				t.Errorf("Expected %d parts, got %d", len(test.expectedParts), len(info.Parts))
				return
			}

			for i, expected := range test.expectedParts {
				if info.Parts[i] != expected {
					t.Errorf("Part %d: expected '%s', got '%s'", i, expected, info.Parts[i])
				}
			}

			if !info.HasArrayAccess {
				t.Errorf("Expected array access, but HasArrayAccess is false")
			}

			if len(info.ArrayIndices) != len(test.expectedArrays) {
				t.Errorf("Expected %d array indices, got %d", len(test.expectedArrays), len(info.ArrayIndices))
			}

			for partIdx, expectedArrayIdx := range test.expectedArrays {
				if actualArrayIdx, exists := info.ArrayIndices[partIdx]; !exists {
					t.Errorf("Missing array index at part %d", partIdx)
				} else if actualArrayIdx != expectedArrayIdx {
					t.Errorf("Part %d: expected array index %d, got %d", partIdx, expectedArrayIdx, actualArrayIdx)
				}
			}
		})
	}
}

func TestPathNormalizer_EdgeCases(t *testing.T) {
	normalizer := NewPathNormalizer()

	tests := []struct {
		name        string
		path        string
		shouldError bool
	}{
		{
			name:        "Leading dot",
			path:        ".user",
			shouldError: false, // Empty part filtered, becomes just "user"
		},
		{
			name:        "Trailing dot",
			path:        "user.",
			shouldError: false, // Empty part filtered, becomes just "user"
		},
		{
			name:        "Multiple consecutive dots",
			path:        "user...department",
			shouldError: false, // Empty parts filtered
		},
		{
			name:        "Only dots",
			path:        "...",
			shouldError: true, // No valid parts remain
		},
		{
			name:        "Array index at start (not allowed as field name)",
			path:        "0.user",
			shouldError: true, // "0" at start is not a valid identifier
		},
		{
			name:        "Empty brackets",
			path:        "user.roles[]",
			shouldError: true,
		},
		{
			name:        "Space in identifier",
			path:        "user name.department",
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := normalizer.NormalizePath(test.path)

			if test.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !test.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestParsePath_Convenience(t *testing.T) {
	// Test the convenience function
	info, err := ParsePath("user.roles[0].name")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(info.Parts) != 3 {
		t.Errorf("Expected 3 parts, got %d", len(info.Parts))
	}

	if !info.HasArrayAccess {
		t.Errorf("Expected array access")
	}

	if info.Raw != "user.roles[0].name" {
		t.Errorf("Expected raw path to be preserved")
	}
}
