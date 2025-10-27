# Policy Enforcement Point (PEP) Package

## 📋 Tổng Quan

Package `pep` cung cấp **Policy Enforcement Point (PEP)** - component đơn giản và hiệu quả để enforce access control policies trong hệ thống ABAC. Package này đã được tối ưu hóa để chỉ giữ lại những components thực sự cần thiết và được sử dụng.

## 🏗️ Kiến Trúc Hiện Tại

### ✅ Active Components

```
pep/
├── simple_pep.go        # Core PEP implementation - MAIN COMPONENT
├── config.go           # Configuration và result types
├── simple_audit.go     # Basic audit logging
└── simple_pep_test.go  # Comprehensive tests
```

### ❌ Removed Components (Không được sử dụng)

Các components sau đã được xóa vì không được sử dụng trong main application:
- `core.go` - Full-featured PEP với advanced features
- `middleware.go` - HTTP middleware (main.go tự implement middleware)
- `interceptor.go` - Method-level interceptors
- `cache.go` - Decision caching
- `rate_limiter.go` - Rate limiting  
- `circuit_breaker.go` - Circuit breaker
- `metrics.go` - Advanced metrics
- `examples.go` - Integration examples
- `middleware_test.go` - Middleware tests

## 🚀 Current Usage

### ✅ SimplePolicyEnforcementPoint

Đây là component chính và duy nhất được sử dụng trong hệ thống hiện tại:

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "abac_go_example/evaluator/core"  // ✅ Sử dụng core package
    "abac_go_example/models"
    "abac_go_example/pep"
    "abac_go_example/storage"
)

