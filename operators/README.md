# Operators Package - Rule Evaluation Engine

## 📋 Tổng Quan

Package `operators` cung cấp **Rule Evaluation Engine** - bộ operators để thực hiện các phép so sánh và logic trong policy rules. Đây là component quan trọng quyết định việc một rule có match hay không dựa trên actual value và expected value.

## 🎯 Trách Nhiệm Chính

1. **Comparison Operations**: Thực hiện các phép so sánh cơ bản (eq, neq, gt, lt, etc.)
2. **Collection Operations**: Xử lý arrays và collections (in, contains, etc.)
3. **Pattern Matching**: Support regex và wildcard patterns
4. **Type Conversion**: Handle different data types safely
5. **Range Operations**: Time ranges, numeric ranges
6. **Extensibility**: Easy to add new operators

## 📁 Cấu Trúc Files

```
operators/
├── operators.go          # Operator implementations
└── operators_test.go     # Unit tests cho operators
```

## 🏗️ Core Architecture

### Operator Interface

```go
type Operator interface {
    Evaluate(actual, expected interface{}) bool
}
```

**Design Principles:**
- **Simple Interface**: Chỉ một method `Evaluate`
- **Type Agnostic**: Handle any data types
- **Stateless**: No internal state, thread-safe
- **Composable**: Có thể combine với negation logic

### OperatorRegistry

```go
type OperatorRegistry struct {
    operators map[string]Operator
}
```

**Registry Pattern:**
- **Centralized Management**: Tất cả operators trong một registry
- **Dynamic Registration**: Có thể add operators at runtime
- **Name-based Lookup**: Access operators by string name
- **Error Handling**: Graceful handling của unknown operators

## 🔧 Available Operators

### 1. Basic Comparison Operators

#### EqualOperator (`eq`)
```go
type EqualOperator struct{}

func (o *EqualOperator) Evaluate(actual, expected interface{}) bool {
    return reflect.DeepEqual(actual, expected)
}
```

**Use Cases:**
- Exact string matching: `"engineering" == "engineering"`
- Number comparison: `5 == 5`
- Boolean comparison: `true == true`

**Examples:**
```json
{
  "operator": "eq",
  "actual_value": "engineering",
  "expected_value": "engineering",
  "result": true
}

{
  "operator": "eq", 
  "actual_value": 3,
  "expected_value": 3,
  "result": true
}
```

#### NotEqualOperator (`neq`)
```go
func (o *NotEqualOperator) Evaluate(actual, expected interface{}) bool {
    return !reflect.DeepEqual(actual, expected)
}
```

**Use Cases:**
- Exclusion checks: `department != "hr"`
- Status validation: `status != "disabled"`

### 2. Numeric Comparison Operators

#### GreaterThanOperator (`gt`)
```go
func (o *GreaterThanOperator) Evaluate(actual, expected interface{}) bool {
    return compareNumbers(actual, expected) > 0
}
```

**Type Conversion Logic:**
```go
func toFloat64(value interface{}) float64 {
    switch v := value.(type) {
    case int:
        return float64(v)
    case int32:
        return float64(v)
    case int64:
        return float64(v)
    case float32:
        return float64(v)
    case float64:
        return v
    case string:
        if f, err := strconv.ParseFloat(v, 64); err == nil {
            return f
        }
    }
    return 0
}
```

**Examples:**
```json
{
  "operator": "gt",
  "actual_value": 5,
  "expected_value": 3,
  "result": true
}

{
  "operator": "gte",
  "actual_value": 5,
  "expected_value": 5,
  "result": true
}
```

#### Complete Numeric Operators:
- `gt`: Greater than (`>`)
- `gte`: Greater than or equal (`>=`)
- `lt`: Less than (`<`)
- `lte`: Less than or equal (`<=`)

### 3. Collection Operators

#### InOperator (`in`)
```go
func (o *InOperator) Evaluate(actual, expected interface{}) bool {
    expectedSlice := toSlice(expected)
    if expectedSlice == nil {
        return false
    }
    
    for _, item := range expectedSlice {
        if reflect.DeepEqual(actual, item) {
            return true
        }
    }
    return false
}
```

