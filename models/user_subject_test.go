package models

import (
	"testing"
	"time"
)

func TestUserSubject_GetAttributes(t *testing.T) {
	// Setup test data
	hireDate := time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC)

	company := &Company{
		ID:          "company-001",
		CompanyCode: "TECH-001",
		CompanyName: "TechCorp",
		Industry:    "Technology",
		Country:     "USA",
	}

	department := &Department{
		ID:             "dept-001",
		DepartmentCode: "ENG",
		DepartmentName: "Engineering",
		Company:        company,
	}

	position := &Position{
		ID:             "pos-001",
		PositionCode:   "DEV-SR",
		PositionName:   "Senior Developer",
		PositionLevel:  3,
		ClearanceLevel: "confidential",
	}

	user := &User{
		ID:         "user-001",
		Username:   "john.doe",
		Email:      "john.doe@techcorp.com",
		FullName:   "John Doe",
		Status:     "active",
		EmployeeID: "EMP-001",
		HireDate:   &hireDate,
		Metadata:   JSONMap{"preferred_name": "John"},
	}

	managerID := "manager-001"
	profile := &UserProfile{
		ID:                "profile-001",
		UserID:            user.ID,
		CompanyID:         company.ID,
		DepartmentID:      department.ID,
		PositionID:        position.ID,
		ManagerID:         &managerID,
		Location:          "New York",
		SecurityClearance: "confidential",
		AccessLevel:       5,
		Company:           company,
		Department:        department,
		Position:          position,
		Attributes:        JSONMap{"custom_field": "value"},
	}

	roles := []Role{
		{
			ID:       "role-001",
			RoleCode: "developer",
			RoleName: "Developer",
			RoleType: "functional",
		},
		{
			ID:       "role-002",
			RoleCode: "reviewer",
			RoleName: "Code Reviewer",
			RoleType: "functional",
		},
	}

	// Create UserSubject
	userSubject := NewUserSubject(user, profile, roles)

	// Test GetID
	if userSubject.GetID() != user.ID {
		t.Errorf("GetID() = %v, want %v", userSubject.GetID(), user.ID)
	}

	// Test GetType
	if userSubject.GetType() != SubjectTypeUser {
		t.Errorf("GetType() = %v, want %v", userSubject.GetType(), SubjectTypeUser)
	}

	// Test GetDisplayName
	if userSubject.GetDisplayName() != user.FullName {
		t.Errorf("GetDisplayName() = %v, want %v", userSubject.GetDisplayName(), user.FullName)
	}

	// Test IsActive
	if !userSubject.IsActive() {
		t.Error("IsActive() = false, want true")
	}

	// Test GetAttributes
	attrs := userSubject.GetAttributes()

	// Verify core attributes
	if attrs["user_id"] != user.ID {
		t.Errorf("attributes[user_id] = %v, want %v", attrs["user_id"], user.ID)
	}

	if attrs["username"] != user.Username {
		t.Errorf("attributes[username] = %v, want %v", attrs["username"], user.Username)
	}

	if attrs["email"] != user.Email {
		t.Errorf("attributes[email] = %v, want %v", attrs["email"], user.Email)
	}

	if attrs["status"] != user.Status {
		t.Errorf("attributes[status] = %v, want %v", attrs["status"], user.Status)
	}

	// Verify company attributes
	if attrs["company_code"] != company.CompanyCode {
		t.Errorf("attributes[company_code] = %v, want %v", attrs["company_code"], company.CompanyCode)
	}

	// Verify department attributes
	if attrs["department_code"] != department.DepartmentCode {
		t.Errorf("attributes[department_code] = %v, want %v", attrs["department_code"], department.DepartmentCode)
	}

	// Verify position attributes
	if attrs["position_level"] != position.PositionLevel {
		t.Errorf("attributes[position_level] = %v, want %v", attrs["position_level"], position.PositionLevel)
	}

	// Verify security clearance
	if attrs["clearance"] != profile.SecurityClearance {
		t.Errorf("attributes[clearance] = %v, want %v", attrs["clearance"], profile.SecurityClearance)
	}

	// Verify access level
	if attrs["access_level"] != profile.AccessLevel {
		t.Errorf("attributes[access_level] = %v, want %v", attrs["access_level"], profile.AccessLevel)
	}

	// Verify roles
	rolesAttr, ok := attrs["roles"].([]string)
	if !ok {
		t.Fatal("attributes[roles] is not []string")
	}
	if len(rolesAttr) != 2 {
		t.Errorf("len(attributes[roles]) = %v, want 2", len(rolesAttr))
	}

	// Verify manager
	if attrs["has_manager"] != true {
		t.Error("attributes[has_manager] = false, want true")
	}
}

func TestUserSubject_HasRole(t *testing.T) {
	roles := []Role{
		{RoleCode: "developer"},
		{RoleCode: "admin"},
	}

	userSubject := &UserSubject{
		User:  &User{ID: "test-user"},
		Roles: roles,
	}

	tests := []struct {
		roleCode string
		expected bool
	}{
		{"developer", true},
		{"admin", true},
		{"manager", false},
		{"DEVELOPER", true}, // Case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.roleCode, func(t *testing.T) {
			result := userSubject.HasRole(tt.roleCode)
			if result != tt.expected {
				t.Errorf("HasRole(%s) = %v, want %v", tt.roleCode, result, tt.expected)
			}
		})
	}
}

func TestUserSubject_HasAnyRole(t *testing.T) {
	roles := []Role{
		{RoleCode: "developer"},
		{RoleCode: "reviewer"},
	}

	userSubject := &UserSubject{
		User:  &User{ID: "test-user"},
		Roles: roles,
	}

	tests := []struct {
		name      string
		roleCodes []string
		expected  bool
	}{
		{"has one", []string{"developer"}, true},
		{"has multiple", []string{"developer", "admin"}, true},
		{"has none", []string{"admin", "manager"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := userSubject.HasAnyRole(tt.roleCodes)
			if result != tt.expected {
				t.Errorf("HasAnyRole(%v) = %v, want %v", tt.roleCodes, result, tt.expected)
			}
		})
	}
}

func TestUserSubject_HasAllRoles(t *testing.T) {
	roles := []Role{
		{RoleCode: "developer"},
		{RoleCode: "reviewer"},
	}

	userSubject := &UserSubject{
		User:  &User{ID: "test-user"},
		Roles: roles,
	}

	tests := []struct {
		name      string
		roleCodes []string
		expected  bool
	}{
		{"has all", []string{"developer", "reviewer"}, true},
		{"missing one", []string{"developer", "admin"}, false},
		{"has none", []string{"admin", "manager"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := userSubject.HasAllRoles(tt.roleCodes)
			if result != tt.expected {
				t.Errorf("HasAllRoles(%v) = %v, want %v", tt.roleCodes, result, tt.expected)
			}
		})
	}
}

func TestNewUserSubject_NilUser(t *testing.T) {
	userSubject := NewUserSubject(nil, nil, nil)
	if userSubject != nil {
		t.Error("NewUserSubject(nil, nil, nil) should return nil")
	}
}
