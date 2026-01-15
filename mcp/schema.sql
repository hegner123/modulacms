-- Team Memory Database Schema
-- Stores shared knowledge, decisions, and context for the development team

-- Core memory table for general knowledge and notes
CREATE TABLE IF NOT EXISTS memories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    category TEXT NOT NULL CHECK(category IN (
        'design-decision',
        'architecture',
        'component',
        'agent-output',
        'issue',
        'best-practice',
        'client-customization',
        'pattern',
        'workflow',
        'implementation-plans',
        'active-implementation-plan',
        'general'
    )),
    tags TEXT, -- Comma-separated tags for flexible categorization
    author TEXT, -- Who added this (developer name or 'AI Agent')
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT -- JSON for additional structured data
);

-- Design decisions with rationale
CREATE TABLE IF NOT EXISTS design_decisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    decision TEXT NOT NULL,
    rationale TEXT NOT NULL,
    alternatives_considered TEXT, -- What else was evaluated
    impact TEXT, -- What this affects
    made_by TEXT,
    date_decided DATETIME DEFAULT CURRENT_TIMESTAMP,
    status TEXT CHECK(status IN ('active', 'deprecated', 'superseded')) DEFAULT 'active',
    superseded_by INTEGER REFERENCES design_decisions(id),
    tags TEXT
);

-- Component contracts from agent workflows
CREATE TABLE IF NOT EXISTS component_contracts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    component_name TEXT NOT NULL,
    contract_json TEXT NOT NULL, -- Full component contract
    design_image_path TEXT, -- Reference to design image if available
    generated_files TEXT, -- JSON array of generated file paths
    validation_status TEXT CHECK(validation_status IN ('pending', 'passed', 'failed', 'needs-review')),
    validation_notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    agent_version TEXT, -- Which agent version generated this
    tags TEXT
);

-- Agent execution history and outputs
CREATE TABLE IF NOT EXISTS agent_executions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_name TEXT NOT NULL,
    task_description TEXT NOT NULL,
    input_summary TEXT,
    output_summary TEXT,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    execution_time_ms INTEGER,
    artifacts TEXT, -- JSON array of generated artifacts
    executed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    executed_by TEXT, -- Developer who ran it
    tags TEXT
);

-- Known issues and workarounds
CREATE TABLE IF NOT EXISTS known_issues (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    workaround TEXT,
    severity TEXT CHECK(severity IN ('low', 'medium', 'high', 'critical')) DEFAULT 'medium',
    status TEXT CHECK(status IN ('open', 'investigating', 'workaround-available', 'resolved')) DEFAULT 'open',
    affected_components TEXT, -- Comma-separated list
    resolution TEXT,
    reported_by TEXT,
    reported_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,
    tags TEXT
);

-- Best practices discovered during development
CREATE TABLE IF NOT EXISTS best_practices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    example_code TEXT, -- Code snippet demonstrating the practice
    context TEXT, -- When/where to apply this
    category TEXT CHECK(category IN (
        'backend',
        'frontend',
        'umbraco',
        'nextjs',
        'accessibility',
        'performance',
        'security',
        'testing',
        'general'
    )),
    learned_from TEXT, -- What situation taught us this
    added_by TEXT,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    tags TEXT
);

-- Client-specific customizations (important for seed project)
CREATE TABLE IF NOT EXISTS client_customizations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    client_name TEXT NOT NULL,
    customization_type TEXT CHECK(customization_type IN (
        'component',
        'theme',
        'feature',
        'workflow',
        'integration',
        'configuration'
    )),
    description TEXT NOT NULL,
    implementation_notes TEXT,
    files_affected TEXT, -- JSON array of file paths
    shared_vs_custom TEXT CHECK(shared_vs_custom IN ('shared', 'custom')),
    -- 'shared' = should be in seed, 'custom' = client-specific
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    tags TEXT
);

-- Architectural patterns and their usage
CREATE TABLE IF NOT EXISTS patterns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pattern_name TEXT NOT NULL,
    description TEXT NOT NULL,
    use_cases TEXT, -- When to use this pattern
    example_implementation TEXT, -- Code or reference
    pros TEXT,
    cons TEXT,
    related_patterns TEXT, -- Comma-separated pattern names
    category TEXT CHECK(category IN (
        'backend',
        'frontend',
        'fullstack',
        'data-access',
        'api',
        'ui',
        'state-management'
    )),
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    tags TEXT
);

-- Search and retrieval indexes
CREATE INDEX IF NOT EXISTS idx_memories_category ON memories(category);
CREATE INDEX IF NOT EXISTS idx_memories_created ON memories(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_decisions_status ON design_decisions(status);
CREATE INDEX IF NOT EXISTS idx_components_name ON component_contracts(component_name);
CREATE INDEX IF NOT EXISTS idx_agents_name ON agent_executions(agent_name);
CREATE INDEX IF NOT EXISTS idx_issues_status ON known_issues(status);
CREATE INDEX IF NOT EXISTS idx_issues_severity ON known_issues(severity);
CREATE INDEX IF NOT EXISTS idx_practices_category ON best_practices(category);
CREATE INDEX IF NOT EXISTS idx_clients_name ON client_customizations(client_name);

-- Full-text search virtual table for memories
CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
    title,
    content,
    tags,
    content=memories,
    content_rowid=id
);

-- Triggers to keep FTS table in sync
CREATE TRIGGER IF NOT EXISTS memories_ai AFTER INSERT ON memories BEGIN
    INSERT INTO memories_fts(rowid, title, content, tags)
    VALUES (new.id, new.title, new.content, new.tags);
END;

CREATE TRIGGER IF NOT EXISTS memories_ad AFTER DELETE ON memories BEGIN
    DELETE FROM memories_fts WHERE rowid = old.id;
END;

CREATE TRIGGER IF NOT EXISTS memories_au AFTER UPDATE ON memories BEGIN
    UPDATE memories_fts
    SET title = new.title, content = new.content, tags = new.tags
    WHERE rowid = new.id;
END;