func main() {
    // Initialize storage
    config := storage.DefaultDatabaseConfig()
    storage, err := storage.NewPostgreSQLStorage(config)
    if err != nil {
        panic(err)
    }

    // Initialize PDP với core package
    pdp := core.NewPolicyDecisionPoint(storage)

    // Initialize audit logger
    auditLogger, err := pep.NewSimpleAuditLogger("./audit.log")
    if err != nil {
        panic(err)
    }

    // Create simple PEP với default config
    pepInstance := pep.NewSimplePolicyEnforcementPoint(pdp, auditLogger, nil)

    // Use PEP to enforce access
    request := &models.EvaluationRequest{
        RequestID:  "req-001",
        SubjectID:  "user-123",
        ResourceID: "/api/users",
        Action:     "read",
        Context:    map[string]interface{}{
            "timestamp": time.Now().UTC().Format(time.RFC3339),
        },
    }

    result, err := pepInstance.EnforceRequest(context.Background(), request)
    if err != nil {
        panic(err)
    }

    if result.Allowed {
        fmt.Println("Access granted!")
    } else {
        fmt.Printf("Access denied: %s\n", result.Reason)
    }
}
```

## ⚠️ Lưu ý về HTTP Middleware

**HTTP Middleware đã được xóa** vì main application tự implement middleware riêng. Thay vào đó, main.go sử dụng trực tiếp `core.PolicyDecisionPoint` trong Gin middleware:

```go
// Trong main.go - Custom ABAC middleware
func (service *ABACService) ABACMiddleware(requiredAction string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Tạo evaluation request
        request := &models.EvaluationRequest{
            SubjectID:  c.GetHeader("X-Subject-ID"),
            ResourceID: c.Request.URL.Path,
            Action:     requiredAction,
            // ... other fields
        }
        
        // Kiểm tra quyền với PDP trực tiếp
        decision, err := service.pdp.Evaluate(request)
        if err != nil || decision.Result != "permit" {
            c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

## 📊 Configuration

### PEPConfig

```go
type PEPConfig struct {
    FailSafeMode      bool          // Default to DENY on errors
    StrictValidation  bool          // Strict input validation  
    AuditEnabled      bool          // Enable audit logging
    EvaluationTimeout time.Duration // Timeout for policy evaluation
}
```

### EnforcementResult

```go
type EnforcementResult struct {
    Allowed           bool                   // Whether access is allowed
    Decision          string                 // "permit" or "deny"
    Reason            string                 // Reason for decision
    MatchedPolicies   []string               // Policies that matched
    EvaluationTime    time.Duration          // Time taken for evaluation
    EvaluationTimeMs  int                    // Time in milliseconds
    CacheHit          bool                   // Whether result came from cache
    Timestamp         time.Time              // When decision was made
    Metadata          map[string]interface{} // Additional metadata
}
```

## 🧪 Testing

### ✅ Current Test Coverage

```bash
# Run PEP tests
go test ./pep/... -v

# Run with coverage
go test ./pep/... -cover
```

### Test Results
- ✅ `TestSimplePolicyEnforcementPoint_EnforceRequest` - Core functionality
- ✅ `TestSimplePolicyEnforcementPoint_Metrics` - Metrics collection  
- ✅ `TestSimplePolicyEnforcementPoint_Validation` - Input validation
- ✅ `TestSimplePolicyEnforcementPoint_Timeout` - Timeout handling

## 📈 Metrics

SimplePEP cung cấp basic metrics:

```go
type SimplePEPMetrics struct {
    TotalRequests    int64 // Total enforcement requests
    PermitDecisions  int64 // Number of permit decisions
    DenyDecisions    int64 // Number of deny decisions  
    ValidationErrors int64 // Number of validation errors
    EvaluationErrors int64 // Number of evaluation errors
}

// Get current metrics
metrics := pepInstance.GetMetrics()
fmt.Printf("Total requests: %d\n", metrics.TotalRequests)
```

## 🔧 Configuration Examples

### Default Configuration
```go
pepInstance := pep.NewSimplePolicyEnforcementPoint(pdp, auditLogger, nil)
// Uses default config: FailSafe=true, StrictValidation=true, Audit=true
```

### Custom Configuration  
```go
config := &pep.PEPConfig{
    FailSafeMode:      true,                    // Deny on errors
    StrictValidation:  true,                    // Validate all inputs
    AuditEnabled:      true,                    // Log all decisions
    EvaluationTimeout: time.Millisecond * 200, // 200ms timeout
}

pepInstance := pep.NewSimplePolicyEnforcementPoint(pdp, auditLogger, config)
```

## 🎯 Current Status

### ✅ What's Working
- **SimplePolicyEnforcementPoint**: Core PEP functionality
- **Basic Configuration**: Essential settings only
- **Simple Audit Logging**: File-based audit trail
- **Comprehensive Testing**: All core features tested
- **Integration Ready**: Works with evaluator/core

### ❌ What's Been Removed
- Advanced caching, rate limiting, circuit breaker
- HTTP middleware (replaced by custom Gin middleware)
- Method interceptors
- Complex examples and demos
- Advanced metrics and monitoring

### 🚀 Future Considerations

Nếu cần advanced features trong tương lai:
1. **Caching**: Implement decision caching for performance
2. **Rate Limiting**: Add rate limiting for DoS protection  
3. **Circuit Breaker**: Add fault tolerance
4. **Advanced Metrics**: Detailed monitoring and alerting
5. **HTTP Middleware**: Generic middleware for different frameworks

## 📝 Migration Notes

Nếu có code cũ sử dụng các components đã xóa:

### Thay thế HTTP Middleware
```go
// ❌ Cũ - Không còn tồn tại
middleware := pep.NewHTTPMiddleware(pepInstance, nil)

// ✅ Mới - Sử dụng custom middleware trong main.go
func (service *ABACService) ABACMiddleware(action string) gin.HandlerFunc {
    // Custom implementation
}
```

### Thay thế Advanced PEP
```go  
// ❌ Cũ - Không còn tồn tại
pep := pep.NewPolicyEnforcementPoint(pdp, auditLogger, advancedConfig)

// ✅ Mới - Sử dụng SimplePEP
pep := pep.NewSimplePolicyEnforcementPoint(pdp, auditLogger, basicConfig)
```

## 🔒 Security Features

### Fail-Safe Mode
- Default deny on any errors
- Timeout protection
- Input validation

### Audit Logging
- All decisions logged
- Configurable audit logger
- Structured audit data

## 📊 Performance

### Optimizations
- Minimal overhead
- Fast evaluation path
- Configurable timeouts
- Basic metrics collection

### Benchmarks
```bash
# Run performance tests
go test ./pep -bench=.
```

## 🤝 Contributing

Khi contribute vào PEP package:

1. Giữ code đơn giản và focused
2. Thêm tests cho new features
3. Update documentation
4. Ensure backward compatibility
5. Follow Go best practices