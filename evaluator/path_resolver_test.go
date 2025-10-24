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
	// Use default shortcuts config
	resolver := NewShortcutResolver(DefaultShortcuts())

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

func TestShortcutResolver_CustomConfig(t *testing.T) {
	// Test with custom shortcut configurations
	customShortcuts := []ShortcutConfig{
		{Prefix: "user", TargetPath: []string{"attributes"}},
		{Prefix: "resource", TargetPath: []string{"metadata"}}, // Different from default
		{Prefix: "environment", TargetPath: []string{"properties"}}, // New shortcut
	}
	resolver := NewShortcutResolver(customShortcuts)

	tests := []struct {
		name     string
		path     string
		context  map[string]interface{}
		expected interface{}
		found    bool
	}{
		{
			name: "resource.x -> resource.metadata.x (custom config)",
			path: "resource.classification",
			context: map[string]interface{}{
				"resource": map[string]interface{}{
					"metadata": map[string]interface{}{
						"classification": "top-secret",
					},
				},
			},
			expected: "top-secret",
			found:    true,
		},
		{
			name: "environment.x -> environment.properties.x (new shortcut)",
			path: "environment.time",
			context: map[string]interface{}{
				"environment": map[string]interface{}{
					"properties": map[string]interface{}{
						"time": "14:30",
					},
				},
			},
			expected: "14:30",
			found:    true,
		},
		{
			name: "Deep target path",
			path: "user.name",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"attributes": map[string]interface{}{
						"name": "John",
					},
				},
			},
			expected: "John",
			found:    true,
		},
		{
			name: "No shortcut for prefix",
			path: "action.type",
			context: map[string]interface{}{
				"action": map[string]interface{}{
					"properties": map[string]interface{}{
						"type": "read",
					},
				},
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

func TestShortcutResolver_DeepTargetPath(t *testing.T) {
	// Test with multi-level target path
	customShortcuts := []ShortcutConfig{
		{Prefix: "user", TargetPath: []string{"data", "extended", "attributes"}},
	}
	resolver := NewShortcutResolver(customShortcuts)

	context := map[string]interface{}{
		"user": map[string]interface{}{
			"data": map[string]interface{}{
				"extended": map[string]interface{}{
					"attributes": map[string]interface{}{
						"department": "Engineering",
					},
				},
			},
		},
	}

	value, found := resolver.Resolve("user.department", context)
	if !found {
		t.Errorf("Expected to find value")
	}
	if value != "Engineering" {
		t.Errorf("Expected 'Engineering', got '%v'", value)
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

func TestArrayAccessResolver(t *testing.T) {
	resolver := NewArrayAccessResolver()

	tests := []struct {
		name     string
		path     string
		context  map[string]interface{}
		expected interface{}
		found    bool
	}{
		{
			name: "Simple array access with brackets",
			path: "user.roles[0]",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"roles": []interface{}{"admin", "user", "developer"},
				},
			},
			expected: "admin",
			found:    true,
		},
		{
			name: "Array access with dot notation",
			path: "user.roles.1",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"roles": []interface{}{"admin", "user", "developer"},
				},
			},
			expected: "user",
			found:    true,
		},
		{
			name: "Array access then property",
			path: "users[0].name",
			context: map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{"name": "Alice", "age": 30},
					map[string]interface{}{"name": "Bob", "age": 25},
				},
			},
			expected: "Alice",
			found:    true,
		},
		{
			name: "Nested array access",
			path: "data.items[1].tags[0]",
			context: map[string]interface{}{
				"data": map[string]interface{}{
					"items": []interface{}{
						map[string]interface{}{"tags": []interface{}{"tag1", "tag2"}},
						map[string]interface{}{"tags": []interface{}{"tag3", "tag4"}},
					},
				},
			},
			expected: "tag3",
			found:    true,
		},
		{
			name: "Array index out of bounds",
			path: "user.roles[10]",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"roles": []interface{}{"admin", "user"},
				},
			},
			expected: nil,
			found:    false,
		},
		{
			name: "Array access on non-array",
			path: "user.name[0]",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
				},
			},
			expected: nil,
			found:    false,
		},
		{
			name: "Path without array access - should not process",
			path: "user.name",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
				},
			},
			expected: nil,
			found:    false,
		},
		{
			name: "Mixed notation",
			path: "data.items[0].tags.1",
			context: map[string]interface{}{
				"data": map[string]interface{}{
					"items": []interface{}{
						map[string]interface{}{"tags": []interface{}{"tag1", "tag2", "tag3"}},
					},
				},
			},
			expected: "tag2",
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

func TestCompositePathResolver_WithArrayAccess(t *testing.T) {
	resolver := NewCompositePathResolver()

	tests := []struct {
		name     string
		path     string
		context  map[string]interface{}
		expected interface{}
		found    bool
	}{
		{
			name: "Array access via composite",
			path: "user.roles[0]",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"roles": []interface{}{"admin", "user"},
				},
			},
			expected: "admin",
			found:    true,
		},
		{
			name: "Regular path still works",
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
			name: "Priority: direct path over array access",
			path: "user.roles[0]",
			context: map[string]interface{}{
				"user.roles[0]": "direct-value",
				"user": map[string]interface{}{
					"roles": []interface{}{"array-value"},
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
