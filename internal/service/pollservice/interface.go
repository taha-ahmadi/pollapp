package pollservice

import (
	"context"
	"pollapp/internal/service/domain"
)

// IPollService is the interface that provides poll operations
type IPollService interface {
	Create(ctx context.Context, poll domain.Poll) error
	GetByID(ctx context.Context, id uint) (domain.Poll, error)
	Get(ctx context.Context, params *domain.GetPollsParams) ([]domain.Poll, error)
	QueueVote(ctx context.Context, pollID uint, vote domain.VoteRequest) error
	Skip(ctx context.Context, pollID uint, skip domain.SkipRequest) error
	GetStats(ctx context.Context, pollID uint) (domain.PollStats, error)
}
