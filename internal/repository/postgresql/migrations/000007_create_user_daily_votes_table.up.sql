-- Create user_daily_votes table to track daily vote limits
CREATE TABLE IF NOT EXISTS user_daily_votes (
    user_id INTEGER NOT NULL,
    vote_date DATE NOT NULL,
    vote_count INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, vote_date)
);

-- Create index for quick lookups
CREATE INDEX IF NOT EXISTS idx_user_daily_votes_user_id_date ON user_daily_votes(user_id, vote_date); 