**Use Cases:**
- Whitelist checking: `data_classification in ["public", "internal"]`
- Role validation: `user_role in ["admin", "manager"]`

**Examples:**
```json
{
  "operator": "in",
  "actual_value": "internal",
  "expected_value": ["public", "internal", "confidential"],
  "result": true
}

{
  "operator": "in",
  "actual_value": "external",
  "expected_value": ["public", "internal"],
  "result": false
}
```

#### ContainsOperator (`contains`)
```go
func (o *ContainsOperator) Evaluate(actual, expected interface{}) bool {
    actualSlice := toSlice(actual)
    if actualSlice == nil {
        return false
    }
    
    for _, item := range actualSlice {
        if reflect.DeepEqual(item, expected) {
            return true
        }
    }
    return false
}
```

**Use Cases:**
- Multi-role users: `user.roles contains "senior_developer"`
- Permission checking: `permissions contains "write"`

**Examples:**
```json
{
  "operator": "contains",
  "actual_value": ["senior_developer", "code_reviewer", "team_lead"],
  "expected_value": "senior_developer",
  "result": true
}

{
  "operator": "contains",
  "actual_value": ["read", "write", "delete"],
  "expected_value": "execute",
  "result": false
}
```

#### NotInOperator (`nin`)
```go
func (o *NotInOperator) Evaluate(actual, expected interface{}) bool {
    inOp := &InOperator{}
    return !inOp.Evaluate(actual, expected)
}
```

**Use Cases:**
- Blacklist checking: `source_ip nin ["192.168.1.100", "10.0.0.1"]`
- Exclusion rules: `department nin ["hr", "legal"]`

### 4. Pattern Matching Operators

#### RegexOperator (`regex`)
```go
func (o *RegexOperator) Evaluate(actual, expected interface{}) bool {
    actualStr := toString(actual)
    expectedStr := toString(expected)
    
    if expectedStr == "" {
        return false
    }
    
    matched, err := regexp.MatchString(expectedStr, actualStr)
    if err != nil {
        return false
    }
    return matched
}
```

**Use Cases:**
- IP pattern matching: `source_ip regex "^10\\."`
- Email validation: `email regex ".*@company\\.com$"`
- Path matching: `resource_path regex "/api/v[0-9]+/.*"`

**Examples:**
```json
{
  "operator": "regex",
  "actual_value": "10.0.1.50",
  "expected_value": "^10\\.",
  "result": true
}

{
  "operator": "regex",
  "actual_value": "john.doe@company.com",
  "expected_value": ".*@company\\.com$",
  "result": true
}
```

**Common Regex Patterns:**
- Internal IP: `^10\\.` hoặc `^192\\.168\\.`
- Email domain: `.*@company\\.com$`
- API versioning: `/api/v[0-9]+/`
- File extensions: `\\.(pdf|doc|docx)$`

### 5. Range Operators

#### BetweenOperator (`between`)
```go
func (o *BetweenOperator) Evaluate(actual, expected interface{}) bool {
    expectedSlice := toSlice(expected)
    if expectedSlice == nil || len(expectedSlice) != 2 {
        return false
    }
    
    // For time-based comparisons
    if actualStr := toString(actual); actualStr != "" {
        if strings.Contains(actualStr, ":") {
            return isTimeBetween(actualStr, toString(expectedSlice[0]), toString(expectedSlice[1]))
        }
    }
    
    // For numeric comparisons
    actualNum := toFloat64(actual)
    lowerNum := toFloat64(expectedSlice[0])
    upperNum := toFloat64(expectedSlice[1])
    
    return actualNum >= lowerNum && actualNum <= upperNum
}
```

