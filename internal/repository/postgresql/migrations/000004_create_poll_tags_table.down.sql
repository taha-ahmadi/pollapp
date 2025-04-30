-- Drop indexes
DROP INDEX IF EXISTS idx_poll_tags_poll_id;
DROP INDEX IF EXISTS idx_poll_tags_tag_id;

-- Drop poll_tags table
DROP TABLE IF EXISTS poll_tags; 