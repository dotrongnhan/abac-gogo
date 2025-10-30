package models

import (
	"fmt"
	"strings"
	"time"
)

const (
	maxAttributeMapSize = 100
)

// UserSubject implements SubjectInterface for user-based authentication
// It wraps a User entity and provides ABAC attributes from relational data
type UserSubject struct {
	User    *User
	Profile *UserProfile
	Roles   []Role
}

// NewUserSubject creates a new UserSubject with all necessary data preloaded
func NewUserSubject(user *User, profile *UserProfile, roles []Role) *UserSubject {
	if user == nil {
		return nil
	}

	return &UserSubject{
		User:    user,
		Profile: profile,
		Roles:   roles,
	}
}

// GetID returns the user's unique identifier
func (us *UserSubject) GetID() string {
	if us.User == nil {
		return ""
	}
	return us.User.ID
}

// GetType returns the subject type as "user"
func (us *UserSubject) GetType() SubjectType {
	return SubjectTypeUser
}

// GetDisplayName returns the user's full name
func (us *UserSubject) GetDisplayName() string {
	if us.User == nil {
		return ""
	}
	return us.User.FullName
}

// IsActive returns whether the user is currently active
func (us *UserSubject) IsActive() bool {
	if us.User == nil {
		return false
	}
	return strings.ToLower(us.User.Status) == "active"
}

// GetAttributes returns all ABAC attributes as a flat map
// This maps relational user data to the flat attribute structure expected by policies
func (us *UserSubject) GetAttributes() map[string]interface{} {
	if us.User == nil {
		return make(map[string]interface{})
	}

	return us.MapToAttributes()
}

// MapToAttributes implements AttributeMapper interface
// Converts user, profile, and role data into flat ABAC attributes
func (us *UserSubject) MapToAttributes() map[string]interface{} {
	attributes := make(map[string]interface{}, maxAttributeMapSize)

	// Core user attributes
	us.addCoreUserAttributes(attributes)

	// Profile-based attributes (company, department, position)
	us.addProfileAttributes(attributes)

	// Role-based attributes
	us.addRoleAttributes(attributes)

	// Custom metadata attributes
	us.addMetadataAttributes(attributes)

	return attributes
}

// addCoreUserAttributes adds basic user information to attributes
func (us *UserSubject) addCoreUserAttributes(attributes map[string]interface{}) {
	if us.User == nil {
		return
	}

	attributes["user_id"] = us.User.ID
	attributes["username"] = us.User.Username
	attributes["email"] = us.User.Email
	attributes["full_name"] = us.User.FullName
	attributes["status"] = us.User.Status
	attributes["subject_type"] = string(SubjectTypeUser)

	if us.User.EmployeeID != "" {
		attributes["employee_id"] = us.User.EmployeeID
	}

	if us.User.HireDate != nil {
		attributes["hire_date"] = us.User.HireDate.Format("2006-01-02")
		attributes["tenure_years"] = calculateTenureYears(us.User.HireDate)
	}

	if us.User.TerminationDate != nil {
		attributes["termination_date"] = us.User.TerminationDate.Format("2006-01-02")
		attributes["is_terminated"] = true
	} else {
		attributes["is_terminated"] = false
	}
}

// addProfileAttributes adds organizational attributes from user profile
func (us *UserSubject) addProfileAttributes(attributes map[string]interface{}) {
	if us.Profile == nil {
		return
	}

	// Company attributes
	if us.Profile.Company != nil {
		attributes["company_id"] = us.Profile.Company.ID
		attributes["company_name"] = us.Profile.Company.CompanyName
		attributes["company_code"] = us.Profile.Company.CompanyCode
		if us.Profile.Company.Industry != "" {
			attributes["industry"] = us.Profile.Company.Industry
		}
		if us.Profile.Company.Country != "" {
			attributes["country"] = us.Profile.Company.Country
		}
	}

	// Department attributes
	if us.Profile.Department != nil {
		attributes["department_id"] = us.Profile.Department.ID
		attributes["department"] = us.Profile.Department.DepartmentName
		attributes["department_code"] = us.Profile.Department.DepartmentCode
		if us.Profile.Department.CostCenter != "" {
			attributes["cost_center"] = us.Profile.Department.CostCenter
		}
	}

	// Position attributes
	if us.Profile.Position != nil {
		attributes["position_id"] = us.Profile.Position.ID
		attributes["position"] = us.Profile.Position.PositionName
		attributes["position_code"] = us.Profile.Position.PositionCode
		attributes["position_level"] = us.Profile.Position.PositionLevel
		if us.Profile.Position.PositionCategory != "" {
			attributes["position_category"] = us.Profile.Position.PositionCategory
		}
		if us.Profile.Position.ClearanceLevel != "" {
			attributes["position_clearance"] = us.Profile.Position.ClearanceLevel
		}
	}

	// Manager attributes
	if us.Profile.ManagerID != nil {
		attributes["manager_id"] = *us.Profile.ManagerID
		attributes["has_manager"] = true
		if us.Profile.Manager != nil {
			attributes["manager_name"] = us.Profile.Manager.FullName
		}
	} else {
		attributes["has_manager"] = false
	}

	// Location attributes
	if us.Profile.Location != "" {
		attributes["location"] = us.Profile.Location
	}
	if us.Profile.OfficeLocation != "" {
		attributes["office_location"] = us.Profile.OfficeLocation
	}

	// Security and access attributes
	if us.Profile.SecurityClearance != "" {
		attributes["clearance"] = us.Profile.SecurityClearance
		attributes["security_clearance"] = us.Profile.SecurityClearance
	}
	attributes["access_level"] = us.Profile.AccessLevel

	// Custom profile attributes (from JSONB field)
	for key, value := range us.Profile.Attributes {
		// Prefix custom attributes to avoid collisions
		attributes["custom_"+key] = value
	}
}

