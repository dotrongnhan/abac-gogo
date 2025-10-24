package evaluator

import (
	"testing"
)

func TestDirectPathResolver(t *testing.T) {
	resolver := &DirectPathResolver{}

	tests := []struct {
		name     string
		path     string
		context  map[string]interface{}
		expected interface{}
		found    bool
	}{
		{
			name: "Existing key",
			path: "user_id",
			context: map[string]interface{}{
				"user_id": "12345",
			},
			expected: "12345",
			found:    true,
		},
		{
			name: "Non-existing key",
			path: "missing",
			context: map[string]interface{}{
				"user_id": "12345",
			},
			expected: nil,
			found:    false,
		},
		{
			name: "Empty string path",
			path: "",
			context: map[string]interface{}{
				"user_id": "12345",
			},
			expected: nil,
			found:    false,
		},
		{
			name: "Colon notation key",
			path: "user:department",
			context: map[string]interface{}{
				"user:department": "Engineering",
			},
			expected: "Engineering",
			found:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, found := resolver.Resolve(test.path, test.context)
			if found != test.found {
				t.Errorf("Expected found=%v, got found=%v", test.found, found)
			}
			if value != test.expected {
				t.Errorf("Expected value=%v, got value=%v", test.expected, value)
			}
		})
	}
}

func TestDotNotationResolver(t *testing.T) {
	resolver := &DotNotationResolver{}

	tests := []struct {
		name     string
		path     string
		context  map[string]interface{}
		expected interface{}
		found    bool
	}{
		{
			name: "Simple nested path",
			path: "user.department",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department": "Engineering",
				},
			},
			expected: "Engineering",
			found:    true,
		},
		{
			name: "Deep nested path",
			path: "user.profile.work.department",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"work": map[string]interface{}{
							"department": "Engineering",
						},
					},
				},
			},
			expected: "Engineering",
			found:    true,
		},
		{
			name: "Non-existent nested path",
			path: "user.profile.missing",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"work": "data",
					},
				},
			},
			expected: nil,
			found:    false,
		},
		{
			name: "Path without dots - should not process",
			path: "user",
			context: map[string]interface{}{
				"user": "data",
			},
			expected: nil,
			found:    false,
		},
		{
			name: "Broken path - intermediate not a map",
			path: "user.profile.work",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": "string-value", // Not a map!
				},
			},
			expected: nil,
			found:    false,
		},
		{
			name: "First part missing",
			path: "missing.field",
			context: map[string]interface{}{
				"user": "data",
			},
			expected: nil,
			found:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, found := resolver.Resolve(test.path, test.context)
			if found != test.found {
				t.Errorf("Expected found=%v, got found=%v", test.found, found)
			}
			if value != test.expected {
				t.Errorf("Expected value=%v, got value=%v", test.expected, value)
			}
		})
	}
}

func TestColonFallbackResolver(t *testing.T) {
	resolver := &ColonFallbackResolver{}

	tests := []struct {
		name     string
		path     string
		context  map[string]interface{}
		expected interface{}
		found    bool
	}{
		{
			name: "Convert dot to colon successfully",
			path: "user.department",
			context: map[string]interface{}{
				"user:department": "Engineering",
			},
			expected: "Engineering",
			found:    true,
		},
		{
			name: "Path without dots - should not process",
			path: "user",
			context: map[string]interface{}{
				"user": "data",
			},
			expected: nil,
			found:    false,
		},
		{
			name: "Non-existent colon path",
			path: "user.missing",
			context: map[string]interface{}{
				"user:department": "Engineering",
			},
			expected: nil,
			found:    false,
		},
		{
			name: "Multiple dots - only first converted",
			path: "user.profile.work",
			context: map[string]interface{}{
				"user:profile.work": "Engineering",
			},
			expected: "Engineering",
			found:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, found := resolver.Resolve(test.path, test.context)
			if found != test.found {
				t.Errorf("Expected found=%v, got found=%v", test.found, found)
			}
			if value != test.expected {
				t.Errorf("Expected value=%v, got value=%v", test.expected, value)
			}
		})
	}
}

