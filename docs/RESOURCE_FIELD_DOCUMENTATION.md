# Tài Liệu Chi Tiết: Field Resource trong Policy JSON

## Mục lục
1. [Tổng quan](#tổng-quan)
2. [Cấu trúc và Định dạng](#cấu-trúc-và-định-dạng)
3. [Resource Pattern Formats](#resource-pattern-formats)
4. [Wildcard Support](#wildcard-support)
5. [Variable Substitution](#variable-substitution)
6. [Hierarchical Resources](#hierarchical-resources)
7. [NotResource - Exclusion Pattern](#notresource---exclusion-pattern)
8. [Logic Matching](#logic-matching)
9. [Ví dụ Chi tiết](#ví-dụ-chi-tiết)
10. [Best Practices](#best-practices)
11. [Lưu ý và Hạn chế](#lưu-ý-và-hạn-chế)
12. [Validation và Error Handling](#validation-và-error-handling)

---

## Tổng quan

Field **Resource** trong Policy JSON xác định **tài nguyên (resources)** mà policy áp dụng. Đây là một trong ba thành phần bắt buộc của một policy statement (cùng với Effect và Action).

### Vai trò của Resource
- **Xác định phạm vi**: Chỉ định cụ thể resource nào được áp dụng policy
- **Granular control**: Cho phép kiểm soát chi tiết từ cấp service đến resource cụ thể
- **Pattern matching**: Hỗ trợ wildcard và variable substitution
- **Hierarchical support**: Hỗ trợ cấu trúc phân cấp (parent/child)

### Vị trí trong Policy
```json
{
  "Statement": [
    {
      "Sid": "ExampleStatement",
      "Effect": "Allow",
      "Action": "...",
      "Resource": "...",           // ← Field Resource ở đây
      "NotResource": "..."         // ← Field NotResource (optional)
    }
  ]
}
```

---

## Cấu trúc và Định dạng

### 1. Hai định dạng được hỗ trợ

Field Resource có thể được định nghĩa theo **hai cách**:

#### a) **Single String** (Chuỗi đơn)
```json
{
  "Resource": "api:documents:owner:user-123/file:doc-456"
}
```

#### b) **Array of Strings** (Mảng chuỗi)
```json
{
  "Resource": [
    "api:documents:owner:user-123/*",
    "api:documents:dept:sales/*"
  ]
}
```

### 2. Implementation trong Code

Field Resource được implement bởi type `JSONActionResource` (dùng chung với Action):

```go
// File: models/types.go:134-139
type JSONActionResource struct {
    Single   string    // Giá trị đơn
    Multiple []string  // Giá trị mảng
    IsArray  bool      // Flag xác định kiểu
}

// Lấy tất cả values dưới dạng slice
func (j JSONActionResource) GetValues() []string {
    if j.IsArray {
        return j.Multiple
    }
    return []string{j.Single}
}
```

### 3. JSON Parsing

Hệ thống tự động phát hiện định dạng khi parse JSON (giống Action field).

---

## Resource Pattern Formats

### 1. Simple Format (Format đơn giản)

Resource pattern đơn giản tuân theo format **3 phần** phân tách bởi dấu hai chấm (`:`):

```
<service>:<resource-type>:<resource-id>
```

**Ví dụ:**
```
api:documents:doc-123
storage:file:file-456
payment:transaction:txn-789
```

### 2. Chi tiết từng phần

#### a) **Service** (Phần 1)
- Tên của service/API
- Định danh hệ thống con
- **Quy tắc**: Không được rỗng

**Ví dụ:**
```
api
storage
payment
auth
notification
```

#### b) **Resource Type** (Phần 2)
- Loại tài nguyên trong service
- Object type của resource
- **Quy tắc**: Không được rỗng

**Ví dụ:**
```
documents
file
transaction
user
profile
folder
```

#### c) **Resource ID** (Phần 3)
- ID cụ thể của resource
- Có thể là wildcard hoặc variable
- **Quy tắc**: Không được rỗng (trừ khi là wildcard `*`)

**Ví dụ:**
```
doc-123
user-456
*
${request:UserId}
```

### 3. Extended Format (Format mở rộng)

Resource có thể có **nhiều hơn 3 phần** để biểu diễn cấu trúc phức tạp hơn:

```
<service>:<resource-type>:<owner-type>:<owner-id>
```

**Ví dụ:**
```
api:documents:owner:user-123
api:documents:dept:sales
api:documents:project:proj-456
storage:file:bucket:public-assets
```

---

## Wildcard Support

### 1. Tổng quan Wildcard

Resource pattern hỗ trợ **wildcard** (`*`) để match nhiều resources. Wildcard có thể xuất hiện ở:
- Từng phần riêng lẻ (service, resource-type, resource-id)
- Toàn bộ resource pattern
- Kết hợp với text (prefix, suffix, middle)

### 2. Full Wildcard

#### Match tất cả resources
```json
{
  "Resource": "*"
}
```
→ Match **BẤT KỲ** resource nào

**Code logic:**
```go
// File: evaluator/matching.go:76-79
func (rm *ResourceMatcher) Match(pattern, resource string, context map[string]interface{}) bool {
    if pattern == "*" {
        return true  // Match tất cả
    }
    // ...
}
```

**Ví dụ:**
```json
{
  "Sid": "AdminFullAccess",
  "Effect": "Allow",
  "Action": "*",
  "Resource": "*"
}
```

### 3. Partial Wildcard

#### a) Wildcard ở Resource ID
```json
{
  "Resource": "api:documents:owner:user-123/*"
}
```
→ Match **TẤT CẢ** documents của owner user-123

**Matches:**
- `api:documents:owner:user-123/doc-1` ✓
- `api:documents:owner:user-123/doc-2` ✓
- `api:documents:owner:user-123/anything` ✓

**Không match:**
- `api:documents:owner:user-456/doc-1` ✗ (owner khác)
- `api:files:owner:user-123/file-1` ✗ (resource-type khác)

#### b) Wildcard ở Resource Type
```json
{
  "Resource": "api:*:owner:user-123"
}
```
→ Match **TẤT CẢ** resource types của owner user-123

**Matches:**
- `api:documents:owner:user-123` ✓
- `api:files:owner:user-123` ✓
- `api:folders:owner:user-123` ✓

#### c) Wildcard ở Service
```json
{
  "Resource": "*:documents:doc-123"
}
```
→ Match document doc-123 trong **TẤT CẢ** services

**Matches:**
- `api:documents:doc-123` ✓
- `storage:documents:doc-123` ✓
- `backup:documents:doc-123` ✓

#### d) Multiple Wildcards
```json
{
  "Resource": "api:documents:*:*"
}
```
→ Match **TẤT CẢ** documents trong api service bất kể owner

**Matches:**
- `api:documents:owner:user-123` ✓
- `api:documents:dept:sales` ✓
- `api:documents:project:proj-456` ✓
- `api:documents:public:shared` ✓

### 4. Pattern Wildcard (Prefix, Suffix, Middle)

#### a) Prefix Wildcard
```json
{
  "Resource": "api:documents:owner:user-*"
}
```
→ Match tất cả documents của owners có ID **bắt đầu** bằng "user-"

**Matches:**
- `api:documents:owner:user-123` ✓
- `api:documents:owner:user-456` ✓
- `api:documents:owner:user-admin` ✓

**Không match:**
- `api:documents:owner:admin-123` ✗

#### b) Suffix Wildcard
```json
{
  "Resource": "api:documents:owner:*-admin"
}
```
→ Match tất cả documents của owners có ID **kết thúc** bằng "-admin"

**Matches:**
- `api:documents:owner:user-admin` ✓
- `api:documents:owner:super-admin` ✓
- `api:documents:owner:system-admin` ✓

#### c) Middle Wildcard
```json
{
  "Resource": "api:*-archive:*"
}
```
→ Match tất cả archive resource types

**Matches:**
- `api:document-archive:doc-1` ✓
- `api:file-archive:file-2` ✓
- `api:data-archive:data-3` ✓

#### d) Wildcard trong Hierarchical Path
```json
{
  "Resource": "api:documents:owner:user-123/*/report-*"
}
```
→ Match tất cả reports trong bất kỳ subfolder nào của user-123

**Matches:**
- `api:documents:owner:user-123/folder-1/report-2024` ✓
- `api:documents:owner:user-123/folder-2/report-sales` ✓

---

## Variable Substitution

### 1. Tổng quan Variable Substitution

Resource pattern hỗ trợ **variable substitution** để tạo dynamic patterns dựa trên context runtime.

**Cú pháp:**
```
${context-key}
```

### 2. Các loại Variables hỗ trợ

#### a) **Request Variables**
```json
{
  "Resource": "api:documents:owner:${request:UserId}/*"
}
```

**Context keys:**
- `${request:UserId}` - User ID từ request
- `${request:ResourceId}` - Resource ID từ request
- `${request:Action}` - Action từ request

#### b) **User/Subject Variables**
```json
{
  "Resource": "api:documents:dept:${user:Department}/*"
}
```

**Context keys:**
- `${user:Department}` - Department của user
- `${user:Role}` - Role của user
- `${user:Team}` - Team của user
- Bất kỳ attribute nào trong `user:*`

#### c) **Resource Variables**
```json
{
  "Resource": "api:documents:owner:${resource:OwnerId}/*"
}
```

**Context keys:**
- `${resource:OwnerId}` - Owner ID của resource
- `${resource:Department}` - Department của resource
- Bất kỳ attribute nào trong `resource:*`

#### d) **Environment Variables**
```json
{
  "Resource": "api:services:region:${environment:region}/*"
}
```

**Context keys:**
- `${environment:region}` - Region từ environment
- `${environment:country}` - Country từ environment

### 3. Variable Substitution Logic

```go
// File: evaluator/matching.go:223-243
func (rm *ResourceMatcher) substituteVariables(pattern string, context map[string]interface{}) string {
    result := pattern

    // Find all ${...} patterns
    re := regexp.MustCompile(`\$\{([^}]+)\}`)
    matches := re.FindAllStringSubmatch(pattern, -1)

    for _, match := range matches {
        if len(match) >= 2 {
            varName := match[1]
            if value, exists := context[varName]; exists {
                if strValue, ok := value.(string); ok {
                    result = strings.ReplaceAll(result, match[0], strValue)
                }
            }
        }
    }

    return result
}
```

### 4. Variable Substitution Flow

```
Original Pattern: "api:documents:owner:${request:UserId}/*"
Context: {"request:UserId": "user-123"}
              ↓
        [Substitute Variables]
              ↓
Expanded Pattern: "api:documents:owner:user-123/*"
              ↓
        [Match with Resource]
              ↓
Resource: "api:documents:owner:user-123/doc-456"
              ↓
Result: MATCH ✓
```

### 5. Ví dụ Chi tiết Variable Substitution

#### Ví dụ 1: Own Documents Access
```json
{
  "Sid": "AccessOwnDocuments",
  "Effect": "Allow",
  "Action": "document-service:file:read",
  "Resource": "api:documents:owner:${request:UserId}/*"
}
```

**Test Cases:**

| Request UserId | Requested Resource | Pattern After Substitution | Match? |
|---------------|-------------------|----------------------------|--------|
| `user-123` | `api:documents:owner:user-123/doc-1` | `api:documents:owner:user-123/*` | ✓ YES |
| `user-123` | `api:documents:owner:user-456/doc-1` | `api:documents:owner:user-123/*` | ✗ NO |
| `user-456` | `api:documents:owner:user-456/doc-2` | `api:documents:owner:user-456/*` | ✓ YES |

#### Ví dụ 2: Department Resources
```json
{
  "Sid": "DepartmentAccess",
  "Effect": "Allow",
  "Action": "document-service:file:*",
  "Resource": "api:documents:dept:${user:Department}/*"
}
```

**Test Cases:**

| User Department | Requested Resource | Pattern After Substitution | Match? |
|----------------|-------------------|----------------------------|--------|
| `sales` | `api:documents:dept:sales/doc-1` | `api:documents:dept:sales/*` | ✓ YES |
| `sales` | `api:documents:dept:hr/doc-1` | `api:documents:dept:sales/*` | ✗ NO |
| `hr` | `api:documents:dept:hr/doc-2` | `api:documents:dept:hr/*` | ✓ YES |

#### Ví dụ 3: Multiple Variables
```json
{
  "Sid": "TeamProjectAccess",
  "Effect": "Allow",
  "Action": "*",
  "Resource": "api:projects:team:${user:Team}/project:${user:CurrentProject}/*"
}
```

**Context:**
```json
{
  "user:Team": "engineering",
  "user:CurrentProject": "proj-alpha"
}
```

**Pattern After Substitution:**
```
api:projects:team:engineering/project:proj-alpha/*
```

#### Ví dụ 4: Wildcard + Variable
```json
{
  "Resource": "api:*:owner:${request:UserId}/*"
}
```
→ Match **TẤT CẢ** resource types thuộc sở hữu của user hiện tại

---

## Hierarchical Resources

### 1. Tổng quan Hierarchical Resources

Hierarchical resources cho phép biểu diễn cấu trúc **parent/child** hoặc **nested resources** bằng cách sử dụng dấu `/`.

**Format:**
```
<parent-resource>/<child-resource>
```

### 2. Single-Level Hierarchy

```json
{
  "Resource": "api:documents:owner:user-123/file:doc-456"
}
```

**Cấu trúc:**
- **Parent**: `api:documents:owner:user-123`
- **Child**: `file:doc-456`

**Giải thích:** Document `doc-456` thuộc owner `user-123`

### 3. Multi-Level Hierarchy

```json
{
  "Resource": "api:storage:bucket:public/folder:images/file:photo.jpg"
}
```

**Cấu trúc:**
- **Level 1**: `api:storage:bucket:public` (Bucket)
- **Level 2**: `folder:images` (Folder trong bucket)
- **Level 3**: `file:photo.jpg` (File trong folder)

### 4. Hierarchical Matching Logic

```go
// File: evaluator/matching.go:120-136
func (rm *ResourceMatcher) matchHierarchical(pattern, resource string) bool {
    // Split by '/' first, then by ':'
    patternParts := rm.parseHierarchical(pattern)
    resourceParts := rm.parseHierarchical(resource)

    if len(patternParts) != len(resourceParts) {
        return false
    }

    for i := 0; i < len(patternParts); i++ {
        if !rm.matchSimple(patternParts[i], resourceParts[i]) {
            return false
        }
    }
    return true
}
```

### 5. Hierarchical với Wildcards

#### a) Wildcard ở Child Level
```json
{
  "Resource": "api:documents:owner:user-123/*"
}
```
→ Match **TẤT CẢ** children của owner user-123

**Matches:**
- `api:documents:owner:user-123/file:doc-1` ✓
- `api:documents:owner:user-123/folder:personal` ✓
- `api:documents:owner:user-123/archive:2024` ✓

#### b) Wildcard ở Parent Level
```json
{
  "Resource": "*/file:doc-456"
}
```
→ Match file `doc-456` trong **BẤT KỲ** parent nào

**Matches:**
- `api:documents:owner:user-123/file:doc-456` ✓
- `storage:bucket:public/file:doc-456` ✓

#### c) Wildcard ở cả Parent và Child
```json
{
  "Resource": "api:documents:owner:*/*"
}
```
→ Match **TẤT CẢ** children trong **TẤT CẢ** owners

**Matches:**
- `api:documents:owner:user-123/file:doc-1` ✓
- `api:documents:owner:user-456/file:doc-2` ✓
- `api:documents:owner:admin/folder:reports` ✓

### 6. Hierarchical với Variables

```json
{
  "Resource": "api:documents:owner:${request:UserId}/folder:${user:DefaultFolder}/*"
}
```

**Context:**
```json
{
  "request:UserId": "user-123",
  "user:DefaultFolder": "personal"
}
```

**After Substitution:**
```
api:documents:owner:user-123/folder:personal/*
```

**Matches:**
- `api:documents:owner:user-123/folder:personal/file:doc-1` ✓
- `api:documents:owner:user-123/folder:personal/report:sales-2024` ✓

### 7. Ví dụ Thực tế: File System Structure

```json
{
  "id": "pol-hierarchical",
  "policy_name": "Hierarchical File Access",
  "statement": [
    {
      "Sid": "OwnFolderFullAccess",
      "Effect": "Allow",
      "Action": "storage:*:*",
      "Resource": "api:storage:bucket:users/folder:${request:UserId}/*"
    },
    {
      "Sid": "SharedFolderRead",
      "Effect": "Allow",
      "Action": "storage:file:read",
      "Resource": "api:storage:bucket:shared/folder:${user:Department}/*"
    },
    {
      "Sid": "PublicRead",
      "Effect": "Allow",
      "Action": "storage:file:read",
      "Resource": "api:storage:bucket:public/*/*"
    }
  ],
  "enabled": true
}
```

**Test Cases:**

| UserId | Department | Resource | Statement Match | Decision |
|--------|-----------|----------|----------------|----------|
| user-123 | sales | `api:storage:bucket:users/folder:user-123/file:doc.pdf` | OwnFolderFullAccess | PERMIT |
| user-123 | sales | `api:storage:bucket:shared/folder:sales/report.xlsx` | SharedFolderRead | PERMIT |
| user-123 | sales | `api:storage:bucket:public/images/logo.png` | PublicRead | PERMIT |
| user-123 | sales | `api:storage:bucket:users/folder:user-456/file:doc.pdf` | None | DENY |

---

## NotResource - Exclusion Pattern

### 1. Tổng quan NotResource

Field **NotResource** cho phép **loại trừ (exclude)** các resources khỏi policy scope. Nó hoạt động như một **negative filter**.

**Cơ chế:**
1. Resource pattern xác định **positive scope**
2. NotResource pattern xác định **negative scope** (loại trừ)
3. Final match = (Match Resource) AND (NOT match NotResource)

### 2. NotResource trong Policy Statement

```go
// File: models/types.go:298-306
type PolicyStatement struct {
    Sid         string             `json:"Sid,omitempty"`
    Effect      string             `json:"Effect"`
    Action      JSONActionResource `json:"Action"`
    Resource    JSONActionResource `json:"Resource"`              // Positive match
    NotResource JSONActionResource `json:"NotResource,omitempty"` // Negative match (exclusion)
    Condition   JSONMap            `json:"Condition,omitempty"`
}
```

### 3. NotResource Evaluation Logic

```go
// File: evaluator/pdp.go:406-425
func (pdp *PolicyDecisionPoint) isResourceMatched(statement models.PolicyStatement, context map[string]interface{}) bool {
    requestedResource, ok := context[ContextKeyRequestResourceID].(string)
    if !ok {
        return false
    }

    // Check positive resource matching
    if !pdp.matchesResourcePatterns(statement.Resource, requestedResource, context) {
        return false
    }

    // Check NotResource exclusions
    return !pdp.matchesNotResourcePatterns(statement.NotResource, requestedResource, context)
}
```

### 4. Ví dụ NotResource

#### Ví dụ 1: Exclude Sensitive Documents
```json
{
  "Sid": "AllDocumentsExceptConfidential",
  "Effect": "Allow",
  "Action": "document-service:file:read",
  "Resource": "api:documents:*",
  "NotResource": "api:documents:sensitivity:confidential/*"
}
```

**Logic:**
- **Allow**: Tất cả documents (`api:documents:*`)
- **EXCEPT**: Documents có sensitivity = confidential

**Test Cases:**

| Requested Resource | Match Resource? | Match NotResource? | Final Match? |
|-------------------|----------------|-------------------|-------------|
| `api:documents:public:doc-1` | ✓ YES | ✗ NO | ✓ YES (Allow) |
| `api:documents:internal:doc-2` | ✓ YES | ✗ NO | ✓ YES (Allow) |
| `api:documents:sensitivity:confidential/doc-3` | ✓ YES | ✓ YES | ✗ NO (Excluded) |

#### Ví dụ 2: All Users Except Admins
```json
{
  "Sid": "RestrictedOperation",
  "Effect": "Deny",
  "Action": "user-service:*:delete",
  "Resource": "api:users:*",
  "NotResource": "api:users:role:admin/*"
}
```

**Logic:**
- **Deny delete**: Tất cả users
- **EXCEPT**: Admin users (admins có thể delete)

**Test Cases:**

| Requested Resource | Match Resource? | Match NotResource? | Final Match? | Effect |
|-------------------|----------------|-------------------|-------------|--------|
| `api:users:role:user/user-123` | ✓ YES | ✗ NO | ✓ YES | DENY |
| `api:users:role:member/user-456` | ✓ YES | ✗ NO | ✓ YES | DENY |
| `api:users:role:admin/admin-1` | ✓ YES | ✓ YES | ✗ NO | No match (Allow if other policy allows) |

#### Ví dụ 3: Multiple NotResource Patterns
```json
{
  "Sid": "AllowExceptSystemAndArchive",
  "Effect": "Allow",
  "Action": "document-service:file:*",
  "Resource": "api:documents:*",
  "NotResource": [
    "api:documents:system/*",
    "api:documents:archive/*"
  ]
}
```

**Logic:**
- **Allow**: Tất cả documents
- **EXCEPT**: System documents AND Archive documents

**Test Cases:**

| Requested Resource | Match Resource? | Match NotResource? | Final Match? |
|-------------------|----------------|-------------------|-------------|
| `api:documents:user:doc-1` | ✓ YES | ✗ NO | ✓ YES |
| `api:documents:dept:sales/doc-2` | ✓ YES | ✗ NO | ✓ YES |
| `api:documents:system/config.json` | ✓ YES | ✓ YES (system) | ✗ NO |
| `api:documents:archive/old-doc` | ✓ YES | ✓ YES (archive) | ✗ NO |

### 5. NotResource Best Practices

#### ✅ Nên:
```json
{
  "Sid": "ExcludeSpecificResources",
  "Resource": "api:documents:*",
  "NotResource": "api:documents:confidential/*"
}
```
→ Rõ ràng, cụ thể

#### ❌ Không nên:
```json
{
  "Sid": "ConfusingExclusion",
  "Resource": "*",
  "NotResource": "*"
}
```
→ Vô nghĩa, loại bỏ tất cả

---

## Logic Matching

### 1. Quy trình Resource Matching trong PDP

```go
// File: evaluator/pdp.go:406-425
func (pdp *PolicyDecisionPoint) isResourceMatched(
    statement models.PolicyStatement,
    context map[string]interface{}
) bool {
    // 1. Get requested resource from context
    requestedResource, ok := context[ContextKeyRequestResourceID].(string)
    if !ok {
        log.Printf("Warning: Missing or invalid resource ID in context")
        return false
    }

    // 2. Validate requested resource
    if requestedResource == "" {
        log.Printf("Warning: Empty resource ID provided")
        return false
    }

    // 3. Check positive resource matching
    if !pdp.matchesResourcePatterns(statement.Resource, requestedResource, context) {
        return false
    }

    // 4. Check NotResource exclusions
    return !pdp.matchesNotResourcePatterns(statement.NotResource, requestedResource, context)
}
```

### 2. ResourceMatcher.Match Logic

```go
// File: evaluator/matching.go:72-101
func (rm *ResourceMatcher) Match(pattern, resource string, context map[string]interface{}) bool {
    // Step 1: Check full wildcard
    if pattern == "*" {
        return true
    }

    // Step 2: Validate resource format
    if !rm.validateResourceFormat(resource) {
        return false
    }

    // Step 3: Substitute variables in pattern
    expandedPattern := rm.substituteVariables(pattern, context)

    // Step 4: Validate expanded pattern
    if !rm.validateResourceFormat(expandedPattern) && expandedPattern != "*" {
        return false
    }

    // Step 5: Handle hierarchical resources
    if strings.Contains(expandedPattern, "/") || strings.Contains(resource, "/") {
        return rm.matchHierarchical(expandedPattern, resource)
    }

    // Step 6: Simple resource matching
    return rm.matchSimple(expandedPattern, resource)
}
```

### 3. Flow Chart

```
Request Resource: "api:documents:owner:user-123/file:doc-456"
Policy Resource:  "api:documents:owner:${request:UserId}/*"
Context: {"request:UserId": "user-123"}

┌─────────────────────────────────────┐
│ 1. Get Resource from context        │
│    → "api:documents:owner:user-123/ │
│       file:doc-456"                 │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 2. Get Resource pattern from policy │
│    → "api:documents:owner:          │
│       ${request:UserId}/*"          │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 3. Substitute Variables             │
│    → "api:documents:owner:          │
│       user-123/*"                   │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 4. Detect Hierarchical (contains /) │
│    → YES, use matchHierarchical     │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 5. Split by '/'                     │
│    Pattern: [api:documents:owner:   │
│              user-123, *]           │
│    Resource: [api:documents:owner:  │
│               user-123, file:doc-456]│
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 6. Match each level:                │
│    [0] "api:documents:owner:        │
│         user-123" == "..."          │
│        → ✓ MATCH                    │
│    [1] "*" matches "file:doc-456"   │
│        → ✓ MATCH                    │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 7. Check NotResource                │
│    → No NotResource specified       │
│    → ✓ NOT EXCLUDED                 │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 8. Result: TRUE                     │
│    → Resource is matched            │
└─────────────────────────────────────┘
```

### 4. Validation Rules

#### a) Resource Format Validation
```go
// File: evaluator/matching.go:173-221
func (rm *ResourceMatcher) validateResourceFormat(resource string) bool {
    if resource == "*" {
        return true
    }

    // Skip validation if contains variables
    if rm.hasVariables(resource) {
        return true
    }

    // Handle hierarchical resources
    if strings.Contains(resource, "/") {
        parts := strings.Split(resource, "/")
        for _, part := range parts {
            if !rm.validateSimpleResourceFormat(part) {
                return false
            }
        }
        return true
    }

    return rm.validateSimpleResourceFormat(resource)
}

func (rm *ResourceMatcher) validateSimpleResourceFormat(resource string) bool {
    parts := strings.Split(resource, ":")

    // Must have at least 3 parts: service:type:id
    if len(parts) < 3 {
        return false
    }

    // No empty segments
    for _, part := range parts {
        if part == "" {
            return false
        }
    }

    return true
}
```

**Quy tắc:**
- Phải có **ít nhất 3 parts** khi split bởi `:`
- Không có **empty segments**
- Cho phép **variables** (`${...}`)
- Cho phép **wildcards** (`*`)

---

## Ví dụ Chi tiết

### 1. Simple Resource - Exact Match

#### Policy:
```json
{
  "id": "pol-001",
  "policy_name": "Specific Document Access",
  "statement": [
    {
      "Sid": "SpecificDoc",
      "Effect": "Allow",
      "Action": "document-service:file:read",
      "Resource": "api:documents:public:doc-123"
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| Requested Resource | Match? | Lý do |
|-------------------|--------|-------|
| `api:documents:public:doc-123` | ✓ YES | Exact match |
| `api:documents:public:doc-456` | ✗ NO | Resource ID khác |
| `api:documents:private:doc-123` | ✗ NO | Resource type khác |
| `storage:documents:public:doc-123` | ✗ NO | Service khác |

---

### 2. Wildcard Resource

#### Policy:
```json
{
  "id": "pol-002",
  "policy_name": "All Public Documents",
  "statement": [
    {
      "Sid": "AllPublicDocs",
      "Effect": "Allow",
      "Action": "document-service:file:read",
      "Resource": "api:documents:public:*"
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| Requested Resource | Match? | Lý do |
|-------------------|--------|-------|
| `api:documents:public:doc-123` | ✓ YES | `*` matches `doc-123` |
| `api:documents:public:doc-456` | ✓ YES | `*` matches `doc-456` |
| `api:documents:public:anything` | ✓ YES | `*` matches `anything` |
| `api:documents:private:doc-123` | ✗ NO | Type khác (private ≠ public) |

---

### 3. Variable Substitution - Own Resources

#### Policy:
```json
{
  "id": "pol-003",
  "policy_name": "Own Documents Access",
  "statement": [
    {
      "Sid": "OwnDocs",
      "Effect": "Allow",
      "Action": "document-service:file:*",
      "Resource": "api:documents:owner:${request:UserId}/*"
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| Request UserId | Requested Resource | Pattern After Sub | Match? |
|---------------|-------------------|-------------------|--------|
| `user-123` | `api:documents:owner:user-123/doc-1` | `api:documents:owner:user-123/*` | ✓ YES |
| `user-123` | `api:documents:owner:user-456/doc-1` | `api:documents:owner:user-123/*` | ✗ NO |
| `user-456` | `api:documents:owner:user-456/doc-2` | `api:documents:owner:user-456/*` | ✓ YES |

---

### 4. Department Resources

#### Policy:
```json
{
  "id": "pol-004",
  "policy_name": "Department Documents",
  "statement": [
    {
      "Sid": "DeptDocs",
      "Effect": "Allow",
      "Action": "document-service:file:read",
      "Resource": "api:documents:dept:${user:Department}/*"
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| User Dept | Requested Resource | Pattern After Sub | Match? |
|----------|-------------------|-------------------|--------|
| `sales` | `api:documents:dept:sales/doc-1` | `api:documents:dept:sales/*` | ✓ YES |
| `sales` | `api:documents:dept:hr/doc-1` | `api:documents:dept:sales/*` | ✗ NO |
| `hr` | `api:documents:dept:hr/report.pdf` | `api:documents:dept:hr/*` | ✓ YES |

---

### 5. Hierarchical Resources

#### Policy:
```json
{
  "id": "pol-005",
  "policy_name": "Hierarchical File Access",
  "statement": [
    {
      "Sid": "UserFiles",
      "Effect": "Allow",
      "Action": "storage:file:*",
      "Resource": "api:storage:bucket:users/folder:${request:UserId}/*"
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| UserId | Requested Resource | Pattern After Sub | Match? |
|--------|-------------------|-------------------|--------|
| `user-123` | `api:storage:bucket:users/folder:user-123/file:doc.pdf` | `api:storage:bucket:users/folder:user-123/*` | ✓ YES |
| `user-123` | `api:storage:bucket:users/folder:user-456/file:doc.pdf` | `api:storage:bucket:users/folder:user-123/*` | ✗ NO |
| `user-456` | `api:storage:bucket:shared/folder:public/file:img.png` | `api:storage:bucket:users/folder:user-456/*` | ✗ NO |

---

### 6. Multiple Resource Patterns

#### Policy:
```json
{
  "id": "pol-006",
  "policy_name": "Multiple Resources Access",
  "statement": [
    {
      "Sid": "MultipleResources",
      "Effect": "Allow",
      "Action": "document-service:file:read",
      "Resource": [
        "api:documents:owner:${request:UserId}/*",
        "api:documents:dept:${user:Department}/*",
        "api:documents:public:*"
      ]
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| UserId | Dept | Requested Resource | Matched Pattern | Match? |
|--------|------|-------------------|----------------|--------|
| user-123 | sales | `api:documents:owner:user-123/doc-1` | Pattern 1 | ✓ YES |
| user-123 | sales | `api:documents:dept:sales/report.pdf` | Pattern 2 | ✓ YES |
| user-123 | sales | `api:documents:public:announcement` | Pattern 3 | ✓ YES |
| user-123 | sales | `api:documents:owner:user-456/doc-1` | None | ✗ NO |

---

### 7. NotResource - Exclusion

#### Policy:
```json
{
  "id": "pol-007",
  "policy_name": "All Docs Except Confidential",
  "statement": [
    {
      "Sid": "ExcludeConfidential",
      "Effect": "Allow",
      "Action": "document-service:file:read",
      "Resource": "api:documents:*",
      "NotResource": "api:documents:sensitivity:confidential/*"
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| Requested Resource | Match Resource? | Match NotResource? | Final? |
|-------------------|----------------|-------------------|--------|
| `api:documents:public:doc-1` | ✓ YES | ✗ NO | ✓ YES |
| `api:documents:internal:doc-2` | ✓ YES | ✗ NO | ✓ YES |
| `api:documents:sensitivity:confidential/secret` | ✓ YES | ✓ YES | ✗ NO |

---

### 8. Complex Real-World Example

#### Scenario: Multi-tier Document Management

```json
{
  "id": "pol-008",
  "policy_name": "Complex Document Access",
  "statement": [
    {
      "Sid": "OwnDocumentsFullAccess",
      "Effect": "Allow",
      "Action": "document-service:file:*",
      "Resource": "api:documents:owner:${request:UserId}/*"
    },
    {
      "Sid": "DepartmentDocsRead",
      "Effect": "Allow",
      "Action": "document-service:file:read",
      "Resource": "api:documents:dept:${user:Department}/*",
      "NotResource": "api:documents:dept:${user:Department}/sensitivity:confidential/*"
    },
    {
      "Sid": "PublicRead",
      "Effect": "Allow",
      "Action": "document-service:file:read",
      "Resource": "api:documents:public:*"
    },
    {
      "Sid": "DenyConfidentialDelete",
      "Effect": "Deny",
      "Action": "document-service:file:delete",
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "resource:Sensitivity": "confidential"
        }
      }
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| User | Dept | Action | Resource | Resource Sensitivity | Match Statements | Decision |
|------|------|--------|----------|---------------------|------------------|----------|
| user-123 | sales | file:read | `api:documents:owner:user-123/doc-1` | public | [OwnDocumentsFullAccess] | PERMIT |
| user-123 | sales | file:write | `api:documents:owner:user-123/doc-1` | public | [OwnDocumentsFullAccess] | PERMIT |
| user-123 | sales | file:delete | `api:documents:owner:user-123/doc-1` | confidential | [OwnDocumentsFullAccess, DenyConfidentialDelete] | DENY |
| user-123 | sales | file:read | `api:documents:dept:sales/report.pdf` | public | [DepartmentDocsRead] | PERMIT |
| user-123 | sales | file:read | `api:documents:dept:sales/sensitivity:confidential/secret.pdf` | confidential | None (NotResource excludes) | DENY |
| user-123 | sales | file:read | `api:documents:public:announcement` | public | [PublicRead] | PERMIT |

---

## Best Practices

### 1. Principle of Least Privilege

#### ❌ Không nên:
```json
{
  "Resource": "*"
}
```
→ Cho phép access **TẤT CẢ** resources

#### ✅ Nên:
```json
{
  "Resource": "api:documents:owner:${request:UserId}/*"
}
```
→ Chỉ cho phép access resources **thuộc sở hữu của user**

---

### 2. Use Variables for Dynamic Access Control

#### ❌ Không linh hoạt:
```json
{
  "Resource": "api:documents:owner:user-123/*"
}
```
→ Hard-coded user ID

#### ✅ Linh hoạt:
```json
{
  "Resource": "api:documents:owner:${request:UserId}/*"
}
```
→ Dynamic, tự động áp dụng cho mỗi user

---

### 3. Use Hierarchical Resources for Complex Structures

#### ✅ Tốt:
```json
{
  "Resource": "api:storage:bucket:users/folder:${request:UserId}/*"
}
```
→ Rõ ràng cấu trúc parent/child

---

### 4. Use NotResource for Exclusions

#### ❌ Phức tạp:
```json
{
  "Resource": [
    "api:documents:type:public/*",
    "api:documents:type:internal/*",
    "api:documents:type:private/*"
  ]
}
```
→ Liệt kê tất cả types trừ confidential

#### ✅ Đơn giản:
```json
{
  "Resource": "api:documents:*",
  "NotResource": "api:documents:type:confidential/*"
}
```
→ Allow all except confidential

---

### 5. Validate Resource Format

#### ✅ Đúng format:
```json
{
  "Resource": "api:documents:owner:user-123"
}
```
→ 3+ parts: service:type:id

#### ❌ Sai format:
```json
{
  "Resource": "api:documents"
}
```
→ Chỉ có 2 parts, không hợp lệ

---

### 6. Use Descriptive Resource Patterns

#### ❌ Không rõ ràng:
```json
{
  "Resource": "api:docs:u123/*"
}
```

#### ✅ Rõ ràng:
```json
{
  "Resource": "api:documents:owner:user-123/*"
}
```

---

### 7. Combine Resource Patterns with Conditions

```json
{
  "Sid": "SensitiveDocAccess",
  "Effect": "Allow",
  "Action": "document-service:file:read",
  "Resource": "api:documents:sensitivity:confidential/*",
  "Condition": {
    "StringEquals": {
      "user:Clearance": "top-secret"
    }
  }
}
```

---

## Lưu ý và Hạn chế

### 1. Case Sensitivity

**Resource matching là CASE-SENSITIVE:**

```json
{
  "Resource": "api:Documents:owner:User-123"
}
```

#### Test:
- Requested: `api:documents:owner:user-123` → ✗ **NO MATCH**
- Requested: `api:Documents:owner:User-123` → ✓ **MATCH**

**Best Practice:** Sử dụng **lowercase** cho consistency.

---

### 2. Minimum 3 Parts Required

Resource phải có **ít nhất 3 parts** (service:type:id):

#### ❌ Không hợp lệ:
```
api:documents          (chỉ 2 parts)
api                    (chỉ 1 part)
```

#### ✓ Hợp lệ:
```
api:documents:doc-123  (3 parts)
api:documents:owner:user-123  (4 parts, OK)
```

---

### 3. Variable Substitution Failures

Nếu variable không tồn tại trong context, nó sẽ **KHÔNG được substitute**:

```json
{
  "Resource": "api:documents:owner:${request:UserId}/*"
}
```

**Context:**
```json
{
  "user:Name": "John"
  // Missing "request:UserId"
}
```

**Result:**
```
Pattern remains: "api:documents:owner:${request:UserId}/*"
→ Will not match any resource (invalid format)
```

**Best Practice:** Đảm bảo tất cả variables có trong context.

---

### 4. Hierarchical Level Count Must Match

Số levels trong pattern và resource phải **bằng nhau**:

#### ❌ Không match:
```
Pattern:  "api:storage:bucket:users/*"           (2 levels: parent / child)
Resource: "api:storage:bucket:users/folder:x/file:y"  (3 levels)
→ NO MATCH
```

#### ✓ Match:
```
Pattern:  "api:storage:bucket:users/*/*"        (3 levels)
Resource: "api:storage:bucket:users/folder:x/file:y"  (3 levels)
→ MATCH
```

---

### 5. Empty Resource Patterns

#### ❌ Không hợp lệ:
```json
{
  "Resource": ""
}
```
→ Empty resource sẽ **fail validation**

```json
{
  "Resource": []
}
```
→ Empty array → **NO MATCH**

---

### 6. NotResource Without Resource

NotResource **phải đi kèm** với Resource:

#### ❌ Không hợp lệ:
```json
{
  "NotResource": "api:documents:confidential/*"
  // Missing Resource field
}
```

#### ✓ Hợp lệ:
```json
{
  "Resource": "api:documents:*",
  "NotResource": "api:documents:confidential/*"
}
```

---

### 7. Wildcard Performance

#### a) Full Wildcard
```json
{
  "Resource": "*"
}
```
→ **NHANH NHẤT** (check ngay)

#### b) Simple Wildcard
```json
{
  "Resource": "api:documents:*"
}
```
→ **NHANH** (string comparison)

#### c) Pattern Wildcard
```json
{
  "Resource": "api:*-archive-*:owner:*-admin/*"
}
```
→ **CHẬM HƠN** (regex matching)

**Best Practice:** Dùng simple wildcards khi có thể.

---

## Validation và Error Handling

### 1. Resource Format Validation

```go
// File: evaluator/matching.go:173-221
func (rm *ResourceMatcher) validateResourceFormat(resource string) bool {
    if resource == "*" {
        return true
    }

    // Skip validation if contains variables
    if rm.hasVariables(resource) {
        return true
    }

    // Handle hierarchical
    if strings.Contains(resource, "/") {
        parts := strings.Split(resource, "/")
        for _, part := range parts {
            if !rm.validateSimpleResourceFormat(part) {
                return false
            }
        }
        return true
    }

    return rm.validateSimpleResourceFormat(resource)
}

func (rm *ResourceMatcher) validateSimpleResourceFormat(resource string) bool {
    parts := strings.Split(resource, ":")

    // Must have at least 3 parts
    if len(parts) < 3 {
        return false
    }

    // No empty segments
    for _, part := range parts {
        if part == "" {
            return false
        }
    }

    return true
}
```

#### Validation Rules:
- Phải có **ít nhất 3 parts** (service:type:id)
- **Không có empty segments**
- Cho phép **wildcards** (`*`)
- Cho phép **variables** (`${...}`)

---

### 2. Context Validation

```go
// File: evaluator/pdp.go:407-416
requestedResource, ok := context[ContextKeyRequestResourceID].(string)
if !ok {
    log.Printf("Warning: Missing or invalid resource ID in context")
    return false
}

if requestedResource == "" {
    log.Printf("Warning: Empty resource ID provided")
    return false
}
```

#### Validation Rules:
- Context phải có `request:ResourceId` key
- ResourceId **không được empty string**

---

### 3. Error Handling Flow

```
Request → Validate Request → Validate Context → Match Resources
   ↓             ↓                  ↓                 ↓
  nil?        missing?          missing key?      no match?
   |             |                  |                 |
   ↓             ↓                  ↓                 ↓
 ERROR        ERROR              false             false
```

---

## Appendix: Code References

### File Locations

| Functionality | File | Lines |
|--------------|------|-------|
| Resource Definition | `models/types.go` | 134-219 |
| Resource Matching | `evaluator/matching.go` | 64-243 |
| Resource Evaluation | `evaluator/pdp.go` | 406-452 |
| Variable Substitution | `evaluator/matching.go` | 223-243 |
| Format Validation | `evaluator/matching.go` | 173-221 |

### Key Constants

```go
// File: evaluator/pdp.go
ContextKeyRequestResourceID = "request:ResourceId"
```

### Key Functions

```go
// Match resource pattern
func (rm *ResourceMatcher) Match(pattern, resource string, context map[string]interface{}) bool

// Match simple resource (non-hierarchical)
func (rm *ResourceMatcher) matchSimple(pattern, resource string) bool

// Match hierarchical resource
func (rm *ResourceMatcher) matchHierarchical(pattern, resource string) bool

// Substitute variables in pattern
func (rm *ResourceMatcher) substituteVariables(pattern string, context map[string]interface{}) string

// Validate resource format
func (rm *ResourceMatcher) validateResourceFormat(resource string) bool

// Check if resource matches patterns
func (pdp *PolicyDecisionPoint) matchesResourcePatterns(
    resourceSpec models.JSONActionResource,
    requestedResource string,
    context map[string]interface{}
) bool

// Check if resource matches NotResource exclusions
func (pdp *PolicyDecisionPoint) matchesNotResourcePatterns(
    notResourceSpec models.JSONActionResource,
    requestedResource string,
    context map[string]interface{}
) bool
```

---

## Summary

### Resource Field Characteristics

1. **Định dạng**: String hoặc Array of Strings
2. **Simple Pattern**: `<service>:<resource-type>:<resource-id>`
3. **Hierarchical Pattern**: `<parent>/<child>/<grandchild>`
4. **Wildcard**: Hỗ trợ `*` ở bất kỳ segment nào
5. **Variables**: Hỗ trợ `${context-key}` cho dynamic patterns
6. **NotResource**: Hỗ trợ exclusion patterns
7. **Validation**: Minimum 3 parts, no empty segments

### Matching Algorithm

1. **Full Wildcard Check**: Pattern `*` → Match all
2. **Format Validation**: Validate resource structure
3. **Variable Substitution**: Replace `${...}` với context values
4. **Hierarchical Detection**: Check for `/` separator
5. **Pattern Matching**: Match pattern với resource
6. **NotResource Check**: Ensure NOT matched by exclusion patterns

### Best Practices Summary

✅ **DO:**
- Use variables for dynamic access control (`${request:UserId}`)
- Use hierarchical resources for complex structures
- Use NotResource for exclusions
- Validate resource format (minimum 3 parts)
- Use descriptive resource patterns
- Combine with Conditions for fine-grained control

❌ **DON'T:**
- Use full wildcard `*` unless absolutely necessary
- Hard-code user IDs or resource IDs
- Leave empty resource patterns
- Use NotResource without Resource
- Forget to validate variable existence in context

---

**Document Version:** 1.0
**Last Updated:** 2025-10-24
**Based on:** `evaluator/pdp.go`, `evaluator/matching.go`, `models/types.go`
