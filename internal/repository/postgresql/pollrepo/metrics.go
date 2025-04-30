package pollrepo

import (
	"context"
	"fmt"
	"pollapp/internal/repository/model"
	"pollapp/internal/service/domain"
	"pollapp/pkg/monitoring"
)

func (m *MetricsRepository) Create(ctx context.Context, poll model.Poll) error {
	defer monitoring.MeasureDuration(monitoring.DBQueryDuration, []string{"create_poll"})()

	err := m.repo.Create(ctx, poll)

	if err != nil {
		monitoring.DBQueriesTotal.WithLabelValues("create_poll", "error").Inc()
	} else {
		monitoring.DBQueriesTotal.WithLabelValues("create_poll", "success").Inc()
		monitoring.PollCreatedTotal.Inc()
	}

	return err
}

func (m *MetricsRepository) Get(ctx context.Context, filter domain.GetPollsParams) ([]domain.Poll, error) {
	defer monitoring.MeasureDuration(monitoring.DBQueryDuration, []string{"get_polls"})()

	polls, err := m.repo.Get(ctx, filter)

	if err != nil {
		monitoring.DBQueriesTotal.WithLabelValues("get_polls", "error").Inc()
	} else {
		monitoring.DBQueriesTotal.WithLabelValues("get_polls", "success").Inc()
	}

	return polls, err
}

func (m *MetricsRepository) GetByID(ctx context.Context, id uint) (domain.Poll, error) {
	defer monitoring.MeasureDuration(monitoring.DBQueryDuration, []string{"get_poll_by_id"})()

	poll, err := m.repo.GetByID(ctx, id)

	if err != nil {
		monitoring.DBQueriesTotal.WithLabelValues("get_poll_by_id", "error").Inc()
	} else {
		monitoring.DBQueriesTotal.WithLabelValues("get_poll_by_id", "success").Inc()
	}

	return poll, err
}

func (m *MetricsRepository) Vote(ctx context.Context, pollID uint, vote model.Vote) error {
	defer monitoring.MeasureDuration(monitoring.DBQueryDuration, []string{"vote"})()

	err := m.repo.Vote(ctx, pollID, vote)

	if err != nil {
		monitoring.DBQueriesTotal.WithLabelValues("vote", "error").Inc()
	} else {
		monitoring.DBQueriesTotal.WithLabelValues("vote", "success").Inc()
		monitoring.VotesReceivedTotal.WithLabelValues(fmt.Sprintf("%d", pollID), fmt.Sprintf("%d", vote.OptionIndex)).Inc()
	}

	return err
}

func (m *MetricsRepository) Skip(ctx context.Context, pollID uint, skip domain.SkipRequest) error {
	defer monitoring.MeasureDuration(monitoring.DBQueryDuration, []string{"skip_poll"})()

	err := m.repo.Skip(ctx, pollID, skip)

	if err != nil {
		monitoring.DBQueriesTotal.WithLabelValues("skip_poll", "error").Inc()
	} else {
		monitoring.DBQueriesTotal.WithLabelValues("skip_poll", "success").Inc()
	}

	return err
}

func (m *MetricsRepository) GetStats(ctx context.Context, pollID uint) (*domain.PollStats, error) {
	defer monitoring.MeasureDuration(monitoring.DBQueryDuration, []string{"get_poll_stats"})()

	stats, err := m.repo.GetStats(ctx, pollID)

	if err != nil {
		monitoring.DBQueriesTotal.WithLabelValues("get_poll_stats", "error").Inc()
	} else {
		monitoring.DBQueriesTotal.WithLabelValues("get_poll_stats", "success").Inc()
	}

	return stats, err
}

func (m *MetricsRepository) CheckDailyVoteLimit(ctx context.Context, userID uint) (int, error) {
	defer monitoring.MeasureDuration(monitoring.DBQueryDuration, []string{"check_daily_vote_limit"})()

	count, err := m.repo.CheckDailyVoteLimit(ctx, userID)

	if err != nil {
		monitoring.DBQueriesTotal.WithLabelValues("check_daily_vote_limit", "error").Inc()
	} else {
		monitoring.DBQueriesTotal.WithLabelValues("check_daily_vote_limit", "success").Inc()
	}

	return count, err
}

func (m *MetricsRepository) GetAllWithTags(ctx context.Context) ([]domain.Poll, error) {
	defer monitoring.MeasureDuration(monitoring.DBQueryDuration, []string{"get_all_with_tags"})()

	polls, err := m.repo.GetAllWithTags(ctx)

	if err != nil {
		monitoring.DBQueriesTotal.WithLabelValues("get_all_with_tags", "error").Inc()
	} else {
		monitoring.DBQueriesTotal.WithLabelValues("get_all_with_tags", "success").Inc()
	}

	return polls, err
}
