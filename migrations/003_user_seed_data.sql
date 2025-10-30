-- Seed Data for User-based ABAC System
-- This file contains sample data for testing the user-based ABAC implementation
-- Created: 2025-10-30

-- ============================================================================
-- COMPANIES
-- ============================================================================

INSERT INTO companies (id, company_code, company_name, industry, country, status) VALUES
('company-001', 'TECH-001', 'TechCorp International', 'Technology', 'USA', 'active'),
('company-002', 'FIN-001', 'FinanceHub Ltd', 'Finance', 'UK', 'active'),
('company-003', 'HEALTH-001', 'HealthCare Systems', 'Healthcare', 'USA', 'active')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- DEPARTMENTS
-- ============================================================================

INSERT INTO departments (id, company_id, department_code, department_name, parent_department_id, cost_center, status) VALUES
('dept-001', 'company-001', 'ENG', 'Engineering', NULL, 'CC-1000', 'active'),
('dept-002', 'company-001', 'ENG-BE', 'Backend Engineering', 'dept-001', 'CC-1001', 'active'),
('dept-003', 'company-001', 'ENG-FE', 'Frontend Engineering', 'dept-001', 'CC-1002', 'active'),
('dept-004', 'company-001', 'FINANCE', 'Finance', NULL, 'CC-2000', 'active'),
('dept-005', 'company-001', 'HR', 'Human Resources', NULL, 'CC-3000', 'active'),
('dept-006', 'company-002', 'FINANCE', 'Finance', NULL, 'CC-2000', 'active'),
('dept-007', 'company-002', 'AUDIT', 'Audit & Compliance', NULL, 'CC-2100', 'active'),
('dept-008', 'company-003', 'MEDICAL', 'Medical Services', NULL, 'CC-4000', 'active')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- POSITIONS
-- ============================================================================

INSERT INTO positions (id, position_code, position_name, position_level, position_category, clearance_level, description) VALUES
('pos-001', 'DEV-JR', 'Junior Developer', 1, 'Engineering', 'basic', 'Entry-level software developer'),
('pos-002', 'DEV-SR', 'Senior Developer', 3, 'Engineering', 'standard', 'Experienced software developer'),
('pos-003', 'DEV-LEAD', 'Lead Developer', 5, 'Engineering', 'confidential', 'Technical lead and architect'),
('pos-004', 'MGR-ENG', 'Engineering Manager', 6, 'Management', 'confidential', 'Engineering team manager'),
('pos-005', 'FIN-ANALYST', 'Financial Analyst', 2, 'Finance', 'standard', 'Financial data analyst'),
('pos-006', 'FIN-MGR', 'Finance Manager', 5, 'Finance', 'secret', 'Finance department manager'),
('pos-007', 'HR-SPEC', 'HR Specialist', 2, 'HR', 'standard', 'Human resources specialist'),
('pos-008', 'EXEC-VP', 'Vice President', 8, 'Executive', 'top_secret', 'Executive leadership'),
('pos-009', 'EXEC-CEO', 'Chief Executive Officer', 10, 'Executive', 'top_secret', 'Company CEO'),
('pos-010', 'ADMIN-SYS', 'System Administrator', 4, 'IT', 'secret', 'System and infrastructure admin')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- ROLES
-- ============================================================================

INSERT INTO roles (id, role_code, role_name, role_type, description, is_system) VALUES
('role-001', 'admin', 'Administrator', 'functional', 'Full system administrator access', true),
('role-002', 'developer', 'Developer', 'functional', 'Software development access', false),
('role-003', 'reviewer', 'Code Reviewer', 'functional', 'Code review permissions', false),
('role-004', 'manager', 'Manager', 'functional', 'Team management access', false),
('role-005', 'finance_viewer', 'Finance Viewer', 'functional', 'View financial data', false),
('role-006', 'finance_editor', 'Finance Editor', 'functional', 'Edit financial data', false),
('role-007', 'hr_admin', 'HR Administrator', 'functional', 'HR system administration', false),
('role-008', 'auditor', 'Auditor', 'functional', 'Audit and compliance access', false),
('role-009', 'readonly', 'Read Only User', 'functional', 'Read-only access to systems', false),
('role-010', 'api_consumer', 'API Consumer', 'functional', 'API access permissions', false)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- USERS
-- ============================================================================