**Time Range Logic:**
```go
func isTimeBetween(timeStr, startStr, endStr string) bool {
    parseTime := func(s string) (int, int, error) {
        parts := strings.Split(s, ":")
        if len(parts) != 2 {
            return 0, 0, fmt.Errorf("invalid time format")
        }
        hour, err1 := strconv.Atoi(parts[0])
        minute, err2 := strconv.Atoi(parts[1])
        if err1 != nil || err2 != nil {
            return 0, 0, fmt.Errorf("invalid time format")
        }
        return hour, minute, nil
    }
    
    timeHour, timeMin, _ := parseTime(timeStr)
    startHour, startMin, _ := parseTime(startStr)
    endHour, endMin, _ := parseTime(endStr)
    
    timeMinutes := timeHour*60 + timeMin
    startMinutes := startHour*60 + startMin
    endMinutes := endHour*60 + endMin
    
    return timeMinutes >= startMinutes && timeMinutes <= endMinutes
}
```

**Use Cases:**
- Business hours: `time_of_day between ["08:00", "18:00"]`
- Age ranges: `age between [18, 65]`
- Score ranges: `performance_score between [80, 100]`

**Examples:**
```json
{
  "operator": "between",
  "actual_value": "14:30",
  "expected_value": ["08:00", "18:00"],
  "result": true
}

{
  "operator": "between",
  "actual_value": 25,
  "expected_value": [18, 65],
  "result": true
}
```

### 6. Existence Operators

#### ExistsOperator (`exists`)
```go
func (o *ExistsOperator) Evaluate(actual, expected interface{}) bool {
    return actual != nil
}
```

**Use Cases:**
- Optional field validation: `manager_id exists`
- Attribute presence: `clearance_level exists`

## 🔄 Operator Registry Management

### Registry Initialization
```go
func NewOperatorRegistry() *OperatorRegistry {
    registry := &OperatorRegistry{
        operators: make(map[string]Operator),
    }
    
    // Register default operators
    registry.Register("eq", &EqualOperator{})
    registry.Register("neq", &NotEqualOperator{})
    registry.Register("in", &InOperator{})
    registry.Register("nin", &NotInOperator{})
    registry.Register("contains", &ContainsOperator{})
    registry.Register("regex", &RegexOperator{})
    registry.Register("gt", &GreaterThanOperator{})
    registry.Register("gte", &GreaterThanEqualOperator{})
    registry.Register("lt", &LessThanOperator{})
    registry.Register("lte", &LessThanEqualOperator{})
    registry.Register("between", &BetweenOperator{})
    registry.Register("exists", &ExistsOperator{})
    
    return registry
}
```

### Dynamic Registration
```go
func (r *OperatorRegistry) Register(name string, operator Operator) {
    r.operators[name] = operator
}

func (r *OperatorRegistry) Get(name string) (Operator, error) {
    operator, exists := r.operators[name]
    if !exists {
        return nil, fmt.Errorf("operator not found: %s", name)
    }
    return operator, nil
}
```

## 🔍 Operator Usage Examples

### Example 1: Engineering Department Access
```json
{
  "rule": {
    "target_type": "subject",
    "attribute_path": "attributes.department",
    "operator": "eq",
    "expected_value": "engineering"
  },
  "context": {
    "subject": {
      "attributes": {
        "department": "engineering"
      }
    }
  },
  "evaluation": {
    "actual_value": "engineering",
    "expected_value": "engineering",
    "operator": "eq",
    "result": true
  }
}
```

### Example 2: Multi-Role Check
```json
{
  "rule": {
    "target_type": "subject",
    "attribute_path": "attributes.role",
    "operator": "contains",
    "expected_value": "senior_developer"
  },
  "context": {
    "subject": {
      "attributes": {
        "role": ["senior_developer", "code_reviewer"]
      }
    }
  },
  "evaluation": {
    "actual_value": ["senior_developer", "code_reviewer"],
    "expected_value": "senior_developer",
    "operator": "contains",
    "result": true
  }
}
```

### Example 3: Business Hours Check
```json
{
  "rule": {
    "target_type": "environment",
    "attribute_path": "time_of_day",
    "operator": "between",
    "expected_value": ["08:00", "18:00"]
  },
  "context": {
    "environment": {
      "time_of_day": "14:30"
    }
  },
  "evaluation": {
    "actual_value": "14:30",
    "expected_value": ["08:00", "18:00"],
    "operator": "between",
    "result": true
  }
}
```

