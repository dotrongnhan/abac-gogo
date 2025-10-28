# 🚀 ABAC Go System - Production-Ready Attribute-Based Access Control

A comprehensive **Attribute-Based Access Control (ABAC)** system implemented in Go 1.23+, featuring advanced policy evaluation, PostgreSQL storage, HTTP service integration, and enterprise-grade security controls.

## ✨ Key Features

- **🎯 Advanced Policy Decision Point (PDP)**: Enhanced condition evaluation with 20+ operators
- **🗄️ PostgreSQL Storage**: Production-ready database with GORM ORM and JSONB support
- **🔧 HTTP Service**: RESTful API with Gin framework and ABAC middleware
- **📊 Rich Condition Support**: Time-based, network-based, and complex logical conditions
- **🔍 Comprehensive Audit**: Detailed logging and compliance tracking
- **⚡ High Performance**: Optimized evaluation with caching and pre-filtering
- **🧪 Extensive Testing**: 85%+ test coverage with integration and benchmark tests

## 🏗️ Architecture Overview

### ABAC Components Flow
```
HTTP Request → PEP (Middleware) → PDP (Evaluator) → PIP (Attributes) → PAP (Storage) → Decision
```

### Project Structure
```
abac_go_example/
├── main.go                     # HTTP service entry point
├── cmd/migrate/                # Database migration tools
├── models/                     # Data models with GORM tags
├── evaluator/                  # Policy Decision Point (PDP)
│   ├── core/                   # Main PDP engine and validation
│   ├── conditions/             # Modular condition evaluators (REFACTORED)
│   │   ├── enhanced_condition_evaluator.go  # Main orchestrator
│   │   ├── string_evaluator.go             # String operations
│   │   ├── numeric_evaluator.go            # Numeric operations
│   │   ├── time_evaluator.go               # Time/Date operations
│   │   ├── array_evaluator.go              # Array operations
│   │   ├── network_evaluator.go            # Network operations
│   │   ├── logical_evaluator.go            # AND/OR/NOT logic
│   │   ├── base_evaluator.go               # Common functionality
│   │   └── interfaces.go                   # Type definitions
│   ├── matchers/               # Action/resource pattern matching
│   └── path/                   # Attribute path resolution
├── attributes/                 # Policy Information Point (PIP)
├── storage/                    # Policy Administration Point (PAP)
│   ├── postgresql_storage.go   # PostgreSQL implementation
│   ├── mock_storage.go         # Testing utilities
│   └── interface.go            # Storage abstraction
├── pep/                        # Policy Enforcement Point
├── operators/                  # Comparison operators
├── audit/                      # Audit logging system
├── constants/                  # System constants and enums (ENHANCED)
│   ├── business_rules.go       # Business logic constants
│   ├── condition_operators.go  # Legacy operator constants
│   ├── context_keys.go         # Context key definitions
│   ├── policy_constants.go     # Policy-related constants
│   └── evaluator_constants.go  # NEW: All evaluator constants
└── docs/                       # Documentation
```

## 🚀 Quick Start

### Prerequisites
- **Go 1.23+**
- **PostgreSQL 12+**
- **Docker** (optional, for database)

### Setup & Run
```bash
# Clone repository
git clone <repository-url>
cd ABAC-gogo-example

# Start PostgreSQL (Docker)
docker-compose up -d

# Or create database manually
createdb abac_db

# Install dependencies and migrate
go mod tidy
go run cmd/migrate/main.go

# Start HTTP service
go run main.go
# → Service runs on http://localhost:8081
```

### Using Makefile (Recommended)
```bash
# Full setup from scratch
make setup

# Run application
make run

# Run all tests
make test

# Run benchmarks
make benchmark
```

## 📋 API Endpoints

| Method | Endpoint | Permission | Description |
|--------|----------|------------|-------------|
| `GET` | `/health` | None | Health check (public) |
| `GET` | `/api/v1/users` | `read` | List users |
| `POST` | `/api/v1/users/create` | `write` | Create user |
| `GET` | `/api/v1/financial` | `read` | Financial data |
| `GET` | `/api/v1/admin` | `admin` | Admin panel |
| `GET` | `/debug/routes` | None | Debug: List all routes |

### Authentication
Use header `X-Subject-ID` to identify the user:
```bash
curl -H "X-Subject-ID: sub-001" http://localhost:8081/api/v1/users
```

