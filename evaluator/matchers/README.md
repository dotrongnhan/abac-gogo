# Matchers Package

Package matchers cung cấp specialized pattern matching cho actions và resources trong ABAC policies.

## Components

### ActionMatcher

Xử lý action pattern matching với hỗ trợ wildcards và hierarchical action structures.

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

Xử lý resource pattern matching với hỗ trợ hierarchical resources, wildcards, và variable substitution.

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
1. **Wildcard Check**: Nếu pattern là `*`, return true
2. **Segment Split**: Split cả pattern và action theo `:`
3. **Length Validation**: Đảm bảo same number of segments
4. **Segment Matching**: Match mỗi segment với wildcard support

### Resource Matching
1. **Wildcard Check**: Nếu pattern là `*`, return true
2. **Format Validation**: Validate resource format
3. **Variable Substitution**: Replace `${variable}` với context values
4. **Hierarchical Detection**: Check cho `/` separator
5. **Pattern Matching**: Apply appropriate matching strategy

### Variable Substitution Algorithm
1. **Pattern Detection**: Find `${...}` patterns sử dụng regex
2. **Context Lookup**: Resolve variable từ context
3. **Replacement**: Replace pattern với resolved value
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
- Wildcard patterns được converted thành regex cho efficient matching
- Compiled regex patterns có thể được cached (future enhancement)

### Early Termination
- Quick checks cho exact matches và full wildcards
- Segment-by-segment matching stops trên first mismatch

### Efficient String Operations
- Minimal string manipulation cho common cases
- Direct string comparison khi có thể

## Error Handling

Package matchers xử lý various error conditions:

### Variable Resolution Errors
- Missing variables trong context được ignored (no substitution)
- Invalid variable syntax được preserved as literal text
- Type mismatches được handled gracefully

### Format Validation Errors
- Invalid resource formats return false (no match)
- Malformed patterns được handled safely
- Empty hoặc null inputs được handled appropriately

## Testing

Comprehensive test coverage bao gồm:

### Action Matcher Tests
- Exact matching scenarios
- Tất cả wildcard combinations
- Edge cases và invalid inputs
- Performance benchmarks

### Resource Matcher Tests
- Simple resource matching
- Hierarchical resource matching
- Variable substitution scenarios
- Format validation tests
- Error condition handling

Chạy matcher tests:

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
1. **Be Specific**: Sử dụng most specific pattern có thể
2. **Consistent Naming**: Tuân theo consistent naming conventions
3. **Logical Hierarchy**: Design hierarchical resources logically
4. **Variable Usage**: Sử dụng variables cho user-specific resources

### Performance
1. **Avoid Deep Hierarchies**: Giới hạn nesting depth để better performance
2. **Cache Contexts**: Reuse context objects khi có thể
3. **Pattern Ordering**: Order patterns từ most đến least specific
4. **Minimize Variables**: Sử dụng variables judiciously để avoid overhead

### Security
1. **Validate Patterns**: Luôn validate pattern syntax
2. **Sanitize Variables**: Đảm bảo variable values are safe
3. **Principle of Least Privilege**: Sử dụng specific patterns over wildcards
4. **Audit Wildcards**: Carefully review wildcard usage

## Cải tiến Tương lai

Các cải tiến được lên kế hoạch:

1. **Pattern Caching**: Cache compiled regex patterns để better performance
2. **Advanced Variables**: Hỗ trợ computed variables và functions
3. **Pattern Optimization**: Automatic pattern optimization và conflict detection
4. **Custom Matchers**: Plugin system cho custom matching logic
5. **Performance Metrics**: Detailed matching performance monitoring
