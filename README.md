# ABAC HTTP Service - Đơn Giản & Hiệu Quả

Hệ thống **Attribute-Based Access Control (ABAC)** được triển khai dưới dạng HTTP service đơn giản, dễ sử dụng và tích hợp.

## 🚀 Khởi Chạy Nhanh

```bash
# Clone repository
git clone <repository-url>
cd ABAC-gogo-example

# Chạy service
go run main.go

# Service sẽ chạy trên http://localhost:8081
```

## 📋 API Endpoints

| Method | Endpoint | Description | Required Permission |
|--------|----------|-------------|-------------------|
| `GET` | `/health` | Health check | None (public) |
| `GET` | `/api/v1/users` | Danh sách users | `read` |
| `POST` | `/api/v1/users/create` | Tạo user mới | `write` |
| `GET` | `/api/v1/financial` | Dữ liệu tài chính | `read` |
| `GET` | `/api/v1/admin` | Admin panel | `admin` |

## 🔑 Authentication

Sử dụng header `X-Subject-ID` để xác định user:

   ```bash
curl -H 'X-Subject-ID: sub-001' http://localhost:8081/api/v1/users
```

### Test Users

| Subject ID | Name | Department | Permissions |
|------------|------|------------|-------------|
| `sub-001` | John Doe | Engineering | Read APIs |
| `sub-002` | Alice Smith | Finance | Read financial data |
| `sub-003` | Payment Service | System | Service account |
| `sub-004` | Bob Wilson | Engineering | Limited (probation) |

## 💡 Ví Dụ Sử Dụng

### 1. Health Check (Không cần auth)
   ```bash
curl http://localhost:8081/health
   ```

### 2. Engineering User Truy Cập Users API
   ```bash
curl -H 'X-Subject-ID: sub-001' http://localhost:8081/api/v1/users
```
**Kết quả:** ✅ PERMIT - Engineering user có quyền đọc API

### 3. Finance User Truy Cập Financial Data
   ```bash
curl -H 'X-Subject-ID: sub-002' http://localhost:8081/api/v1/financial
```
**Kết quả:** ❌ DENY/NOT_APPLICABLE - Tùy thuộc vào policy

### 4. Probation User Cố Gắng Truy Cập
   ```bash
curl -H 'X-Subject-ID: sub-004' http://localhost:8081/api/v1/users
   ```
**Kết quả:** ❌ DENY - User đang bị hạn chế

### 5. Missing Authentication
```bash
curl http://localhost:8081/api/v1/users
```
**Kết quả:** ❌ 401 Unauthorized - Missing X-Subject-ID header

## 🏗️ Kiến Trúc ABAC

### Luồng Hoạt Động Đơn Giản

```
1. User Request → 2. PEP Intercept → 3. PDP Evaluate → 4. PIP Get Attributes → 5. PAP Check Policies → 6. Decision → 7. Enforce
```

### Chi Tiết Từng Bước

1. **User gửi HTTP request** với header `X-Subject-ID`
2. **PEP (Policy Enforcement Point)** - ABAC Middleware chặn request
3. **PDP (Policy Decision Point)** - Đánh giá quyền truy cập
4. **PIP (Policy Information Point)** - Lấy attributes của user/resource
5. **PAP (Policy Administration Point)** - Kiểm tra policies
6. **Decision** - PERMIT/DENY/NOT_APPLICABLE
7. **Enforce** - Cho phép hoặc từ chối request

### Components

- **`main.go`** - HTTP server với ABAC middleware
- **`evaluator/`** - PDP implementation
- **`attributes/`** - PIP implementation  
- **`storage/`** - PAP data access
- **`models/`** - Data structures
- **`*.json`** - Test data (subjects, resources, actions, policies)

## 📊 Response Format

### Successful Response
```json
{
  "message": "Users retrieved successfully",
  "users": [...]
}
```

### Access Denied Response
```json
{
  "error": "Access denied",
  "reason": "No applicable policies found",
  "subject": "sub-004",
  "resource": "/api/v1/users",
  "action": "read"
}
```

### Error Response
```json
{
  "error": "Missing X-Subject-ID header"
}
```

## 🔧 Configuration

Service sử dụng JSON files để cấu hình:

- **`subjects.json`** - Danh sách users và attributes
- **`resources.json`** - Danh sách resources và properties
- **`actions.json`** - Các actions có thể thực hiện
- **`policies.json`** - ABAC policies và rules

## 🚦 ABAC Decision Logic

1. **DENY có ưu tiên cao nhất** - Nếu có policy DENY match → từ chối ngay
2. **PERMIT cần có policy match** - Phải có ít nhất 1 policy PERMIT
3. **NOT_APPLICABLE** - Không có policy nào áp dụng

## 🛠️ Development

### Thêm Endpoint Mới

1. Thêm handler function
2. Đăng ký route với ABAC middleware
3. Thêm resource vào `resources.json`
4. Tạo policy trong `policies.json`

### Thêm User Mới

1. Thêm subject vào `subjects.json`
2. Định nghĩa attributes (department, role, clearance_level, etc.)

### Thêm Policy Mới

1. Thêm policy vào `policies.json`
2. Định nghĩa rules với target_type, attribute_path, operator, expected_value

## 📝 Logs

Service ghi log các ABAC decisions:

```
2025/10/11 19:52:54 ABAC Decision: permit - Subject: sub-001, Resource: /api/v1/users, Action: read, Reason: Access granted by matching permit policies
```

## 🔍 Troubleshooting

### Common Issues

1. **"resource not found"** - Kiểm tra `resources.json` có resource với đúng `resource_id`
2. **"subject not found"** - Kiểm tra `subjects.json` có subject với đúng ID
3. **"Authorization error"** - Kiểm tra policies và rules trong `policies.json`
4. **"Missing X-Subject-ID header"** - Thêm header vào request

### Debug

Sử dụng endpoint `/debug/routes` để xem các routes đã đăng ký:

```bash
curl http://localhost:8081/debug/routes
```

## 🎯 Production Considerations

1. **Authentication** - Thay thế `X-Subject-ID` header bằng JWT token
2. **Database** - Chuyển từ JSON files sang PostgreSQL
3. **Caching** - Thêm Redis cache cho decisions
4. **Monitoring** - Thêm metrics và alerting
5. **Rate Limiting** - Implement rate limiting per user
6. **HTTPS** - Sử dụng TLS trong production

## 📚 Tài Liệu Khác

- [`code_architecture.md`](code_architecture.md) - Chi tiết kiến trúc
- [`ABAC_SYSTEM_DOCUMENTATION.md`](ABAC_SYSTEM_DOCUMENTATION.md) - Tài liệu hệ thống ABAC
- [`API_DOCUMENTATION.md`](API_DOCUMENTATION.md) - Chi tiết API endpoints