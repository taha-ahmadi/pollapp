package pollservice

import (
	"errors"
	"pollapp/internal/adapter/broker"
	"pollapp/internal/repository/postgresql/pollrepo"
	"pollapp/internal/repository/redis/feedrepo"
	"pollapp/internal/repository/redis/voterepo"
)

var (
	ErrDailyVoteLimitExceeded = errors.New("daily vote limit exceeded")
	ErrAlreadyVoted           = errors.New("user has already voted on this poll")
	ErrAlreadySkipped         = errors.New("user has already skipped this poll")
)

// Service implements the PollService interface
type Service struct {
	pollRepo pollrepo.IPollRepository
	voteRepo voterepo.IVoteRepository
	feedRepo feedrepo.IFeedRepository
	broker   broker.MessageBroker
}

// NewPollService creates a new PollServiceImpl
func New(
	pollRepo pollrepo.IPollRepository,
	voteRepo voterepo.IVoteRepository,
	feedRepo feedrepo.IFeedRepository,
	messageBroker broker.MessageBroker,
) IPollService {
	return Service{
		pollRepo: pollRepo,
		voteRepo: voteRepo,
		feedRepo: feedRepo,
		broker:   messageBroker,
	}
}