INSERT INTO users (id, username, email, full_name, status, employee_id, hire_date, metadata) VALUES
('user-001', 'john.doe', 'john.doe@techcorp.com', 'John Doe', 'active', 'EMP-001', '2020-01-15', '{"preferred_name": "John", "timezone": "America/New_York"}'),
('user-002', 'alice.smith', 'alice.smith@techcorp.com', 'Alice Smith', 'active', 'EMP-002', '2019-03-20', '{"preferred_name": "Alice", "timezone": "America/Los_Angeles"}'),
('user-003', 'bob.wilson', 'bob.wilson@techcorp.com', 'Bob Wilson', 'active', 'EMP-003', '2024-01-10', '{"preferred_name": "Bob", "timezone": "America/Chicago"}'),
('user-004', 'carol.brown', 'carol.brown@techcorp.com', 'Carol Brown', 'active', 'EMP-004', '2018-06-01', '{"preferred_name": "Carol", "timezone": "America/New_York"}'),
('user-005', 'david.lee', 'david.lee@financehub.com', 'David Lee', 'active', 'EMP-005', '2021-09-15', '{"preferred_name": "David", "timezone": "Europe/London"}'),
('user-006', 'emma.garcia', 'emma.garcia@healthcare.com', 'Emma Garcia', 'active', 'EMP-006', '2022-02-01', '{"preferred_name": "Emma", "timezone": "America/New_York"}'),
('user-007', 'frank.chen', 'frank.chen@techcorp.com', 'Frank Chen', 'probation', 'EMP-007', '2024-10-01', '{"preferred_name": "Frank", "timezone": "Asia/Shanghai"}')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- USER PROFILES
-- ============================================================================

INSERT INTO user_profiles (id, user_id, company_id, department_id, position_id, manager_id, location, office_location, security_clearance, access_level, attributes) VALUES
('profile-001', 'user-001', 'company-001', 'dept-002', 'pos-003', 'user-004', 'New York, NY', 'NYC-HQ-5F', 'confidential', 5, '{"project_access": ["project-alpha", "project-beta"], "certifications": ["AWS", "Kubernetes"]}'),
('profile-002', 'user-002', 'company-001', 'dept-004', 'pos-006', NULL, 'San Francisco, CA', 'SF-HQ-3F', 'secret', 7, '{"budget_authority": 1000000, "financial_systems": ["SAP", "Oracle"]}'),
('profile-003', 'user-003', 'company-001', 'dept-002', 'pos-001', 'user-001', 'Chicago, IL', 'CHI-OFFICE-2F', 'basic', 1, '{"training_status": "in_progress", "mentor": "user-001"}'),
('profile-004', 'user-004', 'company-001', 'dept-001', 'pos-004', NULL, 'New York, NY', 'NYC-HQ-6F', 'confidential', 6, '{"team_size": 15, "direct_reports": ["user-001", "user-003"]}'),
('profile-005', 'user-005', 'company-002', 'dept-006', 'pos-005', NULL, 'London, UK', 'LON-HQ-4F', 'standard', 3, '{"specialization": "risk_analysis", "markets": ["EU", "US"]}'),
('profile-006', 'user-006', 'company-003', 'dept-008', 'pos-010', NULL, 'Boston, MA', 'BOS-HOSPITAL-1F', 'secret', 4, '{"system_access": ["EHR", "PACS"], "oncall_rotation": true}'),
('profile-007', 'user-007', 'company-001', 'dept-002', 'pos-001', 'user-001', 'Remote', 'REMOTE', 'basic', 1, '{"probation_end": "2025-01-01", "restricted_access": true}')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- USER ROLES ASSIGNMENT
-- ============================================================================

INSERT INTO user_roles (id, user_id, role_id, assigned_by, is_active) VALUES
('ur-001', 'user-001', 'role-002', 'user-004', true),
('ur-002', 'user-001', 'role-003', 'user-004', true),
('ur-003', 'user-002', 'role-005', NULL, true),
('ur-004', 'user-002', 'role-006', NULL, true),
('ur-005', 'user-003', 'role-002', 'user-001', true),
('ur-006', 'user-003', 'role-009', 'user-001', true),
('ur-007', 'user-004', 'role-002', NULL, true),
('ur-008', 'user-004', 'role-004', NULL, true),
('ur-009', 'user-005', 'role-005', NULL, true),
('ur-010', 'user-005', 'role-008', NULL, true),
('ur-011', 'user-006', 'role-001', NULL, true),
('ur-012', 'user-006', 'role-010', NULL, true),
('ur-013', 'user-007', 'role-009', 'user-001', true)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SAMPLE POLICIES FOR USER-BASED ABAC
-- ============================================================================

