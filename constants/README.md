# Constants Package

Package constants cung cấp tất cả system constants được sử dụng trong ABAC system, đảm bảo zero hardcoded values và type safety.

## Overview

Constants package được tổ chức theo functional areas và tuân thủ nguyên tắc **Zero Hardcoded Values**. Tất cả string literals, magic numbers, và configuration values được centralized tại đây.

## Files Structure

```
constants/
├── business_rules.go       # Business logic constants
├── condition_operators.go  # Legacy operator constants  
├── context_keys.go         # Context key definitions
├── policy_constants.go     # Policy-related constants
└── evaluator_constants.go  # Evaluator constants (NEW)
```

## Evaluator Constants (evaluator_constants.go)

### Operator Constants

Tất cả condition operators được định nghĩa như constants để avoid hardcoded strings:

#### String Operators
```go
const (
    OpStringEquals     = "stringequals"
    OpStringNotEquals  = "stringnotequals"
    OpStringLike       = "stringlike"
    OpStringContains   = "stringcontains"
    OpStringStartsWith = "stringstartswith"
    OpStringEndsWith   = "stringendswith"
    OpStringRegex      = "stringregex"
)
```

#### Numeric Operators
```go
const (
    OpNumericEquals              = "numericequals"
    OpNumericNotEquals           = "numericnotequals"
    OpNumericLessThan            = "numericlessthan"
    OpNumericLessThanEquals      = "numericlessthanequals"
    OpNumericGreaterThan         = "numericgreaterthan"
    OpNumericGreaterThanEquals   = "numericgreaterthanequals"
    OpNumericBetween             = "numericbetween"
)
```

#### Time/Date Operators
```go
const (
    OpDateLessThan            = "datelessthan"
    OpTimeLessThan            = "timelessthan"
    OpDateLessThanEquals      = "datelessthanequals"
    OpTimeLessThanEquals      = "timelessthanequals"
    OpDateGreaterThan         = "dategreaterthan"
    OpTimeGreaterThan         = "timegreaterthan"
    OpDateGreaterThanEquals   = "dategreaterthanequals"
    OpTimeGreaterThanEquals   = "timegreaterthanequals"
    OpDateBetween             = "datebetween"
    OpTimeBetween             = "timebetween"
    OpDayOfWeek               = "dayofweek"
    OpTimeOfDay               = "timeofday"
    OpIsBusinessHours         = "isbusinesshours"
)
```

#### Array Operators
```go
const (
    OpArrayContains    = "arraycontains"
    OpArrayNotContains = "arraynotcontains"
    OpArraySize        = "arraysize"
)
```

#### Network Operators
```go
const (
    OpIPInRange      = "ipinrange"
    OpIPNotInRange   = "ipnotinrange"
    OpIsInternalIP   = "isinternalip"
)
```

#### Logical Operators
```go
const (
    OpAnd = "and"
    OpOr  = "or"
    OpNot = "not"
)
```

### Value Constants

#### Time Format Constants
```go
const (
    TimeFormatHourMinute = "15:04"
    TimeFormatDate       = "2006-01-02"
    TimeFormatDateTime   = "2006-01-02 15:04:05"
    TimeFormatISO        = "2006-01-02T15:04:05Z"
)
```

#### Day Constants
```go
const (
    DaySunday    = "sunday"
    DayMonday    = "monday"
    DayTuesday   = "tuesday"
    DayWednesday = "wednesday"
    DayThursday  = "thursday"
    DayFriday    = "friday"
    DaySaturday  = "saturday"
)
```

#### Weekday Number Constants
```go
const (
    WeekdaySunday    = 0
    WeekdayMonday    = 1
    WeekdayTuesday   = 2
    WeekdayWednesday = 3
    WeekdayThursday  = 4
    WeekdayFriday    = 5
    WeekdaySaturday  = 6
    WeekdayInvalid   = -1
)
```

#### Boolean String Constants
```go
const (
    BoolStringTrue  = "true"
    BoolStringOne   = "1"
    BoolStringFalse = "false"
    BoolStringZero  = "0"
)
```

#### Range Constants
```go
const (
    RangeKeyMin = "min"
    RangeKeyMax = "max"
)
```

#### Array Size Operator Constants
```go
const (
    SizeOpEquals             = "eq"
    SizeOpEqualsLong         = "equals"
    SizeOpGreaterThan        = "gt"
    SizeOpGreaterThanLong    = "greaterthan"
    SizeOpGreaterThanEquals  = "gte"
    SizeOpGreaterThanEqualsLong = "greaterthanequals"
    SizeOpLessThan           = "lt"
    SizeOpLessThanLong       = "lessthan"
    SizeOpLessThanEquals     = "lte"
    SizeOpLessThanEqualsLong = "lessthanequals"
)
```

