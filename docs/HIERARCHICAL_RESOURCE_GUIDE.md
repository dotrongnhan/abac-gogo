# Hierarchical Resource Guide

## Tổng Quan

Hierarchical resources cho phép mô hình hóa cấu trúc phân cấp phức tạp trong enterprise applications. Hệ thống hỗ trợ kết hợp Extended Format với Hierarchical Structure để tạo ra các pattern linh hoạt và mạnh mẽ.

## Cú Pháp Cơ Bản

### Standard Format
```
<service>:<resource-type>:<resource-id>
```

### Extended Format  
```
<service>:<resource-type>:<owner-type>:<owner-id>
```

### Hierarchical Format
```
<parent>/<child>/<grandchild>
```

### Hierarchical + Extended Format
```
<service>:<resource-type>:<owner-type>:<owner-id>/<child-type>:<child-id>/<grandchild-type>:<grandchild-id>
```

## Enterprise Use Cases

### 1. Multi-tenant SaaS Platform

**Cấu trúc:** Organization → Department → Project → File

```
api:documents:org:acme-corp/dept:engineering/project:alpha/file:design-spec.pdf
```

**Policy Examples:**
```json
{
  "Effect": "Allow",
  "Action": "document-service:*:*",
  "Resource": "api:documents:org:${user:Organization}/*/*"
}
```

### 2. Healthcare System

**Cấu trúc:** Hospital → Department → Ward → Patient → Record

```
api:medical:hospital:general/dept:cardiology/ward:icu/patient:p-123/record:vitals
```

**Policy Examples:**
```json
{
  "Effect": "Allow",
  "Action": "medical:record:read",
  "Resource": "api:medical:hospital:${user:Hospital}/dept:${user:Department}/*/*"
}
```

### 3. Financial Services

**Cấu trúc:** Bank → Branch → Customer → Account → Transaction

```
api:banking:bank:chase/branch:manhattan/customer:c-456/account:checking/transaction:txn-789
```

**Policy Examples:**
```json
{
  "Effect": "Allow",
  "Action": "banking:account:view",
  "Resource": "api:banking:bank:${user:Bank}/branch:${user:Branch}/*/*"
}
```

### 4. E-commerce Marketplace

**Cấu trúc:** Platform → Vendor → Category → Product → Variant

```
api:marketplace:platform:amazon/vendor:apple/category:electronics/product:iphone/variant:pro-max
```

**Policy Examples:**
```json
{
  "Effect": "Allow",
  "Action": "marketplace:product:*",
  "Resource": "api:marketplace:platform:*/vendor:${user:VendorId}/*/*"
}
```

### 5. Educational Institution

**Cấu trúc:** University → Faculty → Department → Course → Assignment

```
api:education:university:mit/faculty:engineering/dept:cs/course:cs-101/assignment:hw-1
```

**Policy Examples:**
```json
{
  "Effect": "Allow",
  "Action": "education:assignment:submit",
  "Resource": "api:education:university:${user:University}/*/*/course:${user:EnrolledCourses}/*"
}
```

### 6. Enterprise Cloud Storage

**Cấu trúc:** Company → Division → Team → Folder → File

```
api:storage:company:microsoft/division:azure/team:compute/folder:configs/file:prod.json
```

**Policy Examples:**
```json
{
  "Effect": "Allow",
  "Action": "storage:file:read",
  "Resource": "api:storage:company:${user:Company}/division:*/team:${user:Team}/*"
}
```

## Pattern Matching

### Wildcards
- `*` - Matches any single segment
- `api:documents:org:*/dept:*/project:*/file:*`

### Variables
- `${user:attribute}` - Dynamic substitution
- `api:documents:org:${user:Organization}/dept:${user:Department}/*`

### Mixed Patterns
```
api:documents:org:${user:Organization}/dept:*/project:${user:Projects}/file:*
```

## Best Practices

### 1. Consistent Naming Convention
```
✅ Good: api:documents:org:acme/dept:engineering/project:alpha
❌ Bad:  api:docs:organization:acme/department:eng/proj:a
```

### 2. Logical Hierarchy
```
✅ Good: company/division/team/project/file
❌ Bad:  file/project/team/division/company
```

### 3. Security Boundaries
- Enforce organization boundaries at top level
- Use variables for dynamic access control
- Implement least privilege principle

