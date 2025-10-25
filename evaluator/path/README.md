# Path Package

Package path cung cấp flexible và efficient attribute path resolution cho ABAC condition evaluation.

## Components

### PathResolver Interface

Base interface cho tất cả path resolution strategies:

```go
type PathResolver interface {
    Resolve(path string, context map[string]interface{}) (interface{}, bool)
}
```

### CompositePathResolver

Main path resolver kết hợp multiple resolution strategies để maximum flexibility và performance.

#### Resolution Strategies (theo thứ tự execution):
1. **DirectPathResolver**: Direct key lookup cho simple paths
2. **DotNotationResolver**: Nested object access sử dụng dot notation
3. **PathNormalizer**: Advanced path normalization và complex expressions

#### Usage:

```go
import "abac_go_example/evaluator/path"

resolver := path.NewCompositePathResolver()

context := map[string]interface{}{
    "user": map[string]interface{}{
        "profile": map[string]interface{}{
            "department": "engineering",
            "level":      5,
        },
        "roles": []interface{}{"developer", "admin"},
    },
    "simple_key": "simple_value",
}

// Direct path resolution
value, found := resolver.Resolve("simple_key", context)
// value: "simple_value", found: true

// Dot notation resolution
value, found = resolver.Resolve("user.profile.department", context)
// value: "engineering", found: true

// Array access (via normalizer)
value, found = resolver.Resolve("user.roles[0]", context)
// value: "developer", found: true
```

### DotNotationResolver

Specialized resolver cho nested object access sử dụng dot notation.

#### Tính năng:
- **Nested Access**: Navigate through nested maps sử dụng dots
- **Type Safety**: Handles type mismatches gracefully
- **Performance Optimized**: Efficient string splitting và navigation

#### Supported Patterns:
```go
// Simple nested access
"user.department"           // context["user"]["department"]
"resource.metadata.size"    // context["resource"]["metadata"]["size"]

// Deep nesting
"org.team.project.settings.enabled"
```

#### Usage:

```go
resolver := path.NewDotNotationResolver()

context := map[string]interface{}{
    "user": map[string]interface{}{
        "profile": map[string]interface{}{
            "name":       "John Doe",
            "department": "engineering",
        },
    },
}

value, found := resolver.Resolve("user.profile.name", context)
// value: "John Doe", found: true

value, found = resolver.Resolve("user.profile.nonexistent", context)
// value: nil, found: false
```

### PathNormalizer

Advanced path processor xử lý complex path expressions và normalization.

#### Tính năng:
- **Array Access**: Hỗ trợ array indexing với `[index]` syntax
- **Path Validation**: Validates path syntax và structure
- **Expression Parsing**: Xử lý complex path expressions
- **Normalization**: Converts various path formats thành standard form

#### Supported Patterns:

**Array Access:**
```go
"users[0]"              // First user
"users[1].name"         // Name of second user
"roles[0]"              // First role
"permissions[2].action" // Action of third permission
```

**Complex Expressions:**
```go
"user.roles[0].permissions[1].resource"
"organization.teams[0].members[2].profile.department"
```

#### Usage:

```go
normalizer := path.NewPathNormalizer()

context := map[string]interface{}{
    "users": []interface{}{
        map[string]interface{}{"name": "Alice", "role": "admin"},
        map[string]interface{}{"name": "Bob", "role": "user"},
    },
    "roles": []interface{}{"admin", "user", "guest"},
}

// Array access
value, found := normalizer.Resolve("users[0].name", context)
// value: "Alice", found: true

value, found = normalizer.Resolve("roles[1]", context)
// value: "user", found: true

// Complex nested array access
value, found = normalizer.Resolve("users[1].role", context)
// value: "user", found: true
```

## Path Resolution Algorithm

### CompositePathResolver Algorithm
1. **Direct Lookup**: Thử direct key access trước (fastest)
2. **Dot Notation**: Nếu path chứa dots, sử dụng dot notation resolver
3. **Complex Expressions**: Sử dụng path normalizer cho array access và complex expressions
4. **Return Result**: Return first successful resolution

### DotNotationResolver Algorithm
1. **Dot Check**: Chỉ process paths chứa dots
2. **Split Path**: Split path theo dots thành segments
3. **Navigate**: Navigate through nested maps segment by segment
4. **Type Validation**: Đảm bảo mỗi intermediate value là map
5. **Return Value**: Return final value và success status

### PathNormalizer Algorithm
1. **Expression Parsing**: Parse complex path expressions
2. **Array Detection**: Identify array access patterns `[index]`
3. **Segment Processing**: Process mỗi path segment individually
4. **Index Resolution**: Resolve array indices thành actual values
5. **Value Extraction**: Extract final value từ resolved path

## Performance Optimizations

### Composite Resolver Optimizations
- **Strategy Ordering**: Most efficient resolvers thử trước
- **Early Termination**: Stop trên first successful resolution
- **Minimal Overhead**: Direct delegation đến specialized resolvers

### Dot Notation Optimizations
- **Efficient Splitting**: Optimized string splitting cho dot notation
- **Type Checking**: Fast type assertions cho map navigation
- **Memory Efficient**: Minimal memory allocation trong resolution

