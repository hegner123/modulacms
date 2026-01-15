-- Migration: Add implementation-plans and active-implementation-plan categories
-- SQLite doesn't support modifying CHECK constraints, so we need to recreate the table

BEGIN TRANSACTION;

-- Create new table with updated CHECK constraint
CREATE TABLE memories_new (
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
    tags TEXT,
    author TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT
);

-- Copy all data from old table to new table
INSERT INTO memories_new (id, title, content, category, tags, author, created_at, updated_at, metadata)
SELECT id, title, content, category, tags, author, created_at, updated_at, metadata
FROM memories;

-- Drop old table
DROP TABLE memories;

-- Rename new table to original name
ALTER TABLE memories_new RENAME TO memories;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_memories_category ON memories(category);
CREATE INDEX IF NOT EXISTS idx_memories_created ON memories(created_at DESC);

-- Recreate FTS table
DROP TABLE IF EXISTS memories_fts;

CREATE VIRTUAL TABLE memories_fts USING fts5(
    title,
    content,
    tags,
    content=memories,
    content_rowid=id
);

-- Repopulate FTS table
INSERT INTO memories_fts(rowid, title, content, tags)
SELECT id, title, content, tags FROM memories;

-- Recreate triggers
DROP TRIGGER IF EXISTS memories_ai;
DROP TRIGGER IF EXISTS memories_ad;
DROP TRIGGER IF EXISTS memories_au;

CREATE TRIGGER memories_ai AFTER INSERT ON memories BEGIN
    INSERT INTO memories_fts(rowid, title, content, tags)
    VALUES (new.id, new.title, new.content, new.tags);
END;

CREATE TRIGGER memories_ad AFTER DELETE ON memories BEGIN
    DELETE FROM memories_fts WHERE rowid = old.id;
END;

CREATE TRIGGER memories_au AFTER UPDATE ON memories BEGIN
    UPDATE memories_fts
    SET title = new.title, content = new.content, tags = new.tags
    WHERE rowid = new.id;
END;

COMMIT;
