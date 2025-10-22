## **Yêu cầu & Logic xử lý cho ABAC System**

### **1. Core Requirements (Functional)**

#### **1.1 Policy Evaluation Engine**
- **Multi-dimensional matching**: Evaluate dựa trên 4 chiều (Subject + Resource + Action + Environment)
- **Priority-based resolution**: Policy có priority thấp hơn được evaluate trước
- **Conflict resolution**: Deny overrides Permit (hoặc configurable)
- **Short-circuit evaluation**: Dừng ngay khi có Deny policy match
- **Partial matching**: Support wildcards (`/api/v1/*`), regex patterns
- **Hierarchical evaluation**: Child resources inherit parent policies

#### **1.2 Attribute Resolution**
- **Just-in-time fetching**: Lấy attributes khi cần, không pre-load hết
- **Temporal attributes**: Xử lý `valid_from/valid_until` cho time-based access
- **Dynamic attributes**: Tính toán real-time (vd: `years_of_service` từ hire_date)
- **Inherited attributes**: Resource con kế thừa attributes từ parent
- **Multi-value attributes**: Role là array, cần logic "contains any" vs "contains all"
- **JSONB Support**: Handle PostgreSQL JSONB data types với custom Go types

#### **1.3 Context Enrichment**
- **Environment injection**: Tự động thêm time, IP, location vào context
- **Derived attributes**: Tính toán attributes từ raw data (vd: `is_business_hours` từ timestamp)
- **External lookups**: Query external systems cho additional context
- **Caching strategy**: Cache attributes với TTL khác nhau theo loại

### **2. Database Architecture Decisions**

#### **2.1 Storage Layer Design**
```yaml
Decision: Dual Storage Implementation
Status: Implemented
Context: Cần support cả development và production environments
Consequences:
  - PostgreSQL cho production với GORM ORM
  - JSON mock data cho development/testing
  - Unified Storage interface cho consistency
  - Easy switching giữa implementations
```

#### **2.2 PostgreSQL & GORM Selection**
```yaml
Decision: PostgreSQL với GORM ORM
Status: Implemented
Context: Cần production-ready database với complex data types
Alternatives Considered:
  - MySQL: Không có native JSONB support
  - MongoDB: NoSQL không phù hợp với relational policies
  - SQLite: Không scale cho production
Consequences:
  - JSONB support cho complex attributes
  - ACID transactions cho data consistency
  - Advanced indexing cho performance
  - Connection pooling built-in
  - Auto-migration capabilities
```

#### **2.3 Custom JSONB Types**
```yaml
Decision: Custom Go types cho PostgreSQL JSONB
Status: Implemented
Context: Cần store complex Go data structures in database
Implementation:
  - JSONMap: map[string]interface{} → JSONB
  - JSONStringSlice: []string → JSONB  
  - JSONPolicyRules: []PolicyRule → JSONB
Consequences:
  - Type-safe database operations
  - Automatic JSON marshaling/unmarshaling
  - JSONB indexing support
  - Seamless attribute resolution
```

### **3. Non-Functional Requirements**

#### **3.1 Performance với Database**
```yaml
Database Performance Targets:
  - P50: < 5ms (với database caching)
  - P95: < 25ms (including database queries)  
  - P99: < 50ms (complex JSONB queries)

Throughput:
  - 5,000 evaluations/second per instance (database-backed)
  - 15,000 evaluations/second (with in-memory caching)
  - Linear scaling với database connection pooling

Database Optimization:
  - Connection pooling: 20-50 connections per instance
  - JSONB GIN indexes cho attribute queries
  - Query result caching với 5-minute TTL
  - Prepared statements cho frequent queries
```

#### **3.2 Scalability Patterns với Database**
- **Read-heavy optimization**: 99% requests là evaluation, 1% là policy updates
- **Database read replicas**: Multiple read replicas cho load distribution
- **Connection pooling**: Shared connection pools across instances
- **Policy caching**: In-memory policy cache với database sync
- **Graceful degradation**: Fallback to cached decisions khi database fails

### **4. Core Logic Components**

#### **4.1 Policy Matching Logic với Database**
```
1. Filter applicable policies (optimized database query):
   - WHERE enabled = true
   - WHERE actions @> '["requested_action"]' (JSONB contains)
   - WHERE resource_patterns overlap với requested resource
   - ORDER BY priority ASC (database-level sorting)
   
2. For each policy:
   - Evaluate all rules in AND/OR logic
   - Handle JSONB attribute extraction
   - If all conditions match:
     - If effect = "deny" → DENY immediately
     - If effect = "permit" → Remember and continue
   
3. Return final decision:
   - If any DENY found → DENY
   - If any PERMIT found → PERMIT  
   - Otherwise → NOT_APPLICABLE
```

