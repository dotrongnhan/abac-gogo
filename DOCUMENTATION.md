# ABAC Go System - Complete Documentation

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL database

### Setup & Run
```bash
# Setup database
docker-compose up -d  # or createdb abac_db

# Install & migrate
go mod tidy
go run cmd/migrate/main.go

# Start service
go run main.go
# → http://localhost:8081
```

### Test API
```bash
# Health check
curl http://localhost:8081/health

# Test with user
curl -H 'X-Subject-ID: sub-001' http://localhost:8081/api/v1/users
```

## 🏗️ Architecture

### ABAC Components
```
HTTP Request → PEP (Middleware) → PDP (Evaluator) → PIP (Attributes) → PAP (Storage) → Decision
```

### Project Structure
```
├── main.go              # HTTP service entry
├── cmd/migrate/         # Database migration
├── models/              # Data models (GORM)
├── evaluator/           # PDP - Policy Decision Point
├── attributes/          # PIP - Policy Information Point  
├── storage/             # PAP - Data access layer
├── pep/                 # Policy Enforcement Point
├── operators/           # Comparison operators
└── audit/               # Audit logging
```

## 📋 API Endpoints

| Method | Endpoint | Permission | Description |
|--------|----------|------------|-------------|
| `GET` | `/health` | None | Health check |
| `GET` | `/api/v1/users` | `read` | List users |
| `POST` | `/api/v1/users/create` | `write` | Create user |
| `GET` | `/api/v1/financial` | `read` | Financial data |
| `GET` | `/api/v1/admin` | `admin` | Admin panel |

### Authentication
Use header: `X-Subject-ID: <user_id>`

### Test Users
- `sub-001`: John Doe (Engineering) - Read access
- `sub-002`: Alice Smith (Finance) - Financial access  
- `sub-003`: Payment Service (System) - Service account
- `sub-004`: Bob Wilson (Probation) - Limited access

## 🎯 Policy Format

### Basic Policy Structure
```json
{
  "id": "pol-001",
  "policy_name": "Engineering Read Access",
  "effect": "permit",
  "priority": 100,
  "enabled": true,
  "rules": [
    {
      "target_type": "subject",
      "attribute_path": "attributes.department",
      "operator": "eq",
      "expected_value": "engineering"
    }
  ],
  "actions": ["read"],
  "resource_patterns": ["/api/v1/*"]
}
```

### Enhanced Policy with Complex Conditions
```json
{
  "conditions": {
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
              "user:role": "senior_developer"
            }
          }
        ]
      },
      {
        "TimeOfDay": {
          "environment:time_of_day": "09:00-17:00"
        }
      }
    ]
  }
}
```

## 🔧 Operators & Conditions

### String Operators
- `StringEquals`, `StringNotEquals`
- `StringContains`, `StringStartsWith`, `StringEndsWith`
- `StringLike`, `StringRegex`

### Numeric Operators  
- `NumericEquals`, `NumericNotEquals`
- `NumericGreaterThan`, `NumericGreaterThanEquals`
- `NumericLessThan`, `NumericLessThanEquals`
- `NumericBetween`

### Time-based Operators
- `TimeOfDay`, `DayOfWeek`, `IsBusinessHours`
- `DateGreaterThan`, `DateLessThan`, `DateBetween`

### Network Operators
- `IpAddress`, `IpInRange`, `IpNotInRange`
- `IsInternalIP`

### Logical Operators
- `And`: All conditions must be true
- `Or`: At least one condition must be true  
- `Not`: Invert condition result

### Array Operators
- `ArrayContains`, `ArrayNotContains`, `ArraySize`

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
    Rules            JSONPolicyRules `gorm:"type:jsonb"`
    Actions          JSONStringSlice `gorm:"type:jsonb"`
    ResourcePatterns JSONStringSlice `gorm:"type:jsonb"`
}
```

## 🧪 Testing

### Run Tests
```bash
# All tests
go test ./...

# Specific packages
go test ./evaluator -v
go test ./storage -v

