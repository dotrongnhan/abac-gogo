# API Documentation - ABAC HTTP Service

Tài liệu chi tiết về các API endpoints của ABAC HTTP Service.

## 🌐 Base URL

```
http://localhost:8081
```

## 🔐 Authentication

Tất cả protected endpoints yêu cầu header:

```
X-Subject-ID: <subject_id>
```

**Ví dụ:**
```bash
curl -H 'X-Subject-ID: sub-001' http://localhost:8081/api/v1/users
```

## 📋 Endpoints

### 1. Health Check

Kiểm tra trạng thái service.

**Endpoint:** `GET /health`  
**Authentication:** Không cần  
**ABAC:** Không áp dụng  

#### Request
```bash
curl http://localhost:8081/health
```

#### Response
```json
{
  "service": "ABAC Authorization Service",
  "status": "healthy", 
  "timestamp": "2025-10-11T19:52:54+07:00"
}
```

---

### 2. List Users

Lấy danh sách users trong hệ thống.

**Endpoint:** `GET /api/v1/users`  
**Authentication:** Required (`X-Subject-ID`)  
**ABAC Permission:** `read`  
**Resource:** `/api/v1/users`

#### Request
```bash
curl -H 'X-Subject-ID: sub-001' http://localhost:8081/api/v1/users
```

#### Success Response (200)
```json
{
  "message": "Users retrieved successfully",
  "users": [
    {
      "id": "1",
      "name": "John Doe", 
      "department": "Engineering"
    },
    {
      "id": "2",
      "name": "Alice Smith",
      "department": "Finance"
    },
    {
      "id": "3", 
      "name": "Bob Wilson",
      "department": "Engineering"
    }
  ]
}
```

#### Access Denied Response (403)
```json
{
  "error": "Access denied",
  "reason": "No applicable policies found",
  "subject": "sub-004",
  "resource": "/api/v1/users", 
  "action": "read"
}
```

#### ABAC Rules
- ✅ **Engineering users** có thể đọc users API
- ❌ **Probation users** bị từ chối
- ❌ **Finance users** không có quyền truy cập users API

---

### 3. Create User

Tạo user mới trong hệ thống.

**Endpoint:** `POST /api/v1/users/create`  
**Authentication:** Required (`X-Subject-ID`)  
**ABAC Permission:** `write`  
**Resource:** `/api/v1/users/create`

#### Request
```bash
curl -X POST \
  -H 'X-Subject-ID: sub-001' \
  -H 'Content-Type: application/json' \
  -d '{"name": "New User", "department": "Engineering"}' \
  http://localhost:8081/api/v1/users/create
```

#### Success Response (200)
```json
{
  "message": "User created successfully",
  "user_id": "new_user_123"
}
```

#### Access Denied Response (403)
```json
{
  "error": "Access denied",
  "reason": "Denied by policy: Probation Write Restriction",
  "subject": "sub-004",
  "resource": "/api/v1/users/create",
  "action": "write"
}
```

#### ABAC Rules
- ✅ **Senior developers** có thể tạo users
- ❌ **Probation users** không thể write
- ❌ **Read-only users** không có write permission

---

### 4. Financial Data

Truy cập dữ liệu tài chính.

**Endpoint:** `GET /api/v1/financial`  
**Authentication:** Required (`X-Subject-ID`)  
**ABAC Permission:** `read`  
**Resource:** `/api/v1/financial`

#### Request
```bash
curl -H 'X-Subject-ID: sub-002' http://localhost:8081/api/v1/financial
```

#### Success Response (200)
```json
{
  "message": "Financial data retrieved successfully",
  "financial_data": {
    "revenue": "$1,000,000",
    "expenses": "$800,000", 
    "profit": "$200,000",
    "quarter": "Q1 2024"
  }
}
```

#### Access Denied Response (403)
```json
{
  "error": "Access denied",
  "reason": "Insufficient clearance level",
  "subject": "sub-001",
  "resource": "/api/v1/financial",
  "action": "read"
}
```

#### ABAC Rules
- ✅ **Finance department** có thể đọc financial data
- ✅ **High clearance users** (level 3+) có thể truy cập
- ❌ **Engineering users** không có quyền truy cập financial data
- ❌ **Low clearance users** bị từ chối

---

### 5. Admin Panel

Truy cập admin panel với quyền quản trị.

**Endpoint:** `GET /api/v1/admin`  
**Authentication:** Required (`X-Subject-ID`)  
**ABAC Permission:** `admin`  
**Resource:** `/api/v1/admin`

