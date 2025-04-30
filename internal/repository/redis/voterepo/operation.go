package voterepo

import (
	"context"
	"encoding/json"
	"fmt"
	redisdb "pollapp/internal/repository/redis"
	"strconv"
	"time"
)

func (r *Repository) Vote(ctx context.Context, userID uint, pollID uint) error {
	votedKey := fmt.Sprintf(userVotedPollsKey, userID)
	pollIDStr := fmt.Sprintf("%d", pollID)
	err := r.db.Client().SAdd(ctx, votedKey, pollIDStr).Err()
	if err != nil {
		return fmt.Errorf("add to voted set: %w", err)
	}

	statsKey := fmt.Sprintf(pollStatsKey, pollID)
	r.db.Client().Del(ctx, statsKey)

	return nil
}

func (r *Repository) Skip(ctx context.Context, userID uint, pollID uint) error {
	skippedKey := fmt.Sprintf(userSkippedPollsKey, userID)
	pollIDStr := fmt.Sprintf("%d", pollID)
	err := r.db.Client().SAdd(ctx, skippedKey, pollIDStr).Err()
	if err != nil {
		return fmt.Errorf("add to skipped set: %w", err)
	}
	return nil
}

func (r *Repository) HasUserInteractedWithPoll(ctx context.Context, userID uint, pollID uint) (bool, error) {
	votedKey := fmt.Sprintf(userVotedPollsKey, userID)
	pollIDStr := fmt.Sprintf("%d", pollID)
	hasVoted, err := r.db.Client().SIsMember(ctx, votedKey, pollIDStr).Result()
	if err != nil {
		return false, fmt.Errorf("check voted set: %w", err)
	}

	if hasVoted {
		return true, nil
	}

	skippedKey := fmt.Sprintf(userSkippedPollsKey, userID)
	hasSkipped, err := r.db.Client().SIsMember(ctx, skippedKey, pollIDStr).Result()
	if err != nil {
		return false, fmt.Errorf("check skipped set: %w", err)
	}

	return hasSkipped, nil
}

func (r *Repository) GetInteractedPollIDs(ctx context.Context, userID int) ([]int, error) {
	votedKey := fmt.Sprintf(userVotedPollsKey, userID)
	skippedKey := fmt.Sprintf(userSkippedPollsKey, userID)

	interactedPollStrs, err := r.db.Client().SUnion(ctx, votedKey, skippedKey).Result()
	if err != nil {
		return nil, fmt.Errorf("get interacted polls: %w", err)
	}

	interactedPolls := make([]int, 0, len(interactedPollStrs))
	for _, idStr := range interactedPollStrs {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, fmt.Errorf("convert poll ID to int: %w", err)
		}
		interactedPolls = append(interactedPolls, id)
	}

	return interactedPolls, nil
}

func (r *Repository) IncrementDailyCount(ctx context.Context, userID uint) (int, error) {
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf(userDailyVotesKey, userID, today)

	count, err := r.db.Client().Incr(ctx, dailyKey).Result()
	if err != nil {
		return 0, fmt.Errorf("increment daily vote count: %w", err)
	}

	if count == 1 {
		r.db.Client().Expire(ctx, dailyKey, dailyVotesCacheTTL)
	}

	return int(count), nil
}

func (r *Repository) GetDailyCount(ctx context.Context, userID uint) (int, error) {
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf(userDailyVotesKey, userID, today)

	countStr, err := r.db.Client().Get(ctx, dailyKey).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return 0, nil
		}
		return 0, fmt.Errorf("get daily vote count: %w", err)
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("parse vote count: %w", err)
	}

	return count, nil
}

func (r *Repository) StorePollStats(ctx context.Context, pollID uint, stats interface{}) error {
	statsKey := fmt.Sprintf(pollStatsKey, pollID)
	return r.db.Set(ctx, statsKey, stats, statsCacheTTL)
}

func (r *Repository) GetPollStats(ctx context.Context, pollID uint, result interface{}) (bool, error) {
	statsKey := fmt.Sprintf(pollStatsKey, pollID)
	data, err := r.db.Get(ctx, statsKey)

	if err != nil {
		if err == redisdb.NotExist {
			return false, nil
		}
		return false, fmt.Errorf("get poll stats: %w", err)
	}

	err = json.Unmarshal([]byte(data), result)
	if err != nil {
		return false, fmt.Errorf("unmarshal poll stats: %w", err)
	}

	return true, nil
}