### 4. Performance Optimization
- Limit hierarchical depth (≤4 levels recommended)
- Use specific patterns instead of wildcards when possible
- Cache frequently used patterns

## Validation Rules

### Segment Requirements
- Minimum 2 parts per segment: `type:id`
- No empty segments allowed
- Variables must follow `${context:attribute}` format

### Hierarchical Structure
- Segments separated by `/`
- Each segment validated independently
- Support for unlimited depth (with performance considerations)

## Common Patterns

### Organization-based Access
```
api:*:org:${user:Organization}/*/*
```

### Department-level Control
```
api:documents:org:${user:Organization}/dept:${user:Department}/*/*
```

### Cross-department Collaboration
```
api:documents:org:${user:Organization}/dept:*/project:${user:Projects}/*
```

### Public Resource Access
```
api:documents:org:*/dept:*/project:*/file:public-*
```

## Troubleshooting

### Common Issues

1. **Validation Errors**
   - Ensure each segment has at least 2 parts
   - Check for empty segments
   - Verify variable syntax

2. **Pattern Matching Failures**
   - Verify hierarchical structure alignment
   - Check wildcard placement
   - Validate variable substitution

3. **Performance Issues**
   - Reduce hierarchical depth
   - Use more specific patterns
   - Implement caching for complex patterns

### Debug Tips

1. **Test with Simple Patterns First**
   ```
   api:documents:org:test-org/file:test-file
   ```

2. **Gradually Add Complexity**
   ```
   api:documents:org:test-org/dept:test-dept/file:test-file
   ```

3. **Verify Variable Substitution**
   ```
   api:documents:org:${user:Organization}/file:*
   ```

## Integration Examples

### Policy Definition
```json
{
  "id": "hierarchical-access-policy",
  "policy_name": "Hierarchical Resource Access",
  "statement": [
    {
      "Sid": "OrgAdminAccess",
      "Effect": "Allow",
      "Action": "*:*:*",
      "Resource": "api:*:org:${user:Organization}/*/*"
    },
    {
      "Sid": "DeptManagerAccess",
      "Effect": "Allow", 
      "Action": "document-service:*:*",
      "Resource": "api:documents:org:${user:Organization}/dept:${user:Department}/*/*"
    },
    {
      "Sid": "ProjectMemberRead",
      "Effect": "Allow",
      "Action": "document-service:file:read", 
      "Resource": "api:documents:org:${user:Organization}/dept:*/project:${user:Projects}/*"
    }
  ]
}
```

### Code Usage
```go
// Create subject from user ID
subject, err := storage.BuildSubjectFromUser("user-123")
if err != nil {
    return err
}

request := &models.EvaluationRequest{
    Subject:    subject,
    ResourceID: "api:documents:org:acme-corp/dept:engineering/project:alpha/file:spec.pdf",
    Action:     "document-service:file:read",
    Context: map[string]interface{}{
        "user:Organization": "acme-corp",
        "user:Department":   "engineering", 
        "user:Projects":     []string{"alpha", "beta"},
    },
}

decision, err := pdp.Evaluate(request)
```

## Performance Considerations

### Matching Complexity

| Format Type | Complexity | Performance |
|-------------|------------|-------------|
| Simple (3 parts) | O(3) | Fastest |
| Extended (4+ parts) | O(n) | Fast |
| Hierarchical (2 levels) | O(2×n) | Good |
| Hierarchical + Extended | O(levels×segments) | Acceptable |

### Optimization Strategies

1. **Pattern Caching** - Cache compiled patterns for reuse
2. **Early Validation** - Validate format before complex matching
3. **Depth Limits** - Implement configurable depth limits
4. **Index Optimization** - Structure data for efficient lookups

## Security Considerations

### Access Control
- Implement fail-safe defaults (deny by default)
- Use deny-override algorithm for policy combining
- Validate all inputs thoroughly

### Compliance Support
- **HIPAA** - Patient data isolation
- **SOX** - Financial audit trails  
- **PCI-DSS** - Payment data protection
- **GDPR** - Personal data access control

### Audit Requirements
- Log all access attempts
- Track policy evaluations
- Maintain compliance trails
- Monitor unusual patterns

---

*Tài liệu này cung cấp hướng dẫn đầy đủ về Hierarchical Resources cho người mới bắt đầu và advanced users.*
