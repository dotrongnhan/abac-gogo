# Core Package

Package core chứa main Policy Decision Point (PDP) implementation và policy validation components.

## Components

### PolicyDecisionPoint (PDP)

Engine đánh giá chính thực hiện ABAC policy evaluation logic.

#### Tính năng:
- **Deny-Override Algorithm**: Thực hiện AWS IAM-style deny-override policy combining
- **Enhanced Context Building**: Tự động enriches evaluation context với time-based và environmental attributes
- **Structured Attributes**: Hỗ trợ cả flat và nested attribute access patterns
- **Performance Optimized**: Bao gồm validation, caching, và configurable limits

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

PDP tự động enhances evaluation context với:

**Time-based Attributes:**
- `environment:time_of_day` - Current time trong HH:MM format
- `environment:day_of_week` - Current day of the week
- `environment:hour` - Current hour (0-23)
- `environment:is_weekend` - Boolean indicating weekend
- `environment:is_business_hours` - Boolean cho 9 AM - 5 PM, Mon-Fri

**Environmental Attributes:**
- `environment:client_ip` - Client IP address
- `environment:is_internal_ip` - Boolean cho internal IP ranges
- `environment:ip_class` - IP version (ipv4/ipv6)
- `environment:user_agent` - User agent string
- `environment:is_mobile` - Mobile device detection
- `environment:browser` - Browser type detection

**Structured Attributes:**
- `user.*` - Flat user attributes cho backward compatibility
- `user` - Nested user object với structured access
- `resource.*` - Flat resource attributes
- `resource` - Nested resource object

### PolicyValidator

Validates policy documents against ABAC schema và business rules.

#### Tính năng:
- **Syntax Validation**: Đảm bảo proper JSON structure và required fields
- **Semantic Validation**: Validates condition operators và value types
- **Business Rules**: Enforces organizational policy constraints
- **Detailed Error Reporting**: Cung cấp specific validation error messages

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

## Chi tiết Implementation

### Evaluation Algorithm

PDP sử dụng deny-override algorithm:

1. **Policy Retrieval**: Get all enabled policies từ storage
2. **Context Enhancement**: Enrich request context với computed attributes
3. **Statement Evaluation**: Cho mỗi policy statement:
   - Check action matching
   - Check resource matching (bao gồm NotResource exclusions)
   - Evaluate conditions
4. **Decision Logic**:
   - Nếu bất kỳ statement nào với Effect="Deny" matches → DENY
   - Nếu bất kỳ statement nào với Effect="Allow" matches → PERMIT
   - Nếu không có statements match → DENY (implicit deny)

### Performance Optimizations

- **Early Termination**: Stop evaluation trên first deny match
- **Context Validation**: Validate context structure trước evaluation
- **Resource Limits**: Configurable limits trên condition complexity
- **Efficient Matching**: Optimized pattern matching algorithms

### Error Handling

Package core cung cấp comprehensive error handling:

- **Input Validation**: Validates tất cả inputs trước processing
- **Context Errors**: Reports missing hoặc invalid context attributes
- **Storage Errors**: Handles storage backend failures gracefully
- **Evaluation Errors**: Captures và reports evaluation failures

### Testing

Package core bao gồm extensive tests:

- **Unit Tests**: Test individual methods và components
- **Integration Tests**: End-to-end policy evaluation scenarios
- **Performance Tests**: Benchmarking và load testing
- **Error Cases**: Comprehensive error condition testing

Chạy core package tests:

```bash
go test ./evaluator/core
go test ./evaluator/core -bench=.
```

## Configuration

Package core tuân theo configuration constants:

```go
// Từ constants package
const (
    MaxConditionDepth   = 10    // Maximum condition nesting
    MaxConditionKeys    = 100   // Maximum condition keys
    MaxEvaluationTimeMs = 5000  // Maximum evaluation time
)
```

## Cân nhắc Security

- **Deny by Default**: Không có matching policies results in deny
- **Input Sanitization**: Tất cả inputs validated và sanitized
- **DoS Protection**: Limits ngăn chặn resource exhaustion attacks
- **Audit Logging**: Tất cả evaluation decisions được logged

## Cải tiến Tương lai

Các cải tiến được lên kế hoạch cho core package:

1. **Policy Caching**: Intelligent caching của frequently used policies
2. **Parallel Evaluation**: Concurrent evaluation của independent statements
3. **Policy Optimization**: Automatic policy conflict detection và optimization
4. **Enhanced Metrics**: Detailed performance và usage metrics
5. **Distributed Evaluation**: Hỗ trợ distributed policy stores
