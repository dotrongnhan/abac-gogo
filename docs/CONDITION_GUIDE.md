# Condition Field Guide

## Tổng Quan

Field **Condition** trong Policy JSON cho phép thiết lập các điều kiện chi tiết để kiểm soát quyền truy cập dựa trên attributes của user, resource, environment và request context.

## Cách Hoạt Động

- Tất cả conditions trong cùng một block phải đồng thời thỏa mãn (AND logic)
- Nếu có bất kỳ condition nào không thỏa mãn, policy statement sẽ không được áp dụng
- Operator names không phân biệt hoa thường (case-insensitive)
- Attribute values có phân biệt hoa thường (case-sensitive)

## Context Keys

### User Context
- `user:id` - User ID
- `user:role` - User role
- `user:department` - Department
- `user:level` - User level/seniority

### Resource Context
- `resource:owner` - Resource owner
- `resource:type` - Resource type
- `resource:classification` - Security classification

### Environment Context
- `environment:time_of_day` - Current time
- `environment:day_of_week` - Current day
- `environment:client_ip` - Client IP address

### Request Context
- `request:method` - HTTP method
- `request:user_agent` - User agent string

## String Operators

### StringEquals
```json
{
  "StringEquals": {
    "user:role": "admin"
  }
}
```

### StringNotEquals
```json
{
  "StringNotEquals": {
    "user:status": "suspended"
  }
}
```

### StringLike (Pattern với *)
```json
{
  "StringLike": {
    "user:email": "*@company.com"
  }
}
```

### StringContains
```json
{
  "StringContains": {
    "user:groups": "engineering"
  }
}
```

### StringStartsWith
```json
{
  "StringStartsWith": {
    "resource:name": "public-"
  }
}
```

### StringEndsWith
```json
{
  "StringEndsWith": {
    "resource:name": "-prod"
  }
}
```

### StringRegex
```json
{
  "StringRegex": {
    "user:employee_id": "^EMP[0-9]{4}$"
  }
}
```

## Numeric Operators

### NumericEquals
```json
{
  "NumericEquals": {
    "user:level": 5
  }
}
```

### NumericGreaterThan
```json
{
  "NumericGreaterThan": {
    "user:level": 3
  }
}
```

### NumericLessThan
```json
{
  "NumericLessThan": {
    "resource:size": 1000000
  }
}
```

### NumericBetween
```json
{
  "NumericBetween": {
    "user:level": [3, 7]
  }
}
```

## Date/Time Operators

### TimeOfDay
```json
{
  "TimeOfDay": {
    "environment:time_of_day": "09:00-17:00"
  }
}
```

### DayOfWeek
```json
{
  "DayOfWeek": {
    "environment:day_of_week": ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]
  }
}
```

### IsBusinessHours
```json
{
  "IsBusinessHours": {
    "environment:is_business_hours": true
  }
}
```

### DateGreaterThan
```json
{
  "DateGreaterThan": {
    "user:start_date": "2024-01-01"
  }
}
```

## Network/IP Operators

### IPInRange
```json
{
  "IPInRange": {
    "environment:client_ip": ["10.0.0.0/8", "192.168.1.0/24"]
  }
}
```

### IsInternalIP
```json
{
  "IsInternalIP": {
    "environment:client_ip": true
  }
}
```

## Array Operators

### ArrayContains
```json
{
  "ArrayContains": {
    "user:roles": "admin"
  }
}
```

### ArrayNotContains
```json
{
  "ArrayNotContains": {
    "user:restricted_actions": "delete"
  }
}
```

### ArraySize
```json
{
  "ArraySize": {
    "user:projects": 5
  }
}
```

## Logical Operators

### And - Tất cả conditions phải true
```json
{
  "And": [
    {
      "StringEquals": {
        "user:department": "engineering"
      }
    },
    {
      "NumericGreaterThan": {
        "user:level": 3
      }
    }
  ]
}
```

### Or - Ít nhất một condition phải true
```json
{
  "Or": [
    {
      "StringEquals": {
        "user:role": "admin"
      }
    },
    {
      "StringEquals": {
        "user:role": "manager"
      }
    }
  ]
}
```

### Not - Đảo ngược kết quả
```json
{
  "Not": {
    "StringEquals": {
      "user:status": "suspended"
    }
  }
}
```

## Variable Substitution

### Dynamic Values
```json
{
  "StringEquals": {
    "resource:owner": "${user:id}"
  }
}
```

