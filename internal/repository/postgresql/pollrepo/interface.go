package pollrepo

import (
	"context"
	"pollapp/internal/repository/model"
	"pollapp/internal/service/domain"
)

// IPollRepository defines methods for poll-related database operations
type IPollRepository interface {
	Create(ctx context.Context, poll model.Poll) error
	Get(ctx context.Context, filter domain.GetPollsParams) ([]domain.Poll, error)
	GetByID(ctx context.Context, id uint) (domain.Poll, error)
	Vote(ctx context.Context, pollID uint, vote model.Vote) error
	Skip(ctx context.Context, pollID uint, skip domain.SkipRequest) error
	GetStats(ctx context.Context, pollID uint) (*domain.PollStats, error)
	CheckDailyVoteLimit(ctx context.Context, userID uint) (int, error)
	GetAllWithTags(ctx context.Context) ([]domain.Poll, error)
}
