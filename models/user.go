package models

import (
	"time"
)

// Company represents an organizational company
type Company struct {
	ID          string    `json:"id" gorm:"primaryKey;size:255"`
	CompanyCode string    `json:"company_code" gorm:"size:100;not null;uniqueIndex"`
	CompanyName string    `json:"company_name" gorm:"size:255;not null"`
	Industry    string    `json:"industry,omitempty" gorm:"size:100"`
	Country     string    `json:"country,omitempty" gorm:"size:100"`
	Status      string    `json:"status" gorm:"size:50;not null;default:'active';index"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for Company
func (Company) TableName() string {
	return "companies"
}

// Department represents an organizational department
type Department struct {
	ID                 string       `json:"id" gorm:"primaryKey;size:255"`
	CompanyID          string       `json:"company_id" gorm:"size:255;not null;index"`
	DepartmentCode     string       `json:"department_code" gorm:"size:100;not null"`
	DepartmentName     string       `json:"department_name" gorm:"size:255;not null"`
	ParentDepartmentID *string      `json:"parent_department_id,omitempty" gorm:"size:255;index"`
	ManagerID          *string      `json:"manager_id,omitempty" gorm:"size:255"`
	CostCenter         string       `json:"cost_center,omitempty" gorm:"size:100"`
	Status             string       `json:"status" gorm:"size:50;not null;default:'active';index"`
	CreatedAt          time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	Company            *Company     `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	ParentDepartment   *Department  `json:"parent_department,omitempty" gorm:"foreignKey:ParentDepartmentID"`
	SubDepartments     []Department `json:"sub_departments,omitempty" gorm:"foreignKey:ParentDepartmentID"`
}

// TableName specifies the table name for Department
func (Department) TableName() string {
	return "departments"
}

// Position represents a job position or title
type Position struct {
	ID               string    `json:"id" gorm:"primaryKey;size:255"`
	PositionCode     string    `json:"position_code" gorm:"size:100;not null;uniqueIndex"`
	PositionName     string    `json:"position_name" gorm:"size:255;not null"`
	PositionLevel    int       `json:"position_level" gorm:"not null;default:1;index"`
	PositionCategory string    `json:"position_category,omitempty" gorm:"size:100"`
	ClearanceLevel   string    `json:"clearance_level,omitempty" gorm:"size:50;index"`
	Description      string    `json:"description,omitempty" gorm:"type:text"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for Position
func (Position) TableName() string {
	return "positions"
}

// Role represents a functional role for RBAC integration
type Role struct {
	ID          string    `json:"id" gorm:"primaryKey;size:255"`
	RoleCode    string    `json:"role_code" gorm:"size:100;not null;uniqueIndex"`
	RoleName    string    `json:"role_name" gorm:"size:255;not null"`
	RoleType    string    `json:"role_type" gorm:"size:50;not null;default:'functional';index"`
	Description string    `json:"description,omitempty" gorm:"type:text"`
	IsSystem    bool      `json:"is_system" gorm:"default:false;index"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for Role
func (Role) TableName() string {
	return "roles"
}

// User represents a core user entity
type User struct {
	ID              string       `json:"id" gorm:"primaryKey;size:255"`
	Username        string       `json:"username" gorm:"size:255;not null;uniqueIndex"`
	Email           string       `json:"email" gorm:"size:255;not null;uniqueIndex"`
	FullName        string       `json:"full_name" gorm:"size:255;not null"`
	Status          string       `json:"status" gorm:"size:50;not null;default:'active';index"`
	EmployeeID      string       `json:"employee_id,omitempty" gorm:"size:100;uniqueIndex"`
	HireDate        *time.Time   `json:"hire_date,omitempty" gorm:"type:date"`
	TerminationDate *time.Time   `json:"termination_date,omitempty" gorm:"type:date"`
	Metadata        JSONMap      `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt       time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	Profile         *UserProfile `json:"profile,omitempty" gorm:"foreignKey:UserID"`
	Roles           []Role       `json:"roles,omitempty" gorm:"many2many:user_roles;"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// UserProfile represents extended user information
type UserProfile struct {
	ID                string      `json:"id" gorm:"primaryKey;size:255"`
	UserID            string      `json:"user_id" gorm:"size:255;not null;uniqueIndex;index"`
	CompanyID         string      `json:"company_id" gorm:"size:255;not null;index"`
	DepartmentID      string      `json:"department_id" gorm:"size:255;not null;index"`
	PositionID        string      `json:"position_id" gorm:"size:255;not null;index"`
	ManagerID         *string     `json:"manager_id,omitempty" gorm:"size:255;index"`
	Location          string      `json:"location,omitempty" gorm:"size:255"`
	OfficeLocation    string      `json:"office_location,omitempty" gorm:"size:255"`
	PhoneNumber       string      `json:"phone_number,omitempty" gorm:"size:50"`
	MobileNumber      string      `json:"mobile_number,omitempty" gorm:"size:50"`
	EmergencyContact  JSONMap     `json:"emergency_contact,omitempty" gorm:"type:jsonb"`
	SecurityClearance string      `json:"security_clearance,omitempty" gorm:"size:50;index"`
	AccessLevel       int         `json:"access_level" gorm:"default:1;index"`
	Attributes        JSONMap     `json:"attributes,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt         time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
	Company           *Company    `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	Department        *Department `json:"department,omitempty" gorm:"foreignKey:DepartmentID"`
	Position          *Position   `json:"position,omitempty" gorm:"foreignKey:PositionID"`
	Manager           *User       `json:"manager,omitempty" gorm:"foreignKey:ManagerID"`
}

// TableName specifies the table name for UserProfile
func (UserProfile) TableName() string {
	return "user_profiles"
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	ID         string     `json:"id" gorm:"primaryKey;size:255"`
	UserID     string     `json:"user_id" gorm:"size:255;not null;index;uniqueIndex:idx_user_role"`
	RoleID     string     `json:"role_id" gorm:"size:255;not null;index;uniqueIndex:idx_user_role"`
	AssignedBy *string    `json:"assigned_by,omitempty" gorm:"size:255"`
	AssignedAt time.Time  `json:"assigned_at" gorm:"autoCreateTime"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" gorm:"index"`
	IsActive   bool       `json:"is_active" gorm:"default:true;index"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	User       *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Role       *Role      `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	Assigner   *User      `json:"assigner,omitempty" gorm:"foreignKey:AssignedBy"`
}

// TableName specifies the table name for UserRole
func (UserRole) TableName() string {
	return "user_roles"
}

// UserAttributeHistory tracks changes to user attributes for audit purposes
type UserAttributeHistory struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID        string    `json:"user_id" gorm:"size:255;not null;index"`
	AttributeName string    `json:"attribute_name" gorm:"size:255;not null"`
	OldValue      JSONMap   `json:"old_value,omitempty" gorm:"type:jsonb"`
	NewValue      JSONMap   `json:"new_value,omitempty" gorm:"type:jsonb"`
	ChangedBy     *string   `json:"changed_by,omitempty" gorm:"size:255"`
	ChangeReason  string    `json:"change_reason,omitempty" gorm:"type:text"`
	ChangedAt     time.Time `json:"changed_at" gorm:"autoCreateTime;index"`
	User          *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for UserAttributeHistory
func (UserAttributeHistory) TableName() string {
	return "user_attribute_history"
}
