# ABAC Evaluator Package

Package evaluator cung cấp hệ thống đánh giá policy ABAC (Attribute-Based Access Control) toàn diện. Package này sử dụng kiến trúc modular với `evaluator/core` làm engine đánh giá chính và enhanced condition evaluator để xử lý complex logical expressions.

**⚠️ Lưu ý quan trọng**: Package này chỉ là documentation wrapper. Tất cả implementations thực tế nằm trong các subpackages. Sử dụng trực tiếp các subpackages để có type safety và clarity tốt hơn.

## Tổng quan Kiến trúc

Package evaluator được tổ chức thành nhiều subpackage chuyên biệt:

```
evaluator/
├── core/                    # Core PDP và policy validation
├── conditions/              # Condition evaluation engines
├── matchers/               # Action và resource matching
├── path/                   # Path resolution utilities
└── evaluator.go            # Tài liệu package và hướng dẫn sử dụng
```

## Cấu trúc Package

### Core Package (`evaluator/core`) - **ENGINE CHÍNH**

Đây là package chính chứa Policy Decision Point (PDP) và tất cả logic đánh giá policy:

- **PolicyDecisionPoint**: Engine đánh giá chính thực hiện AWS IAM-style deny-override algorithm
- **PolicyDecisionPointInterface**: Interface chuẩn cho policy evaluation
- **Enhanced Context Building**: Tự động enriches evaluation context với time-based và environmental attributes
- **Integration tests**: Testing toàn diện end-to-end với mock storage

#### Tính năng chính:
- **Deny-Override Algorithm**: Nếu có bất kỳ policy nào deny, kết quả là deny ngay lập tức
- **Enhanced Context Building**: Tự động thêm time-based attributes (business hours, weekend, etc.)
- **Environmental Context**: IP detection, user agent parsing, location attributes
- **Structured Attributes**: Hỗ trợ cả flat (`user:department`) và nested (`user.department`) access
- **Performance Optimized**: Input validation, configurable limits, evaluation timeouts

### Conditions Package (`evaluator/conditions`)

Đánh giá condition nâng cao với hỗ trợ complex logical expressions:

- **EnhancedConditionEvaluator**: Engine đánh giá condition chính
- **ExpressionEvaluator**: Đánh giá boolean expression
- **ComplexCondition**: Cấu trúc condition cũ để backward compatibility

#### Các Operator được hỗ trợ:

**String Operators:**
- `StringEquals`, `StringNotEquals`, `StringLike`
- `StringContains`, `StringStartsWith`, `StringEndsWith`
- `StringRegex` (có caching để tối ưu performance)

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
- `ArraySize` (với comparison operators)

**Network Operators:**
- `IPInRange`, `IPNotInRange`
- `IsInternalIP`

**Logical Operators:**
- `And`, `Or`, `Not`

### Matchers Package (`evaluator/matchers`)

Xử lý action và resource pattern matching:

- **ActionMatcher**: Match action patterns với wildcard support
- **ResourceMatcher**: Match resource patterns với hierarchical support và variable substitution

#### Định dạng Pattern:
- Actions: `<service>:<resource-type>:<operation>`
- Resources: `<service>:<resource-type>:<resource-id>`
- Hierarchical: cấu trúc `<parent>/<child>`
- Variables: thay thế `${variable}` từ context

### Path Package (`evaluator/path`)

Cung cấp flexible attribute path resolution:

- **CompositePathResolver**: Kết hợp nhiều resolution strategies
- **DotNotationResolver**: Xử lý nested object access (`user.department`)
- **PathNormalizer**: Normalize và validate attribute paths

## ✅ Ví dụ Sử dụng Hiện tại

### 1. Đánh giá Policy Cơ bản (Main Application Pattern)

```go
import (
    "abac_go_example/evaluator/core"  // ✅ Sử dụng core package
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

### 2. PEP Integration Pattern

```go
import "abac_go_example/evaluator/core"

// Trong PEP service
type ABACService struct {
    pdp     core.PolicyDecisionPointInterface  // ✅ Sử dụng core interface
    storage storage.Storage
}

// PEP middleware sử dụng core PDP
func (service *ABACService) ABACMiddleware(requiredAction string) gin.HandlerFunc {
    return func(c *gin.Context) {
        request := &models.EvaluationRequest{
            SubjectID:  c.GetHeader("X-Subject-ID"),
            ResourceID: c.Request.URL.Path,
            Action:     requiredAction,
        }
        
        decision, err := service.pdp.Evaluate(request)
        // ... handle decision
    }
}
```

### 3. Đánh giá Condition Nâng cao (Enhanced Evaluator Pattern)

```go
import "abac_go_example/evaluator/conditions"

// ✅ Sử dụng enhanced condition evaluator
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