#### **4.2 Rule Evaluation Logic với JSONB**
```
For each rule:
  1. Extract actual_value from target:
     - subject.attributes.{path} (from JSONB column)
     - resource.attributes.{path} (from JSONB column)
     - environment.{path}
     
  2. Handle JSONB to Go type conversion:
     - JSONMap → map[string]interface{}
     - JSONStringSlice → []string
     - Proper type assertions và nil checks
     
  3. Apply operator:
     - eq: exact match
     - in: value in array
     - contains: array contains value
     - regex: pattern match
     - between: range check
     - gte/lte: comparison
     
  4. Handle special cases:
     - NULL values từ JSONB
     - Missing attributes
     - Type mismatches
     - Array vs single value
```

#### **4.3 Attribute Resolution Flow với Database**
```
1. Check local cache (LRU, 1MB per instance)
2. Query PostgreSQL database:
   - SELECT attributes FROM subjects WHERE id = ?
   - JSONB extraction với proper indexing
   - Connection pooling cho efficiency
3. Transform & normalize JSONB data:
   - JSONMap.Scan() method
   - Type assertions và validation
4. Cache with appropriate TTL (5-60 minutes)
5. Return Go-native data structures
```

### **4. Complex Scenarios Handling**

#### **4.1 Hierarchical Resources**
```
Resource: /api/v1/users/123/profile
Parents: ["/api", "/api/v1", "/api/v1/users", "/api/v1/users/123"]

Logic:
- Check policies for exact match first
- Then check parent paths với is_recursive flag
- Accumulate permissions (unless explicit deny)
```

#### **4.2 Time-based Access**
```
Scenarios:
- Business hours only (08:00-18:00)
- Temporary elevated permissions
- Expired clearance levels
- Future-dated access

Implementation:
- Store timezone per location
- Convert all times to UTC for comparison
- Cache time-sensitive decisions với short TTL
```

#### **4.3 Delegation & Impersonation**
```
Requirements:
- Manager can act on behalf of team members
- Service accounts can impersonate users
- Audit trail must capture both real & effective identity

Logic:
- Check delegation policy first
- Evaluate as delegated identity
- Log both identities in audit
```

### **5. Optimization Strategies**

#### **5.1 Policy Indexing**
```
Build indexes:
- By resource_type → HashMap
- By action → HashMap  
- By subject.department → HashMap
- Composite index for common queries

Result: O(1) policy filtering instead of O(n)
```

#### **5.2 Decision Caching**
```
Cache key generation:
hash(subject_id + resource_pattern + action + context_hash)

Context normalization:
- Round timestamps to minute
- Normalize IP to /24 subnet
- Ignore non-relevant headers

Cache invalidation:
- Policy change → Clear all
- Attribute change → Clear by subject
- Time-based → Natural TTL expiry
```

#### **5.3 Batch Evaluation**
```
Use case: Check permissions for 100 resources at once

Optimization:
- Group by resource_type
- Single attribute fetch
- Parallel rule evaluation
- Return map[resource_id]decision
```

### **6. Error Handling & Fallback**

#### **6.1 Failure Modes**
```
PIP unreachable:
→ Use cached attributes if available
→ Fall back to "core attributes only" evaluation
→ Default to DENY for sensitive resources

Policy corruption:
→ Keep last known good policy set in memory
→ Alert operators immediately
→ Rollback within 30 seconds

Timeout scenarios:
→ Evaluation timeout: 100ms hard limit
→ Attribute fetch timeout: 30ms
→ Circuit breaker after 3 failures
```

#### **6.2 Audit & Compliance**
```
Every decision must log:
- Who (subject + delegation)
- What (resource + action)
- When (timestamp with timezone)
- Where (IP + location)
- Why (policies evaluated + decision)
- How long (evaluation time)

Retention:
- Hot storage: 7 days (for debugging)
- Warm storage: 90 days (for compliance)
- Cold storage: 7 years (for audit)
```

### **7. Testing Requirements**

#### **7.1 Unit Testing**
- Each operator logic (eq, in, regex...)
- Attribute type conversions
- Policy priority ordering
- Conflict resolution

#### **7.2 Integration Testing**
- Full evaluation flow
- Cache interactions
- External system failures
- Policy updates during evaluation

#### **7.3 Performance Testing**
```
Scenarios:
- 1000 concurrent evaluations
- Policy set with 10,000 rules
- Cache miss storm (cold start)
- Cascading failures
- Memory leaks over 24 hours
```

### **8. Deployment Considerations**

#### **8.1 Policy Deployment**
```
Requirements:
- Zero-downtime updates
- Instant rollback capability
- A/B testing for new policies
- Gradual rollout by percentage

Implementation:
- Version all policy sets
- Use feature flags for activation
- Shadow mode evaluation
- Canary deployment
```

#### **8.2 Monitoring & Alerting**
```
Key metrics:
- Evaluation latency (P50, P95, P99)
- Cache hit rates
- Policy match rates
- Deny/Permit ratio
- Error rates by component

Alerts:
- Latency > 100ms for P95
- Cache hit rate < 70%
- Error rate > 0.1%
- No policy matches > 5%
```

**Tóm lại:** System cần xử lý được evaluation nhanh, scale tốt, handle nhiều edge cases, và maintain được audit trail đầy đủ. Focus vào caching strategy, value-based storage optimization, và memory efficiency là critical cho performance.