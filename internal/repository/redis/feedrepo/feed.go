package feedrepo

import (
	redisdb "pollapp/internal/repository/redis"
	"pollapp/internal/repository/redis/voterepo"
	"time"
)

const (
	feedCacheKey   = "user:%d:feed:tag:%s:page:%d:limit:%d" // User's feed cache key with tag filter
	feedCacheTTL   = 5 * time.Minute                        // Short TTL for feed cache to stay relatively fresh
	tagPollsKey    = "tag:%s:polls"                         // Sorted set of poll IDs for a tag (by timestamp)
	recentPollsKey = "recent_polls"                         // Sorted set of all recent polls (by timestamp)
	tagListKey     = "tags:list"                            // Set of all tags
)

type Repository struct {
	db       *redisdb.DB
	voteRepo voterepo.IVoteRepository
}

func New(db *redisdb.DB, voteRepo voterepo.IVoteRepository) *Repository {
	return &Repository{
		db:       db,
		voteRepo: voteRepo,
	}
}
