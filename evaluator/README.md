# ABAC Evaluator Package

The evaluator package provides a comprehensive Attribute-Based Access Control (ABAC) policy evaluation system. This package has been refactored to use a clean, modular architecture with the enhanced condition evaluator as the primary evaluation engine.

## Architecture Overview

The evaluator package is organized into several specialized subpackages:

```
evaluator/
├── core/                    # Core PDP and policy validation
├── conditions/              # Condition evaluation engines
├── matchers/               # Action and resource matching
├── path/                   # Path resolution utilities
└── evaluator.go            # Package documentation and usage guide
```

## Package Structure

### Core Package (`evaluator/core`)

Contains the main Policy Decision Point (PDP) and policy validation components:

- **PolicyDecisionPoint**: Main evaluation engine implementing deny-override algorithm
- **PolicyValidator**: Validates policy syntax and structure
- **Integration tests**: Comprehensive end-to-end testing

#### Key Features:
- Deny-override policy combining algorithm
- Enhanced context building with time-based and environmental attributes
- Structured subject and resource attribute handling
- Performance optimizations with configurable limits

### Conditions Package (`evaluator/conditions`)

Advanced condition evaluation with support for complex logical expressions:

- **EnhancedConditionEvaluator**: Primary condition evaluation engine
- **ExpressionEvaluator**: Boolean expression evaluation
- **ComplexCondition**: Legacy condition structure for backward compatibility

#### Supported Operators:

**String Operators:**
- `StringEquals`, `StringNotEquals`, `StringLike`
- `StringContains`, `StringStartsWith`, `StringEndsWith`
- `StringRegex` (with caching for performance)

**Numeric Operators:**
- `NumericEquals`, `NumericNotEquals`
- `NumericLessThan`, `NumericLessThanEquals`
- `NumericGreaterThan`, `NumericGreaterThanEquals`
- `NumericBetween`

**Date/Time Operators:**
- `DateLessThan`, `DateLessThanEquals`
- `DateGreaterThan`, `DateGreaterThanEquals`
- `DateBetween`, `DayOfWeek`, `TimeOfDay`
- `IsBusinessHours`

**Array Operators:**
- `ArrayContains`, `ArrayNotContains`
- `ArraySize` (with comparison operators)

**Network Operators:**
- `IPInRange`, `IPNotInRange`
- `IsInternalIP`

**Logical Operators:**
- `And`, `Or`, `Not`

### Matchers Package (`evaluator/matchers`)

Handles action and resource pattern matching:

- **ActionMatcher**: Matches action patterns with wildcard support
- **ResourceMatcher**: Matches resource patterns with hierarchical support and variable substitution

#### Pattern Formats:
- Actions: `<service>:<resource-type>:<operation>`
- Resources: `<service>:<resource-type>:<resource-id>`
- Hierarchical: `<parent>/<child>` structure
- Variables: `${variable}` substitution from context

### Path Package (`evaluator/path`)

Provides flexible attribute path resolution:

- **CompositePathResolver**: Combines multiple resolution strategies
- **DotNotationResolver**: Handles nested object access (`user.department`)
- **PathNormalizer**: Normalizes and validates attribute paths

## Usage Examples

### Basic Policy Evaluation

```go
import (
    "abac_go_example/evaluator/core"
    "abac_go_example/storage"
)

// Create PDP with storage backend
pdp := core.NewPolicyDecisionPoint(storage)

// Create evaluation request
request := &models.EvaluationRequest{
    SubjectID:  "user-123",
    ResourceID: "api:documents:doc-456",
    Action:     "document-service:file:read",
    Context: map[string]interface{}{
        "department": "engineering",
    },
}

// Evaluate request
decision, err := pdp.Evaluate(request)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Decision: %s, Reason: %s\n", decision.Result, decision.Reason)
```

### Advanced Condition Evaluation

