# Complex Logical Conditions trong ABAC System

## Tổng quan

Hệ thống ABAC hiện đã hỗ trợ các điều kiện logic phức tạp với các toán tử `And`, `Or`, và `Not`. Điều này cho phép tạo ra các policy rules có logic phức tạp và nested conditions.

## Các Toán Tử Logic

### 1. AND Operator (`And`)
- **Mục đích**: Tất cả các điều kiện con phải đúng
- **Cú pháp**: Array của các conditions hoặc left/right format
- **Kết quả**: `true` nếu TẤT CẢ điều kiện con đều `true`

### 2. OR Operator (`Or`)
- **Mục đích**: Ít nhất một điều kiện con phải đúng
- **Cú pháp**: Array của các conditions hoặc left/right format  
- **Kết quả**: `true` nếu ÍT NHẤT MỘT điều kiện con là `true`

### 3. NOT Operator (`Not`)
- **Mục đích**: Đảo ngược kết quả của điều kiện con
- **Cú pháp**: Single condition hoặc operand format
- **Kết quả**: `true` nếu điều kiện con là `false`

## Cú Pháp và Định Dạng

### 1. Array Format (Khuyến nghị)

```json
{
  "And": [
    {
      "StringEquals": {
        "user.department": "engineering"
      }
    },
    {
      "NumericGreaterThan": {
        "user.level": 5
      }
    }
  ]
}
```

### 2. Object Format (Traditional)

```json
{
  "Or": {
    "StringEquals": {
      "user.role": "admin"
    },
    "StringEquals": {
      "user.department": "security"
    }
  }
}
```

### 3. ComplexCondition Struct (Programmatic)

```go
condition := &ComplexCondition{
    Type:     "logical",
    Operator: ConditionAnd,
    Left: &ComplexCondition{
        Type:     "simple",
        Operator: ConditionStringEquals,
        Key:      "user.department",
        Value:    "engineering",
    },
    Right: &ComplexCondition{
        Type:     "simple",
        Operator: ConditionNumericGreaterThan,
        Key:      "user.level",
        Value:    5,
    },
}
```

## Ví Dụ Thực Tế

### 1. Điều Kiện Đơn Giản

#### AND - Cả hai điều kiện phải đúng
```json
{
  "And": [
    {
      "StringEquals": {
        "user.department": "engineering"
      }
    },
    {
      "NumericGreaterThanEquals": {
        "user.level": 5
      }
    }
  ]
}
```

**Kết quả**: `true` nếu user thuộc department "engineering" VÀ có level >= 5

#### OR - Một trong hai điều kiện đúng
```json
{
  "Or": [
    {
      "StringEquals": {
        "user.role": "admin"
      }
    },
    {
      "StringEquals": {
        "user.department": "security"
      }
    }
  ]
}
```

**Kết quả**: `true` nếu user có role "admin" HOẶC thuộc department "security"

#### NOT - Đảo ngược điều kiện
```json
{
  "Not": {
    "Bool": {
      "user.on_probation": true
    }
  }
}
```

**Kết quả**: `true` nếu user KHÔNG trong thời gian thử việc

### 2. Nested Conditions (Lồng nhau)

```json
{
  "And": [
    {
      "Or": [
        {
          "StringEquals": {
            "user.department": "engineering"
          }
        },
        {
          "StringEquals": {
            "user.department": "security"
          }
        }
      ]
    },
    {
      "NumericGreaterThan": {
        "user.level": 3
      }
    },
    {
      "Not": {
        "Bool": {
          "user.on_probation": true
        }
      }
    }
  ]
}
```

**Logic**: (department = "engineering" OR department = "security") AND level > 3 AND NOT on_probation

### 3. Multiple Levels of Nesting

```json
{
  "Or": [
    {
      "And": [
        {
          "StringEquals": {
            "user.role": "admin"
          }
        },
        {
          "IpAddress": {
            "request.sourceIp": ["10.0.0.0/8", "192.168.0.0/16"]
          }
        }
      ]
    },
    {
      "And": [
        {
          "StringEquals": {
            "user.department": "engineering"
          }
        },
        {
          "NumericGreaterThan": {
            "user.level": 7
          }
        },
        {
          "Not": {
            "Bool": {
              "user.on_probation": true
            }
          }
        }
      ]
    }
  ]
}
```

**Logic**: 
- (role = "admin" AND IP trong internal network) 
- OR 
- (department = "engineering" AND level > 7 AND NOT on_probation)

## Use Cases Thực Tế

### 1. Access Control cho Sensitive Data

```json
{
  "And": [
    {
      "Or": [
        {
          "StringEquals": {
            "user.role": "admin"
          }
        },
        {
          "StringEquals": {
            "user.clearance_level": "top_secret"
          }
        }
      ]
    },
    {
      "IpAddress": {
        "request.sourceIp": ["10.0.0.0/8"]
      }
    },
    {
      "Not": {
        "Bool": {
          "user.account_locked": true
        }
      }
    }
  ]
}
```

### 2. Time-based Access với Multiple Conditions

