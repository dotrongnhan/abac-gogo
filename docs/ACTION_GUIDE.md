# Action Field Guide

## Tổng Quan

Field **Action** trong Policy JSON xác định các hành động (operations) mà policy cho phép hoặc từ chối.

## Cú Pháp

### Action Pattern Format
```
<service>:<resource-type>:<operation>
```

**Ví dụ:**
- `document-service:file:read`
- `user-service:profile:update`
- `api:documents:*`

### Định Dạng Hỗ Trợ

#### 1. Single String
```json
{
  "Action": "document-service:file:read"
}
```

#### 2. Array of Strings
```json
{
  "Action": [
    "document-service:file:read",
    "document-service:file:write"
  ]
}
```

## Wildcard Support

### Wildcard `*`
- `*` - Matches bất kỳ giá trị nào
- `document-service:*:*` - Tất cả operations của document-service
- `*:file:read` - Read operation cho tất cả services

### Ví Dụ Wildcard
```json
{
  "Action": "document-service:*:*"    // Tất cả operations của document-service
}
```

```json
{
  "Action": "*:*:read"                // Tất cả read operations
}
```

## Pattern Matching Logic

### Exact Match
```json
{
  "Action": "document-service:file:read"
}
```
✅ Matches: `document-service:file:read`
❌ Không match: `document-service:file:write`

### Wildcard Match
```json
{
  "Action": "document-service:file:*"
}
```
✅ Matches: `document-service:file:read`, `document-service:file:write`
❌ Không match: `user-service:profile:read`

### Array Match
```json
{
  "Action": [
    "document-service:file:read",
    "document-service:file:write"
  ]
}
```
✅ Matches: Bất kỳ action nào trong array

## Ví Dụ Thực Tế

### 1. Read-only Access
```json
{
  "Effect": "Allow",
  "Action": "*:*:read",
  "Resource": "*"
}
```

### 2. Service-specific Access
```json
{
  "Effect": "Allow", 
  "Action": "document-service:*:*",
  "Resource": "api:documents:*"
}
```

### 3. Multiple Operations
```json
{
  "Effect": "Allow",
  "Action": [
    "document-service:file:read",
    "document-service:file:write",
    "document-service:folder:list"
  ],
  "Resource": "api:documents:owner:${user:id}/*"
}
```

### 4. Admin Access
```json
{
  "Effect": "Allow",
  "Action": "*:*:*",
  "Resource": "*",
  "Condition": {
    "StringEquals": {
      "user:role": "admin"
    }
  }
}
```

## Validation Rules

### Format Requirements
- Minimum 3 parts: `service:type:operation`
- Separated by colons `:`
- No empty segments
- Case-sensitive matching

### Valid Examples
- ✅ `document-service:file:read`
- ✅ `api:users:create`
- ✅ `*:*:read`

### Invalid Examples
- ❌ `document-service:file` (missing operation)
- ❌ `document-service::read` (empty segment)
- ❌ `document-service` (insufficient parts)

## Best Practices

### 1. Principle of Least Privilege
```json
// ✅ Good - Specific permissions
{
  "Action": "document-service:file:read"
}

// ❌ Avoid - Too broad
{
  "Action": "*:*:*"
}
```

### 2. Logical Grouping
```json
// ✅ Good - Related operations
{
  "Action": [
    "document-service:file:read",
    "document-service:file:write"
  ]
}
```

### 3. Service Boundaries
```json
// ✅ Good - Service-specific
{
  "Action": "document-service:*:*"
}

// ❌ Avoid - Cross-service wildcards without conditions
{
  "Action": "*:*:*"
}
```

### 4. Consistent Naming
```json
// ✅ Good - Consistent pattern
{
  "Action": [
    "document-service:file:read",
    "document-service:folder:list",
    "document-service:share:create"
  ]
}
```

## Common Patterns

### Read-only User
```json
{
  "Action": "*:*:read"
}
```

### Content Manager
```json
{
  "Action": [
    "document-service:file:*",
    "document-service:folder:*"
  ]
}
```

### Service Admin
```json
{
  "Action": "document-service:*:*"
}
```

### System Admin
```json
{
  "Action": "*:*:*",
  "Condition": {
    "StringEquals": {
      "user:role": "system-admin"
    }
  }
}
```

## Troubleshooting

### Common Issues

1. **Action không match**
   - Kiểm tra format: `service:type:operation`
   - Verify case sensitivity
   - Check wildcard placement

2. **Wildcard không hoạt động**
   - Ensure correct position
   - Verify no typos in pattern

3. **Array matching issues**
   - Check JSON syntax
   - Verify all elements are strings

### Debug Tips

1. **Test với simple patterns trước**
2. **Sử dụng exact match để verify**
3. **Check logs để xem action được parsed như thế nào**

---

*Tài liệu này cung cấp hướng dẫn đầy đủ về Action field cho người mới bắt đầu.*