### Supported Variables
- `${user:attribute}` - User attributes
- `${resource:attribute}` - Resource attributes
- `${environment:attribute}` - Environment context
- `${request:attribute}` - Request context

## Ví Dụ Thực Tế

### 1. Business Hours Access
```json
{
  "And": [
    {
      "StringEquals": {
        "user:department": "finance"
      }
    },
    {
      "TimeOfDay": {
        "environment:time_of_day": "09:00-17:00"
      }
    },
    {
      "DayOfWeek": {
        "environment:day_of_week": ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]
      }
    }
  ]
}
```

### 2. Resource Owner Access
```json
{
  "Or": [
    {
      "StringEquals": {
        "resource:owner": "${user:id}"
      }
    },
    {
      "And": [
        {
          "StringEquals": {
            "user:department": "${resource:department}"
          }
        },
        {
          "NumericGreaterThanEquals": {
            "user:level": 5
          }
        }
      ]
    }
  ]
}
```

### 3. Network-based Restrictions
```json
{
  "And": [
    {
      "StringEquals": {
        "user:role": "admin"
      }
    },
    {
      "IPInRange": {
        "environment:client_ip": ["10.0.0.0/8", "192.168.1.0/24"]
      }
    },
    {
      "Not": {
        "StringEquals": {
          "resource:classification": "top_secret"
        }
      }
    }
  ]
}
```

### 4. Multi-factor Conditions
```json
{
  "And": [
    {
      "StringEquals": {
        "user:department": "engineering"
      }
    },
    {
      "Or": [
        {
          "NumericGreaterThan": {
            "user:level": 5
          }
        },
        {
          "ArrayContains": {
            "user:roles": "senior_developer"
          }
        }
      ]
    },
    {
      "Not": {
        "StringEquals": {
          "user:status": "probation"
        }
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

## Best Practices

### 1. Use Logical Grouping
```json
// ✅ Good - Clear logical structure
{
  "And": [
    {
      "StringEquals": {
        "user:department": "finance"
      }
    },
    {
      "Or": [
        {
          "StringEquals": {
            "user:role": "manager"
          }
        },
        {
          "NumericGreaterThan": {
            "user:level": 7
          }
        }
      ]
    }
  ]
}
```

### 2. Combine Time and User Conditions
```json
{
  "And": [
    {
      "StringEquals": {
        "user:role": "employee"
      }
    },
    {
      "TimeOfDay": {
        "environment:time_of_day": "09:00-17:00"
      }
    }
  ]
}
```

### 3. Use Variables for Dynamic Access
```json
{
  "StringEquals": {
    "resource:owner": "${user:id}"
  }
}
```

### 4. Network Security
```json
{
  "And": [
    {
      "StringEquals": {
        "user:role": "admin"
      }
    },
    {
      "IsInternalIP": {
        "environment:client_ip": true
      }
    }
  ]
}
```

## Common Patterns

### Own Resources Only
```json
{
  "StringEquals": {
    "resource:owner": "${user:id}"
  }
}
```

### Department Access
```json
{
  "StringEquals": {
    "user:department": "${resource:department}"
  }
}
```

### Business Hours Only
```json
{
  "And": [
    {
      "TimeOfDay": {
        "environment:time_of_day": "09:00-17:00"
      }
    },
    {
      "DayOfWeek": {
        "environment:day_of_week": ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]
      }
    }
  ]
}
```

### Senior Staff Access
```json
{
  "Or": [
    {
      "StringEquals": {
        "user:role": "manager"
      }
    },
    {
      "NumericGreaterThan": {
        "user:level": 5
      }
    }
  ]
}
```

## Troubleshooting

### Common Issues

1. **Condition không evaluate đúng**
   - Check attribute names (case-sensitive)
   - Verify context availability
   - Validate operator syntax

2. **Variable substitution fails**
   - Ensure variable syntax: `${context:attribute}`
   - Check if attribute exists in context
   - Verify variable scope

3. **Logical operators không hoạt động**
   - Check JSON structure
   - Verify array syntax for And/Or
   - Ensure proper nesting

### Debug Tips

1. **Test simple conditions trước**
2. **Verify context values**
3. **Check operator case sensitivity**
4. **Use logs để trace evaluation**

---

*Tài liệu này cung cấp hướng dẫn đầy đủ về Condition field cho người mới bắt đầu.*
