-- 1. SUBJECTS (Users, Services, Applications)
CREATE TABLE subjects (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          external_id VARCHAR(255) UNIQUE NOT NULL,
                          subject_type VARCHAR(50) NOT NULL CHECK (subject_type IN ('user', 'service', 'application', 'device')),
                          metadata JSONB DEFAULT '{}',
                          created_at TIMESTAMP DEFAULT NOW(),
                          updated_at TIMESTAMP DEFAULT NOW()
);

-- 2. SUBJECT_ATTRIBUTES (Normalized attributes)
CREATE TABLE subject_attributes (
                                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
                                    attribute_name VARCHAR(100) NOT NULL,
                                    attribute_value TEXT NOT NULL,
                                    data_type VARCHAR(30) CHECK (data_type IN ('string', 'number', 'boolean', 'datetime', 'array')),
                                    valid_from TIMESTAMP DEFAULT NOW(),
                                    valid_until TIMESTAMP,
                                    UNIQUE(subject_id, attribute_name),
                                    INDEX idx_subj_attr_name (attribute_name),
                                    INDEX idx_subj_attr_valid (valid_from, valid_until)
);

-- 3. RESOURCES (APIs, Documents, Data objects)
CREATE TABLE resources (
                           id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           resource_type VARCHAR(100) NOT NULL,
                           resource_id VARCHAR(500) NOT NULL,
                           parent_id UUID REFERENCES resources(id) ON DELETE CASCADE,
                           path LTREE,
                           metadata JSONB DEFAULT '{}',
                           created_at TIMESTAMP DEFAULT NOW(),
                           UNIQUE(resource_type, resource_id),
                           INDEX idx_resource_type (resource_type),
                           INDEX idx_resource_path USING GIST (path)
);

-- 4. RESOURCE_ATTRIBUTES
CREATE TABLE resource_attributes (
                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                     resource_id UUID NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
                                     attribute_name VARCHAR(100) NOT NULL,
                                     attribute_value TEXT NOT NULL,
                                     data_type VARCHAR(30) CHECK (data_type IN ('string', 'number', 'boolean', 'datetime', 'array')),
                                     is_inherited BOOLEAN DEFAULT false,
                                     UNIQUE(resource_id, attribute_name),
                                     INDEX idx_res_attr_name (attribute_name)
);

-- 5. ACTIONS (Operations)
CREATE TABLE actions (
                         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                         action_name VARCHAR(100) UNIQUE NOT NULL,
                         action_category VARCHAR(50),
                         description TEXT,
                         is_system BOOLEAN DEFAULT false,
                         INDEX idx_action_category (action_category)
);

-- 6. POLICIES (Core policy definitions)
CREATE TABLE policies (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          policy_name VARCHAR(255) UNIQUE NOT NULL,
                          description TEXT,
                          effect VARCHAR(10) NOT NULL CHECK (effect IN ('permit', 'deny')),
                          priority INTEGER DEFAULT 100,
                          enabled BOOLEAN DEFAULT true,
                          conditions JSONB DEFAULT '{}', -- complex conditions in JSON
                          parent_policy_id UUID REFERENCES policies(id),
                          version INTEGER DEFAULT 1,
                          created_at TIMESTAMP DEFAULT NOW(),
                          updated_at TIMESTAMP DEFAULT NOW(),
                          INDEX idx_policy_priority (priority, enabled),
                          INDEX idx_policy_parent (parent_policy_id)
);

-- 7. POLICY_RULES (Granular conditions)
CREATE TABLE policy_rules (
                              id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                              policy_id UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
                              target_type VARCHAR(30) CHECK (target_type IN ('subject', 'resource', 'action', 'environment')),
                              attribute_path VARCHAR(255) NOT NULL, -- e.g., 'attributes.department', 'metadata.classification'
                              operator VARCHAR(20) NOT NULL CHECK (operator IN ('eq', 'neq', 'gt', 'gte', 'lt', 'lte', 'in', 'nin', 'contains', 'regex', 'exists')),
                              expected_value JSONB NOT NULL,
                              is_negative BOOLEAN DEFAULT false,
                              rule_order INTEGER DEFAULT 0,
                              INDEX idx_rule_policy (policy_id),
                              INDEX idx_rule_target (target_type, attribute_path)
);

-- 8. POLICY_ACTIONS (Many-to-Many)
CREATE TABLE policy_actions (
                                policy_id UUID REFERENCES policies(id) ON DELETE CASCADE,
                                action_id UUID REFERENCES actions(id) ON DELETE CASCADE,
                                PRIMARY KEY (policy_id, action_id)
);

-- 9. POLICY_RESOURCES (Resource patterns for policies)
CREATE TABLE policy_resources (
                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                  policy_id UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
                                  resource_type VARCHAR(100),
                                  resource_pattern VARCHAR(500), -- Support wildcards: /api/v1/users/*
                                  is_recursive BOOLEAN DEFAULT false,
                                  INDEX idx_pol_res_policy (policy_id),
                                  INDEX idx_pol_res_type (resource_type)
);

-- 10. AUDIT_LOGS (Partitioned by month)
CREATE TABLE audit_logs (
                            id BIGSERIAL,
                            request_id UUID NOT NULL DEFAULT gen_random_uuid(),
                            subject_id UUID REFERENCES subjects(id),
                            resource_id UUID REFERENCES resources(id),
                            action_id UUID REFERENCES actions(id),
                            decision VARCHAR(10) CHECK (decision IN ('permit', 'deny', 'notapplicable')),
                            context JSONB DEFAULT '{}',
                            evaluation_ms INTEGER,
                            created_at TIMESTAMP DEFAULT NOW(),
                            PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Create monthly partitions
CREATE TABLE audit_logs_2024_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Indexes for audit logs
CREATE INDEX idx_audit_created ON audit_logs (created_at DESC);
CREATE INDEX idx_audit_subject ON audit_logs (subject_id);
CREATE INDEX idx_audit_decision ON audit_logs (decision);