package models

// Test helpers for creating mock subjects

// NewMockUserSubject creates a simple mock UserSubject for testing
func NewMockUserSubject(id, username string) SubjectInterface {
	user := &User{
		ID:       id,
		Username: username,
		FullName: username,
		Status:   "active",
	}
	return NewUserSubject(user, nil, nil)
}

// NewMockUserSubjectWithProfile creates a mock UserSubject with profile
func NewMockUserSubjectWithProfile(id, username, department string, level int) SubjectInterface {
	user := &User{
		ID:       id,
		Username: username,
		FullName: username,
		Status:   "active",
	}

	profile := &UserProfile{
		UserID:      id,
		AccessLevel: level,
		Attributes: JSONMap{
			"department": department,
		},
	}

	return NewUserSubject(user, profile, nil)
}

// NewMockServiceSubject creates a simple mock ServiceSubject for testing
func NewMockServiceSubject(id, name string) SubjectInterface {
	return NewServiceSubject(id, name, "default")
}

// CreateMockSubjectFromLegacyID creates a SubjectInterface from a legacy subject ID
// This is a helper for converting old tests
func CreateMockSubjectFromLegacyID(subjectID string, attributes map[string]interface{}) SubjectInterface {
	user := &User{
		ID:       subjectID,
		Username: subjectID,
		FullName: subjectID,
		Status:   "active",
	}

	profile := &UserProfile{
		UserID:     subjectID,
		Attributes: JSONMap(attributes),
	}

	return NewUserSubject(user, profile, nil)
}

// CreateMockSubjectWithAttributes creates a SubjectInterface with properly structured attributes
// that don't use the custom_ prefix. This is useful for testing.
func CreateMockSubjectWithAttributes(userID string, attrs map[string]interface{}) SubjectInterface {
	user := &User{
		ID:       userID,
		Username: userID,
		FullName: userID,
		Status:   "active",
	}

	profile := &UserProfile{
		UserID:      userID,
		AccessLevel: getIntAttribute(attrs, "level", 0),
	}

	// Set department if provided
	if dept, ok := attrs["department"].(string); ok {
		profile.Department = &Department{
			ID:             "dept-" + userID,
			DepartmentName: dept,
			DepartmentCode: dept,
		}
	}

	// Set subject_type in user metadata if provided
	if subjectType, ok := attrs["subject_type"].(string); ok {
		user.Metadata = JSONMap{"subject_type": subjectType}
	}

	// Handle other attributes that need specific handling
	customAttrs := make(JSONMap)
	for key, value := range attrs {
		// Skip attributes that are handled specially
		if key != "department" && key != "level" && key != "subject_type" {
			customAttrs[key] = value
		}
	}

	if len(customAttrs) > 0 {
		profile.Attributes = customAttrs
	}

	return NewUserSubject(user, profile, nil)
}

// Helper function to safely get int from attributes
func getIntAttribute(attrs map[string]interface{}, key string, defaultVal int) int {
	if val, ok := attrs[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return defaultVal
}
