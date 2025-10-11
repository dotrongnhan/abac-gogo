Dựa trên các file thông tin bạn cung cấp, tôi sẽ giải thích cách thiết kế code cho ABAC system với PostgreSQL database và GORM ORM, cùng với fallback mock data từ JSON:

## **Thiết kế Code Architecture**

### **1. Package Structure**
```
abac/
├── models/          # Data models với GORM tags
├── evaluator/       # PDP - Policy Decision Point
├── attributes/      # PIP - Policy Information Point  
├── operators/       # Rule operators (eq, in, regex...)
├── storage/         # Dual storage implementation
│   ├── postgresql_storage.go  # Production PostgreSQL với GORM
│   ├── mock_storage.go       # Development JSON mock
│   └── database.go           # Database connection management
├── audit/          # Audit logging
├── cmd/            # CLI tools
│   └── migrate/    # Database migration và seeding
├── docker-compose.yml # PostgreSQL development environment
└── main.go
```

### **2. Core Components Design**

#### **2.1 Models Package với GORM Support**
- **Subject, Resource, Action, Policy** structs với GORM tags cho auto-migration
- **Custom JSONB Types**: `JSONMap`, `JSONStringSlice`, `JSONPolicyRules` cho PostgreSQL JSONB storage
- **Database Timestamps**: Auto-managed `created_at`, `updated_at` fields
- **Indexes**: Optimized database indexes cho performance

#### **2.2 Storage Package (Dual Implementation)**
``` go
type Storage interface {
    GetSubject(id string) (*Subject, error)
    GetResource(id string) (*Resource, error)
    GetPolicies() ([]*Policy, error)
    GetAction(name string) (*Action, error)
}

// PostgreSQL Implementation (Production)
type PostgreSQLStorage struct {
    db *gorm.DB
}

// Mock Implementation (Development/Testing)
type MockStorage struct {
    subjects  map[string]Subject   // Load từ subjects.json (values, not pointers)
    resources map[string]Resource  // Load từ resources.json (values, not pointers)
    policies  []*Policy            // Load từ policies.json (still pointers)
    actions   map[string]Action    // Load từ actions.json (values, not pointers)
}
```

**PostgreSQL Storage Features:**
- GORM ORM với connection pooling
- Auto-migration cho schema updates
- JSONB support cho complex attributes
- Optimized queries với proper indexing
- Transaction support cho data consistency

#### **2.3 Evaluator Package (PDP)**
**Main flow:**
1. **FilterApplicablePolicies**: Lọc policies phù hợp với request
2. **SortByPriority**: Sắp xếp theo priority (ascending)
3. **EvaluatePolicies**: Loop qua từng policy
4. **ShortCircuit**: Dừng ngay khi gặp DENY
5. **ReturnDecision**: Trả về quyết định cuối cùng

**Key methods:**
- `Evaluate(request *EvaluationRequest) *Decision`
- `matchResourcePattern(pattern, resource string) bool`
- `evaluateRules(policy *Policy, context *Context) bool`

#### **2.4 Attributes Package (PIP) - Enhanced for GORM**
**Responsibilities:**
- Resolve subject attributes từ database hoặc mock data
- Handle GORM custom types (`JSONMap`, `JSONStringSlice`)
- Resolve resource attributes với JSONB queries
- Enrich environment context
- Handle temporal attributes (valid_from/valid_until)
- Process inherited attributes cho hierarchical resources

**Key methods:**
- `GetSubjectAttributes(subjectID string) map[string]interface{}`
- `GetResourceAttributes(resourceID string) map[string]interface{}`
- `EnrichContext(request *EvaluationRequest) *Context`
- `ResolveHierarchy(resourcePath string) []string`
- `GetAttributeValue(target interface{}, path string) interface{}` - **Updated để handle JSONMap**

#### **2.5 Operators Package**
Implement các operator cho rule evaluation:
- **eq**: So sánh bằng
- **in**: Giá trị trong array
- **contains**: Array chứa giá trị
- **regex**: Pattern matching
- **between**: Range check
- **gte/lte**: So sánh lớn hơn/nhỏ hơn

**Interface design:**
``` go
type Operator interface {
    Evaluate(actual, expected interface{}) bool
}
```

### **3. Database Setup & Migration**

#### **Step 1: Environment Setup**
```bash
# Start PostgreSQL với Docker Compose
docker-compose up -d

# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=abac_system
```

#### **Step 2: Migration & Seeding**
```bash
# Install dependencies
go mod tidy

# Run database migration và seed data từ JSON files
go run cmd/migrate/main.go

# Run application với PostgreSQL
go run main.go
```

