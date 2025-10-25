# Conditions Package

The conditions package provides advanced condition evaluation capabilities for ABAC policies, supporting complex logical expressions and a wide range of operators.

## Components

### EnhancedConditionEvaluator

The primary condition evaluation engine that supports complex logical expressions and advanced operators.

#### Features:
- **Comprehensive Operators**: String, numeric, date/time, array, network, and logical operators
- **Regex Caching**: Compiled regex patterns are cached for performance
- **Path Resolution**: Advanced attribute path resolution with dot notation and array access
- **Type Coercion**: Intelligent type conversion for flexible value matching

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

### Supported Operators

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

The conditions package uses advanced path resolution for attribute access:

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
- Compiled regex patterns are cached by pattern string
- Significant performance improvement for repeated evaluations
- Cache is maintained per evaluator instance

### Efficient Type Conversion
- Intelligent type coercion reduces evaluation overhead
- Supports string-to-number, string-to-bool, and time parsing
- Handles multiple time formats automatically

### Path Resolution Optimization
- Composite resolver tries most efficient strategies first
- Direct path lookup before dot notation parsing
- Minimal string manipulation for common cases

## Error Handling

The conditions package provides detailed error information:

- **Invalid Operators**: Reports unknown or unsupported operators
- **Type Mismatches**: Identifies type conversion failures
- **Path Resolution**: Reports attribute path resolution failures
- **Regex Errors**: Captures regex compilation errors

## Testing

Comprehensive test coverage includes:

- **Operator Tests**: Individual operator functionality
- **Complex Logic Tests**: Nested logical expressions
- **Performance Tests**: Benchmarking and caching effectiveness
- **Edge Cases**: Null values, type mismatches, malformed input

Run conditions package tests:

```bash
go test ./evaluator/conditions
go test ./evaluator/conditions -bench=.
```

## Best Practices

### Condition Design
1. **Use Specific Operators**: Choose the most specific operator for your use case
2. **Minimize Nesting**: Keep logical expressions as flat as possible
3. **Cache Regex**: Use StringRegex for complex patterns that will be reused
4. **Type Consistency**: Ensure value types match operator expectations

### Performance
1. **Order Conditions**: Place most selective conditions first in AND operations
2. **Use Caching**: Take advantage of regex and path resolution caching
3. **Avoid Deep Nesting**: Limit condition nesting depth for better performance
4. **Batch Evaluations**: Reuse evaluator instances when possible

### Security
1. **Validate Input**: Always validate condition structure before evaluation
2. **Limit Complexity**: Use MaxConditionDepth to prevent DoS attacks
3. **Sanitize Regex**: Validate regex patterns to prevent ReDoS attacks
4. **Type Safety**: Rely on type coercion rather than unsafe casting

## Migration from Legacy

The legacy ConditionEvaluator has been removed. Migration steps:

1. **Replace Constructor**: Use `NewEnhancedConditionEvaluator()` instead of `NewConditionEvaluator()`
2. **Update Method Calls**: Use `EvaluateConditions()` instead of `Evaluate()`
3. **Enhanced Operators**: Take advantage of new operators like StringRegex, NumericBetween
4. **Path Resolution**: Use dot notation for nested attribute access

## Future Enhancements

Planned improvements:

1. **Custom Operators**: Plugin system for custom condition operators
2. **Condition Optimization**: Automatic condition reordering and optimization
3. **Distributed Caching**: Shared regex and path resolution caches
4. **Schema Validation**: Runtime validation of condition structure
5. **Performance Metrics**: Detailed performance monitoring and reporting
