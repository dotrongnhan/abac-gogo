# Hướng Dẫn Sử Dụng Policy JSON - ABAC System

## 📋 Tổng Quan

Hệ thống ABAC này sử dụng format JSON để định nghĩa các policy kiểm soát truy cập. Policy JSON tuân theo cấu trúc tương tự AWS IAM Policy với các cải tiến cho ABAC.

## 🏗️ Cấu Trúc Policy JSON

### Cấu Trúc Cơ Bản

```json
{
  "policies": [
    {
      "id": "pol-001",
      "policy_name": "Tên Policy",
      "description": "Mô tả policy",
      "version": "2024-10-21",
      "statement": [
        {
          "Sid": "Statement ID",
          "Effect": "Allow|Deny",
          "Action": "service:resource:operation",
          "Resource": "resource_pattern",
          "Condition": {
            "OperatorType": {
              "attribute_path": "expected_value"
            }
          }
        }
      ],
      "enabled": true
    }
  ]
}
```

### Các Thành Phần Chính

#### 1. **Policy Level**
- `id`: ID duy nhất của policy
- `policy_name`: Tên policy (phải unique)
- `description`: Mô tả chức năng policy
- `version`: Phiên bản policy
- `statement`: Mảng các statement
- `enabled`: Trạng thái kích hoạt policy

#### 2. **Statement Level**
- `Sid`: Statement ID (tùy chọn, dùng để debug)
- `Effect`: `"Allow"` hoặc `"Deny"`
- `Action`: Hành động được phép/cấm
- `Resource`: Tài nguyên áp dụng
- `NotResource`: Tài nguyên loại trừ (tùy chọn)
- `Condition`: Điều kiện runtime (tùy chọn)

## 🎯 Action & Resource Patterns

### Action Format
```
service:resource_type:operation
```

**Ví dụ:**
- `document-service:file:read`
- `payment-service:transaction:approve`
- `*:*:*` (tất cả)

### Resource Format
```
api:resource_type:identifier
```

**Ví dụ:**
- `api:documents:owner-${request:UserId}`
- `api:transactions:*`
- `api:departments:${user:Department}/documents:*`

### Variable Substitution
- `${request:UserId}`: ID của user trong request
- `${user:Department}`: Department của user
- `${resource:Sensitivity}`: Thuộc tính của resource

## 🔧 Condition Operators

### String Operators
```json
{
  "StringEquals": {
    "user:Role": "manager"
  },
  "StringNotEquals": {
    "resource:Sensitivity": "confidential"
  },
  "StringLike": {
    "user:Email": "*@company.com"
  },
  "StringContains": {
    "user:Department": "IT"
  },
  "StringStartsWith": {
    "resource:Path": "/public/"
  },
  "StringEndsWith": {
    "resource:Name": ".pdf"
  },
  "StringRegex": {
    "user:Phone": "^\\+84[0-9]{9}$"
  }
}
```

### Numeric Operators
```json
{
  "NumericLessThan": {
    "transaction:Amount": 1000000
  },
  "NumericLessThanEquals": {
    "user:Age": 65
  },
  "NumericGreaterThan": {
    "transaction:Amount": 0
  },
  "NumericGreaterThanEquals": {
    "user:Experience": 5
  },
  "NumericBetween": {
    "transaction:Amount": [100000, 5000000]
  }
}
```

### Date/Time Operators
```json
{
  "DateGreaterThan": {
    "request:TimeOfDay": "09:00:00"
  },
  "DateLessThan": {
    "request:TimeOfDay": "18:00:00"
  },
  "DateBetween": {
    "request:Date": ["2024-01-01", "2024-12-31"]
  },
  "DayOfWeek": {
    "request:DayOfWeek": ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]
  },
  "TimeOfDay": {
    "request:Time": "14:30:00"
  },
  "IsBusinessHours": {
    "request:Timestamp": true
  }
}
```

### Network Operators
```json
{
  "IpAddress": {
    "request:SourceIp": ["10.0.0.0/8", "192.168.1.0/24"]
  },
  "IpInRange": {
    "request:ClientIP": "192.168.1.0/24"
  },
  "IpNotInRange": {
    "request:ClientIP": "10.0.0.0/8"
  },
  "IsInternalIP": {
    "request:SourceIP": true
  }
}
```

### Boolean Operators
```json
{
  "Bool": {
    "request:IsExternal": false,
    "user:IsActive": true
  }
}
```

### Array Operators
```json
{
  "ArrayContains": {
    "user:Roles": "admin"
  },
  "ArrayNotContains": {
    "user:Permissions": "delete_all"
  },
  "ArraySize": {
    "user:Groups": 3
  }
}
```

### Logical Operators
```json
{
  "And": [
    {
      "StringEquals": {
        "user:Department": "Finance"
      }
    },
    {
      "NumericGreaterThan": {
        "user:Level": 3
      }
    }
  ],
  "Or": [
    {
      "StringEquals": {
        "user:Role": "admin"
      }
    },
    {
      "StringEquals": {
        "user:Role": "manager"
      }
    }
  ],
  "Not": {
    "StringEquals": {
      "user:Status": "suspended"
    }
  }
}
```

