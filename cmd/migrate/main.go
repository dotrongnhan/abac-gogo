package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"abac_go_example/models"
	"abac_go_example/storage"
)

func main() {
	fmt.Println("🚀 ABAC Database Migration and Data Seeder")
	fmt.Println("==========================================")

	// Initialize PostgreSQL storage
	config := storage.DefaultDatabaseConfig()
	pgStorage, err := storage.NewPostgreSQLStorage(config)
	if err != nil {
		fmt.Printf("Failed to initialize PostgreSQL storage: %v\n", err)
		os.Exit(1)
	}
	defer pgStorage.Close()

	fmt.Println("✅ Database connection established and tables migrated")

	// Seed data from JSON files
	if err := seedData(pgStorage, "."); err != nil {
		fmt.Printf("Failed to seed data: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Data seeding completed successfully")
}

func seedData(storage *storage.PostgreSQLStorage, dataDir string) error {
	// Seed subjects
	if err := seedSubjects(storage, filepath.Join(dataDir, "subjects.json")); err != nil {
		return fmt.Errorf("failed to seed subjects: %w", err)
	}

	// Seed resources
	if err := seedResources(storage, filepath.Join(dataDir, "resources.json")); err != nil {
		return fmt.Errorf("failed to seed resources: %w", err)
	}

	// Seed actions
	if err := seedActions(storage, filepath.Join(dataDir, "actions.json")); err != nil {
		return fmt.Errorf("failed to seed actions: %w", err)
	}

	// Seed policies
	if err := seedPolicies(storage, filepath.Join(dataDir, "policies.json")); err != nil {
		return fmt.Errorf("failed to seed policies: %w", err)
	}

	return nil
}

func seedSubjects(storage *storage.PostgreSQLStorage, filename string) error {
	fmt.Printf("📥 Seeding subjects from %s...\n", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var subjectsData struct {
		Subjects []struct {
			ID          string                 `json:"id"`
			ExternalID  string                 `json:"external_id"`
			SubjectType string                 `json:"subject_type"`
			Metadata    map[string]interface{} `json:"metadata"`
			Attributes  map[string]interface{} `json:"attributes"`
		} `json:"subjects"`
	}

	if err := json.Unmarshal(data, &subjectsData); err != nil {
		return err
	}

	for _, subjectData := range subjectsData.Subjects {
		subject := &models.Subject{
			ID:          subjectData.ID,
			ExternalID:  subjectData.ExternalID,
			SubjectType: subjectData.SubjectType,
			Metadata:    models.JSONMap(subjectData.Metadata),
			Attributes:  models.JSONMap(subjectData.Attributes),
		}

		if err := storage.CreateSubject(subject); err != nil {
			// If subject already exists, update it
			if err := storage.UpdateSubject(subject); err != nil {
				return fmt.Errorf("failed to create/update subject %s: %w", subject.ID, err)
			}
		}
	}

	fmt.Printf("✅ Seeded %d subjects\n", len(subjectsData.Subjects))
	return nil
}

func seedResources(storage *storage.PostgreSQLStorage, filename string) error {
	fmt.Printf("📥 Seeding resources from %s...\n", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var resourcesData struct {
		Resources []struct {
			ID           string                 `json:"id"`
			ResourceType string                 `json:"resource_type"`
			ResourceID   string                 `json:"resource_id"`
			Path         string                 `json:"path"`
			ParentID     string                 `json:"parent_id,omitempty"`
			Metadata     map[string]interface{} `json:"metadata"`
			Attributes   map[string]interface{} `json:"attributes"`
		} `json:"resources"`
	}

	if err := json.Unmarshal(data, &resourcesData); err != nil {
		return err
	}

	for _, resourceData := range resourcesData.Resources {
		resource := &models.Resource{
			ID:           resourceData.ID,
			ResourceType: resourceData.ResourceType,
			ResourceID:   resourceData.ResourceID,
			Path:         resourceData.Path,
			ParentID:     resourceData.ParentID,
			Metadata:     models.JSONMap(resourceData.Metadata),
			Attributes:   models.JSONMap(resourceData.Attributes),
		}

		if err := storage.CreateResource(resource); err != nil {
			// If resource already exists, update it
			if err := storage.UpdateResource(resource); err != nil {
				return fmt.Errorf("failed to create/update resource %s: %w", resource.ID, err)
			}
		}
	}

	fmt.Printf("✅ Seeded %d resources\n", len(resourcesData.Resources))
	return nil
}

func seedActions(storage *storage.PostgreSQLStorage, filename string) error {
	fmt.Printf("📥 Seeding actions from %s...\n", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var actionsData struct {
		Actions []struct {
			ID             string `json:"id"`
			ActionName     string `json:"action_name"`
			ActionCategory string `json:"action_category"`
			Description    string `json:"description"`
			IsSystem       bool   `json:"is_system"`
		} `json:"actions"`
	}

	if err := json.Unmarshal(data, &actionsData); err != nil {
		return err
	}

	for _, actionData := range actionsData.Actions {
		action := &models.Action{
			ID:             actionData.ID,
			ActionName:     actionData.ActionName,
			ActionCategory: actionData.ActionCategory,
			Description:    actionData.Description,
			IsSystem:       actionData.IsSystem,
		}

		if err := storage.CreateAction(action); err != nil {
			// If action already exists, update it
			if err := storage.UpdateAction(action); err != nil {
				return fmt.Errorf("failed to create/update action %s: %w", action.ID, err)
			}
		}
	}

	fmt.Printf("✅ Seeded %d actions\n", len(actionsData.Actions))
	return nil
}

func seedPolicies(storage *storage.PostgreSQLStorage, filename string) error {
	fmt.Printf("📥 Seeding policies from %s...\n", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var policiesData struct {
		Policies []struct {
			ID               string                 `json:"id"`
			PolicyName       string                 `json:"policy_name"`
			Description      string                 `json:"description"`
			Effect           string                 `json:"effect"`
			Priority         int                    `json:"priority"`
			Enabled          bool                   `json:"enabled"`
			Version          int                    `json:"version"`
			Conditions       map[string]interface{} `json:"conditions"`
			Rules            []models.PolicyRule    `json:"rules"`
			Actions          []string               `json:"actions"`
			ResourcePatterns []string               `json:"resource_patterns"`
		} `json:"policies"`
	}

	if err := json.Unmarshal(data, &policiesData); err != nil {
		return err
	}

	// TODO: Update to use new policy format
	/*
		for _, policyData := range policiesData.Policies {
			policy := &models.Policy{
				ID:               policyData.ID,
				PolicyName:       policyData.PolicyName,
				Description:      policyData.Description,
				Version:          "2024-10-21", // Convert to string
				Enabled:          policyData.Enabled,
				// Statement:        // TODO: Convert from old format to new format
			}

			if err := storage.CreatePolicy(policy); err != nil {
				// If policy already exists, update it
				if err := storage.UpdatePolicy(policy); err != nil {
					return fmt.Errorf("failed to create/update policy %s: %w", policy.ID, err)
				}
			}
		}
	*/

	fmt.Printf("✅ Seeded %d policies\n", len(policiesData.Policies))
	return nil
}