### Test Users (from migration data)
- **sub-001**: John Doe (Engineering) - Read access to APIs
- **sub-002**: Alice Smith (Finance) - Financial data access  
- **sub-003**: Payment Service (System) - Service account
- **sub-004**: Bob Wilson (Probation) - Limited access

## 🎯 Policy Format & Examples

### Enhanced Policy Structure
```json
{
  "id": "pol-001",
  "policy_name": "Engineering Read Access",
  "version": "2024-10-21",
  "enabled": true,
  "statement": [
    {
      "Sid": "EngineeringReadAccess",
      "Effect": "Allow",
      "Action": "document-service:file:read",
      "Resource": "api:documents:dept-${user:Department}/*",
      "Condition": {
        "And": [
          {
            "StringEquals": {
              "user:department": "engineering"
            }
          },
          {
            "TimeOfDay": {
              "environment:time_of_day": "09:00-17:00"
            }
          },
          {
            "IsBusinessHours": {
              "environment:is_business_hours": true
            }
          }
        ]
      }
    }
  ]
}
```

### Complex Logical Conditions
```json
{
  "Condition": {
    "And": [
      {
        "StringEquals": {
          "user:department": "engineering"
        }
      },
      {
        "Or": [
          {
            "NumericGreaterThan": {
              "user:level": 5
            }
          },
          {
            "ArrayContains": {
              "user:roles": "senior_developer"
            }
          }
        ]
      },
      {
        "Not": {
          "StringEquals": {
            "user:status": "probation"
          }
        }
      }
    ]
  }
}
```

## 🔧 Supported Operators & Conditions

### String Operators
- `StringEquals`, `StringNotEquals`, `StringLike`
- `StringContains`, `StringStartsWith`, `StringEndsWith`
- `StringRegex` (with pattern caching)

### Numeric Operators  
- `NumericEquals`, `NumericNotEquals`
- `NumericGreaterThan`, `NumericGreaterThanEquals`
- `NumericLessThan`, `NumericLessThanEquals`
- `NumericBetween`

### Time-based Operators
- `TimeOfDay` - Time range (e.g., "09:00-17:00")
- `DayOfWeek` - Specific days (e.g., ["Monday", "Friday"])
- `IsBusinessHours` - Business hours detection
- `DateGreaterThan`, `DateLessThan`, `DateBetween`

### Network Operators
- `IPInRange`, `IPNotInRange` - CIDR range matching
- `IsInternalIP` - Internal IP detection

### Array Operators
- `ArrayContains`, `ArrayNotContains`
- `ArraySize` - Array length comparison

### Logical Operators
- `And` - All conditions must be true
- `Or` - At least one condition must be true  
- `Not` - Invert condition result

## 🗄️ Database Configuration

### PostgreSQL Setup
```go
config := &storage.DatabaseConfig{
    Host:         "localhost",
    Port:         5432,
    User:         "postgres", 
    Password:     "password",
    DatabaseName: "abac_db",
    SSLMode:      "disable",
    TimeZone:     "UTC",
}
```

### Environment Variables
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=abac_db
```

### Data Models (GORM)
The system uses JSONB for flexible attribute storage:
```go
type Subject struct {
    ID          string    `gorm:"primaryKey"`
    SubjectType string    `gorm:"index"`
    Attributes  JSONMap   `gorm:"type:jsonb"`
    CreatedAt   time.Time `gorm:"autoCreateTime"`
}

type Policy struct {
    ID               string          `gorm:"primaryKey"`
    PolicyName       string          `gorm:"index"`
    Effect           string          `gorm:"index"`
    Priority         int             `gorm:"index"`
    Enabled          bool            `gorm:"index;default:true"`
    Statement        JSONStatements  `gorm:"type:jsonb"`
}
```

## ⚡ Performance & Benchmarks

### Performance Metrics
- **Mock Storage**: ~4.5µs per evaluation
- **PostgreSQL**: ~2.6ms per evaluation  
- **Throughput**: 5,000+ evaluations/second
- **Memory**: 1,856-3,500 bytes per evaluation

### Optimization Features
- **Regex Caching**: Compiled patterns cached for reuse
- **Path Resolution**: Composite resolver with efficient strategies
- **Context Validation**: Early validation prevents unnecessary processing
- **Configurable Limits**: Protection against DoS attacks

### Performance Limits
```go
const (
    MaxConditionDepth   = 10    // Maximum nesting depth
    MaxConditionKeys    = 100   // Maximum condition keys per policy
    MaxEvaluationTimeMs = 5000  // Maximum evaluation time
)
```

## 🧪 Testing

### Run Tests
```bash
# All tests
make test
# or: go test ./...

