# Hướng Dẫn Chi Tiết về Field Condition trong Policy JSON

## Mục Lục

1. [Tổng Quan](#tổng-quan)
2. [Cấu Trúc Condition](#cấu-trúc-condition)
3. [Context Keys - Các Khóa Ngữ Cảnh](#context-keys---các-khóa-ngữ-cảnh)
4. [Các Toán Tử String](#các-toán-tử-string)
5. [Các Toán Tử Numeric](#các-toán-tử-numeric)
6. [Các Toán Tử Boolean](#các-toán-tử-boolean)
7. [Các Toán Tử Date/Time](#các-toán-tử-datetime)
8. [Các Toán Tử Network/IP](#các-toán-tử-networkip)
9. [Các Toán Tử Array](#các-toán-tử-array)
10. [Các Toán Tử Logic (And/Or/Not)](#các-toán-tử-logic-andornot)
11. [Kết Hợp Nhiều Conditions](#kết-hợp-nhiều-conditions)
12. [Best Practices](#best-practices)

---

## Tổng Quan

Field `Condition` trong policy JSON cho phép bạn thiết lập các điều kiện chi tiết để kiểm soát quyền truy cập dựa trên các thuộc tính của user, resource, environment và request context.

### Cách Hoạt Động

- **Tất cả conditions** trong cùng một block phải **đồng thời thỏa mãn (AND logic)** để policy được áp dụng
- Nếu có bất kỳ condition nào không thỏa mãn, policy statement đó sẽ không được áp dụng
- Nếu không có condition nào được chỉ định, statement sẽ luôn được áp dụng (nếu Action và Resource match)

### Ví Dụ Cơ Bản

```json
{
  "Sid": "AllowDepartmentAccess",
  "Effect": "Allow",
  "Action": "document:read",
  "Resource": "api:documents:*",
  "Condition": {
    "StringEquals": {
      "user:Department": "Engineering"
    }
  }
}
```

---

## Cấu Trúc Condition

Condition có cấu trúc như sau:

```json
"Condition": {
  "<OperatorType>": {
    "<ContextKey>": "<Value>"
  }
}
```

Trong đó:
- `<OperatorType>`: Loại toán tử (StringEquals, NumericLessThan, v.v.)
- `<ContextKey>`: Khóa để truy xuất giá trị từ context
- `<Value>`: Giá trị mong đợi để so sánh

---

## Context Keys - Các Khóa Ngữ Cảnh

Hệ thống hỗ trợ các loại context keys sau:

### 1. Request Context Keys (request:*)

Các thuộc tính từ request hiện tại:

```
request:UserId          - ID của user thực hiện request
request:Action          - Action được yêu cầu
request:ResourceId      - ID của resource được truy cập
request:Time            - Thời gian request (RFC3339 format)
request:TimeOfDay       - Giờ trong ngày (HH:MM format)
request:DayOfWeek       - Ngày trong tuần (Monday, Tuesday, v.v.)
request:SourceIp        - IP address của client
```

### 2. User/Subject Context Keys (user:*)

Các thuộc tính từ user/subject:

```
user:Department         - Department của user
user:Role               - Role của user
user:Level              - Level của user
user:Manager            - Manager của user
user:SubjectType        - Loại subject (user, service, v.v.)
user:<custom-attribute> - Bất kỳ thuộc tính custom nào
```

### 3. Resource Context Keys (resource:*)

Các thuộc tính từ resource:

```
resource:Department     - Department sở hữu resource
resource:Owner          - Owner của resource
resource:Sensitivity    - Độ nhạy cảm (public, internal, confidential, v.v.)
resource:ResourceType   - Loại resource
resource:ResourceId     - ID của resource
resource:<custom>       - Thuộc tính custom của resource
```

### 4. Environment Context Keys (environment:*)

Các thuộc tính từ environment:

```
environment:client_ip       - IP của client
environment:user_agent      - User agent string
environment:country         - Country code
environment:region          - Region
environment:time_of_day     - Giờ hiện tại (HH:MM)
environment:day_of_week     - Ngày trong tuần
environment:hour            - Giờ (0-23)
environment:minute          - Phút (0-59)
environment:is_weekend      - true nếu là cuối tuần
environment:is_business_hours - true nếu trong giờ làm việc
environment:is_internal_ip  - true nếu IP nội bộ
environment:ip_class        - "ipv4" hoặc "ipv6"
environment:is_mobile       - true nếu từ mobile device
environment:browser         - Tên browser
```

### 5. Nested/Structured Access

Hệ thống hỗ trợ truy xuất giá trị lồng nhau (nested) bằng dấu chấm:

```
user.attributes.department
resource.attributes.owner
environment.location.country
```

---

## Các Toán Tử String

### StringEquals

So sánh chuỗi **chính xác** (case-sensitive).

**Cú pháp:**
```json
"Condition": {
  "StringEquals": {
    "<context-key>": "<expected-value>"
  }
}
```

**Ví dụ 1: Kiểm tra role**
```json
"Condition": {
  "StringEquals": {
    "user:Role": "admin"
  }
}
```

**Ví dụ 2: Kiểm tra nhiều thuộc tính**
```json
"Condition": {
  "StringEquals": {
    "user:Department": "Engineering",
    "user:Role": "developer"
  }
}
```
> Lưu ý: Tất cả điều kiện phải đồng thời thỏa mãn (AND)

**Ví dụ 3: Kiểm tra với array values**
```json
"Condition": {
  "StringEquals": {
    "user:Role": ["admin", "manager"]
  }
}
```
> User phải có role là "admin" HOẶC "manager"

---

### StringNotEquals

So sánh chuỗi **không bằng**.

**Ví dụ: Loại trừ department**
```json
"Condition": {
  "StringNotEquals": {
    "user:Department": "External"
  }
}
```

**Ví dụ: Loại trừ resource sensitivity**
```json
"Condition": {
  "StringNotEquals": {
    "resource:Sensitivity": "confidential"
  }
}
```

---

### StringLike

So sánh chuỗi với **wildcard pattern** (%, _).

- `%` = khớp với 0 hoặc nhiều ký tự
- `_` = khớp với chính xác 1 ký tự

**Ví dụ 1: Prefix matching**
```json
"Condition": {
  "StringLike": {
    "user:Email": "%.@company.com"
  }
}
```
> Khớp với bất kỳ email nào kết thúc bằng @company.com

**Ví dụ 2: Pattern matching**
```json
"Condition": {
  "StringLike": {
    "resource:Name": "project_%_report"
  }
}
```
> Khớp với: project_Q1_report, project_Q2_report, v.v.

---

### StringContains (Enhanced)

Kiểm tra chuỗi **có chứa** substring.

**Ví dụ:**
```json
"Condition": {
  "StringContains": {
    "user:Email": "@company.com"
  }
}
```

---

### StringStartsWith (Enhanced)

Kiểm tra chuỗi **bắt đầu** bằng prefix.

**Ví dụ:**
```json
"Condition": {
  "StringStartsWith": {
    "resource:ResourceId": "api:documents:"
  }
}
```

---

### StringEndsWith (Enhanced)

Kiểm tra chuỗi **kết thúc** bằng suffix.

**Ví dụ:**
```json
"Condition": {
  "StringEndsWith": {
    "user:Email": "@company.com"
  }
}
```

---

### StringRegex (Enhanced)

Kiểm tra chuỗi với **regular expression**.

**Ví dụ 1: Email validation**
```json
"Condition": {
  "StringRegex": {
    "user:Email": "^[a-zA-Z0-9._%+-]+@company\\.com$"
  }
}
```

**Ví dụ 2: Phone number validation**
```json
"Condition": {
  "StringRegex": {
    "user:Phone": "^\\+84[0-9]{9}$"
  }
}
```

---

## Các Toán Tử Numeric

### NumericEquals (Enhanced)

So sánh số **bằng**.

**Ví dụ:**
```json
"Condition": {
  "NumericEquals": {
    "user:Level": 5
  }
}
```

---

### NumericNotEquals (Enhanced)

So sánh số **không bằng**.

**Ví dụ:**
```json
"Condition": {
  "NumericNotEquals": {
    "resource:Version": 0
  }
}
```

---

### NumericLessThan

So sánh số **nhỏ hơn** (<).

**Ví dụ 1: Giới hạn số tiền**
```json
"Condition": {
  "NumericLessThan": {
    "transaction:Amount": 1000000
  }
}
```
> Chỉ cho phép transaction có amount < 1,000,000

**Ví dụ 2: Giới hạn theo level**
```json
"Condition": {
  "NumericLessThan": {
    "user:Level": 10
  }
}
```

---

### NumericLessThanEquals

So sánh số **nhỏ hơn hoặc bằng** (<=).

**Ví dụ:**
```json
"Condition": {
  "NumericLessThanEquals": {
    "transaction:Amount": 1000000
  }
}
```

---

### NumericGreaterThan

So sánh số **lớn hơn** (>).

**Ví dụ:**
```json
"Condition": {
  "NumericGreaterThan": {
    "user:Experience": 5
  }
}
```
> User phải có > 5 năm kinh nghiệm

---

### NumericGreaterThanEquals

So sánh số **lớn hơn hoặc bằng** (>=).

**Ví dụ: Yêu cầu manager cho giao dịch lớn**
```json
"Condition": {
  "NumericGreaterThanEquals": {
    "transaction:Amount": 1000000
  },
  "StringEquals": {
    "user:Role": "manager"
  }
}
```

---

### NumericBetween (Enhanced)

Kiểm tra số **nằm trong khoảng** [min, max].

**Cú pháp - Array:**
```json
"Condition": {
  "NumericBetween": {
    "transaction:Amount": [100000, 500000]
  }
}
```
> Amount phải từ 100,000 đến 500,000

**Cú pháp - Object:**
```json
"Condition": {
  "NumericBetween": {
    "user:Age": {
      "min": 18,
      "max": 65
    }
  }
}
```

---

## Các Toán Tử Boolean

### Bool / Boolean

Kiểm tra giá trị **boolean**.

**Ví dụ 1: Kiểm tra flag**
```json
"Condition": {
  "Bool": {
    "request:IsExternal": false
  }
}
```
> Chỉ cho phép internal requests

**Ví dụ 2: Kiểm tra MFA**
```json
"Condition": {
  "Bool": {
    "user:MFAEnabled": true
  }
}
```

**Ví dụ 3: Business hours**
```json
"Condition": {
  "Bool": {
    "environment:is_business_hours": true
  }
}
```

---

## Các Toán Tử Date/Time

### DateLessThan / TimeLessThan

Kiểm tra thời gian **trước** thời điểm chỉ định.

**Ví dụ 1: Thời gian trong ngày**
```json
"Condition": {
  "DateLessThan": {
    "request:TimeOfDay": "18:00:00"
  }
}
```
> Chỉ cho phép trước 6 PM

**Ví dụ 2: Expiry date**
```json
"Condition": {
  "DateLessThan": {
    "resource:ExpiryDate": "2025-12-31T23:59:59Z"
  }
}
```

---

### DateGreaterThan / TimeGreaterThan

Kiểm tra thời gian **sau** thời điểm chỉ định.

**Ví dụ:**
```json
"Condition": {
  "DateGreaterThan": {
    "request:TimeOfDay": "09:00:00"
  }
}
```
> Chỉ cho phép sau 9 AM

---

### DateLessThanEquals / TimeLessThanEquals (Enhanced)

Kiểm tra thời gian **trước hoặc bằng**.

**Ví dụ:**
```json
"Condition": {
  "TimeLessThanEquals": {
    "request:TimeOfDay": "17:59:59"
  }
}
```

---

### DateGreaterThanEquals / TimeGreaterThanEquals (Enhanced)

Kiểm tra thời gian **sau hoặc bằng**.

**Ví dụ:**
```json
"Condition": {
  "TimeGreaterThanEquals": {
    "request:TimeOfDay": "09:00:00"
  }
}
```

---

### DateBetween / TimeBetween (Enhanced)

Kiểm tra thời gian **nằm trong khoảng**.

**Ví dụ: Business hours (9 AM - 6 PM)**
```json
"Condition": {
  "TimeBetween": {
    "request:TimeOfDay": ["09:00:00", "18:00:00"]
  }
}
```

**Ví dụ: Date range**
```json
"Condition": {
  "DateBetween": {
    "request:Time": ["2025-01-01T00:00:00Z", "2025-12-31T23:59:59Z"]
  }
}
```

---

### DayOfWeek (Enhanced)

Kiểm tra **ngày trong tuần**.

**Ví dụ 1: Chặn cuối tuần**
```json
"Condition": {
  "DayOfWeek": {
    "request:DayOfWeek": ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]
  }
}
```

**Ví dụ 2: Chỉ cho phép cuối tuần**
```json
"Condition": {
  "DayOfWeek": {
    "environment:day_of_week": ["Saturday", "Sunday"]
  }
}
```

---

### TimeOfDay (Enhanced)

Kiểm tra **giờ chính xác** trong ngày.

**Ví dụ:**
```json
"Condition": {
  "TimeOfDay": {
    "environment:time_of_day": "14:30"
  }
}
```

---

### IsBusinessHours (Enhanced)

Kiểm tra có phải **giờ làm việc** (9 AM - 5 PM, Monday-Friday).

**Ví dụ:**
```json
"Condition": {
  "IsBusinessHours": {
    "environment:is_business_hours": true
  }
}
```

---

## Các Toán Tử Network/IP

### IpAddress

Kiểm tra IP address **thuộc CIDR range**.

**Ví dụ 1: Cho phép internal network**
```json
"Condition": {
  "IpAddress": {
    "request:SourceIp": ["10.0.0.0/8", "192.168.1.0/24"]
  }
}
```

**Ví dụ 2: Whitelist specific IPs**
```json
"Condition": {
  "IpAddress": {
    "environment:client_ip": "203.0.113.0/24"
  }
}
```

---

### IPInRange (Enhanced)

Tương tự `IpAddress`, kiểm tra IP **trong range**.

**Ví dụ:**
```json
"Condition": {
  "IPInRange": {
    "environment:client_ip": ["10.0.0.0/8", "172.16.0.0/12"]
  }
}
```

---

### IPNotInRange (Enhanced)

Kiểm tra IP **không nằm trong range**.

**Ví dụ: Block external IPs**
```json
"Condition": {
  "IPNotInRange": {
    "environment:client_ip": ["0.0.0.0/0"]
  }
}
```

---

### IsInternalIP (Enhanced)

Kiểm tra IP có phải **internal/private IP**.

**Ví dụ:**
```json
"Condition": {
  "IsInternalIP": {
    "environment:is_internal_ip": true
  }
}
```

Private IP ranges được kiểm tra:
- 10.0.0.0/8
- 172.16.0.0/12
- 192.168.0.0/16
- 127.0.0.0/8

---

## Các Toán Tử Array

### ArrayContains (Enhanced)

Kiểm tra array **có chứa** giá trị.

**Ví dụ 1: Kiểm tra role trong danh sách**
```json
"Condition": {
  "ArrayContains": {
    "user:Roles": "admin"
  }
}
```

**Ví dụ 2: Kiểm tra permission**
```json
"Condition": {
  "ArrayContains": {
    "user:Permissions": "documents:write"
  }
}
```

---

### ArrayNotContains (Enhanced)

Kiểm tra array **không chứa** giá trị.

**Ví dụ:**
```json
"Condition": {
  "ArrayNotContains": {
    "user:RestrictedGroups": "blacklisted"
  }
}
```

---

### ArraySize (Enhanced)

Kiểm tra **kích thước** của array.

**Cú pháp 1: Exact size**
```json
"Condition": {
  "ArraySize": {
    "user:Roles": 2
  }
}
```

**Cú pháp 2: With operators**
```json
"Condition": {
  "ArraySize": {
    "user:Permissions": {
      "gte": 5
    }
  }
}
```

**Các operators hỗ trợ:**
- `eq` / `equals`: Bằng
- `gt` / `greaterthan`: Lớn hơn
- `gte` / `greaterthanequals`: Lớn hơn hoặc bằng
- `lt` / `lessthan`: Nhỏ hơn
- `lte` / `lessthanequals`: Nhỏ hơn hoặc bằng

**Ví dụ: Yêu cầu ít nhất 3 roles**
```json
"Condition": {
  "ArraySize": {
    "user:Roles": {
      "gte": 3
    }
  }
}
```

---

## Các Toán Tử Logic (And/Or/Not)

### And

Tất cả các conditions phải **đồng thời thỏa mãn**.

**Cú pháp:**
```json
"Condition": {
  "And": [
    {
      "StringEquals": {
        "user:Department": "Engineering"
      }
    },
    {
      "NumericGreaterThan": {
        "user:Level": 3
      }
    }
  ]
}
```

---

### Or

**Ít nhất một** condition phải thỏa mãn.

**Ví dụ: Admin hoặc Manager**
```json
"Condition": {
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
  ]
}
```

**Ví dụ phức tạp: Multiple criteria**
```json
"Condition": {
  "Or": [
    {
      "StringEquals": {
        "user:Department": "Security"
      }
    },
    {
      "And": [
        {
          "StringEquals": {
            "user:Role": "manager"
          }
        },
        {
          "NumericGreaterThan": {
            "user:Level": 5
          }
        }
      ]
    }
  ]
}
```
> Cho phép: Security department HOẶC (Manager với Level > 5)

---

### Not

**Phủ định** condition.

**Ví dụ 1: Không phải external user**
```json
"Condition": {
  "Not": {
    "StringEquals": {
      "user:Type": "external"
    }
  }
}
```

**Ví dụ 2: Phức tạp hơn**
```json
"Condition": {
  "Not": {
    "Or": [
      {
        "StringEquals": {
          "user:Department": "External"
        }
      },
      {
        "Bool": {
          "user:Suspended": true
        }
      }
    ]
  }
}
```
> Cho phép: KHÔNG phải (External department HOẶC Suspended user)

---

## Kết Hợp Nhiều Conditions

### Ví Dụ 1: Business Hours Access

Chỉ cho phép truy cập trong giờ làm việc từ internal network:

```json
{
  "Sid": "BusinessHoursInternalAccess",
  "Effect": "Allow",
  "Action": "document:*",
  "Resource": "api:documents:*",
  "Condition": {
    "DateGreaterThan": {
      "request:TimeOfDay": "09:00:00"
    },
    "DateLessThan": {
      "request:TimeOfDay": "18:00:00"
    },
    "StringEquals": {
      "request:DayOfWeek": ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]
    },
    "IsInternalIP": {
      "environment:is_internal_ip": true
    }
  }
}
```

---

### Ví Dụ 2: Tiered Transaction Approval

Phê duyệt giao dịch theo level:

```json
{
  "policies": [
    {
      "id": "pol-transaction-approval",
      "policy_name": "Transaction Approval Based on Amount",
      "statement": [
        {
          "Sid": "SmallTransactions",
          "Effect": "Allow",
          "Action": "payment:transaction:approve",
          "Resource": "api:transactions:*",
          "Condition": {
            "NumericLessThan": {
              "transaction:Amount": 100000
            }
          }
        },
        {
          "Sid": "MediumTransactionsRequireManager",
          "Effect": "Allow",
          "Action": "payment:transaction:approve",
          "Resource": "api:transactions:*",
          "Condition": {
            "NumericBetween": {
              "transaction:Amount": [100000, 1000000]
            },
            "StringEquals": {
              "user:Role": "manager"
            }
          }
        },
        {
          "Sid": "LargeTransactionsRequireDirector",
          "Effect": "Allow",
          "Action": "payment:transaction:approve",
          "Resource": "api:transactions:*",
          "Condition": {
            "NumericGreaterThanEquals": {
              "transaction:Amount": 1000000
            },
            "Or": [
              {
                "StringEquals": {
                  "user:Role": "director"
                }
              },
              {
                "StringEquals": {
                  "user:Role": "ceo"
                }
              }
            ]
          }
        }
      ],
      "enabled": true
    }
  ]
}
```

---

### Ví Dụ 3: Document Access Control

Kiểm soát truy cập document theo department và sensitivity:

```json
{
  "Sid": "DepartmentDocumentAccess",
  "Effect": "Allow",
  "Action": ["document:read", "document:list"],
  "Resource": "api:documents:dept-${user:Department}/*",
  "Condition": {
    "StringEquals": {
      "user:Department": "Engineering"
    },
    "StringNotEquals": {
      "resource:Sensitivity": "confidential"
    },
    "Or": [
      {
        "StringEquals": {
          "resource:Owner": "${request:UserId}"
        }
      },
      {
        "NumericGreaterThanEquals": {
          "user:Level": 5
        }
      }
    ]
  }
}
```
> Cho phép Engineering đọc documents (không phải confidential) nếu:
> - Họ là owner, HOẶC
> - Level >= 5

---

### Ví Dụ 4: Geo-Location Based Access

Kiểm soát truy cập theo location và device:

```json
{
  "Sid": "GeoLocationAccess",
  "Effect": "Allow",
  "Action": "sensitive-data:*",
  "Resource": "*",
  "Condition": {
    "StringEquals": {
      "environment:country": ["VN", "SG", "US"]
    },
    "Bool": {
      "user:MFAEnabled": true
    },
    "Not": {
      "Bool": {
        "environment:is_mobile": true
      }
    }
  }
}
```
> Cho phép truy cập sensitive data nếu:
> - Từ VN, SG, hoặc US
> - MFA đã bật
> - KHÔNG phải từ mobile device

---

### Ví Dụ 5: Complex Admin Access

Kiểm soát admin access với nhiều điều kiện:

```json
{
  "Sid": "SecureAdminAccess",
  "Effect": "Allow",
  "Action": "admin:*",
  "Resource": "api:admin:*",
  "Condition": {
    "And": [
      {
        "Or": [
          {
            "StringEquals": {
              "user:Role": "admin"
            }
          },
          {
            "StringEquals": {
              "user:Role": "super-admin"
            }
          }
        ]
      },
      {
        "Bool": {
          "user:MFAEnabled": true
        }
      },
      {
        "IsInternalIP": {
          "environment:is_internal_ip": true
        }
      },
      {
        "IsBusinessHours": {
          "environment:is_business_hours": true
        }
      },
      {
        "Not": {
          "Bool": {
            "user:Suspended": true
          }
        }
      }
    ]
  }
}
```

---

## Best Practices

### 1. Sử Dụng StringEquals Thay Vì StringLike Khi Có Thể

```json
// TỐT
"Condition": {
  "StringEquals": {
    "user:Role": "admin"
  }
}

// TRÁNH (nếu không cần wildcard)
"Condition": {
  "StringLike": {
    "user:Role": "admin"
  }
}
```

### 2. Kết Hợp Conditions Hiệu Quả

Đặt các điều kiện dễ fail trước (performance):

```json
// TỐT - Check role trước (fast), check IP sau
"Condition": {
  "StringEquals": {
    "user:Role": "admin"
  },
  "IsInternalIP": {
    "environment:is_internal_ip": true
  }
}
```

### 3. Sử Dụng Array Values Cho Multiple Options

```json
// TỐT
"Condition": {
  "StringEquals": {
    "user:Role": ["admin", "manager", "supervisor"]
  }
}

// TRÁNH (phức tạp không cần thiết)
"Condition": {
  "Or": [
    {"StringEquals": {"user:Role": "admin"}},
    {"StringEquals": {"user:Role": "manager"}},
    {"StringEquals": {"user:Role": "supervisor"}}
  ]
}
```

### 4. Kiểm Tra Null/Empty Values

Luôn đảm bảo attributes tồn tại trước khi so sánh:

```json
{
  "Condition": {
    "StringNotEquals": {
      "user:Department": ""
    },
    "StringEquals": {
      "user:Department": "Engineering"
    }
  }
}
```

### 5. Sử Dụng NumericBetween Cho Range Checks

```json
// TỐT
"Condition": {
  "NumericBetween": {
    "transaction:Amount": [1000, 10000]
  }
}

// TRÁNH
"Condition": {
  "NumericGreaterThanEquals": {
    "transaction:Amount": 1000
  },
  "NumericLessThanEquals": {
    "transaction:Amount": 10000
  }
}
```

### 6. Documentation và Comments

Sử dụng `Sid` và `description` để document policy:

```json
{
  "Sid": "AllowEngineeringDepartmentDocumentAccess",
  "description": "Allows Engineering department to read non-confidential documents",
  "Effect": "Allow",
  "Action": "document:read",
  "Resource": "api:documents:*",
  "Condition": {
    "StringEquals": {
      "user:Department": "Engineering"
    },
    "StringNotEquals": {
      "resource:Sensitivity": "confidential"
    }
  }
}
```

### 7. Test Conditions Thoroughly

Luôn test với:
- ✅ Happy path (should allow)
- ✅ Edge cases (boundary values)
- ✅ Negative cases (should deny)
- ✅ Missing attributes
- ✅ Invalid values

### 8. Tránh Điều Kiện Quá Phức Tạp

```json
// TRÁNH - Quá phức tạp, khó maintain
"Condition": {
  "Or": [
    {
      "And": [
        {"StringEquals": {"user:Role": "admin"}},
        {"Or": [
          {"NumericGreaterThan": {"user:Level": 5}},
          {"Bool": {"user:VIP": true}}
        ]}
      ]
    },
    {
      "And": [
        {"StringEquals": {"user:Department": "Engineering"}},
        {"Not": {"Bool": {"user:Suspended": true}}}
      ]
    }
  ]
}

// TỐT - Chia thành nhiều statements đơn giản hơn
```

### 9. Variable Substitution

Sử dụng `${...}` để tham chiếu context values trong conditions:

```json
{
  "Resource": "api:documents:owner-${request:UserId}",
  "Condition": {
    "StringEquals": {
      "resource:Department": "${user:Department}"
    }
  }
}
```

### 10. Security Best Practices

- Luôn verify user identity và authentication status
- Sử dụng Deny statements cho security-critical rules
- Implement least privilege principle
- Log và audit policy decisions

```json
{
  "Sid": "DenyConfidentialAccessFromExternal",
  "Effect": "Deny",
  "Action": "*",
  "Resource": "*",
  "Condition": {
    "StringEquals": {
      "resource:Sensitivity": "confidential"
    },
    "Bool": {
      "request:IsExternal": true
    }
  }
}
```

---

## Tham Khảo Thêm

- [ACTION_FIELD_GUIDE.md](./ACTION_FIELD_GUIDE.md) - Hướng dẫn về Action field
- [RESOURCE_FIELD_GUIDE.md](./RESOURCE_FIELD_GUIDE.md) - Hướng dẫn về Resource field
- Xem code implementation tại:
  - `evaluator/conditions.go` - Condition evaluator cơ bản
  - `evaluator/enhanced_condition_evaluator.go` - Enhanced condition evaluator
  - `evaluator/pdp.go` - Policy Decision Point

---

**Lưu Ý:** Tài liệu này dựa trên code logic hiện tại. Nếu có thắc mắc về implementation cụ thể, vui lòng tham khảo source code hoặc liên hệ team development.
