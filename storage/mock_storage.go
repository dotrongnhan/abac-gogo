package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"abac_go_example/models"
)

// Storage interface defines the contract for data access
type Storage interface {
	GetSubject(id string) (*models.Subject, error)
	GetResource(id string) (*models.Resource, error)
	GetAction(name string) (*models.Action, error)
	GetPolicies() ([]*models.Policy, error)
	GetAllSubjects() ([]*models.Subject, error)
	GetAllResources() ([]*models.Resource, error)
	GetAllActions() ([]*models.Action, error)
}

// MockStorage implements Storage interface using JSON files
type MockStorage struct {
	subjects  map[string]models.Subject
	resources map[string]models.Resource
	actions   map[string]models.Action
	policies  []*models.Policy
}

// NewMockStorage creates a new MockStorage instance
func NewMockStorage(dataDir string) (*MockStorage, error) {
	storage := &MockStorage{
		subjects:  make(map[string]models.Subject),
		resources: make(map[string]models.Resource),
		actions:   make(map[string]models.Action),
	}

	if err := storage.loadData(dataDir); err != nil {
		return nil, fmt.Errorf("failed to load data: %w", err)
	}

	return storage, nil
}

func (s *MockStorage) loadData(dataDir string) error {
	// Load subjects
	if err := s.loadSubjects(filepath.Join(dataDir, "subjects.json")); err != nil {
		return fmt.Errorf("failed to load subjects: %w", err)
	}

	// Load resources
	if err := s.loadResources(filepath.Join(dataDir, "resources.json")); err != nil {
		return fmt.Errorf("failed to load resources: %w", err)
	}

	// Load actions
	if err := s.loadActions(filepath.Join(dataDir, "actions.json")); err != nil {
		return fmt.Errorf("failed to load actions: %w", err)
	}

	// Load policies
	if err := s.loadPolicies(filepath.Join(dataDir, "policies.json")); err != nil {
		return fmt.Errorf("failed to load policies: %w", err)
	}

	return nil
}

func (s *MockStorage) loadSubjects(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var subjectsData struct {
		Subjects []models.Subject `json:"subjects"`
	}

	if err := json.Unmarshal(data, &subjectsData); err != nil {
		return err
	}

	for _, subject := range subjectsData.Subjects {
		s.subjects[subject.ID] = subject
	}

	return nil
}

func (s *MockStorage) loadResources(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var resourcesData struct {
		Resources []models.Resource `json:"resources"`
	}

	if err := json.Unmarshal(data, &resourcesData); err != nil {
		return err
	}

	for _, resource := range resourcesData.Resources {
		s.resources[resource.ID] = resource
	}

	return nil
}

func (s *MockStorage) loadActions(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var actionsData struct {
		Actions []models.Action `json:"actions"`
	}

	if err := json.Unmarshal(data, &actionsData); err != nil {
		return err
	}

	for _, action := range actionsData.Actions {
		s.actions[action.ActionName] = action
	}

	return nil
}

func (s *MockStorage) loadPolicies(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var policiesData struct {
		Policies []*models.Policy `json:"policies"`
	}

	if err := json.Unmarshal(data, &policiesData); err != nil {
		return err
	}

	s.policies = policiesData.Policies
	return nil
}

// GetSubject retrieves a subject by ID
func (s *MockStorage) GetSubject(id string) (*models.Subject, error) {
	subject, exists := s.subjects[id]
	if !exists {
		return nil, fmt.Errorf("subject not found: %s", id)
	}
	return &subject, nil
}

// GetResource retrieves a resource by ID
func (s *MockStorage) GetResource(id string) (*models.Resource, error) {
	resource, exists := s.resources[id]
	if !exists {
		// Try to find by ResourceID (path) if not found by ID
		for _, res := range s.resources {
			if res.ResourceID == id {
				return &res, nil
			}
		}
		return nil, fmt.Errorf("resource not found: %s", id)
	}
	return &resource, nil
}

// GetAction retrieves an action by name
func (s *MockStorage) GetAction(name string) (*models.Action, error) {
	action, exists := s.actions[name]
	if !exists {
		return nil, fmt.Errorf("action not found: %s", name)
	}
	return &action, nil
}

// GetPolicies retrieves all policies
func (s *MockStorage) GetPolicies() ([]*models.Policy, error) {
	return s.policies, nil
}

// GetAllSubjects retrieves all subjects
func (s *MockStorage) GetAllSubjects() ([]*models.Subject, error) {
	subjects := make([]*models.Subject, 0, len(s.subjects))
	for _, subject := range s.subjects {
		subjectCopy := subject
		subjects = append(subjects, &subjectCopy)
	}
	return subjects, nil
}

// GetAllResources retrieves all resources
func (s *MockStorage) GetAllResources() ([]*models.Resource, error) {
	resources := make([]*models.Resource, 0, len(s.resources))
	for _, resource := range s.resources {
		resourceCopy := resource
		resources = append(resources, &resourceCopy)
	}
	return resources, nil
}

// GetAllActions retrieves all actions
func (s *MockStorage) GetAllActions() ([]*models.Action, error) {
	actions := make([]*models.Action, 0, len(s.actions))
	for _, action := range s.actions {
		actionCopy := action
		actions = append(actions, &actionCopy)
	}
	return actions, nil
}
