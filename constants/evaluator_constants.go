package constants

import "time"

// Operator string constants
const (
	// String operators
	OpStringEquals     = "stringequals"
	OpStringNotEquals  = "stringnotequals"
	OpStringLike       = "stringlike"
	OpStringContains   = "stringcontains"
	OpStringStartsWith = "stringstartswith"
	OpStringEndsWith   = "stringendswith"
	OpStringRegex      = "stringregex"

	// Numeric operators
	OpNumericEquals            = "numericequals"
	OpNumericNotEquals         = "numericnotequals"
	OpNumericLessThan          = "numericlessthan"
	OpNumericLessThanEquals    = "numericlessthanequals"
	OpNumericGreaterThan       = "numericgreaterthan"
	OpNumericGreaterThanEquals = "numericgreaterthanequals"
	OpNumericBetween           = "numericbetween"

	// Date/Time operators
	OpDateLessThan          = "datelessthan"
	OpTimeLessThan          = "timelessthan"
	OpDateLessThanEquals    = "datelessthanequals"
	OpTimeLessThanEquals    = "timelessthanequals"
	OpDateGreaterThan       = "dategreaterthan"
	OpTimeGreaterThan       = "timegreaterthan"
	OpDateGreaterThanEquals = "dategreaterthanequals"
	OpTimeGreaterThanEquals = "timegreaterthanequals"
	OpDateBetween           = "datebetween"
	OpTimeBetween           = "timebetween"
	OpDayOfWeek             = "dayofweek"
	OpTimeOfDay             = "timeofday"
	OpIsBusinessHours       = "isbusinesshours"

	// Array operators
	OpArrayContains    = "arraycontains"
	OpArrayNotContains = "arraynotcontains"
	OpArraySize        = "arraysize"

	// Network operators
	OpIPInRange    = "ipinrange"
	OpIPNotInRange = "ipnotinrange"
	OpIsInternalIP = "isinternalip"

	// Boolean operators
	OpBool    = "bool"
	OpBoolean = "boolean"

	// Logical operators
	OpAnd = "and"
	OpOr  = "or"
	OpNot = "not"
)

// Time format constants
const (
	TimeFormatHourMinute = "15:04"
	TimeFormatDate       = "2006-01-02"
	TimeFormatDateTime   = "2006-01-02 15:04:05"
	TimeFormatISO        = "2006-01-02T15:04:05Z"
)

// Day of week constants
const (
	DaySunday    = "sunday"
	DayMonday    = "monday"
	DayTuesday   = "tuesday"
	DayWednesday = "wednesday"
	DayThursday  = "thursday"
	DayFriday    = "friday"
	DaySaturday  = "saturday"
)

// Weekday number constants
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

// Boolean string constants
const (
	BoolStringTrue  = "true"
	BoolStringOne   = "1"
	BoolStringFalse = "false"
	BoolStringZero  = "0"
)

// Range constants
const (
	RangeKeyMin = "min"
	RangeKeyMax = "max"
)

// Array size operator constants
const (
	SizeOpEquals                = "eq"
	SizeOpEqualsLong            = "equals"
	SizeOpGreaterThan           = "gt"
	SizeOpGreaterThanLong       = "greaterthan"
	SizeOpGreaterThanEquals     = "gte"
	SizeOpGreaterThanEqualsLong = "greaterthanequals"
	SizeOpLessThan              = "lt"
	SizeOpLessThanLong          = "lessthan"
	SizeOpLessThanEquals        = "lte"
	SizeOpLessThanEqualsLong    = "lessthanequals"
)

// Context key constants
const (
	ContextKeyEnvironmentHour      = "environment.hour"
	ContextKeyEnvironmentDayOfWeek = "environment.day_of_week"
)

// Default values
const (
	DefaultEmptyString = ""
	DefaultZeroFloat   = 0.0
	DefaultZeroInt     = 0
	DefaultFalse       = false
	DefaultTrue        = true
)

// GetAllTimeFormats returns all supported time formats in order of preference
func GetAllTimeFormats() []string {
	return []string{
		time.RFC3339,
		TimeFormatISO,
		TimeFormatDateTime,
		TimeFormatHourMinute,
		TimeFormatDate,
	}
}

// GetDayOfWeekNumber converts day string to weekday number
func GetDayOfWeekNumber(dayStr string) int {
	switch dayStr {
	case DaySunday:
		return WeekdaySunday
	case DayMonday:
		return WeekdayMonday
	case DayTuesday:
		return WeekdayTuesday
	case DayWednesday:
		return WeekdayWednesday
	case DayThursday:
		return WeekdayThursday
	case DayFriday:
		return WeekdayFriday
	case DaySaturday:
		return WeekdaySaturday
	default:
		return WeekdayInvalid
	}
}
