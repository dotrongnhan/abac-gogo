# Conditions Package

Package conditions cung cấp khả năng đánh giá condition nâng cao cho ABAC policies, hỗ trợ complex logical expressions và nhiều loại operators.

## Components

### EnhancedConditionEvaluator

Engine đánh giá condition chính hỗ trợ complex logical expressions và advanced operators.

#### Tính năng:
- **Comprehensive Operators**: String, numeric, date/time, array, network, và logical operators
- **Regex Caching**: Compiled regex patterns được cache để tối ưu performance
- **Path Resolution**: Advanced attribute path resolution với dot notation và array access
- **Type Coercion**: Intelligent type conversion để flexible value matching

#### Usage:

```go
import "abac_go_example/evaluator/conditions"

evaluator := conditions.NewEnhancedConditionEvaluator()

conditions := map[string]interface{}{
    "StringEquals": map[string]interface{}{
        "user.department": "engineering",
    },
    "NumericGreaterThan": map[string]interface{}{
        "user.level": 3,
    },
}

context := map[string]interface{}{
    "user": map[string]interface{}{
        "department": "engineering",
        "level":      5,
    },
}

result := evaluator.EvaluateConditions(conditions, context)
```

### Các Operators được hỗ trợ

#### String Operators

**StringEquals / StringNotEquals**
```json
{
    "StringEquals": {
        "user.department": "engineering"
    }
}
```

**StringLike** - SQL-style pattern matching
```json
{
    "StringLike": {
        "user.email": "%@company.com"
    }
}
```

**StringContains / StringStartsWith / StringEndsWith**
```json
{
    "StringContains": {
        "resource.tags": "confidential"
    }
}
```

**StringRegex** - Regular expression matching with caching
```json
{
    "StringRegex": {
        "user.email": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    }
}
```

#### Numeric Operators

**Basic Comparisons**
```json
{
    "NumericGreaterThan": {
        "user.level": 3
    },
    "NumericLessThanEquals": {
        "transaction.amount": 10000
    }
}
```

**NumericBetween** - Range checking
```json
{
    "NumericBetween": {
        "user.age": [18, 65]
    }
}
```

#### Date/Time Operators

**Basic Date Comparisons**
```json
{
    "DateGreaterThan": {
        "user.created_at": "2023-01-01T00:00:00Z"
    }
}
```

**TimeBetween** - Time range checking
```json
{
    "TimeBetween": {
        "request.timestamp": ["09:00", "17:00"]
    }
}
```

**DayOfWeek** - Day-based restrictions
```json
{
    "DayOfWeek": {
        "environment.day_of_week": ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]
    }
}
```

**IsBusinessHours** - Business hours checking
```json
{
    "IsBusinessHours": {
        "environment.current_time": true
    }
}
```

#### Array Operators

**ArrayContains / ArrayNotContains**
```json
{
    "ArrayContains": {
        "user.roles": "admin"
    }
}
```

**ArraySize** - Array length checking
```json
{
    "ArraySize": {
        "user.permissions": {
            "GreaterThan": 5
        }
    }
}
```

#### Network Operators

**IPInRange / IPNotInRange** - CIDR-based IP matching
```json
{
    "IPInRange": {
        "environment.client_ip": ["192.168.1.0/24", "10.0.0.0/8"]
    }
}
```

**IsInternalIP** - Internal IP detection
```json
{
    "IsInternalIP": {
        "environment.client_ip": true
    }
}
```

#### Logical Operators

**And** - All conditions must be true
```json
{
    "And": [
        {
            "StringEquals": {
                "user.department": "engineering"
            }
        },
        {
            "NumericGreaterThan": {
                "user.level": 3
            }
        }
    ]
}
```

**Or** - At least one condition must be true
```json
{
    "Or": [
        {
            "StringEquals": {
                "user.role": "admin"
            }
        },
        {
            "StringEquals": {
                "user.department": "security"
            }
        }
    ]
}
```

**Not** - Negates the condition
```json
{
    "Not": {
        "StringEquals": {
            "user.status": "suspended"
        }
    }
}
```

### ExpressionEvaluator

Provides boolean expression evaluation with custom operators.

#### Features:
- **Custom Operators**: Register custom evaluation functions
- **Nested Expressions**: Support for complex nested boolean expressions
- **Type Safety**: Strong type checking and validation

#### Usage:

