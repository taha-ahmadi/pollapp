package pollservice

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pollapp/internal/adapter/broker"
	"pollapp/internal/repository/model"
	"pollapp/internal/repository/postgresql/pollrepo"
	"pollapp/internal/service/domain"
	"pollapp/internal/worker"

	"github.com/google/uuid"
)

func (s Service) Create(ctx context.Context, poll domain.Poll) error {
	err := s.pollRepo.Create(ctx, model.ToModel(poll))
	if err != nil {
		return fmt.Errorf("create poll in postgres: %w", err)
	}

	for _, tag := range poll.Tags {
		err = s.feedRepo.AddPollToTag(ctx, tag, int(poll.ID), time.Now())
		if err != nil {
			// Log but don't fail if Redis update fails
			// A background process could sync Redis with Postgres
			fmt.Printf("Warning: failed to add poll to tag in Redis: %v\n", err)
		}
	}

	return nil
}

func (s Service) GetByID(ctx context.Context, id uint) (domain.Poll, error) {
	pollModel, err := s.pollRepo.GetByID(ctx, id)
	if err != nil {
		return domain.Poll{}, err
	}

	return pollModel, nil
}

func (s Service) Get(ctx context.Context, params *domain.GetPollsParams) ([]domain.Poll, error) {
	polls, err := s.pollRepo.Get(ctx, domain.GetPollsParams{
		UserID: params.UserID,
		Tag:    params.Tag,
		Page:   params.Page,
		Limit:  params.Limit,
	})
	if err != nil {
		return nil, err
	}

	domainPolls := make([]domain.Poll, len(polls))
	for i, poll := range polls {
		domainPolls[i] = poll
	}

	return domainPolls, nil
}

func (s Service) Skip(ctx context.Context, pollID uint, skip domain.SkipRequest) error {
	hasInteracted, err := s.voteRepo.HasUserInteractedWithPoll(
		ctx,
		skip.UserID,
		pollID,
	)
	if err == nil && hasInteracted {
		return ErrAlreadySkipped
	}

	err = s.pollRepo.Skip(ctx, pollID, skip)
	if err != nil {
		return fmt.Errorf("skip in postgres: %w", err)
	}

	err = s.voteRepo.Skip(
		ctx,
		skip.UserID,
		pollID,
	)
	if err != nil {
		// Log but don't fail if Redis update fails
		fmt.Printf("Warning: failed to record skip in Redis: %v\n", err)
	}

	err = s.feedRepo.InvalidateUserFeeds(ctx, skip.UserID)
	if err != nil {
		// Log but don't fail if Redis update fails
		fmt.Printf("Warning: failed to invalidate feed cache in Redis: %v\n", err)
	}

	return nil
}

func (s Service) GetStats(ctx context.Context, pollID uint) (domain.PollStats, error) {
	var stats domain.PollStats
	found, err := s.voteRepo.GetPollStats(ctx, pollID, &stats)
	if err != nil {
		fmt.Printf("Warning: failed to get poll stats from Redis: %v\n", err)
	}

	if found {
		return stats, nil
	}

	dbStats, err := s.pollRepo.GetStats(ctx, pollID)
	if err != nil {
		return domain.PollStats{}, fmt.Errorf("get poll stats: %w", err)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		// We're using a separate context to avoid leaking the parent context
		err := s.voteRepo.StorePollStats(ctx, pollID, *dbStats)
		if err != nil {
			fmt.Printf("Warning: failed to cache poll stats in Redis: %v\n", err)
		}
	}()

	return *dbStats, nil
}

func (s Service) QueueVote(ctx context.Context, pollID uint, vote domain.VoteRequest) error {
	hasInteracted, err := s.voteRepo.HasUserInteractedWithPoll(
		ctx,
		vote.UserID,
		pollID,
	)
	if err == nil && hasInteracted {
		return ErrAlreadyVoted
	}

	count, err := s.voteRepo.GetDailyCount(ctx, vote.UserID)
	if err == nil && count >= pollrepo.VoteLimit {
		return ErrDailyVoteLimitExceeded
	}

	voteJob := domain.VoteJob{
		ID:          uuid.New().String(),
		PollID:      pollID,
		UserID:      vote.UserID,
		OptionIndex: vote.OptionIndex,
		CreatedAt:   time.Now(),
	}

	jobBytes, err := json.Marshal(voteJob)
	if err != nil {
		return fmt.Errorf("failed to serialize vote job: %w", err)
	}

	message := broker.Message{
		ID:          voteJob.ID,
		Body:        jobBytes,
		MessageType: "vote",
		QueueName:   worker.VoteQueueName,
	}

	err = s.broker.Publish(message)
	if err != nil {
		return fmt.Errorf("failed to publish vote job: %w", err)
	}

	err = s.voteRepo.Vote(
		ctx,
		vote.UserID,
		pollID,
	)
	if err != nil {
		fmt.Printf("Warning: failed to record vote in Redis: %v\n", err)
	}

	_, err = s.voteRepo.IncrementDailyCount(ctx, vote.UserID)
	if err != nil {
		// Log but don't fail if Redis update fails
		fmt.Printf("Warning: failed to increment daily vote count in Redis: %v\n", err)
	}

	return nil
}
