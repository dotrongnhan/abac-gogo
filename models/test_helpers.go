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
