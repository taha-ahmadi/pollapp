package feedrepo

import (
	"context"
	"encoding/json"
	"fmt"
	redisdb "pollapp/internal/repository/redis"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func (r *Repository) Cache(ctx context.Context, userID int, tag string, page int, limit int, feed interface{}) error {
	key := fmt.Sprintf(feedCacheKey, userID, tag, page, limit)
	return r.db.Set(ctx, key, feed, feedCacheTTL)
}

func (r *Repository) GetCached(ctx context.Context, userID int, tag string, page int, limit int, result interface{}) (bool, error) {
	key := fmt.Sprintf(feedCacheKey, userID, tag, page, limit)
	data, err := r.db.Get(ctx, key)

	if err != nil {
		if err == redisdb.NotExist {
			return false, nil
		}
		return false, fmt.Errorf("get cached feed: %w", err)
	}

	err = json.Unmarshal([]byte(data), result)
	if err != nil {
		return false, fmt.Errorf("unmarshal feed: %w", err)
	}

	return true, nil
}

func (r *Repository) AddPollToTag(ctx context.Context, tag string, pollID int, timestamp time.Time) error {
	err := r.db.Client().SAdd(ctx, tagListKey, tag).Err()
	if err != nil {
		return fmt.Errorf("add tag to list: %w", err)
	}

	pollIDStr := strconv.Itoa(pollID)

	score := float64(timestamp.Unix())
	tagKey := fmt.Sprintf(tagPollsKey, tag)
	err = r.db.Client().ZAdd(ctx, tagKey, redis.Z{
		Score:  score,
		Member: pollIDStr,
	}).Err()
	if err != nil {
		return fmt.Errorf("add poll to tag: %w", err)
	}

	err = r.db.Client().ZAdd(ctx, recentPollsKey, redis.Z{
		Score:  score,
		Member: pollIDStr,
	}).Err()
	if err != nil {
		return fmt.Errorf("add poll to recent: %w", err)
	}

	return nil
}

func (r *Repository) GetPollIDsByTag(ctx context.Context, tag string, page int, limit int) ([]int, error) {
	offset := (page - 1) * limit

	var key string
	if tag == "" {
		key = recentPollsKey
	} else {
		key = fmt.Sprintf(tagPollsKey, tag)
	}

	pollIDStrs, err := r.db.Client().ZRevRange(ctx, key, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("get polls by tag: %w", err)
	}

	pollIDs := make([]int, 0, len(pollIDStrs))
	for _, idStr := range pollIDStrs {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, fmt.Errorf("convert poll ID to int: %w", err)
		}
		pollIDs = append(pollIDs, id)
	}

	return pollIDs, nil
}

// GetRecentPollIDs gets the most recent poll IDs regardless of tag
func (r *Repository) GetRecentPollIDs(ctx context.Context, page int, limit int) ([]int, error) {
	offset := (page - 1) * limit

	pollIDStrs, err := r.db.Client().ZRevRange(ctx, recentPollsKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("get recent polls: %w", err)
	}

	pollIDs := make([]int, 0, len(pollIDStrs))
	for _, idStr := range pollIDStrs {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, fmt.Errorf("convert poll ID to int: %w", err)
		}
		pollIDs = append(pollIDs, id)
	}

	return pollIDs, nil
}

// GetFilteredFeedPollIDs gets poll IDs for a feed filtered by tag and excluding
func (r *Repository) GetFilteredFeedPollIDs(ctx context.Context, userID int, tag string, page int, limit int) ([]int, error) {
	interactedPolls, err := r.voteRepo.GetInteractedPollIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get interacted polls: %w", err)
	}

	var pollIDs []int
	if tag == "" {
		pollIDs, err = r.GetRecentPollIDs(ctx, page, limit*2) // Get twice the limit to account for filtering
	} else {
		pollIDs, err = r.GetPollIDsByTag(ctx, tag, page, limit*2) // Get twice the limit to account for filtering
	}
	if err != nil {
		return nil, fmt.Errorf("get poll IDs: %w", err)
	}

	interactedMap := make(map[int]bool)
	for _, id := range interactedPolls {
		interactedMap[id] = true
	}

	filteredPolls := make([]int, 0, limit)
	for _, id := range pollIDs {
		if !interactedMap[id] {
			filteredPolls = append(filteredPolls, id)
			if len(filteredPolls) >= limit {
				break
			}
		}
	}

	return filteredPolls, nil
}

// InvalidateUserFeeds invalidates all cached feeds for a user
func (r *Repository) InvalidateUserFeeds(ctx context.Context, userID uint) error {
	pattern := fmt.Sprintf("user:%d:feed:*", userID)
	keys, err := r.db.Client().Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("find feed keys: %w", err)
	}

	if len(keys) > 0 {
		err = r.db.Client().Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("delete feed keys: %w", err)
		}
	}

	return nil
}
