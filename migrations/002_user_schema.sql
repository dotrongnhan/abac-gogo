-- Migration 002: User-based ABAC Schema
-- This migration introduces user-centric tables to replace flat Subject attributes
-- Created: 2025-10-30

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- CORE ENTITY TABLES
-- ============================================================================

-- Companies table: Root organizational entity
CREATE TABLE IF NOT EXISTS companies (
    id VARCHAR(255) PRIMARY KEY,
    company_code VARCHAR(100) NOT NULL UNIQUE,
    company_name VARCHAR(255) NOT NULL,
    industry VARCHAR(100),
    country VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_companies_status ON companies(status);
CREATE INDEX idx_companies_company_code ON companies(company_code);

-- Departments table: Organizational units within companies
CREATE TABLE IF NOT EXISTS departments (
    id VARCHAR(255) PRIMARY KEY,
    company_id VARCHAR(255) NOT NULL,
    department_code VARCHAR(100) NOT NULL,
    department_name VARCHAR(255) NOT NULL,
    parent_department_id VARCHAR(255),
    manager_id VARCHAR(255),
    cost_center VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_department_id) REFERENCES departments(id) ON DELETE SET NULL,
    UNIQUE (company_id, department_code)
);

CREATE INDEX idx_departments_company_id ON departments(company_id);
CREATE INDEX idx_departments_parent_id ON departments(parent_department_id);
CREATE INDEX idx_departments_status ON departments(status);