# Specific packages
go test ./evaluator/core -v
go test ./storage -v

# Integration tests
make test-integration

# Benchmarks
make benchmark
# or: go test -bench=. -benchmem
```

### Test Coverage
- **Overall**: 85%+ coverage across core packages
- **Core Evaluator**: >90% coverage
- **Storage Layer**: >85% coverage
- **Integration Tests**: End-to-end scenarios

### Key Test Scenarios
- Policy evaluation (permit/deny/not_applicable)
- Complex condition evaluation (And/Or/Not logic)
- Time-based access control
- Network-based restrictions
- Resource pattern matching
- Performance benchmarks

## 🔒 Security Features

### Access Control
- **Fail-safe Defaults**: Deny by default policy
- **Deny-Override Algorithm**: AWS IAM-style policy combining
- **Input Validation**: Comprehensive validation of all inputs
- **DoS Protection**: Configurable limits and timeouts

### Security Scenarios Tested
- Probation user access blocking
- After-hours access restrictions  
- External IP access prevention
- Cross-department data isolation
- Confidential resource protection

## 🚀 Production Deployment

### Docker Deployment
```bash
# Build and run
docker build -t abac-service .
docker-compose up -d
```

### Production Considerations
1. **Authentication**: Replace `X-Subject-ID` with JWT tokens
2. **Database**: Connection pooling, read replicas, proper indexing
3. **Caching**: Redis for policy and decision caching
4. **Monitoring**: Metrics, alerting, distributed tracing
5. **Security**: HTTPS, rate limiting, input sanitization

## 🔧 Integration Examples

### HTTP Middleware Integration
```go
import "abac_go_example/evaluator/core"

type ABACService struct {
    pdp     core.PolicyDecisionPointInterface
    storage storage.Storage
}