```go
import "abac_go_example/evaluator/conditions"

evaluator := conditions.NewEnhancedConditionEvaluator()

conditions := map[string]interface{}{
    "And": []interface{}{
        map[string]interface{}{
            "StringEquals": map[string]interface{}{
                "user.department": "engineering",
            },
        },
        map[string]interface{}{
            "NumericGreaterThan": map[string]interface{}{
                "user.level": 3,
            },
        },
        map[string]interface{}{
            "IsBusinessHours": map[string]interface{}{
                "environment.current_time": true,
            },
        },
    },
}

context := map[string]interface{}{
    "user": map[string]interface{}{
        "department": "engineering",
        "level":      5,
    },
    "environment": map[string]interface{}{
        "current_time": time.Now(),
    },
}

result := evaluator.EvaluateConditions(conditions, context)
```

### Action and Resource Matching

```go
import "abac_go_example/evaluator/matchers"

// Action matching
actionMatcher := matchers.NewActionMatcher()
matches := actionMatcher.Match("document-service:file:*", "document-service:file:read")

// Resource matching with variables
resourceMatcher := matchers.NewResourceMatcher()
context := map[string]interface{}{
    "request:UserId": "user-123",
}
matches = resourceMatcher.Match("api:documents:owner-${request:UserId}", "api:documents:owner-user-123", context)
```

## Configuration and Constants

The system uses constants defined in the `constants` package:

- **Policy Effects**: `EffectAllow`, `EffectDeny`
- **Decision Results**: `ResultPermit`, `ResultDeny`
- **Context Keys**: Standardized context key prefixes and names
- **Condition Operators**: All supported condition operator types

## Performance Considerations

### Optimizations Implemented:

1. **Regex Caching**: Compiled regex patterns are cached in the enhanced evaluator
2. **Path Resolution**: Composite resolver tries most efficient strategies first
3. **Context Validation**: Early validation prevents unnecessary processing
4. **Configurable Limits**: Maximum condition depth, keys, and evaluation time

### Performance Limits:

```go
const (
    MaxConditionDepth   = 10    // Maximum nesting depth
    MaxConditionKeys    = 100   // Maximum condition keys per policy
    MaxEvaluationTimeMs = 5000  // Maximum evaluation time
)
```

## Testing

Each package includes comprehensive tests:

- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end policy evaluation
- **Performance Tests**: Benchmarking and load testing

Run tests for specific packages:

```bash
# Test all evaluator components
go test ./evaluator/...

# Test specific packages
go test ./evaluator/core
go test ./evaluator/conditions
go test ./evaluator/matchers
go test ./evaluator/path
```

## Migration from Legacy Evaluator

The legacy `ConditionEvaluator` has been completely removed. All condition evaluation now uses the `EnhancedConditionEvaluator`:

### Breaking Changes:
- Removed `NewConditionEvaluator()` - use `conditions.NewEnhancedConditionEvaluator()`
- Removed `evaluateConditionsLegacy()` method
- Updated package structure requires import path changes

### Migration Steps:
1. Update imports to use specific subpackages
2. Replace `NewConditionEvaluator()` with `conditions.NewEnhancedConditionEvaluator()`
3. Update any direct references to internal methods (now properly encapsulated)

## Error Handling

The evaluator provides detailed error information:

- **Validation Errors**: Policy syntax and structure issues
- **Evaluation Errors**: Runtime evaluation problems
- **Context Errors**: Missing or invalid context attributes

## Security Considerations

- **Input Validation**: All inputs are validated before processing
- **DoS Protection**: Configurable limits prevent resource exhaustion
- **Secure Defaults**: Deny-by-default policy combining algorithm
- **Audit Trail**: Comprehensive logging of evaluation decisions

## Future Enhancements

Planned improvements include:

1. **Policy Caching**: Intelligent policy caching for improved performance
2. **Distributed Evaluation**: Support for distributed policy evaluation
3. **Policy Optimization**: Automatic policy optimization and conflict detection
4. **Enhanced Metrics**: Detailed performance and usage metrics
5. **Policy Templates**: Reusable policy templates and inheritance

## Contributing

When contributing to the evaluator package:

1. Follow the established package structure
2. Add comprehensive tests for new features
3. Update documentation for any API changes
4. Ensure backward compatibility where possible
5. Follow Go best practices and the project's coding standards

For detailed implementation examples, see the test files in each subpackage.