#### Context Key Constants
```go
const (
    ContextKeyEnvironmentHour      = "environment.hour"
    ContextKeyEnvironmentDayOfWeek = "environment.day_of_week"
)
```

#### Default Value Constants
```go
const (
    DefaultEmptyString = ""
    DefaultZeroFloat   = 0.0
    DefaultZeroInt     = 0
    DefaultFalse       = false
    DefaultTrue        = true
)
```

### Helper Functions

#### GetAllTimeFormats()
Trả về tất cả supported time formats theo thứ tự ưu tiên:

```go
func GetAllTimeFormats() []string {
    return []string{
        time.RFC3339,
        TimeFormatISO,
        TimeFormatDateTime,
        TimeFormatHourMinute,
        TimeFormatDate,
    }
}
```

#### GetDayOfWeekNumber()
Convert day string thành weekday number:

```go
func GetDayOfWeekNumber(dayStr string) int {
    switch dayStr {
    case DaySunday:
        return WeekdaySunday
    case DayMonday:
        return WeekdayMonday
    // ... other cases
    default:
        return WeekdayInvalid
    }
}
```

## Usage Examples

### Basic Usage
```go
import "abac_go_example/constants"

// Instead of hardcoded strings
conditions := map[string]interface{}{
    constants.OpStringEquals: map[string]interface{}{
        "user.department": "engineering",
    },
    constants.OpNumericGreaterThan: map[string]interface{}{
        "user.level": 3,
    },
}
```

### Time Operations
```go
// Using time format constants
timeStr := time.Now().Format(constants.TimeFormatHourMinute)

// Using day constants
if dayStr == constants.DayMonday {
    // Monday logic
}

// Using helper function
weekdayNum := constants.GetDayOfWeekNumber(dayStr)
```

### Array Size Operations
```go
sizeCondition := map[string]interface{}{
    constants.SizeOpGreaterThan: 5,
}
```

### Range Operations
```go
rangeCondition := map[string]interface{}{
    constants.RangeKeyMin: 10,
    constants.RangeKeyMax: 100,
}
```

## Benefits

### 1. **Zero Hardcoded Values**
- Tất cả string literals được centralized
- Dễ dàng thay đổi values từ một nơi
- Reduced risk của typos và inconsistencies

### 2. **Type Safety**
- Constants có compile-time type checking
- IDE autocomplete support
- Refactoring safety

### 3. **Maintainability**
- Single source of truth cho all values
- Easy to update và maintain
- Clear documentation của all constants

### 4. **Performance**
- Compiler optimizations với constant folding
- Reduced memory allocation
- Faster string comparisons

### 5. **Documentation**
- Constants tự document code
- Clear naming conventions
- Grouped by functionality

## Best Practices

### 1. **Always Use Constants**
```go
// ✅ Good
case constants.OpStringEquals:

// ❌ Bad  
case "stringequals":
```

### 2. **Import Constants Package**
```go
import "abac_go_example/constants"
```

### 3. **Use Helper Functions**
```go
// ✅ Good
formats := constants.GetAllTimeFormats()

// ❌ Bad
formats := []string{"15:04", "2006-01-02", ...}
```

### 4. **Group Related Constants**
```go
// ✅ Good - grouped by functionality
const (
    OpStringEquals   = "stringequals"
    OpStringContains = "stringcontains"
)

// ❌ Bad - mixed functionality
const (
    OpStringEquals = "stringequals"
    OpNumericGreaterThan = "numericgreaterthan"
)
```

### 5. **Use Descriptive Names**
```go
// ✅ Good
const TimeFormatHourMinute = "15:04"

// ❌ Bad
const TimeFormat1 = "15:04"
```

## Migration from Hardcoded Values

### Before
```go
switch operator {
case "stringequals":
    // logic
case "numericgreaterthan":
    // logic
}

if strings.ToLower(v) == "true" || v == "1" {
    return true
}
```

### After
```go
switch operator {
case constants.OpStringEquals:
    // logic
case constants.OpNumericGreaterThan:
    // logic
}

if strings.ToLower(v) == constants.BoolStringTrue || v == constants.BoolStringOne {
    return true
}
```

## Testing

Constants package có comprehensive tests để ensure:
- All constants có correct values
- Helper functions work correctly
- No duplicate constants
- Consistent naming conventions

```bash
go test ./constants -v
```

## Future Enhancements

1. **Configuration Integration**: Load constants từ config files
2. **Validation Functions**: Add validation cho constant values
3. **Internationalization**: Support cho multiple languages
4. **Environment-specific**: Different constants cho different environments