// addRoleAttributes adds role-based attributes
func (us *UserSubject) addRoleAttributes(attributes map[string]interface{}) {
	if len(us.Roles) == 0 {
		attributes["roles"] = []string{}
		attributes["role_count"] = 0
		return
	}

	roleNames := make([]string, 0, len(us.Roles))
	roleCodes := make([]string, 0, len(us.Roles))
	roleTypes := make(map[string]bool)

	for _, role := range us.Roles {
		roleNames = append(roleNames, role.RoleName)
		roleCodes = append(roleCodes, role.RoleCode)
		roleTypes[role.RoleType] = true
	}

	attributes["roles"] = roleCodes
	attributes["role_names"] = roleNames
	attributes["role_count"] = len(us.Roles)

	// Add role type flags
	for roleType := range roleTypes {
		attributes[fmt.Sprintf("has_%s_role", roleType)] = true
	}

	// Helper method to check specific roles
	attributes["is_admin"] = us.hasRole("admin")
	attributes["is_manager"] = us.hasRole("manager")
	attributes["is_developer"] = us.hasRole("developer")
}

// addMetadataAttributes adds custom metadata from user entity
func (us *UserSubject) addMetadataAttributes(attributes map[string]interface{}) {
	if us.User == nil || len(us.User.Metadata) == 0 {
		return
	}

	// Add metadata with prefix to avoid collisions
	for key, value := range us.User.Metadata {
		attributes["metadata_"+key] = value
	}
}

// hasRole checks if the user has a specific role code
func (us *UserSubject) hasRole(roleCode string) bool {
	roleCodeLower := strings.ToLower(roleCode)
	for _, role := range us.Roles {
		if strings.ToLower(role.RoleCode) == roleCodeLower {
			return true
		}
	}
	return false
}

// HasRole is a public method to check if user has a specific role
func (us *UserSubject) HasRole(roleCode string) bool {
	return us.hasRole(roleCode)
}

// HasAnyRole checks if the user has any of the specified roles
func (us *UserSubject) HasAnyRole(roleCodes []string) bool {
	for _, roleCode := range roleCodes {
		if us.hasRole(roleCode) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the user has all of the specified roles
func (us *UserSubject) HasAllRoles(roleCodes []string) bool {
	for _, roleCode := range roleCodes {
		if !us.hasRole(roleCode) {
			return false
		}
	}
	return true
}

// GetDepartmentCode returns the user's department code
func (us *UserSubject) GetDepartmentCode() string {
	if us.Profile != nil && us.Profile.Department != nil {
		return us.Profile.Department.DepartmentCode
	}
	return ""
}

// GetPositionLevel returns the user's position level
func (us *UserSubject) GetPositionLevel() int {
	if us.Profile != nil && us.Profile.Position != nil {
		return us.Profile.Position.PositionLevel
	}
	return 0
}

// GetSecurityClearance returns the user's security clearance
func (us *UserSubject) GetSecurityClearance() string {
	if us.Profile != nil {
		return us.Profile.SecurityClearance
	}
	return ""
}

// calculateTenureYears calculates years of tenure from hire date
func calculateTenureYears(hireDate *time.Time) int {
	if hireDate == nil {
		return 0
	}
	// Simple calculation: current year - hire year
	// For production, consider using a more sophisticated calculation
	return time.Now().Year() - hireDate.Year()
}
