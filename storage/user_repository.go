package storage

import (
	"errors"
	"fmt"
	"time"

	"abac_go_example/models"

	"gorm.io/gorm"
)

const (
	maxPreloadDepth = 3
)

// UserRepository handles user-related database operations
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// GetUserByID retrieves a user by ID without relations
func (ur *UserRepository) GetUserByID(id string) (*models.User, error) {
	var user models.User
	result := ur.db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (ur *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	result := ur.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %s", username)
		}
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (ur *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := ur.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %s", email)
		}
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}
	return &user, nil
}

// GetUserWithRelations retrieves a user with all related data (profile, roles, etc.)
// This is optimized with GORM Preload to minimize database queries
func (ur *UserRepository) GetUserWithRelations(id string) (*models.User, error) {
	var user models.User

	result := ur.db.
		Preload("Profile").
		Preload("Profile.Company").
		Preload("Profile.Department").
		Preload("Profile.Department.Company").
		Preload("Profile.Position").
		Preload("Profile.Manager").
		Preload("Roles").
		Where("id = ?", id).
		First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get user with relations: %w", result.Error)
	}

	return &user, nil
}

// GetUserProfile retrieves the profile for a specific user
func (ur *UserRepository) GetUserProfile(userID string) (*models.UserProfile, error) {
	var profile models.UserProfile
	result := ur.db.
		Preload("Company").
		Preload("Department").
		Preload("Position").
		Preload("Manager").
		Where("user_id = ?", userID).
		First(&profile)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user profile not found for user: %s", userID)
		}
		return nil, fmt.Errorf("failed to get user profile: %w", result.Error)
	}
	return &profile, nil
}

// GetUserRoles retrieves all active roles for a user
func (ur *UserRepository) GetUserRoles(userID string) ([]models.Role, error) {
	var roles []models.Role

	// Join with user_roles to get only active roles
	result := ur.db.
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND user_roles.is_active = ?", userID, true).
		Where("(user_roles.expires_at IS NULL OR user_roles.expires_at > ?)", time.Now()).
		Find(&roles)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", result.Error)
	}

	return roles, nil
}

// GetUserAttributes builds ABAC attributes from user data
// This method loads user, profile, and roles, then returns flat attributes
func (ur *UserRepository) GetUserAttributes(userID string) (map[string]interface{}, error) {
	user, err := ur.GetUserWithRelations(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user with relations: %w", err)
	}

	// Extract profile and roles
	var profile *models.UserProfile
	var roles []models.Role

	if user.Profile != nil {
		profile = user.Profile
	}

	if len(user.Roles) > 0 {
		roles = user.Roles
	}

	// Create UserSubject to map attributes
	userSubject := models.NewUserSubject(user, profile, roles)
	if userSubject == nil {
		return nil, fmt.Errorf("failed to create user subject")
	}

	return userSubject.GetAttributes(), nil
}

// CreateUser creates a new user
func (ur *UserRepository) CreateUser(user *models.User) error {
	result := ur.db.Create(user)
	if result.Error != nil {
		return fmt.Errorf("failed to create user: %w", result.Error)
	}
	return nil
}

// UpdateUser updates an existing user
func (ur *UserRepository) UpdateUser(user *models.User) error {
	result := ur.db.Save(user)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	return nil
}

// DeleteUser deletes a user by ID
func (ur *UserRepository) DeleteUser(id string) error {
	result := ur.db.Delete(&models.User{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	return nil
}

// CreateUserProfile creates a new user profile
func (ur *UserRepository) CreateUserProfile(profile *models.UserProfile) error {
	result := ur.db.Create(profile)
	if result.Error != nil {
		return fmt.Errorf("failed to create user profile: %w", result.Error)
	}
	return nil
}

// UpdateUserProfile updates an existing user profile
func (ur *UserRepository) UpdateUserProfile(profile *models.UserProfile) error {
	result := ur.db.Save(profile)
	if result.Error != nil {
		return fmt.Errorf("failed to update user profile: %w", result.Error)
	}
	return nil
}

// AssignRole assigns a role to a user
func (ur *UserRepository) AssignRole(userID, roleID, assignedBy string) error {
	userRole := &models.UserRole{
		ID:         fmt.Sprintf("ur_%s_%s", userID, roleID),
		UserID:     userID,
		RoleID:     roleID,
		IsActive:   true,
		AssignedAt: time.Now(),
	}
	if assignedBy != "" {
		userRole.AssignedBy = &assignedBy
	}

	result := ur.db.Create(userRole)
	if result.Error != nil {
		return fmt.Errorf("failed to assign role: %w", result.Error)
	}
	return nil
}

// RevokeRole revokes a role from a user
func (ur *UserRepository) RevokeRole(userID, roleID string) error {
	result := ur.db.Model(&models.UserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to revoke role: %w", result.Error)
	}
	return nil
}

// GetAllUsers retrieves all users with optional filters
func (ur *UserRepository) GetAllUsers(status string, limit, offset int) ([]*models.User, error) {
	var users []*models.User

	query := ur.db.Model(&models.User{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	result := query.Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all users: %w", result.Error)
	}

	return users, nil
}

// GetUsersByDepartment retrieves all users in a specific department
func (ur *UserRepository) GetUsersByDepartment(departmentID string) ([]*models.User, error) {
	var users []*models.User

	result := ur.db.
		Joins("JOIN user_profiles ON user_profiles.user_id = users.id").
		Where("user_profiles.department_id = ?", departmentID).
		Find(&users)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get users by department: %w", result.Error)
	}

	return users, nil
}

// GetUsersByRole retrieves all users with a specific role
func (ur *UserRepository) GetUsersByRole(roleID string) ([]*models.User, error) {
	var users []*models.User

	result := ur.db.
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Where("user_roles.role_id = ? AND user_roles.is_active = ?", roleID, true).
		Find(&users)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", result.Error)
	}

	return users, nil
}

// GetCompanyByID retrieves a company by ID
func (ur *UserRepository) GetCompanyByID(id string) (*models.Company, error) {
	var company models.Company
	result := ur.db.Where("id = ?", id).First(&company)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("company not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get company: %w", result.Error)
	}
	return &company, nil
}

// GetDepartmentByID retrieves a department by ID
func (ur *UserRepository) GetDepartmentByID(id string) (*models.Department, error) {
	var department models.Department
	result := ur.db.Preload("Company").Where("id = ?", id).First(&department)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("department not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get department: %w", result.Error)
	}
	return &department, nil
}

// GetPositionByID retrieves a position by ID
func (ur *UserRepository) GetPositionByID(id string) (*models.Position, error) {
	var position models.Position
	result := ur.db.Where("id = ?", id).First(&position)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("position not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get position: %w", result.Error)
	}
	return &position, nil
}

// GetRoleByID retrieves a role by ID
func (ur *UserRepository) GetRoleByID(id string) (*models.Role, error) {
	var role models.Role
	result := ur.db.Where("id = ?", id).First(&role)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get role: %w", result.Error)
	}
	return &role, nil
}

// GetRoleByCode retrieves a role by its code
func (ur *UserRepository) GetRoleByCode(code string) (*models.Role, error) {
	var role models.Role
	result := ur.db.Where("role_code = ?", code).First(&role)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role not found: %s", code)
		}
		return nil, fmt.Errorf("failed to get role: %w", result.Error)
	}
	return &role, nil
}