### Example 4: IP Range Validation
```json
{
  "rule": {
    "target_type": "environment",
    "attribute_path": "source_ip",
    "operator": "regex",
    "expected_value": "^10\\."
  },
  "context": {
    "environment": {
      "source_ip": "10.0.1.50"
    }
  },
  "evaluation": {
    "actual_value": "10.0.1.50",
    "expected_value": "^10\\.",
    "operator": "regex",
    "result": true
  }
}
```

## 🔧 Helper Functions

### Type Conversion Functions

#### toSlice - Array Conversion
```go
func toSlice(value interface{}) []interface{} {
    if value == nil {
        return nil
    }
    
    v := reflect.ValueOf(value)
    if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
        return nil
    }
    
    result := make([]interface{}, v.Len())
    for i := 0; i < v.Len(); i++ {
        result[i] = v.Index(i).Interface()
    }
    return result
}
```

#### toString - String Conversion
```go
func toString(value interface{}) string {
    if value == nil {
        return ""
    }
    
    switch v := value.(type) {
    case string:
        return v
    case fmt.Stringer:
        return v.String()
    default:
        return fmt.Sprintf("%v", v)
    }
}
```

#### compareNumbers - Numeric Comparison
```go
func compareNumbers(actual, expected interface{}) int {
    actualNum := toFloat64(actual)
    expectedNum := toFloat64(expected)
    
    if actualNum > expectedNum {
        return 1
    } else if actualNum < expectedNum {
        return -1
    }
    return 0
}
```

## 🎯 Custom Operator Development

### Creating Custom Operators

```go
// Example: Custom operator cho email domain validation
type EmailDomainOperator struct{}

func (o *EmailDomainOperator) Evaluate(actual, expected interface{}) bool {
    email := toString(actual)
    expectedDomain := toString(expected)
    
    if email == "" || expectedDomain == "" {
        return false
    }
    
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return false
    }
    
    return parts[1] == expectedDomain
}

// Registration
registry.Register("email_domain", &EmailDomainOperator{})
```

**Usage:**
```json
{
  "target_type": "subject",
  "attribute_path": "metadata.email",
  "operator": "email_domain",
  "expected_value": "company.com"
}
```

### Advanced Custom Operators

```go
// Geographic distance operator
type GeoDistanceOperator struct{}

func (o *GeoDistanceOperator) Evaluate(actual, expected interface{}) bool {
    // actual: {"lat": 10.762622, "lng": 106.660172}
    // expected: {"center": {"lat": 10.762622, "lng": 106.660172}, "radius_km": 10}
    
    actualMap, ok1 := actual.(map[string]interface{})
    expectedMap, ok2 := expected.(map[string]interface{})
    
    if !ok1 || !ok2 {
        return false
    }
    
    // Calculate distance và compare với radius
    distance := calculateDistance(actualMap, expectedMap["center"])
    radius := expectedMap["radius_km"].(float64)
    
    return distance <= radius
}
```

## 🧪 Testing Strategies

### Unit Tests for Each Operator
```go
func TestEqualOperator(t *testing.T) {
    op := &EqualOperator{}
    
    // Test cases
    testCases := []struct {
        actual   interface{}
        expected interface{}
        result   bool
    }{
        {"engineering", "engineering", true},
        {"engineering", "finance", false},
        {5, 5, true},
        {5, 3, false},
        {true, true, true},
        {true, false, false},
    }
    
    for _, tc := range testCases {
        result := op.Evaluate(tc.actual, tc.expected)
        assert.Equal(t, tc.result, result)
    }
}
```

### Integration Tests
```go
func TestOperatorRegistry(t *testing.T) {
    registry := NewOperatorRegistry()
    
    // Test operator retrieval
    op, err := registry.Get("eq")
    assert.NoError(t, err)
    assert.NotNil(t, op)
    
    // Test evaluation
    result := op.Evaluate("test", "test")
    assert.True(t, result)
}
```