func (service *ABACService) ABACMiddleware(requiredAction string) gin.HandlerFunc {
    return func(c *gin.Context) {
        request := &models.EvaluationRequest{
            SubjectID:  c.GetHeader("X-Subject-ID"),
            ResourceID: c.Request.URL.Path,
            Action:     requiredAction,
            Context: map[string]interface{}{
                "method":    c.Request.Method,
                "client_ip": c.ClientIP(),
            },
        }
        
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

### Service Integration
```go
import "abac_go_example/evaluator/core"

type SecureService struct {
    pdp core.PolicyDecisionPointInterface
}

func (s *SecureService) GetUser(ctx context.Context, subjectID, userID string) error {
    request := &models.EvaluationRequest{
        SubjectID:  subjectID,
        ResourceID: fmt.Sprintf("user:%s", userID),
        Action:     "read",
    }
    
    decision, err := s.pdp.Evaluate(request)
    if err != nil || decision.Result != "permit" {
        return fmt.Errorf("access denied: %s", decision.Reason)
    }
    
    // Business logic here
    return nil
}
```

## 📚 Documentation

### Component Documentation
- **[Evaluator](evaluator/README.md)** - Policy Decision Point implementation
- **[Storage](storage/README.md)** - Data access layer and database
- **[PEP](pep/README.md)** - Policy Enforcement Point patterns
- **[Audit](audit/README.md)** - Logging and compliance
- **[Models](models/README.md)** - Data models and types

### Additional Documentation
- **[Action Field Guide](docs/ACTION_FIELD_DOCUMENTATION.md)** - Action pattern documentation
- **[Condition Guide](docs/CONDITION_FIELD_GUIDE.md)** - Condition operator reference
- **[Resource Field Guide](docs/RESOURCE_FIELD_DOCUMENTATION.md)** - Resource pattern documentation

## 🛠️ Development

### Adding New Endpoints
1. Create handler function in `main.go`
2. Register route with ABAC middleware
3. Add test subjects and policies to migration
4. Test with appropriate subject IDs

### Adding New Operators

#### Modern Approach (Recommended)
1. **Add constant** in `constants/evaluator_constants.go`
2. **Choose appropriate evaluator**:
   - String operations → `string_evaluator.go`
   - Numeric operations → `numeric_evaluator.go`
   - Time operations → `time_evaluator.go`
   - Array operations → `array_evaluator.go`
   - Network operations → `network_evaluator.go`
   - Logical operations → `logical_evaluator.go`
3. **Implement method** in chosen evaluator
4. **Register operator** in `enhanced_condition_evaluator.go`
5. **Add comprehensive tests** for the specific evaluator
6. **Update documentation** and interfaces

#### Legacy Approach (Deprecated)
1. ~~Implement operator in monolithic file~~ (No longer recommended)
2. Add operator constant in `constants/condition_operators.go`
3. Add comprehensive tests
4. Update documentation

#### Example: Adding Custom String Operator
```go
// 1. Add constant
const OpStringCustomMatch = "stringcustommatch"

// 2. Implement in string_evaluator.go
func (se *StringConditionEvaluator) EvaluateCustomMatch(conditions interface{}, context map[string]interface{}) bool {
    return se.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
        // Custom logic here
        return true
    })
}

// 3. Register in enhanced_condition_evaluator.go
case constants.OpStringCustomMatch:
    return ece.stringEvaluator.EvaluateCustomMatch(operatorConditions, context)
```

### Policy Management
- Policies stored in PostgreSQL `policies` table
- Support for policy versioning via `version` field
- Enable/disable policies with `enabled` flag
- Priority-based evaluation order

## 📊 Monitoring & Metrics

### Performance Metrics
- Evaluation latency (P50, P95, P99)
- Throughput (requests/second)
- Policy cache hit rates
- Error rates by type

### Security Metrics  
- Deny rate percentage
- Policy violations by user/resource
- Unusual access patterns
- Audit trail completeness

## 🎯 Use Cases & Examples

### Time-based Access Control
```json
{
  "And": [
    {
      "StringEquals": {"user:department": "finance"}
    },
    {
      "TimeOfDay": {"environment:time_of_day": "09:00-17:00"}
    },
    {
      "DayOfWeek": {"environment:day_of_week": ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]}
    }
  ]
}
```

### Resource Owner Access
```json
{
  "Or": [
    {
      "StringEquals": {"resource:owner": "${user:id}"}
    },
    {
      "And": [
        {
          "StringEquals": {"user:department": "${resource:department}"}
        },
        {
          "NumericGreaterThanEquals": {"user:level": 5}
        }
      ]
    }
  ]
}
```

### Network-based Restrictions
```json
{
  "And": [
    {
      "StringEquals": {"user:role": "admin"}
    },
    {
      "IPInRange": {"environment:client_ip": ["10.0.0.0/8", "192.168.1.0/24"]}
    },
    {
      "Not": {
        "StringEquals": {"resource:classification": "top_secret"}
      }
    }
  ]
}
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Add comprehensive tests for new functionality
4. Ensure all tests pass (`make test`)
5. Follow Go best practices and project coding standards
6. Update documentation for any API changes
7. Submit a pull request

### Code Standards
- Follow the repository's `.cursorrules` for coding standards
- Maximum 50 lines per function (Single Responsibility)
- No deep nesting (max 3 levels)
- Meaningful names, no abbreviations
- Comments explain "why", not "what"

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🔗 Related Projects & Standards

- **[XACML 3.0](http://docs.oasis-open.org/xacml/3.0/)** - OASIS XACML standard
- **[NIST ABAC Guide](https://csrc.nist.gov/publications/detail/sp/800-162/final)** - NIST SP 800-162
- **[Open Policy Agent](https://www.openpolicyagent.org/)** - Policy-based control for cloud native
- **[Casbin](https://casbin.org/)** - Authorization library with multiple models
- **[AWS IAM](https://aws.amazon.com/iam/)** - Cloud access management reference

---

## 🎉 Conclusion

This ABAC system provides:

✅ **Production-Ready**: PostgreSQL storage, comprehensive testing, performance optimization  
✅ **Flexible**: Rich policy language with 20+ operators and complex conditions  
✅ **Secure**: Fail-safe defaults, audit logging, comprehensive access control  
✅ **Scalable**: Stateless design, connection pooling, horizontal scaling support  
✅ **Maintainable**: Clean architecture, extensive documentation, 85%+ test coverage  

The system successfully demonstrates enterprise-grade ABAC implementation suitable for production deployment with both development-friendly (mock storage) and production-ready (PostgreSQL) configurations.

**Ready to secure your applications with fine-grained access control!** 🚀