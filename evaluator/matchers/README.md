# Matchers Package

The matchers package provides specialized pattern matching for actions and resources in ABAC policies.

## Components

### ActionMatcher

Handles action pattern matching with support for wildcards and hierarchical action structures.

#### Action Format
Actions follow the format: `<service>:<resource-type>:<operation>`

Examples:
- `document-service:file:read`
- `payment-service:transaction:create`
- `user-service:profile:update`

#### Wildcard Support

**Full Wildcard**
```go
matcher.Match("*", "any:action:here") // true
```

**Service Wildcard**
```go
matcher.Match("*:file:read", "document-service:file:read") // true
matcher.Match("*:file:read", "payment-service:file:read") // true
```

**Resource Type Wildcard**
```go
matcher.Match("document-service:*:read", "document-service:file:read") // true
matcher.Match("document-service:*:read", "document-service:folder:read") // true
```

**Operation Wildcard**
```go
matcher.Match("document-service:file:*", "document-service:file:read") // true
matcher.Match("document-service:file:*", "document-service:file:write") // true
```

#### Usage

```go
import "abac_go_example/evaluator/matchers"

matcher := matchers.NewActionMatcher()

// Exact match
matches := matcher.Match("document-service:file:read", "document-service:file:read")

// Wildcard match
matches = matcher.Match("document-service:file:*", "document-service:file:read")

// Multiple wildcards
matches = matcher.Match("*:*:read", "document-service:file:read")
```

### ResourceMatcher

Handles resource pattern matching with support for hierarchical resources, wildcards, and variable substitution.

#### Resource Format
Resources follow the format: `<service>:<resource-type>:<resource-id>`

Examples:
- `api:documents:doc-123`
- `api:users:user-456`
- `payment:transactions:tx-789`

#### Hierarchical Resources
Resources can be hierarchical using `/` as a separator:
- `api:departments:eng/api:documents:doc-123`
- `api:organizations:org-1/api:teams:team-2/api:projects:proj-3`

#### Variable Substitution
Resources support variable substitution from context:

```go
context := map[string]interface{}{
    "request:UserId": "user-123",
    "user:Department": "engineering",
}

// Pattern with variables
pattern := "api:documents:owner-${request:UserId}"
resource := "api:documents:owner-user-123"

matches := matcher.Match(pattern, resource, context) // true
```

#### Wildcard Support

**Full Wildcard**
```go
matcher.Match("*", "any:resource:id", context) // true
```

**Service Wildcard**
```go
matcher.Match("*:documents:doc-123", "api:documents:doc-123", context) // true
```

**Resource Type Wildcard**
```go
matcher.Match("api:*:doc-123", "api:documents:doc-123", context) // true
```

**Resource ID Wildcard**
```go
matcher.Match("api:documents:*", "api:documents:doc-123", context) // true
```

**Prefix/Suffix Wildcards**
```go
matcher.Match("api:documents:admin-*", "api:documents:admin-123", context) // true
matcher.Match("api:documents:*-temp", "api:documents:doc-temp", context) // true
```

#### Usage

```go
import "abac_go_example/evaluator/matchers"

matcher := matchers.NewResourceMatcher()

context := map[string]interface{}{
    "request:UserId": "user-123",
    "user:Department": "engineering",
}

// Exact match
matches := matcher.Match("api:documents:doc-123", "api:documents:doc-123", context)

// Wildcard match
matches = matcher.Match("api:documents:*", "api:documents:doc-123", context)

// Variable substitution
matches = matcher.Match("api:documents:owner-${request:UserId}", "api:documents:owner-user-123", context)

// Hierarchical match
matches = matcher.Match("api:departments:${user:Department}/api:documents:*", "api:departments:engineering/api:documents:doc-123", context)
```

## Pattern Matching Algorithm

### Action Matching
1. **Wildcard Check**: If pattern is `*`, return true
2. **Segment Split**: Split both pattern and action by `:`
3. **Length Validation**: Ensure same number of segments
4. **Segment Matching**: Match each segment with wildcard support

### Resource Matching
1. **Wildcard Check**: If pattern is `*`, return true
2. **Format Validation**: Validate resource format
3. **Variable Substitution**: Replace `${variable}` with context values
4. **Hierarchical Detection**: Check for `/` separator
5. **Pattern Matching**: Apply appropriate matching strategy