-- Positions table: Job positions/titles
CREATE TABLE IF NOT EXISTS positions (
    id VARCHAR(255) PRIMARY KEY,
    position_code VARCHAR(100) NOT NULL UNIQUE,
    position_name VARCHAR(255) NOT NULL,
    position_level INTEGER NOT NULL DEFAULT 1,
    position_category VARCHAR(100),
    clearance_level VARCHAR(50),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_positions_level ON positions(position_level);
CREATE INDEX idx_positions_clearance ON positions(clearance_level);
CREATE INDEX idx_positions_code ON positions(position_code);

-- Roles table: Functional roles for RBAC integration
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(255) PRIMARY KEY,
    role_code VARCHAR(100) NOT NULL UNIQUE,
    role_name VARCHAR(255) NOT NULL,
    role_type VARCHAR(50) NOT NULL DEFAULT 'functional',
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_roles_type ON roles(role_type);
CREATE INDEX idx_roles_system ON roles(is_system);
CREATE INDEX idx_roles_code ON roles(role_code);

-- ============================================================================
-- USER TABLES
-- ============================================================================

-- Users table: Core user entity
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    employee_id VARCHAR(100) UNIQUE,
    hire_date DATE,
    termination_date DATE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_employee_id ON users(employee_id);

-- User profiles table: Extended user information
CREATE TABLE IF NOT EXISTS user_profiles (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL UNIQUE,
    company_id VARCHAR(255) NOT NULL,
    department_id VARCHAR(255) NOT NULL,
    position_id VARCHAR(255) NOT NULL,
    manager_id VARCHAR(255),
    location VARCHAR(255),
    office_location VARCHAR(255),
    phone_number VARCHAR(50),
    mobile_number VARCHAR(50),
    emergency_contact JSONB,
    security_clearance VARCHAR(50),
    access_level INTEGER DEFAULT 1,
    attributes JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE RESTRICT,
    FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE RESTRICT,
    FOREIGN KEY (position_id) REFERENCES positions(id) ON DELETE RESTRICT,
    FOREIGN KEY (manager_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX idx_user_profiles_company_id ON user_profiles(company_id);
CREATE INDEX idx_user_profiles_department_id ON user_profiles(department_id);
CREATE INDEX idx_user_profiles_position_id ON user_profiles(position_id);
CREATE INDEX idx_user_profiles_manager_id ON user_profiles(manager_id);
CREATE INDEX idx_user_profiles_clearance ON user_profiles(security_clearance);
CREATE INDEX idx_user_profiles_access_level ON user_profiles(access_level);

-- User roles junction table: Many-to-many relationship
CREATE TABLE IF NOT EXISTS user_roles (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    role_id VARCHAR(255) NOT NULL,
    assigned_by VARCHAR(255),
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE (user_id, role_id)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_active ON user_roles(is_active);
CREATE INDEX idx_user_roles_expires ON user_roles(expires_at);

-- ============================================================================
-- AUDIT & HISTORY TABLES
-- ============================================================================

-- User attribute change history
CREATE TABLE IF NOT EXISTS user_attribute_history (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    attribute_name VARCHAR(255) NOT NULL,
    old_value JSONB,
    new_value JSONB,
    changed_by VARCHAR(255),
    change_reason TEXT,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_attr_history_user_id ON user_attribute_history(user_id);
CREATE INDEX idx_user_attr_history_changed_at ON user_attribute_history(changed_at);

-- ============================================================================
-- VIEWS FOR EASIER QUERYING
-- ============================================================================

-- Comprehensive user view with all related data
CREATE OR REPLACE VIEW v_users_full AS
SELECT 
    u.id AS user_id,
    u.username,
    u.email,
    u.full_name,
    u.status AS user_status,
    u.employee_id,
    u.hire_date,
    up.company_id,
    c.company_name,
    c.company_code,
    up.department_id,
    d.department_name,
    d.department_code,
    up.position_id,
    p.position_name,
    p.position_code,
    p.position_level,
    p.clearance_level AS position_clearance,
    up.manager_id,
    m.full_name AS manager_name,
    up.location,
    up.office_location,
    up.security_clearance AS user_clearance,
    up.access_level,
    ARRAY_AGG(DISTINCT r.role_code) FILTER (WHERE r.role_code IS NOT NULL) AS role_codes,
    ARRAY_AGG(DISTINCT r.role_name) FILTER (WHERE r.role_name IS NOT NULL) AS role_names,
    u.created_at,
    u.updated_at
FROM users u
LEFT JOIN user_profiles up ON u.id = up.user_id
LEFT JOIN companies c ON up.company_id = c.id
LEFT JOIN departments d ON up.department_id = d.id
LEFT JOIN positions p ON up.position_id = p.id
LEFT JOIN users m ON up.manager_id = m.id
LEFT JOIN user_roles ur ON u.id = ur.user_id AND ur.is_active = TRUE
LEFT JOIN roles r ON ur.role_id = r.id
GROUP BY 
    u.id, u.username, u.email, u.full_name, u.status, u.employee_id, u.hire_date,
    up.company_id, c.company_name, c.company_code,
    up.department_id, d.department_name, d.department_code,
    up.position_id, p.position_name, p.position_code, p.position_level, p.clearance_level,
    up.manager_id, m.full_name,
    up.location, up.office_location, up.security_clearance, up.access_level,
    u.created_at, u.updated_at;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_companies_updated_at BEFORE UPDATE ON companies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_departments_updated_at BEFORE UPDATE ON departments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_positions_updated_at BEFORE UPDATE ON positions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_profiles_updated_at BEFORE UPDATE ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE companies IS 'Organizational companies in the system';
COMMENT ON TABLE departments IS 'Departments within companies, supports hierarchical structure';
COMMENT ON TABLE positions IS 'Job positions/titles with levels and clearances';
COMMENT ON TABLE roles IS 'Functional roles for RBAC integration';
COMMENT ON TABLE users IS 'Core user entities';
COMMENT ON TABLE user_profiles IS 'Extended user information with organizational context';
COMMENT ON TABLE user_roles IS 'Many-to-many relationship between users and roles';
COMMENT ON TABLE user_attribute_history IS 'Audit trail for user attribute changes';
COMMENT ON VIEW v_users_full IS 'Comprehensive view of users with all related data';

