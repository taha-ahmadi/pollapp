-- Create skips table
CREATE TABLE IF NOT EXISTS skips (
    id SERIAL PRIMARY KEY,
    poll_id INTEGER NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(poll_id, user_id)
);

-- Create indexes for quick lookups
CREATE INDEX IF NOT EXISTS idx_skips_user_id ON skips(user_id);
CREATE INDEX IF NOT EXISTS idx_skips_poll_id ON skips(poll_id); 