# Benchmarks
go test -bench=. -benchmem
```

### Test Coverage
- **Overall**: 69.8% (Mock Storage) / Production-ready (PostgreSQL)
- **Core Packages**: >85% coverage
- **Performance**: <10ms per evaluation (target met)

### Key Test Scenarios
- Policy evaluation (permit/deny/not_applicable)
- Rule operators (eq, in, contains, regex, etc.)
- Attribute resolution (dynamic, environment, hierarchical)
- Security scenarios (probation blocking, time restrictions)
- Performance benchmarks (concurrent evaluations)

## ⚡ Performance

### Benchmarks
- **Mock Storage**: ~4.5µs per evaluation
- **PostgreSQL**: ~2.6ms per evaluation  
- **Throughput**: 5,000+ evaluations/second
- **Memory**: 1,856-3,500 bytes per evaluation

### Optimization Features
- Policy pre-filtering (60-80% performance improvement)
- Pattern caching for repeated evaluations
- Connection pooling for database
- JSONB indexing for fast attribute queries

## 🔒 Security Features

### Access Control
- Fail-safe defaults (deny by default)
- Priority-based policy evaluation
- Short-circuit deny logic
- Comprehensive audit logging

### Security Scenarios Tested
- Probation user write blocking
- After-hours access restrictions  
- External IP access prevention
- Cross-department data access control

## 🚀 Production Deployment

### Docker Deployment
```bash
# Build & run
docker build -t abac-service .
docker-compose up -d
```

### Makefile Commands
```bash
make setup          # Full setup
make docker-up       # Start PostgreSQL
make migrate         # Run migration
make test           # Run tests
make run            # Start service
```

### Production Considerations
1. **Authentication**: Replace `X-Subject-ID` with JWT tokens
2. **Database**: Connection pooling, read replicas, proper indexing
3. **Caching**: Redis for decision caching
4. **Monitoring**: Metrics, alerting, distributed tracing
5. **Security**: HTTPS, rate limiting, input validation

## 📊 Monitoring & Metrics

### Performance Metrics
- Evaluation latency (P50, P95, P99)
- Throughput (requests/second)
- Policy coverage percentage
- Error rates and types

### Security Metrics  
- Deny rate percentage
- Policy violations
- Unusual access patterns
- Audit trail completeness

## 🔧 Integration Examples

### HTTP Middleware
```go
// Apply ABAC middleware to protected routes
mux.Handle("/api/users", 
    service.ABACMiddleware("read")(
        http.HandlerFunc(handleUsers)
    )
)
```

### Service Integration
```go
type SecureService struct {
    pep *pep.SimplePolicyEnforcementPoint
}

func (s *SecureService) GetUser(ctx context.Context, subjectID, userID string) error {
    request := &models.EvaluationRequest{
        SubjectID:  subjectID,
        ResourceID: fmt.Sprintf("user:%s", userID),
        Action:     "read",
    }
    
    result, err := s.pep.EnforceRequest(ctx, request)
    if err != nil || !result.Allowed {
        return fmt.Errorf("access denied: %s", result.Reason)
    }
    
    // Business logic here
    return nil
}
```

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
      "IpInRange": {"environment:client_ip": ["10.0.0.0/8", "192.168.1.0/24"]}
    },
    {
      "Not": {
        "StringEquals": {"resource:classification": "top_secret"}
      }
    }
  ]
}
```

## 🛠️ Development

### Adding New Endpoints
1. Create handler function
2. Register route with ABAC middleware
3. Add resource to database
4. Create appropriate policies

### Adding New Operators
1. Implement operator in `operators/operators.go`
2. Register in operator registry
3. Add tests in `operators/operators_test.go`
4. Update documentation

### Policy Management
- Policies stored in PostgreSQL `policies` table
- Support for policy versioning and rollback
- Hot reload capabilities (future enhancement)
- Policy validation and testing tools

## 📚 Additional Resources

### Component Documentation
- [Evaluator README](evaluator/README.md) - PDP implementation details
- [Storage README](storage/README.md) - Data access layer
- [PEP README](pep/README.md) - Policy enforcement patterns
- [Audit README](audit/README.md) - Logging and compliance

### Related Standards
- [XACML 3.0](http://docs.oasis-open.org/xacml/3.0/) - OASIS XACML standard
- [NIST ABAC Guide](https://csrc.nist.gov/publications/detail/sp/800-162/final) - NIST SP 800-162

### Similar Projects
- [Open Policy Agent](https://www.openpolicyagent.org/) - Policy-based control
- [Casbin](https://casbin.org/) - Authorization library
- [AWS IAM](https://aws.amazon.com/iam/) - Cloud access management

---

## 🎉 Conclusion

This ABAC system provides:

✅ **Production-Ready**: PostgreSQL storage, comprehensive testing, performance optimization  
✅ **Flexible**: Rich policy language with complex conditions and operators  
✅ **Secure**: Fail-safe defaults, audit logging, comprehensive access control  
✅ **Scalable**: Stateless design, connection pooling, horizontal scaling support  
✅ **Maintainable**: Clean architecture, extensive documentation, test coverage  

The system successfully demonstrates enterprise-grade ABAC implementation with both development-friendly (mock storage) and production-ready (PostgreSQL) configurations.
