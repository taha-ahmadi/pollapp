-- Drop indexes
DROP INDEX IF EXISTS idx_votes_user_id;
DROP INDEX IF EXISTS idx_votes_poll_id;

-- Drop votes table
DROP TABLE IF EXISTS votes; 