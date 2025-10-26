package matchers

import (
	"testing"
)

func TestActionMatcher_Match(t *testing.T) {
	matcher := NewActionMatcher()

	tests := []struct {
		name     string
		pattern  string
		action   string
		expected bool
	}{
		// Basic wildcard tests
		{
			name:     "universal wildcard matches everything",
			pattern:  "*",
			action:   "admin:user:delete",
			expected: true,
		},
		{
			name:     "exact match",
			pattern:  "admin:user:delete",
			action:   "admin:user:delete",
			expected: true,
		},
		{
			name:     "exact mismatch",
			pattern:  "admin:user:create",
			action:   "admin:user:delete",
			expected: false,
		},

		// Trailing wildcard tests - the main fix
		{
			name:     "trailing wildcard matches longer action - case 1",
			pattern:  "admin:*",
			action:   "admin:user:delete",
			expected: true,
		},
		{
			name:     "trailing wildcard matches longer action - case 2",
			pattern:  "admin:*",
			action:   "admin:role:create:bulk",
			expected: true,
		},
		{
			name:     "trailing wildcard matches exact length",
			pattern:  "admin:*",
			action:   "admin:user",
			expected: true,
		},
		{
			name:     "trailing wildcard fails on shorter action",
			pattern:  "admin:user:*",
			action:   "admin",
			expected: false,
		},
		{
			name:     "trailing wildcard fails on prefix mismatch",
			pattern:  "admin:*",
			action:   "user:role:delete",
			expected: false,
		},

		// Multiple segment prefix with trailing wildcard
		{
			name:     "multi-segment prefix with trailing wildcard",
			pattern:  "admin:user:*",
			action:   "admin:user:delete:force",
			expected: true,
		},
		{
			name:     "multi-segment prefix mismatch with trailing wildcard",
			pattern:  "admin:role:*",
			action:   "admin:user:delete",
			expected: false,
		},

		// Standard segment wildcard tests (existing functionality)
		{
			name:     "middle segment wildcard",
			pattern:  "admin:*:delete",
			action:   "admin:user:delete",
			expected: true,
		},
		{
			name:     "first segment wildcard",
			pattern:  "*:user:delete",
			action:   "admin:user:delete",
			expected: true,
		},

		// Edge cases
		{
			name:     "empty pattern and action",
			pattern:  "",
			action:   "",
			expected: true,
		},
		{
			name:     "single segment with trailing wildcard",
			pattern:  "*",
			action:   "admin:user:delete",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.Match(tt.pattern, tt.action)
			if result != tt.expected {
				t.Errorf("ActionMatcher.Match(%q, %q) = %v, expected %v",
					tt.pattern, tt.action, result, tt.expected)
			}
		})
	}
}

func TestActionMatcher_MatchSegment(t *testing.T) {
	matcher := NewActionMatcher()

	tests := []struct {
		name     string
		pattern  string
		value    string
		expected bool
	}{
		{
			name:     "exact match",
			pattern:  "admin",
			value:    "admin",
			expected: true,
		},
		{
			name:     "wildcard match",
			pattern:  "*",
			value:    "anything",
			expected: true,
		},
		{
			name:     "prefix wildcard match",
			pattern:  "admin-*",
			value:    "admin-user",
			expected: true,
		},
		{
			name:     "suffix wildcard match",
			pattern:  "*-service",
			value:    "user-service",
			expected: true,
		},
		{
			name:     "middle wildcard match",
			pattern:  "admin-*-service",
			value:    "admin-user-service",
			expected: true,
		},
		{
			name:     "no match",
			pattern:  "admin",
			value:    "user",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.matchSegment(tt.pattern, tt.value)
			if result != tt.expected {
				t.Errorf("ActionMatcher.matchSegment(%q, %q) = %v, expected %v",
					tt.pattern, tt.value, result, tt.expected)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkActionMatcher_Match(b *testing.B) {
	matcher := NewActionMatcher()

	testCases := []struct {
		pattern string
		action  string
	}{
		{"admin:*", "admin:user:delete"},
		{"admin:user:*", "admin:user:delete:force:confirm"},
		{"*:user:delete", "admin:user:delete"},
		{"admin:user:delete", "admin:user:delete"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			matcher.Match(tc.pattern, tc.action)
		}
	}
}