```json
{
  "And": [
    {
      "StringEquals": {
        "user.department": "finance"
      }
    },
    {
      "Or": [
        {
          "And": [
            {
              "NumericGreaterThanEquals": {
                "time.hour": 9
              }
            },
            {
              "NumericLessThanEquals": {
                "time.hour": 17
              }
            }
          ]
        },
        {
          "StringEquals": {
            "user.role": "manager"
          }
        }
      ]
    }
  ]
}
```

**Logic**: Finance department users có thể access trong giờ hành chính (9-17h) HOẶC nếu là manager thì access bất kỳ lúc nào.

### 3. Resource-based Access với Complex Logic

```json
{
  "Or": [
    {
      "StringEquals": {
        "resource.owner": "${user.id}"
      }
    },
    {
      "And": [
        {
          "StringEquals": {
            "user.department": "${resource.department}"
          }
        },
        {
          "NumericGreaterThanEquals": {
            "user.level": 5
          }
        }
      ]
    },
    {
      "And": [
        {
          "StringEquals": {
            "user.role": "admin"
          }
        },
        {
          "Not": {
            "StringEquals": {
              "resource.classification": "top_secret"
            }
          }
        }
      ]
    }
  ]
}
```

**Logic**: User có thể access resource nếu:
- Là owner của resource, HOẶC
- Cùng department và có level >= 5, HOẶC  
- Là admin và resource không phải top_secret

## Best Practices

### 1. Cấu Trúc Conditions

- **Sử dụng array format** cho AND/OR operators để dễ đọc
- **Giới hạn độ sâu nesting** (khuyến nghị tối đa 3-4 levels)
- **Đặt điều kiện đơn giản trước** để optimize performance
- **Sử dụng NOT một cách thận trọng** để tránh logic phức tạp

### 2. Performance Optimization

```json
{
  "And": [
    {
      "StringEquals": {
        "user.department": "engineering"
      }
    },
    {
      "Or": [
        {
          "NumericGreaterThan": {
            "user.level": 7
          }
        },
        {
          "StringEquals": {
            "user.role": "senior"
          }
        }
      ]
    }
  ]
}
```

**Lý do**: Đặt điều kiện department trước vì nó có thể filter nhanh nhiều users.

### 3. Readability và Maintainability

```json
{
  "And": [
    {
      "StringEquals": {
        "user.department": "engineering"
      }
    },
    {
      "Not": {
        "Bool": {
          "user.on_probation": true
        }
      }
    },
    {
      "Or": [
        {
          "NumericGreaterThanEquals": {
            "user.level": 5
          }
        },
        {
          "StringEquals": {
            "user.role": "team_lead"
          }
        }
      ]
    }
  ]
}
```

**Tốt hơn là**:

```json
{
  "And": [
    {
      "StringEquals": {
        "user.department": "engineering"
      }
    },
    {
      "Bool": {
        "user.on_probation": false
      }
    },
    {
      "Or": [
        {
          "NumericGreaterThanEquals": {
            "user.level": 5
          }
        },
        {
          "StringEquals": {
            "user.role": "team_lead"
          }
        }
      ]
    }
  ]
}
```

## Error Handling và Validation

### 1. Invalid Conditions

Hệ thống sẽ validate:
- **Structure**: Array/Object format đúng
- **Operators**: Các operators hợp lệ
- **Nesting**: Không có circular references
- **Types**: Data types phù hợp với operators

### 2. Edge Cases

- **Empty arrays**: `And: []` → `true`, `Or: []` → `false`
- **Invalid conditions**: Bỏ qua và log warning
- **Missing context**: Trả về `false` cho missing attributes

## Testing Complex Conditions

### 1. Unit Tests

```go
func TestComplexLogicalConditions(t *testing.T) {
    ce := NewConditionEvaluator()
    
    condition := map[string]interface{}{
        "And": []interface{}{
            map[string]interface{}{
                "StringEquals": map[string]interface{}{
                    "user.department": "engineering",
                },
            },
            map[string]interface{}{
                "Or": []interface{}{
                    map[string]interface{}{
                        "NumericGreaterThan": map[string]interface{}{
                            "user.level": 5,
                        },
                    },
                    map[string]interface{}{
                        "StringEquals": map[string]interface{}{
                            "user.role": "senior",
                        },
                    },
                },
            },
        },
    }
    
    context := map[string]interface{}{
        "user": map[string]interface{}{
            "department": "engineering",
            "level":      7,
            "role":       "developer",
        },
    }
    
    result := ce.Evaluate(condition, context)
    assert.True(t, result)
}
```

### 2. Integration Tests

Test với real policy data và complex scenarios để đảm bảo performance và correctness.

## Migration từ Simple Conditions

### Before (Simple)
```json
{
  "StringEquals": {
    "user.department": "engineering"
  },
  "NumericGreaterThan": {
    "user.level": 5
  }
}
```

### After (Complex)
```json
{
  "And": [
    {
      "StringEquals": {
        "user.department": "engineering"
      }
    },
    {
      "NumericGreaterThan": {
        "user.level": 5
      }
    }
  ]
}
```

**Lưu ý**: Hệ thống vẫn backward compatible với simple conditions format.