#### Request
```bash
curl -H 'X-Subject-ID: sub-001' http://localhost:8081/api/v1/admin
```

#### Success Response (200)
```json
{
  "message": "Admin panel accessed",
  "admin_functions": [
    "user_management",
    "system_config", 
    "audit_logs"
  ]
}
```

#### Access Denied Response (403)
```json
{
  "error": "Access denied",
  "reason": "Admin privileges required",
  "subject": "sub-004",
  "resource": "/api/v1/admin",
  "action": "admin"
}
```

#### ABAC Rules
- ✅ **Admin users** có thể truy cập admin panel
- ✅ **Manager role** có admin privileges
- ❌ **Regular users** không có admin quyền
- ❌ **Service accounts** không thể truy cập admin functions

---

## 🚨 Error Responses

### 401 Unauthorized
Thiếu authentication header.

```json
{
  "error": "Missing X-Subject-ID header"
}
```

### 403 Forbidden  
ABAC từ chối quyền truy cập.

```json
{
  "error": "Access denied",
  "reason": "<specific_reason>",
  "subject": "<subject_id>",
  "resource": "<resource_path>", 
  "action": "<required_action>"
}
```

### 500 Internal Server Error
Lỗi server hoặc ABAC evaluation.

```json
{
  "error": "Authorization error"
}
```

## 🔍 Debug Endpoint

### Debug Routes

Xem danh sách routes đã đăng ký.

**Endpoint:** `GET /debug/routes`  
**Authentication:** Không cần

#### Request
```bash
curl http://localhost:8081/debug/routes
```

#### Response
```json
{
  "routes": [
    "/health",
    "/api/v1/users", 
    "/api/v1/users/create",
    "/api/v1/financial",
    "/api/v1/admin"
  ]
}
```

## 📊 ABAC Decision Flow

Mỗi protected request đi qua luồng ABAC:

1. **Extract Subject** - Lấy `X-Subject-ID` từ header
2. **Identify Resource** - Resource = URL path  
3. **Determine Action** - Action = required permission
4. **Enrich Context** - PIP lấy attributes
5. **Evaluate Policies** - PDP check rules
6. **Make Decision** - PERMIT/DENY/NOT_APPLICABLE
7. **Enforce** - Allow request hoặc return 403

## 🧪 Test Scenarios

### Scenario 1: Engineering User - Success
```bash
# John Doe (Engineering) đọc users API
curl -H 'X-Subject-ID: sub-001' http://localhost:8081/api/v1/users
# Expected: 200 OK với danh sách users
```

### Scenario 2: Finance User - Financial Data
```bash  
# Alice Smith (Finance) đọc financial data
curl -H 'X-Subject-ID: sub-002' http://localhost:8081/api/v1/financial
# Expected: 200 OK với financial data
```

### Scenario 3: Probation User - Denied
```bash
# Bob Wilson (Probation) cố gắng write
curl -X POST -H 'X-Subject-ID: sub-004' http://localhost:8081/api/v1/users/create  
# Expected: 403 Forbidden
```

### Scenario 4: Cross-Department Access - Denied
```bash
# Engineering user cố truy cập financial data
curl -H 'X-Subject-ID: sub-001' http://localhost:8081/api/v1/financial
# Expected: 403 Forbidden  
```

### Scenario 5: Missing Auth - Unauthorized
```bash
# Request không có X-Subject-ID header
curl http://localhost:8081/api/v1/users
# Expected: 401 Unauthorized
```

## 🔧 Integration Examples

### JavaScript/Fetch
```javascript
const response = await fetch('http://localhost:8081/api/v1/users', {
  headers: {
    'X-Subject-ID': 'sub-001',
    'Content-Type': 'application/json'
  }
});

const data = await response.json();
```

### Python/Requests
```python
import requests

headers = {'X-Subject-ID': 'sub-001'}
response = requests.get('http://localhost:8081/api/v1/users', headers=headers)
data = response.json()
```

### cURL Scripts
```bash
#!/bin/bash
SUBJECT_ID="sub-001"
BASE_URL="http://localhost:8081"

# Test all endpoints
curl -H "X-Subject-ID: $SUBJECT_ID" "$BASE_URL/api/v1/users"
curl -H "X-Subject-ID: $SUBJECT_ID" "$BASE_URL/api/v1/financial" 
curl -H "X-Subject-ID: $SUBJECT_ID" "$BASE_URL/api/v1/admin"
```