### Variable Substitution Algorithm
1. **Pattern Detection**: Find `${...}` patterns using regex
2. **Context Lookup**: Resolve variable from context
3. **Replacement**: Replace pattern with resolved value
4. **Validation**: Validate final pattern format

## Validation

### Resource Format Validation
Resources must follow the standard format:
- Minimum 3 segments: `service:type:id`
- No empty segments (except wildcards)
- Variables in `${...}` format are allowed

Invalid examples:
- `users:123` (missing service)
- `api:users` (missing resource-id)
- `api::123` (empty resource-type)
- `""` (empty string)

### Pattern Validation
Patterns are validated for:
- Proper segment structure
- Valid wildcard usage
- Correct variable syntax
- Hierarchical format compliance

## Performance Optimizations

### Regex Compilation
- Wildcard patterns are converted to regex for efficient matching
- Compiled regex patterns could be cached (future enhancement)

### Early Termination
- Quick checks for exact matches and full wildcards
- Segment-by-segment matching stops on first mismatch

### Efficient String Operations
- Minimal string manipulation for common cases
- Direct string comparison when possible

## Error Handling

The matchers package handles various error conditions:

### Variable Resolution Errors
- Missing variables in context are ignored (no substitution)
- Invalid variable syntax is preserved as literal text
- Type mismatches are handled gracefully

### Format Validation Errors
- Invalid resource formats return false (no match)
- Malformed patterns are handled safely
- Empty or null inputs are handled appropriately

## Testing

Comprehensive test coverage includes:

### Action Matcher Tests
- Exact matching scenarios
- All wildcard combinations
- Edge cases and invalid inputs
- Performance benchmarks

### Resource Matcher Tests
- Simple resource matching
- Hierarchical resource matching
- Variable substitution scenarios
- Format validation tests
- Error condition handling

Run matcher tests:

```bash
go test ./evaluator/matchers
go test ./evaluator/matchers -bench=.
```

## Usage Patterns

### Common Action Patterns
```go
// Service-specific actions
"document-service:*:*"          // All document service actions
"*:file:read"                   // Read files in any service
"*:*:read"                      // All read operations

// Administrative actions
"admin-service:*:*"             // All admin operations
"*:user:create"                 // Create users in any service
"security-service:audit:*"      // All audit operations
```

### Common Resource Patterns
```go
// User-owned resources
"api:documents:owner-${request:UserId}"
"api:projects:creator-${request:UserId}"

// Department-scoped resources
"api:departments:${user:Department}/*"
"api:resources:dept-${user:Department}-*"

// Hierarchical resources
"api:organizations:${user:OrgId}/api:teams:${user:TeamId}/*"
"api:projects:${project:Id}/api:documents:*"
```

### Security Patterns
```go
// Sensitive resource protection
"api:secrets:*"                 // All secrets (typically denied)
"api:admin:*"                   // Admin resources
"api:audit:*"                   // Audit logs

// Environment-based access
"api:prod:*"                    // Production resources
"api:dev:*"                     // Development resources
"api:test:*"                    // Test resources
```

## Best Practices

### Pattern Design
1. **Be Specific**: Use the most specific pattern possible
2. **Consistent Naming**: Follow consistent naming conventions
3. **Logical Hierarchy**: Design hierarchical resources logically
4. **Variable Usage**: Use variables for user-specific resources

### Performance
1. **Avoid Deep Hierarchies**: Limit nesting depth for better performance
2. **Cache Contexts**: Reuse context objects when possible
3. **Pattern Ordering**: Order patterns from most to least specific
4. **Minimize Variables**: Use variables judiciously to avoid overhead

### Security
1. **Validate Patterns**: Always validate pattern syntax
2. **Sanitize Variables**: Ensure variable values are safe
3. **Principle of Least Privilege**: Use specific patterns over wildcards
4. **Audit Wildcards**: Carefully review wildcard usage

## Future Enhancements

Planned improvements:

1. **Pattern Caching**: Cache compiled regex patterns for better performance
2. **Advanced Variables**: Support for computed variables and functions
3. **Pattern Optimization**: Automatic pattern optimization and conflict detection
4. **Custom Matchers**: Plugin system for custom matching logic
5. **Performance Metrics**: Detailed matching performance monitoring