## 📝 Ví Dụ Policy Hoàn Chỉnh

### 1. Document Access Policy
```json
{
  "id": "pol-001",
  "policy_name": "Department Document Access",
  "description": "Allow users to access documents in their department",
  "version": "2024-10-21",
  "statement": [
    {
      "Sid": "OwnDocumentsFullAccess",
      "Effect": "Allow",
      "Action": "document-service:file:*",
      "Resource": "api:documents:owner-${request:UserId}"
    },
    {
      "Sid": "DepartmentDocumentsRead",
      "Effect": "Allow",
      "Action": [
        "document-service:file:read",
        "document-service:file:list"
      ],
      "Resource": "api:documents:dept-${user:Department}",
      "Condition": {
        "StringNotEquals": {
          "resource:Sensitivity": "confidential"
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

### 2. Time-Based Access Policy
```json
{
  "id": "pol-003",
  "policy_name": "Business Hours Access",
  "description": "Time-based access control",
  "version": "2024-10-21",
  "statement": [
    {
      "Sid": "BusinessHoursOnly",
      "Effect": "Allow",
      "Action": "payment-service:transaction:create",
      "Resource": "api:transactions:*",
      "Condition": {
        "And": [
          {
            "DateGreaterThan": {
              "request:TimeOfDay": "09:00:00"
            }
          },
          {
            "DateLessThan": {
              "request:TimeOfDay": "18:00:00"
            }
          }
        ]
      }
    },
    {
      "Sid": "DenyWeekendAccess",
      "Effect": "Deny",
      "Action": "payment-service:*:*",
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
```

### 3. Complex Conditional Policy
```json
{
  "id": "pol-005",
  "policy_name": "Manager Override Policy",
  "description": "Managers can override certain restrictions",
  "version": "2024-10-21",
  "statement": [
    {
      "Sid": "ManagerOverride",
      "Effect": "Allow",
      "Action": "*:*:*",
      "Resource": "*",
      "Condition": {
        "And": [
          {
            "StringEquals": {
              "user:Role": "manager"
            }
          },
          {
            "Or": [
              {
                "StringEquals": {
                  "user:Department": "IT"
                }
              },
              {
                "NumericGreaterThan": {
                  "user:Level": 5
                }
              }
            ]
          },
          {
            "Not": {
              "StringEquals": {
                "resource:Classification": "top-secret"
              }
            }
          }
        ]
      }
    }
  ],
  "enabled": true
}
```

## 🔍 Context Variables

### Request Context
- `request:UserId`: ID của user thực hiện request
- `request:SourceIp`: IP address của client
- `request:TimeOfDay`: Thời gian trong ngày (HH:mm:ss)
- `request:DayOfWeek`: Thứ trong tuần
- `request:Timestamp`: Timestamp của request

### User Context
- `user:Department`: Phòng ban của user
- `user:Role`: Vai trò của user
- `user:Level`: Cấp độ của user
- `user:Email`: Email của user
- `user:Groups`: Các nhóm user thuộc về

### Resource Context
- `resource:Sensitivity`: Mức độ nhạy cảm
- `resource:Owner`: Chủ sở hữu resource
- `resource:Department`: Phòng ban sở hữu resource
- `resource:Classification`: Phân loại bảo mật

### Transaction Context
- `transaction:Amount`: Số tiền giao dịch
- `transaction:Type`: Loại giao dịch
- `transaction:Currency`: Đơn vị tiền tệ

## ⚡ Best Practices

### 1. **Policy Organization**
- Sử dụng naming convention rõ ràng
- Nhóm các statement liên quan
- Sử dụng Sid để dễ debug

### 2. **Performance**
- Đặt condition đơn giản trước
- Sử dụng index cho các attribute thường xuyên query
- Tránh regex phức tạp trong condition

### 3. **Security**
- Luôn có explicit deny cho sensitive resources
- Sử dụng principle of least privilege
- Test kỹ các edge cases

### 4. **Maintainability**
- Version control cho policy changes
- Document các business rules
- Sử dụng meaningful descriptions

## 🚨 Lưu Ý Quan Trọng

1. **Effect Priority**: `Deny` luôn có priority cao hơn `Allow`
2. **Default Behavior**: Nếu không có policy nào match, default là `Deny`
3. **Variable Substitution**: Chỉ hoạt động trong runtime evaluation
4. **Case Sensitivity**: Operator names không phân biệt hoa thường
5. **Array Handling**: Action và Resource có thể là string hoặc array

## 📚 Tài Liệu Liên Quan

- [Database Setup Guide](DATABASE_SETUP.md)
- [API Documentation](API_DOCUMENTATION.md)
- [Test Coverage](TEST_COVERAGE.md)
- [Complex Logical Conditions](COMPLEX_LOGICAL_CONDITIONS.md)