### Performance Tests
```go
func BenchmarkOperatorEvaluation(b *testing.B) {
    op := &EqualOperator{}
    
    for i := 0; i < b.N; i++ {
        op.Evaluate("engineering", "engineering")
    }
    // Target: < 100ns per evaluation
}
```

## ⚡ Performance Optimizations

### 1. Type-Specific Optimizations
```go
// Fast path cho string comparisons
func (o *EqualOperator) Evaluate(actual, expected interface{}) bool {
    // Fast path for strings
    if actualStr, ok1 := actual.(string); ok1 {
        if expectedStr, ok2 := expected.(string); ok2 {
            return actualStr == expectedStr
        }
    }
    
    // Fallback to reflect.DeepEqual
    return reflect.DeepEqual(actual, expected)
}
```

### 2. Compiled Regex Caching
```go
type RegexOperator struct {
    cache map[string]*regexp.Regexp
    mutex sync.RWMutex
}

func (o *RegexOperator) Evaluate(actual, expected interface{}) bool {
    pattern := toString(expected)
    
    o.mutex.RLock()
    regex, exists := o.cache[pattern]
    o.mutex.RUnlock()
    
    if !exists {
        compiled, err := regexp.Compile(pattern)
        if err != nil {
            return false
        }
        
        o.mutex.Lock()
        o.cache[pattern] = compiled
        o.mutex.Unlock()
        
        regex = compiled
    }
    
    return regex.MatchString(toString(actual))
}
```

### 3. Memory Pool for Conversions
```go
var stringPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 64)
    },
}

func toStringOptimized(value interface{}) string {
    buf := stringPool.Get().([]byte)
    defer stringPool.Put(buf[:0])
    
    // Use buffer for string conversion
    // ... implementation
}
```

## 🔒 Security Considerations

### 1. Regex DoS Protection
```go
func (o *RegexOperator) Evaluate(actual, expected interface{}) bool {
    pattern := toString(expected)
    
    // Limit pattern complexity
    if len(pattern) > 1000 {
        return false
    }
    
    // Timeout protection
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    // Use context-aware regex matching
    return o.matchWithTimeout(ctx, pattern, toString(actual))
}
```

### 2. Input Validation
```go
func (o *BetweenOperator) Evaluate(actual, expected interface{}) bool {
    expectedSlice := toSlice(expected)
    
    // Validate input format
    if expectedSlice == nil || len(expectedSlice) != 2 {
        return false
    }
    
    // Validate range order
    if toFloat64(expectedSlice[0]) > toFloat64(expectedSlice[1]) {
        return false // Invalid range
    }
    
    // ... rest of implementation
}
```

### 3. Type Safety
```go
func safeTypeConversion(value interface{}, targetType reflect.Type) (interface{}, error) {
    if value == nil {
        return nil, nil
    }
    
    sourceType := reflect.TypeOf(value)
    if sourceType == targetType {
        return value, nil
    }
    
    // Safe conversion logic
    // ... implementation
}
```

## 📊 Monitoring & Metrics

### Key Metrics
- **Evaluation Latency**: Time per operator evaluation
- **Error Rate**: Failed evaluations per operator type
- **Usage Distribution**: Most/least used operators
- **Cache Hit Rate**: For regex pattern caching

### Performance Targets
- **Basic Operators**: < 100ns per evaluation
- **Regex Operators**: < 1ms per evaluation
- **Collection Operators**: < 1μs per evaluation
- **Memory Usage**: < 10MB for operator registry

## 🎯 Best Practices

1. **Operator Selection**: Choose most efficient operator cho use case
2. **Type Consistency**: Maintain consistent types trong comparisons
3. **Error Handling**: Graceful handling của type mismatches
4. **Performance**: Profile và optimize hot paths
5. **Security**: Validate inputs để prevent DoS attacks
6. **Testing**: Comprehensive test coverage cho all operators

Package `operators` cung cấp flexible và performant rule evaluation engine, supporting complex policy logic với type-safe operations.
