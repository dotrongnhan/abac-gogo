# Core Package

The core package contains the main Policy Decision Point (PDP) implementation and policy validation components.

## Components

### PolicyDecisionPoint (PDP)

The main evaluation engine that implements the ABAC policy evaluation logic.

#### Features:
- **Deny-Override Algorithm**: Implements AWS IAM-style deny-override policy combining
- **Enhanced Context Building**: Automatically enriches evaluation context with time-based and environmental attributes
- **Structured Attributes**: Supports both flat and nested attribute access patterns
- **Performance Optimized**: Includes validation, caching, and configurable limits

#### Usage:

```go
import (
    "abac_go_example/evaluator/core"
    "abac_go_example/storage"
)

// Create PDP instance
pdp := core.NewPolicyDecisionPoint(storageInstance)

// Evaluate access request
request := &models.EvaluationRequest{
    SubjectID:  "user-123",
    ResourceID: "api:documents:doc-456", 
    Action:     "document-service:file:read",
    Context: map[string]interface{}{
        "department": "engineering",
    },
}

decision, err := pdp.Evaluate(request)
```

#### Context Enhancement

The PDP automatically enhances the evaluation context with:

**Time-based Attributes:**
- `environment:time_of_day` - Current time in HH:MM format
- `environment:day_of_week` - Current day of the week
- `environment:hour` - Current hour (0-23)
- `environment:is_weekend` - Boolean indicating weekend
- `environment:is_business_hours` - Boolean for 9 AM - 5 PM, Mon-Fri

**Environmental Attributes:**
- `environment:client_ip` - Client IP address
- `environment:is_internal_ip` - Boolean for internal IP ranges
- `environment:ip_class` - IP version (ipv4/ipv6)
- `environment:user_agent` - User agent string
- `environment:is_mobile` - Mobile device detection
- `environment:browser` - Browser type detection

**Structured Attributes:**
- `user.*` - Flat user attributes for backward compatibility
- `user` - Nested user object with structured access
- `resource.*` - Flat resource attributes
- `resource` - Nested resource object

### PolicyValidator

Validates policy documents against the ABAC schema and business rules.

#### Features:
- **Syntax Validation**: Ensures proper JSON structure and required fields
- **Semantic Validation**: Validates condition operators and value types
- **Business Rules**: Enforces organizational policy constraints
- **Detailed Error Reporting**: Provides specific validation error messages

#### Usage:

```go
validator := core.NewPolicyValidator()

policy := &models.Policy{
    ID:      "test-policy",
    Enabled: true,
    Statement: []models.PolicyStatement{
        {
            Effect: "Allow",
            Action: models.JSONActionResource{Single: "service:resource:action"},
            Resource: models.JSONActionResource{Single: "api:documents:*"},
            Condition: map[string]interface{}{
                "StringEquals": map[string]interface{}{
                    "user.department": "engineering",
                },
            },
        },
    },
}

result := validator.ValidatePolicy(policy)
if !result.IsValid {
    for _, err := range result.Errors {
        fmt.Printf("Validation error: %s\n", err.Message)
    }
}
```

#### Validation Rules

**Basic Policy Structure:**
- Policy ID must be non-empty
- At least one statement required
- Statement effect must be "Allow" or "Deny"

**Action/Resource Validation:**
- Must follow format: `service:type:operation`
- Wildcards supported: `*`, `prefix*`, `*suffix`
- Variables supported: `${variable}`

**Condition Validation:**
- Operator must be valid condition operator type
- Value types must match operator requirements
- Nested conditions properly structured
- Maximum nesting depth enforced

## Implementation Details

### Evaluation Algorithm

The PDP uses a deny-override algorithm:

1. **Policy Retrieval**: Get all enabled policies from storage
2. **Context Enhancement**: Enrich request context with computed attributes
3. **Statement Evaluation**: For each policy statement:
   - Check action matching
   - Check resource matching (including NotResource exclusions)
   - Evaluate conditions
4. **Decision Logic**:
   - If any statement with Effect="Deny" matches → DENY
   - If any statement with Effect="Allow" matches → PERMIT
   - If no statements match → DENY (implicit deny)

### Performance Optimizations

- **Early Termination**: Stop evaluation on first deny match
- **Context Validation**: Validate context structure before evaluation
- **Resource Limits**: Configurable limits on condition complexity
- **Efficient Matching**: Optimized pattern matching algorithms

### Error Handling

The core package provides comprehensive error handling:

- **Input Validation**: Validates all inputs before processing
- **Context Errors**: Reports missing or invalid context attributes
- **Storage Errors**: Handles storage backend failures gracefully
- **Evaluation Errors**: Captures and reports evaluation failures

### Testing

The core package includes extensive tests:

- **Unit Tests**: Test individual methods and components
- **Integration Tests**: End-to-end policy evaluation scenarios
- **Performance Tests**: Benchmarking and load testing
- **Error Cases**: Comprehensive error condition testing

Run core package tests:

```bash
go test ./evaluator/core
go test ./evaluator/core -bench=.
```

## Configuration

The core package respects configuration constants:

```go
// From constants package
const (
    MaxConditionDepth   = 10    // Maximum condition nesting
    MaxConditionKeys    = 100   // Maximum condition keys
    MaxEvaluationTimeMs = 5000  // Maximum evaluation time
)
```

## Security Considerations

- **Deny by Default**: No matching policies results in deny
- **Input Sanitization**: All inputs validated and sanitized
- **DoS Protection**: Limits prevent resource exhaustion attacks
- **Audit Logging**: All evaluation decisions are logged

## Future Enhancements

Planned improvements for the core package:

1. **Policy Caching**: Intelligent caching of frequently used policies
2. **Parallel Evaluation**: Concurrent evaluation of independent statements
3. **Policy Optimization**: Automatic policy conflict detection and optimization
4. **Enhanced Metrics**: Detailed performance and usage metrics
5. **Distributed Evaluation**: Support for distributed policy stores