-- Note: These policies use the new user attributes from UserSubject
-- Policy: Allow developers to read API resources
INSERT INTO policies (id, policy_name, description, effect, version, statement, enabled) VALUES
('policy-user-001', 'Developer API Read Access', 'Allow developers to read API resources', 'permit', '1.0', 
'[{
  "Sid": "AllowDeveloperAPIRead",
  "Effect": "Allow",
  "Action": "read",
  "Resource": "/api/v1/*",
  "Condition": {
    "StringEquals": {
      "user.roles": "developer"
    },
    "NumericGreaterThanEquals": {
      "user.position_level": 1
    }
  }
}]'::jsonb, true)
ON CONFLICT (id) DO NOTHING;

-- Policy: Allow Finance department to access financial data
INSERT INTO policies (id, policy_name, description, effect, version, statement, enabled) VALUES
('policy-user-002', 'Finance Department Access', 'Allow finance department to access financial data', 'permit', '1.0',
'[{
  "Sid": "AllowFinanceDeptFinancialData",
  "Effect": "Allow",
  "Action": ["read", "write"],
  "Resource": "/api/v1/financial*",
  "Condition": {
    "StringEquals": {
      "user.department_code": "FINANCE"
    },
    "NumericGreaterThanEquals": {
      "user.access_level": 3
    }
  }
}]'::jsonb, true)
ON CONFLICT (id) DO NOTHING;

-- Policy: Deny access for users on probation
INSERT INTO policies (id, policy_name, description, effect, version, statement, enabled) VALUES
('policy-user-003', 'Deny Probation Users', 'Deny sensitive access for users on probation', 'deny', '1.0',
'[{
  "Sid": "DenyProbationUsers",
  "Effect": "Deny",
  "Action": "*",
  "Resource": ["/api/v1/admin*", "/api/v1/financial*"],
  "Condition": {
    "StringEquals": {
      "user.status": "probation"
    }
  }
}]'::jsonb, true)
ON CONFLICT (id) DO NOTHING;

-- Policy: Admin role full access
INSERT INTO policies (id, policy_name, description, effect, version, statement, enabled) VALUES
('policy-user-004', 'Admin Full Access', 'System administrators have full access', 'permit', '1.0',
'[{
  "Sid": "AllowAdminFullAccess",
  "Effect": "Allow",
  "Action": "*",
  "Resource": "*",
  "Condition": {
    "StringEquals": {
      "user.roles": "admin"
    }
  }
}]'::jsonb, true)
ON CONFLICT (id) DO NOTHING;

-- Policy: Managers can access their department resources
INSERT INTO policies (id, policy_name, description, effect, version, statement, enabled) VALUES
('policy-user-005', 'Manager Department Access', 'Managers can access their department resources', 'permit', '1.0',
'[{
  "Sid": "AllowManagerDepartmentAccess",
  "Effect": "Allow",
  "Action": ["read", "write"],
  "Resource": "/api/v1/users*",
  "Condition": {
    "StringEquals": {
      "user.roles": "manager"
    },
    "NumericGreaterThanEquals": {
      "user.position_level": 4
    }
  }
}]'::jsonb, true)
ON CONFLICT (id) DO NOTHING;

-- Policy: High clearance required for sensitive data
INSERT INTO policies (id, policy_name, description, effect, version, statement, enabled) VALUES
('policy-user-006', 'High Clearance Sensitive Data', 'Require high security clearance for sensitive data', 'permit', '1.0',
'[{
  "Sid": "AllowHighClearanceSensitiveData",
  "Effect": "Allow",
  "Action": "read",
  "Resource": "/api/v1/sensitive/*",
  "Condition": {
    "StringIn": {
      "user.security_clearance": ["secret", "top_secret"]
    }
  }
}]'::jsonb, true)
ON CONFLICT (id) DO NOTHING;

