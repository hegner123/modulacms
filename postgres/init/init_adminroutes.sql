CREATE TABLE IF NOT EXISTS adminroutes (
    id SERIAL PRIMARY KEY,                        -- Auto-incrementing ID for PostgreSQL
    author TEXT NOT NULL,                         -- Not null constraint for important fields
    authorid UUID NOT NULL,                       -- UUID for better foreign key referencing
    slug TEXT UNIQUE NOT NULL,                    -- UNIQUE constraint for the slug field
    title TEXT NOT NULL,                          -- Title should not be null
    status SMALLINT NOT NULL,                     -- Use SMALLINT for status to save space
    datecreated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Use TIMESTAMP for date fields with default
    datemodified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,         -- Auto-update timestamp on modification
    content TEXT,                                 -- No constraints on content
    template TEXT,                                -- No constraints on template
    CONSTRAINT fk_author FOREIGN KEY (authorid) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT check_status CHECK (status IN (0, 1, 2)) -- Example constraint on status
);

-- Create index for fast lookups on commonly queried columns
CREATE INDEX idx_adminroutes_authorid ON adminroutes(authorid);
CREATE INDEX idx_adminroutes_status ON adminroutes(status);
CREATE INDEX idx_adminroutes_datecreated ON adminroutes(datecreated);