```go
evaluator := conditions.NewExpressionEvaluator()

// Register custom operator
evaluator.RegisterOperator("custom_check", func(left, right interface{}) bool {
    // Custom logic here
    return true
})

expression := &models.BooleanExpression{
    Type: "compound",
    Operator: "and",
    Left: &models.BooleanExpression{
        Type: "simple",
        Condition: &models.SimpleCondition{
            AttributePath: "user.department",
            Operator: "eq",
            Value: "engineering",
        },
    },
    Right: &models.BooleanExpression{
        Type: "simple", 
        Condition: &models.SimpleCondition{
            AttributePath: "user.level",
            Operator: "gt",
            Value: 3,
        },
    },
}

result := evaluator.EvaluateExpression(expression, attributes)
```

### ComplexCondition

Legacy condition structure maintained for backward compatibility.

```go
type ComplexCondition struct {
    Type       string                `json:"type"`
    Operator   string                `json:"operator,omitempty"`
    Key        string                `json:"key,omitempty"`
    Value      interface{}           `json:"value,omitempty"`
    Left       *ComplexCondition     `json:"left,omitempty"`
    Right      *ComplexCondition     `json:"right,omitempty"`
    Operand    *ComplexCondition     `json:"operand,omitempty"`
    Conditions []ComplexCondition    `json:"conditions,omitempty"`
}
```

## Path Resolution

Package conditions sử dụng advanced path resolution để attribute access:

### Dot Notation
```json
{
    "StringEquals": {
        "user.profile.department": "engineering"
    }
}
```

### Array Access
```json
{
    "ArrayContains": {
        "user.roles[0]": "admin"
    }
}
```

### Nested Objects
```json
{
    "NumericGreaterThan": {
        "resource.metadata.size": 1000000
    }
}
```

## Performance Optimizations

### Regex Caching
- Compiled regex patterns được cache theo pattern string
- Significant performance improvement cho repeated evaluations
- Cache được maintain per evaluator instance

### Efficient Type Conversion
- Intelligent type coercion giảm evaluation overhead
- Hỗ trợ string-to-number, string-to-bool, và time parsing
- Xử lý multiple time formats automatically

### Path Resolution Optimization
- Composite resolver thử efficient strategies trước
- Direct path lookup trước dot notation parsing
- Minimal string manipulation cho common cases

## Error Handling

Package conditions cung cấp thông tin error chi tiết:

- **Invalid Operators**: Báo cáo unknown hoặc unsupported operators
- **Type Mismatches**: Xác định type conversion failures
- **Path Resolution**: Báo cáo attribute path resolution failures
- **Regex Errors**: Capture regex compilation errors

## Testing

Comprehensive test coverage bao gồm:

- **Operator Tests**: Individual operator functionality
- **Complex Logic Tests**: Nested logical expressions
- **Performance Tests**: Benchmarking và caching effectiveness
- **Edge Cases**: Null values, type mismatches, malformed input

Chạy conditions package tests:

```bash
go test ./evaluator/conditions
go test ./evaluator/conditions -bench=.
```

## Best Practices

### Condition Design
1. **Use Specific Operators**: Chọn operator cụ thể nhất cho use case của bạn
2. **Minimize Nesting**: Giữ logical expressions càng flat càng tốt
3. **Cache Regex**: Sử dụng StringRegex cho complex patterns sẽ được reused
4. **Type Consistency**: Đảm bảo value types match operator expectations

### Performance
1. **Order Conditions**: Đặt most selective conditions trước trong AND operations
2. **Use Caching**: Tận dụng regex và path resolution caching
3. **Avoid Deep Nesting**: Giới hạn condition nesting depth để better performance
4. **Batch Evaluations**: Reuse evaluator instances khi có thể

### Security
1. **Validate Input**: Luôn validate condition structure trước evaluation
2. **Limit Complexity**: Sử dụng MaxConditionDepth để ngăn DoS attacks
3. **Sanitize Regex**: Validate regex patterns để ngăn ReDoS attacks
4. **Type Safety**: Dựa vào type coercion thay vì unsafe casting

## Migration từ Legacy

Legacy ConditionEvaluator đã được removed. Các bước migration:

1. **Replace Constructor**: Sử dụng `NewEnhancedConditionEvaluator()` thay vì `NewConditionEvaluator()`
2. **Update Method Calls**: Sử dụng `EvaluateConditions()` thay vì `Evaluate()`
3. **Enhanced Operators**: Tận dụng new operators như StringRegex, NumericBetween
4. **Path Resolution**: Sử dụng dot notation cho nested attribute access

## Cải tiến Tương lai

Các cải tiến được lên kế hoạch:

1. **Custom Operators**: Plugin system cho custom condition operators
2. **Condition Optimization**: Automatic condition reordering và optimization
3. **Distributed Caching**: Shared regex và path resolution caches
4. **Schema Validation**: Runtime validation của condition structure
5. **Performance Metrics**: Detailed performance monitoring và reporting
