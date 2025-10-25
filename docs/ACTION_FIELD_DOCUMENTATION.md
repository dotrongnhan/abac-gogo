# Tài Liệu Chi Tiết: Field Action trong Policy JSON

## Mục lục
1. [Tổng quan](#tổng-quan)
2. [Cấu trúc và Định dạng](#cấu-trúc-và-định-dạng)
3. [Cú pháp Action Pattern](#cú-pháp-action-pattern)
4. [Wildcard Support](#wildcard-support)
5. [Logic Matching](#logic-matching)
6. [Ví dụ Chi tiết](#ví-dụ-chi-tiết)
7. [Best Practices](#best-practices)
8. [Lưu ý và Hạn chế](#lưu-ý-và-hạn-chế)
9. [Validation và Error Handling](#validation-và-error-handling)

---

## Tổng quan

Field **Action** trong Policy JSON xác định **các hành động (operations)** mà policy cho phép hoặc từ chối. Đây là một trong ba thành phần bắt buộc của một policy statement (cùng với Effect và Resource).

### Vai trò của Action
- **Xác định quyền**: Chỉ định cụ thể operation nào được cho phép/từ chối
- **Granular control**: Cho phép kiểm soát chi tiết từ cấp service đến operation cụ thể
- **Pattern matching**: Hỗ trợ wildcard để áp dụng policy cho nhiều actions

### Vị trí trong Policy
```json
{
  "Statement": [
    {
      "Sid": "ExampleStatement",
      "Effect": "Allow",
      "Action": "...",           // ← Field Action ở đây
      "Resource": "..."
    }
  ]
}
```

---

## Cấu trúc và Định dạng

### 1. Hai định dạng được hỗ trợ

Field Action có thể được định nghĩa theo **hai cách**:

#### a) **Single String** (Chuỗi đơn)
```json
{
  "Action": "document-service:file:read"
}
```

#### b) **Array of Strings** (Mảng chuỗi)
```json
{
  "Action": [
    "document-service:file:read",
    "document-service:file:list"
  ]
}
```

### 2. Implementation trong Code

Field Action được implement bởi type `JSONActionResource`:

```go
// File: models/types.go
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

Hệ thống tự động phát hiện định dạng khi parse JSON:

```go
// Thử parse array trước
var arr []string
if err := json.Unmarshal(data, &arr); err == nil {
    j.Multiple = arr
    j.IsArray = true
    return nil
}

// Nếu không được, parse single string
var str string
if err := json.Unmarshal(data, &str); err == nil {
    j.Single = str
    j.IsArray = false
    return nil
}
```

---

## Cú pháp Action Pattern

### 1. Format Chuẩn

Action pattern tuân theo format **3 phần** phân tách bởi dấu hai chấm (`:`):

```
<service>:<resource-type>:<operation>
```

**Ví dụ:**
```
document-service:file:read
payment-service:transaction:approve
user-service:profile:update
```

### 2. Chi tiết từng phần

#### a) **Service** (Phần 1)
- Tên của service/microservice
- Định danh hệ thống con
- **Quy tắc**: Không được rỗng, có thể chứa `-`

**Ví dụ:**
```
document-service
payment-service
auth-service
notification-service
```

#### b) **Resource Type** (Phần 2)
- Loại tài nguyên trong service
- Object type mà operation tác động lên
- **Quy tắc**: Không được rỗng

**Ví dụ:**
```
file
transaction
user
profile
document
```

#### c) **Operation** (Phần 3)
- Hành động cụ thể
- CRUD operations hoặc business operations
- **Quy tắc**: Không được rỗng

**Ví dụ:**
```
read
write
delete
create
update
list
approve
reject
```

### 3. Ví dụ Action Patterns hợp lệ

```json
{
  "Action": "document-service:file:read"
}
```
→ Cho phép **đọc file** trong **document service**

```json
{
  "Action": "payment-service:transaction:approve"
}
```
→ Cho phép **approve transaction** trong **payment service**

```json
{
  "Action": "user-service:profile:update"
}
```
→ Cho phép **update profile** trong **user service**

---

## Wildcard Support

### 1. Tổng quan Wildcard

Action pattern hỗ trợ **wildcard** (`*`) để match nhiều giá trị. Wildcard có thể xuất hiện ở:
- Từng phần riêng lẻ (service, resource-type, operation)
- Toàn bộ action pattern
- Kết hợp với text (prefix, suffix, middle)

### 2. Full Wildcard

#### a) Match tất cả actions
```json
{
  "Action": "*"
}
```
→ Match **BẤT KỲ** action nào

**Code logic:**
```go
// File: evaluator/matchers/matching.go:19-21
func (am *ActionMatcher) Match(pattern, action string) bool {
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

#### a) Wildcard ở phần Operation
```json
{
  "Action": "document-service:file:*"
}
```
→ Match **TẤT CẢ** operations trên file trong document-service

**Matches:**
- `document-service:file:read` ✓
- `document-service:file:write` ✓
- `document-service:file:delete` ✓
- `document-service:file:list` ✓

**Không match:**
- `document-service:folder:read` ✗ (resource-type khác)
- `payment-service:file:read` ✗ (service khác)

#### b) Wildcard ở phần Resource Type
```json
{
  "Action": "document-service:*:read"
}
```
→ Match operation **read** trên **MỌI** resource type trong document-service

**Matches:**
- `document-service:file:read` ✓
- `document-service:folder:read` ✓
- `document-service:document:read` ✓

**Không match:**
- `document-service:file:write` ✗ (operation khác)

#### c) Wildcard ở phần Service
```json
{
  "Action": "*:file:read"
}
```
→ Match operation **read** trên **file** trong **MỌI** service

**Matches:**
- `document-service:file:read` ✓
- `storage-service:file:read` ✓
- `backup-service:file:read` ✓

**Không match:**
- `document-service:folder:read` ✗ (resource-type khác)

#### d) Multiple Wildcards
```json
{
  "Action": "payment-service:*:*"
}
```
→ Match **TẤT CẢ** operations trên **TẤT CẢ** resource types trong payment-service

**Matches:**
- `payment-service:transaction:approve` ✓
- `payment-service:transaction:reject` ✓
- `payment-service:account:read` ✓
- `payment-service:report:generate` ✓

```json
{
  "Action": "*:*:delete"
}
```
→ Match operation **delete** trên mọi resource trong mọi service

### 4. Pattern Wildcard (Prefix, Suffix, Middle)

#### a) Prefix Wildcard
```json
{
  "Action": "document-service:file:read-*"
}
```
→ Match tất cả operations **bắt đầu** bằng "read-"

**Matches:**
- `document-service:file:read-public` ✓
- `document-service:file:read-private` ✓
- `document-service:file:read-shared` ✓

**Không match:**
- `document-service:file:write-public` ✗

#### b) Suffix Wildcard
```json
{
  "Action": "document-service:file:*-read"
}
```
→ Match tất cả operations **kết thúc** bằng "-read"

**Matches:**
- `document-service:file:full-read` ✓
- `document-service:file:partial-read` ✓
- `document-service:file:meta-read` ✓

#### c) Middle Wildcard
```json
{
  "Action": "document-service:*-archive:read"
}
```
→ Match read operations trên resource types **kết thúc** bằng "-archive"

**Matches:**
- `document-service:file-archive:read` ✓
- `document-service:document-archive:read` ✓

#### d) Complex Pattern
```json
{
  "Action": "*-service:file-*:read-*"
}
```
→ Pattern phức tạp kết hợp nhiều wildcards

**Matches:**
- `document-service:file-archive:read-public` ✓
- `storage-service:file-temp:read-shared` ✓

### 5. Wildcard Implementation

```go
// File: evaluator/matchers/matching.go:50-62
func (am *ActionMatcher) matchWildcard(pattern, value string) bool {
    // Chuyển wildcard pattern thành regex
    regexPattern := strings.ReplaceAll(pattern, "*", ".*")
    regexPattern = "^" + regexPattern + "$"

    regex, err := regexp.Compile(regexPattern)
    if err != nil {
        return false
    }

    return regex.MatchString(value)
}
```

**Ví dụ chuyển đổi:**
- Pattern: `read-*` → Regex: `^read-.*$`
- Pattern: `*-service` → Regex: `^.*-service$`
- Pattern: `*-middle-*` → Regex: `^.*-middle-.*$`

---

## Logic Matching

### 1. Quy trình Matching trong PDP

```go
// File: evaluator/core/pdp.go:373-402
func (pdp *PolicyDecisionPoint) isActionMatched(
    actionSpec models.JSONActionResource,
    context map[string]interface{}
) bool {
    // 1. Lấy requested action từ context
    requestedAction, ok := context[ContextKeyRequestAction].(string)
    if !ok {
        log.Printf("Warning: Missing or invalid action in context")
        return false
    }

    // 2. Validate requested action không rỗng
    if requestedAction == "" {
        log.Printf("Warning: Empty action provided")
        return false
    }

    // 3. Lấy tất cả action patterns từ policy
    actionValues := actionSpec.GetValues()
    if len(actionValues) == 0 {
        log.Printf("Warning: No action patterns specified")
        return false
    }

    // 4. Thử match với từng pattern
    for _, actionPattern := range actionValues {
        if actionPattern == "" {
            log.Printf("Warning: Empty action pattern found")
            continue
        }
        if pdp.actionMatcher.Match(actionPattern, requestedAction) {
            return true  // Match ngay khi tìm thấy pattern phù hợp
        }
    }

    return false  // Không có pattern nào match
}
```

### 2. ActionMatcher.Match Logic

```go
// File: evaluator/matchers/matching.go:16-37
func (am *ActionMatcher) Match(pattern, action string) bool {
    // Step 1: Check full wildcard
    if pattern == "*" {
        return true
    }

    // Step 2: Split pattern và action thành các parts
    patternParts := strings.Split(pattern, ":")
    actionParts := strings.Split(action, ":")

    // Step 3: Validate số lượng parts phải bằng nhau
    if len(patternParts) != len(actionParts) {
        return false
    }

    // Step 4: Match từng segment
    for i := 0; i < len(patternParts); i++ {
        if !am.matchSegment(patternParts[i], actionParts[i]) {
            return false
        }
    }

    return true
}
```

### 3. Segment Matching

```go
// File: evaluator/matchers/matching.go:39-48
func (am *ActionMatcher) matchSegment(pattern, value string) bool {
    // Case 1: Segment là wildcard "*"
    if pattern == "*" {
        return true
    }

    // Case 2: Exact match (không có wildcard)
    if !strings.Contains(pattern, "*") {
        return pattern == value
    }

    // Case 3: Pattern wildcard (prefix, suffix, middle)
    return am.matchWildcard(pattern, value)
}
```

### 4. Flow Chart

```
Request Action: "document-service:file:read"
Policy Action:  ["document-service:file:*", "payment-service:*:*"]

┌─────────────────────────────────────┐
│ 1. Get Action from context          │
│    → "document-service:file:read"   │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 2. Get Action patterns from policy  │
│    → ["document-service:file:*",    │
│       "payment-service:*:*"]        │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 3. Try match pattern #1:            │
│    "document-service:file:*"        │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 4. Split by ":"                     │
│    Pattern: [document-service,      │
│              file, *]               │
│    Action:  [document-service,      │
│              file, read]            │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 5. Match segments:                  │
│    [0] "document-service" == "..."  │
│        → ✓ MATCH                    │
│    [1] "file" == "file"             │
│        → ✓ MATCH                    │
│    [2] "*" matches "read"           │
│        → ✓ MATCH                    │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 6. Result: TRUE                     │
│    → Action is matched              │
└─────────────────────────────────────┘
```

### 5. Validation Rules

#### a) Requested Action Validation
```go
// Must be non-empty string
if requestedAction == "" {
    return false
}

// Must exist in context
if _, ok := context[ContextKeyRequestAction]; !ok {
    return false
}
```

#### b) Policy Action Pattern Validation
```go
// Must have at least one pattern
if len(actionValues) == 0 {
    return false
}

// Each pattern must be non-empty
if actionPattern == "" {
    continue  // Skip empty patterns
}
```

#### c) Structure Validation
```go
// Pattern và Action phải có cùng số parts (khi split by ":")
if len(patternParts) != len(actionParts) {
    return false
}
```

**Ví dụ không hợp lệ:**
- Pattern: `service:resource:operation:extra` (4 parts)
- Action: `service:resource:operation` (3 parts)
- → **KHÔNG MATCH** vì số parts khác nhau

---

## Ví dụ Chi tiết

### 1. Single Action - Exact Match

#### Policy:
```json
{
  "id": "pol-001",
  "policy_name": "Read Own Documents",
  "statement": [
    {
      "Sid": "ReadOwnDocs",
      "Effect": "Allow",
      "Action": "document-service:file:read",
      "Resource": "api:documents:owner:${request:UserId}/*"
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| Requested Action | Match? | Lý do |
|-----------------|--------|-------|
| `document-service:file:read` | ✓ YES | Exact match |
| `document-service:file:write` | ✗ NO | Operation khác (write ≠ read) |
| `document-service:folder:read` | ✗ NO | Resource type khác (folder ≠ file) |
| `storage-service:file:read` | ✗ NO | Service khác |

---

### 2. Multiple Actions - Array

#### Policy:
```json
{
  "id": "pol-002",
  "policy_name": "Department Document Read Access",
  "statement": [
    {
      "Sid": "DepartmentDocsRead",
      "Effect": "Allow",
      "Action": [
        "document-service:file:read",
        "document-service:file:list",
        "document-service:folder:list"
      ],
      "Resource": "api:documents:dept:${user:Department}/*"
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| Requested Action | Match? | Matched Pattern |
|-----------------|--------|----------------|
| `document-service:file:read` | ✓ YES | `document-service:file:read` |
| `document-service:file:list` | ✓ YES | `document-service:file:list` |
| `document-service:folder:list` | ✓ YES | `document-service:folder:list` |
| `document-service:file:write` | ✗ NO | Không có trong array |
| `document-service:file:delete` | ✗ NO | Không có trong array |

---

### 3. Wildcard - All Operations

#### Policy:
```json
{
  "id": "pol-003",
  "policy_name": "Own Documents Full Access",
  "statement": [
    {
      "Sid": "OwnDocumentsFullAccess",
      "Effect": "Allow",
      "Action": "document-service:file:*",
      "Resource": "api:documents:owner:${request:UserId}/*"
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| Requested Action | Match? | Lý do |
|-----------------|--------|-------|
| `document-service:file:read` | ✓ YES | `*` matches `read` |
| `document-service:file:write` | ✓ YES | `*` matches `write` |
| `document-service:file:delete` | ✓ YES | `*` matches `delete` |
| `document-service:file:list` | ✓ YES | `*` matches `list` |
| `document-service:file:share` | ✓ YES | `*` matches `share` |
| `document-service:folder:read` | ✗ NO | Resource type khác |
| `payment-service:file:read` | ✗ NO | Service khác |

---

### 4. Wildcard - All Resource Types

#### Policy:
```json
{
  "id": "pol-004",
  "policy_name": "Read All Resource Types",
  "statement": [
    {
      "Sid": "ReadAllTypes",
      "Effect": "Allow",
      "Action": "document-service:*:read",
      "Resource": "api:documents:*"
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| Requested Action | Match? | Lý do |
|-----------------|--------|-------|
| `document-service:file:read` | ✓ YES | `*` matches `file` |
| `document-service:folder:read` | ✓ YES | `*` matches `folder` |
| `document-service:document:read` | ✓ YES | `*` matches `document` |
| `document-service:archive:read` | ✓ YES | `*` matches `archive` |
| `document-service:file:write` | ✗ NO | Operation khác (write ≠ read) |
| `storage-service:file:read` | ✗ NO | Service khác |

---

### 5. Wildcard - All Services

#### Policy:
```json
{
  "id": "pol-005",
  "policy_name": "Approve Transactions Across Services",
  "statement": [
    {
      "Sid": "ApproveAllServices",
      "Effect": "Allow",
      "Action": "*:transaction:approve",
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "user:Role": "approver"
        }
      }
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| Requested Action | Match? | Lý do |
|-----------------|--------|-------|
| `payment-service:transaction:approve` | ✓ YES | `*` matches `payment-service` |
| `banking-service:transaction:approve` | ✓ YES | `*` matches `banking-service` |
| `invoice-service:transaction:approve` | ✓ YES | `*` matches `invoice-service` |
| `payment-service:transaction:reject` | ✗ NO | Operation khác |
| `payment-service:account:approve` | ✗ NO | Resource type khác |

---

### 6. Multiple Wildcards

#### Policy:
```json
{
  "id": "pol-006",
  "policy_name": "Payment Service Full Access",
  "statement": [
    {
      "Sid": "PaymentFullAccess",
      "Effect": "Allow",
      "Action": "payment-service:*:*",
      "Resource": "*"
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| Requested Action | Match? | Lý do |
|-----------------|--------|-------|
| `payment-service:transaction:approve` | ✓ YES | Both `*` match |
| `payment-service:transaction:reject` | ✓ YES | Both `*` match |
| `payment-service:account:read` | ✓ YES | Both `*` match |
| `payment-service:report:generate` | ✓ YES | Both `*` match |
| `document-service:file:read` | ✗ NO | Service khác |

---

### 7. Pattern Wildcard - Prefix

#### Policy:
```json
{
  "id": "pol-007",
  "policy_name": "All Read Operations",
  "statement": [
    {
      "Sid": "AllReadOps",
      "Effect": "Allow",
      "Action": "document-service:file:read-*",
      "Resource": "api:documents:*"
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| Requested Action | Match? | Lý do |
|-----------------|--------|-------|
| `document-service:file:read-public` | ✓ YES | Pattern matches `read-*` |
| `document-service:file:read-private` | ✓ YES | Pattern matches `read-*` |
| `document-service:file:read-shared` | ✓ YES | Pattern matches `read-*` |
| `document-service:file:read` | ✗ NO | Exact `read` không match `read-*` |
| `document-service:file:write-public` | ✗ NO | Prefix khác |

---

### 8. Pattern Wildcard - Suffix

#### Policy:
```json
{
  "id": "pol-008",
  "policy_name": "All Archive Operations",
  "statement": [
    {
      "Sid": "ArchiveOps",
      "Effect": "Allow",
      "Action": "document-service:*-archive:*",
      "Resource": "api:archives:*"
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| Requested Action | Match? | Lý do |
|-----------------|--------|-------|
| `document-service:file-archive:read` | ✓ YES | Matches `*-archive:*` |
| `document-service:document-archive:write` | ✓ YES | Matches `*-archive:*` |
| `document-service:temp-archive:delete` | ✓ YES | Matches `*-archive:*` |
| `document-service:file:read` | ✗ NO | Resource type không end với `-archive` |
| `document-service:archive:read` | ✗ NO | Exact `archive` không match `*-archive` |

---

### 9. Full Wildcard (Admin Access)

#### Policy:
```json
{
  "id": "pol-009",
  "policy_name": "Administrator Full Access",
  "statement": [
    {
      "Sid": "AdminAccess",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "user:Role": "admin"
        }
      }
    }
  ],
  "enabled": true
}
```

#### Test Cases:

| Requested Action | Match? | Lý do |
|-----------------|--------|-------|
| `document-service:file:read` | ✓ YES | `*` matches all |
| `payment-service:transaction:approve` | ✓ YES | `*` matches all |
| `user-service:profile:update` | ✓ YES | `*` matches all |
| `ANY:THING:HERE` | ✓ YES | `*` matches all |

---

### 10. Deny Statement với Action

#### Policy:
```json
{
  "id": "pol-010",
  "policy_name": "Deny Confidential Delete",
  "statement": [
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

| Requested Action | Resource Sensitivity | Match? | Decision |
|-----------------|---------------------|--------|----------|
| `document-service:file:delete` | `confidential` | ✓ YES | **DENY** |
| `document-service:file:delete` | `public` | ✓ YES | No match (condition fails) |
| `document-service:file:read` | `confidential` | ✗ NO | No match (action khác) |

---

### 11. Combining Multiple Actions with Deny Override

#### Policy Set:
```json
{
  "policies": [
    {
      "id": "pol-011-allow",
      "policy_name": "Allow Payment Operations",
      "statement": [
        {
          "Sid": "AllowPaymentOps",
          "Effect": "Allow",
          "Action": "payment-service:transaction:*",
          "Resource": "api:transactions:*"
        }
      ],
      "enabled": true
    },
    {
      "id": "pol-011-deny",
      "policy_name": "Deny Weekend Transactions",
      "statement": [
        {
          "Sid": "DenyWeekendTransactions",
          "Effect": "Deny",
          "Action": "payment-service:transaction:*",
          "Resource": "*",
          "Condition": {
            "StringEquals": {
              "request:DayOfWeek": ["Saturday", "Sunday"]
            }
          }
        }
      ],
      "enabled": true
    }
  ]
}
```

#### Test Cases:

| Requested Action | Day | Allow Policy Match | Deny Policy Match | Final Decision |
|-----------------|-----|-------------------|------------------|----------------|
| `payment-service:transaction:approve` | Monday | ✓ YES | ✗ NO | **PERMIT** |
| `payment-service:transaction:approve` | Saturday | ✓ YES | ✓ YES | **DENY** (Deny Override) |
| `payment-service:transaction:create` | Sunday | ✓ YES | ✓ YES | **DENY** (Deny Override) |

**Code Logic:**
```go
// File: evaluator/core/pdp.go:288-294
// Deny-Override algorithm
if strings.ToLower(statement.Effect) == EffectDeny {
    return &models.Decision{
        Result: ResultDeny,
        Reason: fmt.Sprintf(ReasonDeniedByStatement, statement.Sid),
    }
}
```

---

### 12. Complex Real-World Example

#### Scenario: Multi-tier Document Access Control

```json
{
  "id": "pol-012",
  "policy_name": "Complex Document Access Control",
  "statement": [
    {
      "Sid": "OwnDocumentsFullAccess",
      "Effect": "Allow",
      "Action": "document-service:file:*",
      "Resource": "api:documents:owner:${request:UserId}/*"
    },
    {
      "Sid": "DepartmentDocumentsRead",
      "Effect": "Allow",
      "Action": [
        "document-service:file:read",
        "document-service:file:list"
      ],
      "Resource": "api:documents:dept:${user:Department}/*",
      "Condition": {
        "StringNotEquals": {
          "resource:Sensitivity": "confidential"
        }
      }
    },
    {
      "Sid": "ManagerApprovalAccess",
      "Effect": "Allow",
      "Action": [
        "document-service:file:approve",
        "document-service:file:reject"
      ],
      "Resource": "api:documents:dept:${user:Department}/*",
      "Condition": {
        "StringEquals": {
          "user:Role": "manager"
        }
      }
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

| User | Action | Resource | Resource Owner | Sensitivity | Match Statements | Decision |
|------|--------|----------|---------------|-------------|------------------|----------|
| user-123 | `document-service:file:read` | `api:documents:owner:user-123/doc1` | user-123 | public | [Sid: OwnDocumentsFullAccess] | **PERMIT** |
| user-123 | `document-service:file:write` | `api:documents:owner:user-123/doc1` | user-123 | public | [Sid: OwnDocumentsFullAccess] | **PERMIT** |
| user-123 | `document-service:file:delete` | `api:documents:owner:user-123/doc1` | user-123 | confidential | [Sid: OwnDocumentsFullAccess, DenyConfidentialDelete] | **DENY** |
| user-123 (dept: sales) | `document-service:file:read` | `api:documents:dept:sales/doc2` | user-456 | public | [Sid: DepartmentDocumentsRead] | **PERMIT** |
| user-123 (dept: sales) | `document-service:file:read` | `api:documents:dept:sales/doc3` | user-456 | confidential | None (condition fails) | **DENY** (implicit) |
| user-123 (role: manager, dept: sales) | `document-service:file:approve` | `api:documents:dept:sales/doc4` | user-456 | public | [Sid: ManagerApprovalAccess] | **PERMIT** |

---

## Best Practices

### 1. Principle of Least Privilege

#### ❌ Không nên:
```json
{
  "Action": "*",
  "Resource": "*"
}
```
→ Cho phép **TẤT CẢ** actions trên **TẤT CẢ** resources

#### ✅ Nên:
```json
{
  "Action": [
    "document-service:file:read",
    "document-service:file:list"
  ],
  "Resource": "api:documents:dept:${user:Department}/*"
}
```
→ Chỉ cho phép operations **cần thiết** trên resources **cụ thể**

---

### 2. Use Specific Actions When Possible

#### ❌ Không nên (quá rộng):
```json
{
  "Action": "payment-service:*:*"
}
```

#### ✅ Nên (cụ thể):
```json
{
  "Action": [
    "payment-service:transaction:read",
    "payment-service:transaction:list"
  ]
}
```

---

### 3. Group Related Actions

#### ✅ Tốt:
```json
{
  "Sid": "ReadOperations",
  "Action": [
    "document-service:file:read",
    "document-service:file:list",
    "document-service:folder:list"
  ]
}
```

#### ✅ Tốt hơn (nếu logic cho phép):
```json
{
  "Sid": "ReadAllResources",
  "Action": "document-service:*:read"
}
```

---

### 4. Use Descriptive Sid

#### ❌ Không rõ ràng:
```json
{
  "Sid": "Statement1",
  "Action": "document-service:file:delete"
}
```

#### ✅ Rõ ràng:
```json
{
  "Sid": "DenyConfidentialFileDelete",
  "Effect": "Deny",
  "Action": "document-service:file:delete",
  "Condition": {
    "StringEquals": {
      "resource:Sensitivity": "confidential"
    }
  }
}
```

---

### 5. Combine with Conditions for Fine-Grained Control

```json
{
  "Sid": "ApproveSmallTransactions",
  "Effect": "Allow",
  "Action": "payment-service:transaction:approve",
  "Resource": "api:transactions:*",
  "Condition": {
    "NumericLessThan": {
      "transaction:Amount": 1000000
    }
  }
}
```

---

### 6. Use Deny Statements Sparingly

#### Quy tắc:
- **Allow by default**: Không, hệ thống là **Deny by default**
- **Explicit Allow**: Chỉ cho phép khi có Allow statement
- **Deny Override**: Deny luôn thắng Allow

#### ✅ Sử dụng Deny khi:
```json
{
  "Sid": "DenyExternalAPIAccess",
  "Effect": "Deny",
  "Action": "*:*:*",
  "Resource": "*",
  "Condition": {
    "Bool": {
      "request:IsExternal": true
    }
  }
}
```
→ Đảm bảo không ai có thể access từ external, dù có Allow statement nào

---

### 7. Document Complex Action Patterns

#### ✅ Tốt:
```json
{
  "Sid": "AllArchiveOperations",
  "Description": "Allows all operations on archive resource types (file-archive, doc-archive, etc.)",
  "Effect": "Allow",
  "Action": "document-service:*-archive:*",
  "Resource": "api:archives:*"
}
```

---

### 8. Test Action Patterns Thoroughly

#### Test Matrix Template:

| Test Case | Action Pattern | Requested Action | Expected | Actual | Status |
|-----------|---------------|------------------|----------|--------|--------|
| TC-001 | `service:file:read` | `service:file:read` | MATCH | MATCH | ✓ |
| TC-002 | `service:file:*` | `service:file:write` | MATCH | MATCH | ✓ |
| TC-003 | `service:*:read` | `service:folder:read` | MATCH | MATCH | ✓ |
| TC-004 | `*:file:read` | `other-service:file:read` | MATCH | MATCH | ✓ |

---

## Lưu ý và Hạn chế

### 1. Case Sensitivity

**Action matching là CASE-SENSITIVE:**

```json
{
  "Action": "document-service:file:Read"
}
```

#### Test:
- Requested: `document-service:file:read` → ✗ **NO MATCH** (Read ≠ read)
- Requested: `document-service:file:Read` → ✓ **MATCH**

**Best Practice:** Sử dụng **lowercase** cho tất cả actions để tránh nhầm lẫn.

---

### 2. Exact Part Count Required

Pattern và Action phải có **cùng số parts** (phân tách bởi `:`):

#### ❌ Không match:
```
Pattern: "service:resource:operation:extra"  (4 parts)
Action:  "service:resource:operation"        (3 parts)
→ NO MATCH
```

#### ✓ Match:
```
Pattern: "service:resource:operation"        (3 parts)
Action:  "service:resource:operation"        (3 parts)
→ MATCH
```

---

### 3. Empty Action Patterns

#### ❌ Không hợp lệ:
```json
{
  "Action": ""
}
```
→ Empty action sẽ bị **skip**

```json
{
  "Action": []
}
```
→ Empty array → **NO MATCH**

#### Code:
```go
// File: evaluator/pdp.go:387-391
actionValues := actionSpec.GetValues()
if len(actionValues) == 0 {
    log.Printf("Warning: No action patterns specified")
    return false
}
```

---

### 4. Wildcard Không Hỗ trợ Regex

Wildcard chỉ hỗ trợ `*`, **KHÔNG** hỗ trợ regex phức tạp:

#### ✅ Hợp lệ:
```
"read-*"
"*-service"
"*-middle-*"
```

#### ❌ Không hỗ trợ:
```
"read-[0-9]+"         # Regex character class
"file-(read|write)"   # Regex alternation
"doc.+"               # Regex quantifier
```

**Lý do:** Wildcard `*` được convert thành regex `.*`, không parse regex phức tạp.

---

### 5. Deny Override Luôn Thắng

#### Quy tắc:
1. Evaluate tất cả policies
2. Nếu **BẤT KỲ** statement nào có `Effect: Deny` và match → **DENY**
3. Nếu có ít nhất một statement có `Effect: Allow` và match → **PERMIT**
4. Nếu không có statement nào match → **DENY** (implicit deny)

#### Ví dụ:
```json
{
  "policies": [
    {
      "statement": [
        {
          "Sid": "AllowRead",
          "Effect": "Allow",
          "Action": "document-service:file:read",
          "Resource": "*"
        }
      ]
    },
    {
      "statement": [
        {
          "Sid": "DenyAll",
          "Effect": "Deny",
          "Action": "*",
          "Resource": "*"
        }
      ]
    }
  ]
}
```

**Result:** `document-service:file:read` → **DENY** (vì DenyAll thắng)

---

### 6. Performance Considerations

#### a) Full Wildcard
```json
{
  "Action": "*"
}
```
→ **NHANH NHẤT** (check ngay, không cần split)

#### b) Exact Match
```json
{
  "Action": "service:resource:operation"
}
```
→ **NHANH** (string comparison cho từng part)

#### c) Pattern Wildcard
```json
{
  "Action": "service:*-archive-*:operation-*"
}
```
→ **CHẬM HỔN** (compile regex, match từng segment)

**Best Practice:** Sử dụng exact match hoặc simple wildcard khi có thể.

---

### 7. Multiple Action Patterns Evaluation

Khi có nhiều patterns trong array, hệ thống **match lần lượt** cho đến khi tìm thấy match:

```go
// File: evaluator/core/pdp.go:392-400
for _, actionPattern := range actionValues {
    if actionPattern == "" {
        continue
    }
    if pdp.actionMatcher.Match(actionPattern, requestedAction) {
        return true  // Return ngay khi match
    }
}
return false
```

#### Impact:
- Pattern đầu tiên match → return ngay, không check patterns sau
- **Optimization:** Đặt patterns phổ biến nhất lên đầu

#### ❌ Không tối ưu:
```json
{
  "Action": [
    "rare-service:rare:operation",
    "common-service:file:read",      // ← Phổ biến nhưng ở sau
    "common-service:file:write"      // ← Phổ biến nhưng ở sau
  ]
}
```

#### ✅ Tối ưu:
```json
{
  "Action": [
    "common-service:file:read",      // ← Phổ biến, đặt trước
    "common-service:file:write",     // ← Phổ biến, đặt trước
    "rare-service:rare:operation"
  ]
}
```

---

## Validation và Error Handling

### 1. Request Validation

```go
// File: evaluator/core/pdp.go:99-104
if request == nil {
    return nil, fmt.Errorf("evaluation request cannot be nil")
}
if request.SubjectID == "" || request.ResourceID == "" || request.Action == "" {
    return nil, fmt.Errorf("invalid request: missing required fields")
}
```

#### Validation Rules:
- Request **không được nil**
- `Action` **không được empty string**
- `SubjectID` **không được empty**
- `ResourceID` **không được empty**

---

### 2. Context Validation

```go
// File: evaluator/core/pdp.go:341-371
func (pdp *PolicyDecisionPoint) isValidEvaluationContext(
    context map[string]interface{}
) bool {
    if context == nil {
        return false
    }

    // Check essential keys
    essentialKeys := []string{
        ContextKeyRequestAction,
        ContextKeyRequestResourceID,
    }

    for _, key := range essentialKeys {
        if _, exists := context[key]; !exists {
            log.Printf("Warning: Missing essential key: %s", key)
            return false
        }
    }

    // Validate context size
    if len(context) > MaxConditionKeys {
        log.Printf("Warning: Context size exceeds maximum")
        return false
    }

    return true
}
```

#### Validation Rules:
- Context **không được nil**
- Phải có `request:Action` key
- Phải có `request:ResourceId` key
- Context size ≤ **100 keys** (MaxConditionKeys)

---

### 3. Action Pattern Validation

```go
// File: evaluator/core/pdp.go:387-391
actionValues := actionSpec.GetValues()
if len(actionValues) == 0 {
    log.Printf("Warning: No action patterns specified")
    return false
}

for _, actionPattern := range actionValues {
    if actionPattern == "" {
        log.Printf("Warning: Empty action pattern found")
        continue  // Skip empty patterns
    }
    // ...
}
```

#### Validation Rules:
- Phải có **ít nhất 1 pattern** (array không rỗng)
- Empty patterns sẽ bị **skip**
- Logging warnings cho invalid patterns

---

### 4. Error Handling Flow

```
Request → Validate Request → Validate Context → Match Actions
   ↓             ↓                  ↓                 ↓
  nil?        missing?          missing key?      no match?
   |             |                  |                 |
   ↓             ↓                  ↓                 ↓
 ERROR        ERROR              false             false
```

#### Error Response:
```go
return nil, fmt.Errorf("evaluation request cannot be nil")
return nil, fmt.Errorf("invalid request: missing required fields")
return nil, fmt.Errorf("failed to enrich context: %w", err)
```

#### No Match Response:
```go
return &models.Decision{
    Result:          ResultDeny,
    MatchedPolicies: []string{},
    Reason:          ReasonImplicitDeny,
}
```

---

## Appendix: Code References

### File Locations

| Functionality | File | Lines |
|--------------|------|-------|
| Action Definition | `models/types.go` | 134-219 |
| Action Matching | `evaluator/matchers/matching.go` | 8-62 |
| Action Evaluation | `evaluator/core/pdp.go` | 373-402 |
| Policy Statement | `models/types.go` | 298-306 |
| Evaluation Request | `models/types.go` | 315-324 |

### Key Constants

```go
// File: evaluator/core/pdp.go:45-46
ContextKeyRequestAction = "request:Action"
```

### Key Functions

```go
// Get action patterns from JSONActionResource
func (j JSONActionResource) GetValues() []string

// Match action pattern
func (am *ActionMatcher) Match(pattern, action string) bool

// Match segment with wildcard
func (am *ActionMatcher) matchSegment(pattern, value string) bool

// Convert wildcard to regex
func (am *ActionMatcher) matchWildcard(pattern, value string) bool

// Check if action is matched in policy evaluation
func (pdp *PolicyDecisionPoint) isActionMatched(
    actionSpec models.JSONActionResource,
    context map[string]interface{}
) bool
```

---

## Summary

### Action Field Characteristics

1. **Định dạng**: String hoặc Array of Strings
2. **Pattern**: `<service>:<resource-type>:<operation>`
3. **Wildcard**: Hỗ trợ `*` ở bất kỳ segment nào
4. **Matching**: Case-sensitive, exact part count
5. **Performance**: Full wildcard > Exact match > Pattern wildcard
6. **Validation**: Non-empty, proper structure, context keys required

### Decision Algorithm

1. **Input Validation**: Request và Context hợp lệ
2. **Pattern Matching**: Thử match với từng pattern
3. **First Match Wins**: Return ngay khi tìm thấy match
4. **Deny Override**: Deny statement luôn thắng Allow
5. **Implicit Deny**: Mặc định deny nếu không có match

### Best Practices Summary

✅ **DO:**
- Sử dụng specific actions thay vì wildcards
- Group related actions
- Use descriptive Sid
- Combine với Conditions
- Test thoroughly
- Document complex patterns

❌ **DON'T:**
- Dùng full wildcard `*` trừ khi thực sự cần
- Mix uppercase/lowercase
- Leave empty patterns
- Forget deny override rules

---

**Document Version:** 1.1
**Last Updated:** 2025-10-25
**Based on:** `evaluator/core/pdp.go`, `evaluator/matchers/matching.go`, `models/types.go`
**Updated:** Cập nhật theo cấu trúc package mới và enhanced evaluator
