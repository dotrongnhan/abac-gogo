# Resource Field Guide

## Tổng Quan

Field **Resource** trong Policy JSON xác định tài nguyên (resources) mà policy áp dụng.

## Cú Pháp

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

## Định Dạng Hỗ Trợ

#### 1. Single String
```json
{
  "Resource": "api:documents:doc-123"
}
```

#### 2. Array of Strings
```json
{
  "Resource": [
    "api:documents:doc-123",
    "api:documents:doc-456"
  ]
}
```

## Pattern Types

### 1. Standard Pattern
```
api:documents:doc-123
api:users:user-456
```

### 2. Extended Pattern
```
api:documents:owner:user-123
api:files:project:proj-alpha
```

### 3. Hierarchical Pattern
```
api:documents:owner:user-123/file:doc-456
api:storage:org:acme/dept:engineering/project:alpha/file:spec.pdf
```

## Wildcard Support

### Wildcard `*`
- `*` - Matches bất kỳ segment nào
- `api:documents:*` - Tất cả documents
- `api:*:owner:user-123` - Tất cả resources của user-123

### Ví Dụ Wildcard
```json
{
  "Resource": "api:documents:*"           // Tất cả documents
}
```

```json
{
  "Resource": "api:*:owner:${user:id}"    // Tất cả resources của user
}
```

## Variable Substitution

### Context Variables
```json
{
  "Resource": "api:documents:owner:${user:id}"
}
```

### Supported Contexts
- `${user:attribute}` - User attributes
- `${resource:attribute}` - Resource attributes  
- `${environment:attribute}` - Environment context
- `${request:attribute}` - Request context

### Ví Dụ Variables
```json
{
  "Resource": "api:documents:owner:${user:id}/file:*"
}
```

```json
{
  "Resource": "api:storage:org:${user:organization}/*"
}
```

## Hierarchical Resources

### Cấu Trúc Phân Cấp
```
Organization → Department → Project → File
api:documents:org:acme/dept:engineering/project:alpha/file:spec.pdf
```

### Use Cases

#### 1. Multi-tenant SaaS
```
api:documents:org:${user:organization}/dept:${user:department}/*
```

#### 2. Healthcare System
```
api:medical:hospital:${user:hospital}/dept:${user:department}/patient:*
```

#### 3. Financial Services
```
api:banking:bank:${user:bank}/branch:${user:branch}/customer:*
```

## NotResource - Exclusion Pattern

### Cú Pháp
```json
{
  "Effect": "Allow",
  "Action": "*:*:*",
  "Resource": "*",
  "NotResource": [
    "api:admin:*",
    "api:system:*"
  ]
}
```

### Logic
- **Resource**: Định nghĩa resources được INCLUDE
- **NotResource**: Định nghĩa resources được EXCLUDE
- Nếu cả hai được chỉ định: `(Resource match) AND NOT (NotResource match)`

## Pattern Matching Logic

### Exact Match
```json
{
  "Resource": "api:documents:doc-123"
}
```
✅ Matches: `api:documents:doc-123`
❌ Không match: `api:documents:doc-456`

### Wildcard Match
```json
{
  "Resource": "api:documents:*"
}
```
✅ Matches: `api:documents:doc-123`, `api:documents:doc-456`
❌ Không match: `api:users:user-123`

### Hierarchical Match
```json
{
  "Resource": "api:documents:org:acme/dept:*/project:alpha/*"
}
```
✅ Matches: `api:documents:org:acme/dept:engineering/project:alpha/file:spec.pdf`
❌ Không match: `api:documents:org:other/dept:engineering/project:alpha/file:spec.pdf`

## Ví Dụ Thực Tế

### 1. Own Resources Only
```json
{
  "Effect": "Allow",
  "Action": "document-service:*:*",
  "Resource": "api:documents:owner:${user:id}/*"
}
```

### 2. Department Access
```json
{
  "Effect": "Allow", 
  "Action": "document-service:file:read",
  "Resource": "api:documents:org:${user:organization}/dept:${user:department}/*"
}
```

### 3. Public Resources
```json
{
  "Effect": "Allow",
  "Action": "document-service:file:read",
  "Resource": "api:documents:visibility:public/*"
}
```

### 4. Admin Exclusion
```json
{
  "Effect": "Allow",
  "Action": "*:*:read",
  "Resource": "*",
  "NotResource": [
    "api:admin:*",
    "api:system:config:*"
  ]
}
```

## Validation Rules

### Format Requirements
- Minimum 2 parts per segment: `type:id`
- Separated by colons `:` within segments
- Hierarchical levels separated by `/`
- No empty segments
- Variables must follow `${context:attribute}` format

### Valid Examples
- ✅ `api:documents:doc-123`
- ✅ `api:documents:owner:user-123`
- ✅ `api:documents:org:acme/dept:engineering/file:*`
- ✅ `api:*:owner:${user:id}`

### Invalid Examples
- ❌ `api:documents` (insufficient parts)
- ❌ `api::doc-123` (empty segment)
- ❌ `api:documents:owner:` (empty value)
- ❌ `${invalid-variable}` (invalid variable format)

## Best Practices

### 1. Principle of Least Privilege
```json
// ✅ Good - Specific resources
{
  "Resource": "api:documents:owner:${user:id}/*"
}

// ❌ Avoid - Too broad
{
  "Resource": "*"
}
```

### 2. Use Variables for Dynamic Access
```json
// ✅ Good - Dynamic based on user
{
  "Resource": "api:documents:org:${user:organization}/*"
}

// ❌ Avoid - Hardcoded values
{
  "Resource": "api:documents:org:hardcoded-org/*"
}
```

### 3. Logical Hierarchy
```json
// ✅ Good - Logical structure
{
  "Resource": "api:storage:company:${user:company}/division:${user:division}/*"
}

// ❌ Avoid - Illogical structure
{
  "Resource": "api:storage:file:*/division:*/company:*"
}
```

### 4. Security Boundaries
```json
// ✅ Good - Organization boundary
{
  "Resource": "api:*:org:${user:organization}/*"
}
```

## Common Patterns

### Own Resources
```json
{
  "Resource": "api:documents:owner:${user:id}/*"
}
```

### Department Resources
```json
{
  "Resource": "api:documents:org:${user:organization}/dept:${user:department}/*"
}
```

### Project Collaboration
```json
{
  "Resource": "api:documents:org:${user:organization}/dept:*/project:${user:projects}/*"
}
```

### Public Access
```json
{
  "Resource": "api:documents:visibility:public/*"
}
```

## Troubleshooting

### Common Issues

1. **Resource không match**
   - Kiểm tra format segments
   - Verify variable substitution
   - Check hierarchical structure

2. **Variable substitution fails**
   - Verify variable syntax: `${context:attribute}`
   - Check context availability
   - Ensure attribute exists

3. **Hierarchical matching issues**
   - Verify segment alignment
   - Check wildcard placement
   - Validate hierarchical structure

### Debug Tips

1. **Test với simple patterns trước**
2. **Verify variable values trong context**
3. **Check logs để xem resource được parsed như thế nào**
4. **Use exact match để isolate issues**

---

*Tài liệu này cung cấp hướng dẫn đầy đủ về Resource field cho người mới bắt đầu.*
