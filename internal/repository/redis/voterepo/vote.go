package voterepo

import (
	redisdb "pollapp/internal/repository/redis"
	"time"
)

const (
	userVotedPollsKey   = "user:%d:voted_polls"    // Set of poll IDs the user has voted on
	userSkippedPollsKey = "user:%d:skipped_polls"  // Set of poll IDs the user has skipped
	userDailyVotesKey   = "user:%d:daily_votes:%s" // Daily vote count for a user (with date)
	pollStatsKey        = "poll:%d:stats"          // Poll statistics (vote counts)
	cacheTTL            = 24 * time.Hour           // Cache TTL for most items
	statsCacheTTL       = 1 * time.Hour            // Cache TTL for poll statistics
	dailyVotesCacheTTL  = 48 * time.Hour           // Cache TTL for daily votes (longer to ensure we don't lose counts)
)

type Repository struct {
	db *redisdb.DB
}

func New(db *redisdb.DB) *Repository {
	return &Repository{db: db}
}