// ✅ Sử dụng EvaluateConditions method
result := evaluator.EvaluateConditions(conditions, context)
```

### Action và Resource Matching

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

## Configuration và Constants

Hệ thống sử dụng constants được định nghĩa trong package `constants`:

- **Policy Effects**: `EffectAllow`, `EffectDeny`
- **Decision Results**: `ResultPermit`, `ResultDeny`
- **Context Keys**: Standardized context key prefixes và names
- **Condition Operators**: Tất cả supported condition operator types

## Cân nhắc Performance

### Optimizations đã triển khai:

1. **Regex Caching**: Compiled regex patterns được cache trong enhanced evaluator
2. **Path Resolution**: Composite resolver thử efficient strategies trước
3. **Context Validation**: Early validation ngăn chặn unnecessary processing
4. **Configurable Limits**: Maximum condition depth, keys, và evaluation time

### Performance Limits:

```go
const (
    MaxConditionDepth   = 10    // Maximum nesting depth
    MaxConditionKeys    = 100   // Maximum condition keys per policy
    MaxEvaluationTimeMs = 5000  // Maximum evaluation time
)
```

## Testing

Mỗi package bao gồm comprehensive tests:

- **Unit Tests**: Testing từng component riêng lẻ
- **Integration Tests**: End-to-end policy evaluation
- **Performance Tests**: Benchmarking và load testing

Chạy tests cho specific packages:

```bash
# Test all evaluator components
go test ./evaluator/...

# Test specific packages
go test ./evaluator/core
go test ./evaluator/conditions
go test ./evaluator/matchers
go test ./evaluator/path
```

## ✅ Migration Hoàn thành - Sử dụng `evaluator/core`

**Tất cả code đã được migrate sang sử dụng `evaluator/core` package.**

### ✅ Những gì đã thay đổi:
- **Đã xóa**: `evaluator/pdp.go` (file duplicate)
- **Sử dụng**: `evaluator/core` làm package chính
- **Cập nhật**: Tất cả imports trong main.go, examples, PEP, và tests

### ✅ Import Pattern mới:
```go
// ✅ ĐÚNG - Sử dụng core package
import "abac_go_example/evaluator/core"
pdp := core.NewPolicyDecisionPoint(storage)

// ❌ SAI - Không còn tồn tại
import "abac_go_example/evaluator"
pdp := evaluator.NewPolicyDecisionPoint(storage)
```

### ✅ Files đã được cập nhật:
- `main.go` - Main application
- `examples/improved_pdp/main.go` - Example application  
- `examples/complex_conditions_demo.go` - Condition demo (cập nhật để sử dụng enhanced evaluator)
- `pep/core.go`, `pep/simple_pep.go`, `pep/examples.go` - Tất cả PEP components
- `pep/simple_pep_test.go`, `pep/middleware_test.go` - Tất cả tests
- `benchmark_test.go` - Benchmark tests

### ✅ Verification:
```bash
# Tất cả builds thành công
go build -o test_main ./main.go                    ✅
go build -o test_example ./examples/improved_pdp/  ✅  
go build -o test_demo ./examples/complex_conditions_demo.go ✅

# Tất cả tests pass
go test ./pep/... -v                              ✅
go test ./evaluator/core -v                       ✅

# Không có linter errors
golangci-lint run                                  ✅
```

## Error Handling

Evaluator cung cấp thông tin error chi tiết:

- **Validation Errors**: Policy syntax và structure issues
- **Evaluation Errors**: Runtime evaluation problems
- **Context Errors**: Missing hoặc invalid context attributes

## Cân nhắc Security

- **Input Validation**: Tất cả inputs được validate trước khi processing
- **DoS Protection**: Configurable limits ngăn chặn resource exhaustion
- **Secure Defaults**: Deny-by-default policy combining algorithm
- **Audit Trail**: Comprehensive logging của evaluation decisions

## ✅ Tình trạng hiện tại

### ✅ Đã hoàn thành:
- **Clean Architecture**: Package structure rõ ràng với separation of concerns
- **Enhanced Condition Evaluation**: Hỗ trợ đầy đủ logical operators và data types
- **Performance Optimization**: Regex caching, input validation, configurable limits
- **Comprehensive Testing**: Unit tests, integration tests, benchmarks
- **Migration Complete**: Tất cả code sử dụng `evaluator/core` consistently

### 🚀 Cải tiến Tương lai:
1. **Policy Caching**: Intelligent policy caching để cải thiện performance
2. **Distributed Evaluation**: Hỗ trợ distributed policy evaluation
3. **Policy Optimization**: Automatic policy optimization và conflict detection
4. **Enhanced Metrics**: Detailed performance và usage metrics
5. **Policy Templates**: Reusable policy templates và inheritance

## Contributing

Khi contribute vào evaluator package:

1. Tuân theo established package structure
2. Thêm comprehensive tests cho new features
3. Cập nhật documentation cho bất kỳ API changes nào
4. Đảm bảo backward compatibility khi có thể
5. Tuân theo Go best practices và project's coding standards

Để xem detailed implementation examples, hãy xem test files trong mỗi subpackage.