#### **Step 3: Development Workflow**
- **Production**: Sử dụng PostgreSQL với environment variables
- **Development**: Fallback to JSON mock data nếu database không available
- **Testing**: Isolated test database hoặc in-memory SQLite

### **4. Evaluation Logic Flow (Updated for Database)**

#### **Step 1: Load Request Context**
- Parse evaluation request
- Fetch subject từ PostgreSQL/mock storage với GORM queries
- Fetch resource từ PostgreSQL/mock storage
- Handle JSONB attributes conversion
- Enrich environment context (time, IP, location)

#### **Step 2: Filter Policies**
- Match resource patterns (support wildcards)
- Match actions
- Check enabled = true
- Build policy candidate list

#### **Step 3: Evaluate Each Policy**
- Sort by priority (low to high)
- For each policy:
    - Evaluate ALL rules (AND logic)
    - Check subject rules
    - Check resource rules
    - Check environment rules
    - If all match → Apply effect

#### **Step 4: Conflict Resolution**
- DENY overrides PERMIT
- First DENY → Return DENY immediately
- No DENY + Has PERMIT → Return PERMIT
- No match → Return NOT_APPLICABLE

### **4. Special Cases Handling**

#### **4.1 Hierarchical Resources**
- Parse resource path (e.g., `/api/v1/users/123`)
- Generate parent paths
- Check policies với `is_recursive` flag
- Accumulate permissions từ parent

#### **4.2 Multi-value Attributes**
- Role là array → Use "contains" operator
- Check "any of" vs "all of" logic
- Handle type mismatches gracefully

#### **4.3 Time-based Access**
- Parse `time_of_day` từ timestamp
- Compare với business hours range
- Check `valid_from/valid_until` cho temporal attrs

#### **4.4 Wildcard Patterns**
- Support `*` wildcard (e.g., `/api/v1/*`)
- Convert to regex for matching
- Cache compiled regex patterns

### **5. Performance Optimizations (Database-Aware)**

#### **5.1 Database Indexing**
```sql
-- Auto-generated indexes từ GORM tags
CREATE INDEX idx_subjects_external_id ON subjects(external_id);
CREATE INDEX idx_subjects_subject_type ON subjects(subject_type);
CREATE INDEX idx_resources_resource_type ON resources(resource_type);
CREATE INDEX idx_policies_enabled ON policies(enabled);
CREATE INDEX idx_policies_priority ON policies(priority);

-- JSONB indexes cho attribute queries
CREATE INDEX idx_subjects_attributes ON subjects USING GIN(attributes);
CREATE INDEX idx_resources_attributes ON resources USING GIN(attributes);
```

#### **5.2 GORM Query Optimization**
- Use preloading cho related entities
- Batch queries cho multiple evaluations
- Connection pooling với proper limits
- Query result caching với TTL

#### **5.3 Memory vs Database Trade-offs**
- Cache frequently accessed policies in memory
- Use database cho large datasets và complex queries
- Fallback to mock data cho development/testing

#### **5.4 Rule Evaluation Cache**
- Cache evaluated rule results trong request lifecycle
- Avoid re-evaluation của same rules

#### **5.5 Batch Processing với Database**
- Group similar requests by entity type
- Single database query cho multiple evaluations
- Parallel rule evaluation với goroutines
- Use database transactions cho consistency

### **6. Audit & Logging**

**Every evaluation logs:**
- Request ID
- Subject/Resource/Action
- Decision (Permit/Deny)
- Policies evaluated
- Evaluation time (ms)
- Context details

### **7. Testing Strategy**

#### **Unit Tests:**
- Test each operator individually
- Test pattern matching logic
- Test priority ordering
- Test conflict resolution

#### **Integration Tests:**
- Full evaluation flow với mock data
- Test các scenarios từ evaluation_requests.json
- Verify expected decisions

#### **Performance Tests:**
- Measure evaluation latency
- Test với 1000+ policies
- Concurrent evaluation tests

### **8. Error Handling**

- Missing attributes → Use defaults or skip rule
- Invalid data types → Type coercion hoặc fail gracefully
- Circular dependencies → Detection và prevention
- Timeout protection → Max evaluation time limit

**Key Design Principles:**
1. **Stateless evaluation** - Không lưu state giữa requests
2. **Fail-safe defaults** - Default DENY cho sensitive resources
3. **Clear audit trail** - Log mọi decision để debug
4. **Performance first** - Optimize hot paths với value-based storage
5. **Memory efficiency** - Use values instead of pointers cho better cache locality
6. **Extensible operators** - Dễ thêm operators mới

Thiết kế này cho phép bạn implement ABAC system hiệu quả với mock data, sau này dễ dàng thay thế bằng real database khi cần.