### Path Normalizer Optimizations
- **Regex Caching**: Compiled regex patterns cho array access detection
- **Index Parsing**: Efficient integer parsing cho array indices
- **Path Caching**: Có thể cache normalized paths (future enhancement)

## Error Handling

### Resolution Errors
Package path xử lý various error conditions gracefully:

**Type Mismatches:**
- Non-map values trong dot notation paths return `(nil, false)`
- Non-array values với array access return `(nil, false)`
- Invalid type assertions được handled safely

**Index Errors:**
- Out-of-bounds array access returns `(nil, false)`
- Invalid array indices (non-numeric) return `(nil, false)`
- Negative indices được treated as invalid

**Path Errors:**
- Empty paths return `(nil, false)`
- Malformed expressions return `(nil, false)`
- Missing intermediate values return `(nil, false)`

### Validation

**Path Syntax Validation:**
```go
// Valid paths
"user.department"           ✓
"users[0].name"            ✓
"roles[1]"                 ✓
"org.teams[0].members[1]"  ✓

// Invalid paths
"user."                    ✗ (trailing dot)
"users[]"                  ✗ (empty brackets)
"users[-1]"                ✗ (negative index)
"users[abc]"               ✗ (non-numeric index)
```

## Testing

Comprehensive test coverage bao gồm:

### Unit Tests
- Individual resolver functionality
- Path parsing và validation
- Error condition handling
- Type safety verification

### Integration Tests
- Composite resolver behavior
- Complex nested structures
- Performance benchmarking
- Edge case scenarios

### Performance Tests
- Resolution speed benchmarks
- Memory usage analysis
- Scalability testing
- Cache effectiveness (khi implemented)

Chạy path package tests:

```bash
go test ./evaluator/path
go test ./evaluator/path -bench=.
```

## Usage Examples

### Simple Attribute Access
```go
context := map[string]interface{}{
    "user_id": "123",
    "department": "engineering",
}

resolver := path.NewCompositePathResolver()
value, _ := resolver.Resolve("user_id", context)        // "123"
value, _ = resolver.Resolve("department", context)      // "engineering"
```

### Nested Object Access
```go
context := map[string]interface{}{
    "user": map[string]interface{}{
        "profile": map[string]interface{}{
            "name": "John Doe",
            "email": "john@company.com",
        },
        "settings": map[string]interface{}{
            "theme": "dark",
            "notifications": true,
        },
    },
}

value, _ := resolver.Resolve("user.profile.name", context)           // "John Doe"
value, _ = resolver.Resolve("user.settings.theme", context)          // "dark"
value, _ = resolver.Resolve("user.settings.notifications", context)  // true
```

### Array Access
```go
context := map[string]interface{}{
    "permissions": []interface{}{
        map[string]interface{}{
            "resource": "documents",
            "actions": []interface{}{"read", "write"},
        },
        map[string]interface{}{
            "resource": "users",
            "actions": []interface{}{"read"},
        },
    },
}

value, _ := resolver.Resolve("permissions[0].resource", context)      // "documents"
value, _ = resolver.Resolve("permissions[0].actions[1]", context)     // "write"
value, _ = resolver.Resolve("permissions[1].resource", context)       // "users"
```

### Complex Nested Structures
```go
context := map[string]interface{}{
    "organization": map[string]interface{}{
        "teams": []interface{}{
            map[string]interface{}{
                "name": "Engineering",
                "members": []interface{}{
                    map[string]interface{}{
                        "name": "Alice",
                        "role": "Lead",
                    },
                    map[string]interface{}{
                        "name": "Bob", 
                        "role": "Developer",
                    },
                },
            },
        },
    },
}

// Access team name
value, _ := resolver.Resolve("organization.teams[0].name", context)
// "Engineering"

// Access member role
value, _ = resolver.Resolve("organization.teams[0].members[1].role", context)
// "Developer"
```

## Best Practices

### Path Design
1. **Use Consistent Naming**: Tuân theo consistent attribute naming conventions
2. **Minimize Nesting**: Giữ nesting depth reasonable cho performance
3. **Prefer Dot Notation**: Sử dụng dot notation cho nested object access
4. **Document Structure**: Document expected context structure

### Performance
1. **Cache Resolvers**: Reuse resolver instances khi có thể
2. **Optimize Context**: Structure context cho efficient access patterns
3. **Avoid Deep Nesting**: Giới hạn nesting depth để better performance
4. **Use Direct Access**: Sử dụng simple keys khi có thể

### Error Handling
1. **Check Return Values**: Luôn check `found` boolean return value
2. **Handle Missing Paths**: Gracefully handle missing attributes
3. **Validate Paths**: Validate path syntax trước resolution
4. **Provide Defaults**: Sử dụng default values cho missing attributes

## Cải tiến Tương lai

Các cải tiến được lên kế hoạch:

1. **Path Caching**: Cache resolved paths để better performance
2. **Advanced Expressions**: Hỗ trợ computed expressions và functions
3. **Path Validation**: Enhanced path syntax validation và error reporting
4. **Custom Resolvers**: Plugin system cho custom resolution strategies
5. **Performance Monitoring**: Detailed resolution performance metrics
6. **Path Optimization**: Automatic path optimization và suggestion