func TestShortcutResolver(t *testing.T) {
	resolver := &ShortcutResolver{}

	tests := []struct {
		name     string
		path     string
		context  map[string]interface{}
		expected interface{}
		found    bool
	}{
		{
			name: "user.x -> user.attributes.x",
			path: "user.department",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"attributes": map[string]interface{}{
						"department": "Engineering",
					},
				},
			},
			expected: "Engineering",
			found:    true,
		},
		{
			name: "resource.x -> resource.attributes.x",
			path: "resource.classification",
			context: map[string]interface{}{
				"resource": map[string]interface{}{
					"attributes": map[string]interface{}{
						"classification": "confidential",
					},
				},
			},
			expected: "confidential",
			found:    true,
		},
		{
			name: "Deep nested in attributes",
			path: "user.profile.work",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"attributes": map[string]interface{}{
						"profile": map[string]interface{}{
							"work": "Engineering",
						},
					},
				},
			},
			expected: "Engineering",
			found:    true,
		},
		{
			name: "Non-shortcut prefix - should not process",
			path: "environment.time",
			context: map[string]interface{}{
				"environment": map[string]interface{}{
					"attributes": map[string]interface{}{
						"time": "14:00",
					},
				},
			},
			expected: nil,
			found:    false,
		},
		{
			name: "Missing attributes map",
			path: "user.department",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"id": "123",
				},
			},
			expected: nil,
			found:    false,
		},
		{
			name: "Attributes is not a map",
			path: "user.department",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"attributes": "string-value",
				},
			},
			expected: nil,
			found:    false,
		},
		{
			name: "Path without dots",
			path: "user",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"attributes": map[string]interface{}{
						"department": "Engineering",
					},
				},
			},
			expected: nil,
			found:    false,
		},
		{
			name: "User context is not a map",
			path: "user.department",
			context: map[string]interface{}{
				"user": "string-value",
			},
			expected: nil,
			found:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, found := resolver.Resolve(test.path, test.context)
			if found != test.found {
				t.Errorf("Expected found=%v, got found=%v", test.found, found)
			}
			if value != test.expected {
				t.Errorf("Expected value=%v, got value=%v", test.expected, value)
			}
		})
	}
}

func TestCompositePathResolver(t *testing.T) {
	resolver := NewCompositePathResolver()

	tests := []struct {
		name     string
		path     string
		context  map[string]interface{}
		expected interface{}
		found    bool
	}{
		{
			name: "Direct path - highest priority",
			path: "user.department",
			context: map[string]interface{}{
				"user.department": "Direct",
				"user": map[string]interface{}{
					"department": "Nested",
				},
			},
			expected: "Direct",
			found:    true,
		},
		{
			name: "Dot notation - second priority",
			path: "user.department",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department": "Engineering",
				},
			},
			expected: "Engineering",
			found:    true,
		},
		{
			name: "Colon fallback - third priority",
			path: "user.department",
			context: map[string]interface{}{
				"user:department": "Engineering",
			},
			expected: "Engineering",
			found:    true,
		},
		{
			name: "Shortcut resolver - fourth priority",
			path: "user.department",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"attributes": map[string]interface{}{
						"department": "Engineering",
					},
				},
			},
			expected: "Engineering",
			found:    true,
		},
		{
			name: "None found",
			path: "missing.path",
			context: map[string]interface{}{
				"user": "data",
			},
			expected: nil,
			found:    false,
		},
		{
			name: "Priority test - direct wins over nested",
			path: "resource.type",
			context: map[string]interface{}{
				"resource.type": "direct-value",
				"resource": map[string]interface{}{
					"type": "nested-value",
					"attributes": map[string]interface{}{
						"type": "shortcut-value",
					},
				},
			},
			expected: "direct-value",
			found:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, found := resolver.Resolve(test.path, test.context)
			if found != test.found {
				t.Errorf("Expected found=%v, got found=%v", test.found, found)
			}
			if value != test.expected {
				t.Errorf("Expected value=%v, got value=%v", test.expected, value)
			}
		})
	}
}

func TestNavigateNestedMap(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		startMap map[string]interface{}
		expected interface{}
		found    bool
	}{
		{
			name:  "Single part",
			parts: []string{"department"},
			startMap: map[string]interface{}{
				"department": "Engineering",
			},
			expected: "Engineering",
			found:    true,
		},
		{
			name:  "Multiple parts",
			parts: []string{"profile", "work", "department"},
			startMap: map[string]interface{}{
				"profile": map[string]interface{}{
					"work": map[string]interface{}{
						"department": "Engineering",
					},
				},
			},
			expected: "Engineering",
			found:    true,
		},
		{
			name:  "Missing intermediate part",
			parts: []string{"profile", "missing", "department"},
			startMap: map[string]interface{}{
				"profile": map[string]interface{}{
					"work": "data",
				},
			},
			expected: nil,
			found:    false,
		},
		{
			name:  "Intermediate part not a map",
			parts: []string{"profile", "work"},
			startMap: map[string]interface{}{
				"profile": "string-value",
			},
			expected: nil,
			found:    false,
		},
		{
			name:     "Empty parts",
			parts:    []string{},
			startMap: map[string]interface{}{},
			expected: nil,
			found:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, found := navigateNestedMap(test.parts, test.startMap)
			if found != test.found {
				t.Errorf("Expected found=%v, got found=%v", test.found, found)
			}
			if value != test.expected {
				t.Errorf("Expected value=%v, got value=%v", test.expected, value)
			}
		})
	}
}
