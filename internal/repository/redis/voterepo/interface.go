package voterepo

import (
	"context"
)

// VoteRepositoryInterface defines the interface for vote repository operations
type IVoteRepository interface {
	Vote(ctx context.Context, userID uint, pollID uint) error
	Skip(ctx context.Context, userID uint, pollID uint) error
	HasUserInteractedWithPoll(ctx context.Context, userID uint, pollID uint) (bool, error)
	GetInteractedPollIDs(ctx context.Context, userID int) ([]int, error)
	IncrementDailyCount(ctx context.Context, userID uint) (int, error)
	GetDailyCount(ctx context.Context, userID uint) (int, error)
	StorePollStats(ctx context.Context, pollID uint, stats interface{}) error
	GetPollStats(ctx context.Context, pollID uint, result interface{}) (bool, error)
}
