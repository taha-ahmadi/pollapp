-- Create poll_tags table
CREATE TABLE IF NOT EXISTS poll_tags (
    poll_id INTEGER NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (poll_id, tag_id)
);

-- Create index for quick lookups
CREATE INDEX IF NOT EXISTS idx_poll_tags_poll_id ON poll_tags(poll_id); 