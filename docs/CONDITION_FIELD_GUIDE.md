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
12. [Ví Dụ Policy Hoàn Chỉnh](#ví-dụ-policy-hoàn-chỉnh)
13. [Variable Substitution](#variable-substitution)
14. [Xử Lý Giá Trị và Type Conversion](#xử-lý-giá-trị-và-type-conversion)
15. [Troubleshooting](#troubleshooting)
16. [Best Practices](#best-practices)

---

## Tổng Quan

Field `Condition` trong policy JSON cho phép bạn thiết lập các điều kiện chi tiết để kiểm soát quyền truy cập dựa trên các thuộc tính của user, resource, environment và request context.

### Cách Hoạt Động

- **Tất cả conditions** trong cùng một block phải **đồng thời thỏa mãn (AND logic)** để policy được áp dụng
- Nếu có bất kỳ condition nào không thỏa mãn, policy statement đó sẽ không được áp dụng
- Nếu không có condition nào được chỉ định, statement sẽ luôn được áp dụng (nếu Action và Resource match)
- Operator names **không phân biệt** hoa thường (case-insensitive)
- Attribute values **có phân biệt** hoa thường (case-sensitive)

### Quick Reference - Tất Cả Operators

| Category | Operator | Mô Tả | Ví Dụ Value |
|----------|----------|-------|-------------|
| **String** | StringEquals | So sánh bằng | `"admin"` hoặc `["admin", "user"]` |
| | StringNotEquals | So sánh không bằng | `"guest"` |
| | StringLike | Pattern với * | `"*@company.com"` |
| | StringContains | Chứa substring | `"@company"` |
| | StringStartsWith | Bắt đầu bằng | `"admin-"` |
| | StringEndsWith | Kết thúc bằng | `"-prod"` |
| | StringRegex | Regular expression | `"^[A-Z][0-9]{3}$"` |
| **Numeric** | NumericEquals | Bằng | `100` |
| | NumericNotEquals | Không bằng | `0` |
| | NumericLessThan | Nhỏ hơn (<) | `1000` |
| | NumericLessThanEquals | Nhỏ hơn hoặc bằng (<=) | `1000` |
| | NumericGreaterThan | Lớn hơn (>) | `100` |
| | NumericGreaterThanEquals | Lớn hơn hoặc bằng (>=) | `100` |
| | NumericBetween | Trong khoảng | `[100, 1000]` hoặc `{"min": 100, "max": 1000}` |
| **Boolean** | Bool / Boolean | So sánh boolean | `true` hoặc `false` |
| **Date/Time** | DateLessThan / TimeLessThan | Trước thời điểm | `"18:00:00"` |
| | DateGreaterThan / TimeGreaterThan | Sau thời điểm | `"09:00:00"` |
| | DateLessThanEquals / TimeLessThanEquals | Trước hoặc bằng | `"17:59:59"` |
| | DateGreaterThanEquals / TimeGreaterThanEquals | Sau hoặc bằng | `"09:00:00"` |
| | DateBetween / TimeBetween | Trong khoảng thời gian | `["09:00:00", "18:00:00"]` |
| | DayOfWeek | Ngày trong tuần | `"Monday"` hoặc `["Monday", "Tuesday"]` |
| | TimeOfDay | Giờ chính xác | `"14:30"` |
| | IsBusinessHours | Giờ làm việc (9-17, Mon-Fri) | `true` |
| **Network/IP** | IpAddress | IP trong CIDR range | `"10.0.0.0/8"` hoặc `["10.0.0.0/8", "192.168.1.0/24"]` |
| | IPInRange | IP trong range | `["172.16.0.0/12"]` |
| | IPNotInRange | IP không trong range | `["0.0.0.0/0"]` |
| | IsInternalIP | IP nội bộ | `true` |
| **Array** | ArrayContains | Array chứa giá trị | `"admin"` |
| | ArrayNotContains | Array không chứa | `"guest"` |
| | ArraySize | Kích thước array | `3` hoặc `{"gte": 3}` |
| **Logic** | And | Tất cả phải thỏa mãn | `[{condition1}, {condition2}]` |
| | Or | Ít nhất một thỏa mãn | `[{condition1}, {condition2}]` |
| | Not | Phủ định | `{condition}` |

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

So sánh chuỗi với **wildcard pattern** sử dụng `*`.

- `*` = khớp với 0 hoặc nhiều ký tự bất kỳ
- Pattern matching hỗ trợ:
  - `*` - khớp tất cả
  - `prefix*` - bắt đầu bằng prefix
  - `*suffix` - kết thúc bằng suffix
  - `*middle*` - chứa middle

**Ví dụ 1: Suffix matching (kết thúc bằng)**
```json
"Condition": {
  "StringLike": {
    "user:Email": "*@company.com"
  }
}
```
> Khớp với bất kỳ email nào kết thúc bằng @company.com

**Ví dụ 2: Prefix matching (bắt đầu bằng)**
```json
"Condition": {
  "StringLike": {
    "resource:Name": "project-*"
  }
}
```
> Khớp với: project-alpha, project-beta, v.v.

**Ví dụ 3: Contains (chứa)**
```json
"Condition": {
  "StringLike": {
    "user:Email": "*@company.*"
  }
}
```
> Khớp với bất kỳ email domain nào chứa "company"

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

## Ví Dụ Policy Hoàn Chỉnh

### Complete Policy với Nhiều Use Cases

Dưới đây là ví dụ một policy file hoàn chỉnh với nhiều use cases thực tế:

```json
{
  "policies": [
    {
      "id": "pol-document-management-001",
      "policy_name": "Document Management Access Control",
      "description": "Comprehensive document access control policy",
      "version": "1.0.0",
      "statement": [
        {
          "Sid": "AllowOwnDocuments",
          "Effect": "Allow",
          "Action": "document:*",
          "Resource": "api:documents:owner-${request:UserId}/*",
          "Condition": {}
        },
        {
          "Sid": "AllowDepartmentRead",
          "Effect": "Allow",
          "Action": ["document:read", "document:list"],
          "Resource": "api:documents:dept-${user:Department}/*",
          "Condition": {
            "StringNotEquals": {
              "resource:Sensitivity": "confidential"
            },
            "IsBusinessHours": {
              "environment:is_business_hours": true
            }
          }
        },
        {
          "Sid": "AllowManagerConfidentialAccess",
          "Effect": "Allow",
          "Action": "document:read",
          "Resource": "api:documents:*",
          "Condition": {
            "StringEquals": {
              "resource:Sensitivity": "confidential",
              "user:Role": "manager"
            },
            "NumericGreaterThanEquals": {
              "user:Level": 5
            }
          }
        },
        {
          "Sid": "DenyExternalAccessToConfidential",
          "Effect": "Deny",
          "Action": "*",
          "Resource": "*",
          "Condition": {
            "StringEquals": {
              "resource:Sensitivity": "confidential"
            },
            "Not": {
              "IsInternalIP": {
                "environment:is_internal_ip": true
              }
            }
          }
        },
        {
          "Sid": "AllowAdminFullAccess",
          "Effect": "Allow",
          "Action": "document:*",
          "Resource": "api:documents:*",
          "Condition": {
            "StringEquals": {
              "user:Role": "admin"
            },
            "Bool": {
              "user:MFAEnabled": true
            }
          }
        }
      ],
      "enabled": true
    },
    {
      "id": "pol-transaction-approval-001",
      "policy_name": "Transaction Approval Rules",
      "description": "Tiered approval based on transaction amount",
      "version": "1.0.0",
      "statement": [
        {
          "Sid": "SmallTransactionAnyEmployee",
          "Effect": "Allow",
          "Action": "transaction:approve",
          "Resource": "api:transactions:*",
          "Condition": {
            "NumericLessThan": {
              "transaction:Amount": 50000
            },
            "IsBusinessHours": {
              "environment:is_business_hours": true
            }
          }
        },
        {
          "Sid": "MediumTransactionRequiresManager",
          "Effect": "Allow",
          "Action": "transaction:approve",
          "Resource": "api:transactions:*",
          "Condition": {
            "NumericBetween": {
              "transaction:Amount": [50000, 500000]
            },
            "Or": [
              {
                "StringEquals": {
                  "user:Role": "manager"
                }
              },
              {
                "StringEquals": {
                  "user:Role": "director"
                }
              }
            ]
          }
        },
        {
          "Sid": "LargeTransactionRequiresDirector",
          "Effect": "Allow",
          "Action": "transaction:approve",
          "Resource": "api:transactions:*",
          "Condition": {
            "NumericGreaterThanEquals": {
              "transaction:Amount": 500000
            },
            "StringEquals": {
              "user:Role": "director"
            },
            "Bool": {
              "user:MFAEnabled": true
            }
          }
        },
        {
          "Sid": "DenyWeekendLargeTransactions",
          "Effect": "Deny",
          "Action": "transaction:approve",
          "Resource": "*",
          "Condition": {
            "NumericGreaterThan": {
              "transaction:Amount": 100000
            },
            "DayOfWeek": {
              "environment:day_of_week": ["Saturday", "Sunday"]
            }
          }
        }
      ],
      "enabled": true
    }
  ]
}
```

**Giải thích flow:**

1. **pol-document-management-001** - Quản lý documents:
   - User luôn có full access với documents của chính họ
   - Trong business hours, có thể read documents của department (trừ confidential)
   - Manager level 5+ có thể read confidential documents
   - **Deny** external access tới confidential documents (rule này override tất cả Allow rules)
   - Admin với MFA có full access

2. **pol-transaction-approval-001** - Phê duyệt transactions:
   - < 50k: Bất kỳ employee nào trong business hours
   - 50k-500k: Manager hoặc Director
   - >= 500k: Director với MFA
   - **Deny** transactions > 100k vào cuối tuần

---

## Variable Substitution

### Cách Hoạt Động

Hệ thống hỗ trợ **variable substitution** sử dụng cú pháp `${prefix:key}` trong cả Resource và Condition values.

**Syntax:**
```
${request:UserId}       - ID của user hiện tại
${user:Department}      - Department của user
${user:<any-attribute>} - Bất kỳ attribute nào của user
${resource:<attribute>} - Attribute của resource
${environment:<key>}    - Environment variable
```

### Ví Dụ 1: Resource Pattern với Variable

```json
{
  "Sid": "AccessOwnDepartmentDocuments",
  "Effect": "Allow",
  "Action": "document:read",
  "Resource": "api:documents:dept-${user:Department}/*",
  "Condition": {}
}
```

**Flow:**
1. User với `Department: "Engineering"` request access
2. Resource pattern được expand thành: `api:documents:dept-Engineering/*`
3. Match với resource: `api:documents:dept-Engineering/doc-123`

### Ví Dụ 2: Condition Value với Variable

```json
{
  "Sid": "OwnerCanDelete",
  "Effect": "Allow",
  "Action": "document:delete",
  "Resource": "api:documents:*",
  "Condition": {
    "StringEquals": {
      "resource:Owner": "${request:UserId}"
    }
  }
}
```

**Flow:**
1. User `user-123` request delete `document-456`
2. Condition được expand: `resource:Owner` must equal `user-123`
3. Nếu `document-456` có `Owner: "user-123"` → Allow

### Ví Dụ 3: Nested Variable Access

```json
{
  "Sid": "SameDepartmentManagerAccess",
  "Effect": "Allow",
  "Action": "employee:view-salary",
  "Resource": "api:employees:*",
  "Condition": {
    "StringEquals": {
      "resource:Department": "${user:Department}",
      "user:Role": "manager"
    }
  }
}
```

### Ví Dụ 4: Complex Variable Substitution

```json
{
  "Resource": "api:projects:${user:Department}/${user:Team}/docs/*",
  "Condition": {
    "StringEquals": {
      "resource:OwnerDepartment": "${user:Department}",
      "resource:OwnerTeam": "${user:Team}"
    }
  }
}
```

**Lưu ý quan trọng:**
- Variables được resolve **tại runtime** khi evaluate policy
- Nếu attribute không tồn tại, variable sẽ được thay bằng chuỗi rỗng `""`
- Variable substitution hoạt động với cả **nested attributes** (e.g., `${user:organization.name}`)

---

## Xử Lý Giá Trị và Type Conversion

### Case Sensitivity

**Operator Names:**
- Operators **KHÔNG phân biệt** hoa thường (case-insensitive)
- Ví dụ: `StringEquals`, `stringequals`, `STRINGEQUALS` đều hợp lệ

```json
// Tất cả đều hợp lệ
"Condition": {
  "StringEquals": { ... }      // OK
  "stringequals": { ... }      // OK
  "STRINGEQUALS": { ... }      // OK
}
```

**Attribute Values:**
- String comparison **phân biệt** hoa thường (case-sensitive)

```json
"Condition": {
  "StringEquals": {
    "user:Role": "Admin"     // Chỉ khớp với "Admin", không khớp "admin"
  }
}
```

### Type Conversion

Hệ thống tự động convert types khi cần:

**String Conversion:**
```go
nil        → ""
123        → "123"
true       → "true"
"hello"    → "hello"
```

**Numeric Conversion:**
```go
"123"      → 123
"123.45"   → 123.45
true       → 1
false      → 0
nil        → 0
```

**Boolean Conversion:**
```go
true       → true
"true"     → true
"1"        → true
1          → true
false      → false
"false"    → false
"0"        → false
0          → false
nil        → false
```

**Time Parsing:**

Hỗ trợ các formats:
```go
"2025-01-15T14:30:00Z"           // RFC3339
"2025-01-15 14:30:00"            // DateTime
"15:04"                          // Time of day (HH:MM)
"2025-01-15"                     // Date only
```

### Missing Attributes

**Khi attribute không tồn tại trong context:**

1. **String operators:**
   ```json
   "StringEquals": {
     "user:NonExistentField": "value"
   }
   ```
   - Context value = `nil` → converted to `""`
   - Result: `"" != "value"` → **false** (condition fails)

2. **Numeric operators:**
   ```json
   "NumericGreaterThan": {
     "user:NonExistentField": 5
   }
   ```
   - Context value = `nil` → converted to `0`
   - Result: `0 > 5` → **false** (condition fails)

3. **Boolean operators:**
   ```json
   "Bool": {
     "user:NonExistentField": true
   }
   ```
   - Context value = `nil` → converted to `false`
   - Result: `false != true` → **false** (condition fails)

**Best Practice:** Luôn đảm bảo attributes được populate trong context trước khi evaluate.

### Array Values trong Conditions

**StringEquals với array sử dụng OR logic:**

```json
"Condition": {
  "StringEquals": {
    "user:Role": ["admin", "manager", "supervisor"]
  }
}
```

Điều này có nghĩa: **Role phải là admin HOẶC manager HOẶC supervisor**

Tương đương với:
```json
"Condition": {
  "Or": [
    {"StringEquals": {"user:Role": "admin"}},
    {"StringEquals": {"user:Role": "manager"}},
    {"StringEquals": {"user:Role": "supervisor"}}
  ]
}
```

### Nested Value Access

**Dot notation cho nested objects:**

```json
// Context có cấu trúc:
{
  "user": {
    "attributes": {
      "department": "Engineering",
      "location": {
        "country": "VN",
        "city": "HCM"
      }
    }
  }
}

// Truy cập nested values:
"Condition": {
  "StringEquals": {
    "user.attributes.department": "Engineering",
    "user.attributes.location.country": "VN"
  }
}
```

**Fallback mechanism:**
1. Thử truy cập với dot notation: `user.attributes.department`
2. Nếu fail, thử convert sang colon: `user:attributes:department`
3. Nếu fail, thử structured access: `user` → `attributes` → `department`

---

## Troubleshooting

### Common Issues

#### 1. Condition Không Hoạt Động Như Mong Đợi

**Problem:** Policy không match mặc dù tưởng là đúng

**Debugging steps:**

```json
// ❌ SAI - Thiếu context key prefix
"Condition": {
  "StringEquals": {
    "Department": "Engineering"    // Missing prefix
  }
}

// ✅ ĐÚNG
"Condition": {
  "StringEquals": {
    "user:Department": "Engineering"  // With prefix
  }
}
```

**Checklist:**
- ✅ Có dùng đúng prefix không? (`user:`, `resource:`, `environment:`, `request:`)
- ✅ Attribute có tồn tại trong context không?
- ✅ Type có đúng không? (string vs number vs boolean)
- ✅ Case sensitivity có đúng không?

#### 2. Variable Substitution Không Hoạt Động

**Problem:** Variable không được thay thế

```json
// ❌ SAI - Syntax sai
"Resource": "api:documents:owner-{request:UserId}"

// ✅ ĐÚNG
"Resource": "api:documents:owner-${request:UserId}"
```

**Checklist:**
- ✅ Dùng đúng syntax `${...}` (không phải `{...}`)
- ✅ Attribute key có đúng không?
- ✅ Attribute có tồn tại trong context không?

#### 3. Time-Based Conditions Không Chính Xác

**Problem:** Time conditions không match

```json
// ❌ SAI - Format không đúng
"Condition": {
  "DateGreaterThan": {
    "request:TimeOfDay": "9:00 AM"
  }
}

// ✅ ĐÚNG
"Condition": {
  "DateGreaterThan": {
    "request:TimeOfDay": "09:00:00"
  }
}
```

**Supported formats:**
- Time of day: `"15:04"` hoặc `"15:04:05"`
- Date: `"2025-01-15"`
- DateTime: `"2025-01-15T14:30:00Z"` (RFC3339)

#### 4. Array Conditions Không Hoạt Động

**Problem:** ArrayContains không tìm thấy giá trị

```json
// Context:
{
  "user:Roles": "admin"  // ❌ String, không phải array
}

// Condition:
"ArrayContains": {
  "user:Roles": "admin"
}
```

**Solution:** Đảm bảo attribute là array trong context:

```json
{
  "user:Roles": ["admin", "user"]  // ✅ Array
}
```

#### 5. IP Address Conditions

**Problem:** IP check không hoạt động

```json
// ❌ SAI - Thiếu CIDR notation
"Condition": {
  "IpAddress": {
    "environment:client_ip": "192.168.1.100"
  }
}

// ✅ ĐÚNG - Với CIDR
"Condition": {
  "IpAddress": {
    "environment:client_ip": "192.168.1.100/32"
  }
}

// ✅ ĐÚNG - Range
"Condition": {
  "IpAddress": {
    "environment:client_ip": "192.168.1.0/24"
  }
}
```

### Testing Conditions

**Ví dụ test case:**

```json
{
  "test_cases": [
    {
      "name": "Manager can approve medium transactions",
      "request": {
        "subject_id": "user-123",
        "action": "transaction:approve",
        "resource_id": "api:transactions:tx-456",
        "context": {
          "transaction:Amount": 250000
        }
      },
      "subject_attributes": {
        "Role": "manager",
        "Department": "Finance"
      },
      "expected_result": "permit"
    },
    {
      "name": "Regular employee cannot approve medium transactions",
      "request": {
        "subject_id": "user-789",
        "action": "transaction:approve",
        "resource_id": "api:transactions:tx-456",
        "context": {
          "transaction:Amount": 250000
        }
      },
      "subject_attributes": {
        "Role": "employee",
        "Department": "Finance"
      },
      "expected_result": "deny"
    }
  ]
}
```

### Debug Logging

Để debug conditions, kiểm tra logs:

```
Debug: Enhanced condition evaluation failed for conditions: map[...]
Warning: Missing essential context key: request:Action
Info: UserId not provided in context
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
- Kiểm tra MFA cho sensitive operations
- Restrict external access với IP checks
- Implement time-based restrictions cho critical actions

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

### 11. Common Mistakes và Cách Tránh

#### ❌ Quên prefix cho context keys
```json
// SAI
"StringEquals": { "Department": "Engineering" }

// ĐÚNG
"StringEquals": { "user:Department": "Engineering" }
```

#### ❌ Dùng sai wildcard syntax
```json
// SAI - SQL LIKE syntax
"StringLike": { "user:Email": "%@company.com" }

// ĐÚNG - Dùng *
"StringLike": { "user:Email": "*@company.com" }
```

#### ❌ Quên CIDR notation cho IP
```json
// SAI
"IpAddress": { "environment:client_ip": "192.168.1.100" }

// ĐÚNG
"IpAddress": { "environment:client_ip": "192.168.1.100/32" }
```

#### ❌ Nhầm lẫn AND vs OR logic
```json
// Trong cùng một operator = AND
"StringEquals": {
  "user:Department": "Engineering",  // AND
  "user:Role": "admin"                // AND
}

// Muốn OR thì phải dùng array hoặc Or operator
"StringEquals": {
  "user:Role": ["admin", "manager"]  // OR
}
```

---

## Common Use Cases - Cheat Sheet

### Use Case 1: Ownership Check
```json
"Condition": {
  "StringEquals": {
    "resource:Owner": "${request:UserId}"
  }
}
```

### Use Case 2: Department Access
```json
"Condition": {
  "StringEquals": {
    "user:Department": "${resource:Department}"
  }
}
```

### Use Case 3: Business Hours Only
```json
"Condition": {
  "TimeBetween": {
    "environment:time_of_day": ["09:00", "18:00"]
  },
  "DayOfWeek": {
    "environment:day_of_week": ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]
  }
}
```

### Use Case 4: Internal Network Only
```json
"Condition": {
  "IsInternalIP": {
    "environment:is_internal_ip": true
  }
}
```

### Use Case 5: MFA Required for Sensitive Actions
```json
"Condition": {
  "Bool": {
    "user:MFAEnabled": true
  }
}
```

### Use Case 6: Role-Based with Level Check
```json
"Condition": {
  "StringEquals": {
    "user:Role": ["manager", "director"]
  },
  "NumericGreaterThanEquals": {
    "user:Level": 5
  }
}
```

### Use Case 7: Amount-Based Approval
```json
"Condition": {
  "NumericBetween": {
    "transaction:Amount": [100000, 1000000]
  },
  "StringEquals": {
    "user:Role": "manager"
  }
}
```

### Use Case 8: Geo-Restriction
```json
"Condition": {
  "StringEquals": {
    "environment:country": ["VN", "SG", "TH"]
  }
}
```

### Use Case 9: Non-Confidential Access
```json
"Condition": {
  "StringNotEquals": {
    "resource:Sensitivity": ["confidential", "top-secret"]
  }
}
```

### Use Case 10: Multi-Factor Checks
```json
"Condition": {
  "And": [
    {
      "StringEquals": {
        "user:Role": "admin"
      }
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
    }
  ]
}
```

---

## Summary

### Key Takeaways

1. **Operators Coverage:**
   - ✅ 7 String operators (equals, like, contains, regex, v.v.)
   - ✅ 7 Numeric operators (comparison, between)
   - ✅ 1 Boolean operator
   - ✅ 8 Date/Time operators (comparison, business hours, day of week)
   - ✅ 4 Network/IP operators (range checks, internal IP)
   - ✅ 3 Array operators (contains, size)
   - ✅ 3 Logic operators (and, or, not)

2. **Context Keys:**
   - `request:*` - Request information
   - `user:*` - User/Subject attributes
   - `resource:*` - Resource attributes
   - `environment:*` - Environmental context

3. **Advanced Features:**
   - ✅ Variable substitution với `${...}`
   - ✅ Nested value access với dot notation
   - ✅ Automatic type conversion
   - ✅ Array values với OR logic
   - ✅ Case-insensitive operators

4. **Best Practices:**
   - ✅ Luôn dùng prefix cho context keys
   - ✅ Test thoroughly với multiple scenarios
   - ✅ Dùng Deny cho security-critical rules
   - ✅ Document policies với Sid và description
   - ✅ Avoid overly complex conditions

---

## Tham Khảo Thêm

### Related Documentation
- [ACTION_FIELD_GUIDE.md](./ACTION_FIELD_GUIDE.md) - Hướng dẫn chi tiết về Action field
- [RESOURCE_FIELD_GUIDE.md](./RESOURCE_FIELD_GUIDE.md) - Hướng dẫn chi tiết về Resource field

### Code Implementation
Xem chi tiết implementation tại:
- `evaluator/conditions/conditions.go` - Condition evaluator cơ bản với traditional operators
- `evaluator/conditions/enhanced_condition_evaluator.go` - Enhanced evaluator với advanced operators
- `evaluator/core/pdp.go` - Policy Decision Point với full evaluation logic

### Example Policies
- `policy_examples_corrected.json` - Các ví dụ policy được validate

---

## Changelog

**Version 1.1.0** (2025-10-25)
- Updated cho enhanced condition evaluator architecture
- Cập nhật file paths theo cấu trúc package mới
- Thêm thông tin về EnhancedConditionEvaluator
- Cập nhật examples với enhanced operators
- Improved troubleshooting guide

---

**Lưu Ý:**
- Tài liệu này dựa trên code logic hiện tại của hệ thống ABAC
- Nếu có thắc mắc về implementation cụ thể, vui lòng tham khảo source code
- Để report issues hoặc contribute, liên hệ team development

**Happy Policy Writing! 🚀**
