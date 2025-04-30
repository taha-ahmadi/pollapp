package voterepo

import (
	"context"
	"pollapp/pkg/monitoring"
	"time"
)

// ** MetricsRepository wraps the base vote repository with metrics **
type MetricsRepository struct {
	repo IVoteRepository
}

func NewWithMetrics(repo IVoteRepository) IVoteRepository {
	return &MetricsRepository{
		repo: repo,
	}
}

func (m *MetricsRepository) Vote(ctx context.Context, userID uint, pollID uint) error {
	start := time.Now()
	err := m.repo.Vote(ctx, userID, pollID)
	duration := time.Since(start).Seconds()

	monitoring.CacheOperationDuration.WithLabelValues("vote").Observe(duration)
	if err != nil {
		monitoring.CacheMissesTotal.WithLabelValues("vote").Inc()
	} else {
		monitoring.CacheHitsTotal.WithLabelValues("vote").Inc()
	}

	return err
}

func (m *MetricsRepository) Skip(ctx context.Context, userID uint, pollID uint) error {
	start := time.Now()
	err := m.repo.Skip(ctx, userID, pollID)
	duration := time.Since(start).Seconds()

	monitoring.CacheOperationDuration.WithLabelValues("skip").Observe(duration)
	if err != nil {
		monitoring.CacheMissesTotal.WithLabelValues("skip").Inc()
	} else {
		monitoring.CacheHitsTotal.WithLabelValues("skip").Inc()
	}

	return err
}

func (m *MetricsRepository) HasUserInteractedWithPoll(ctx context.Context, userID uint, pollID uint) (bool, error) {
	start := time.Now()
	result, err := m.repo.HasUserInteractedWithPoll(ctx, userID, pollID)
	duration := time.Since(start).Seconds()

	monitoring.CacheOperationDuration.WithLabelValues("check_interaction").Observe(duration)
	if err != nil {
		monitoring.CacheMissesTotal.WithLabelValues("check_interaction").Inc()
	} else {
		monitoring.CacheHitsTotal.WithLabelValues("check_interaction").Inc()
	}

	return result, err
}

func (m *MetricsRepository) GetInteractedPollIDs(ctx context.Context, userID int) ([]int, error) {
	start := time.Now()
	result, err := m.repo.GetInteractedPollIDs(ctx, userID)
	duration := time.Since(start).Seconds()

	monitoring.CacheOperationDuration.WithLabelValues("get_interacted_polls").Observe(duration)
	if err != nil {
		monitoring.CacheMissesTotal.WithLabelValues("get_interacted_polls").Inc()
	} else {
		monitoring.CacheHitsTotal.WithLabelValues("get_interacted_polls").Inc()
	}

	return result, err
}

func (m *MetricsRepository) IncrementDailyCount(ctx context.Context, userID uint) (int, error) {
	start := time.Now()
	result, err := m.repo.IncrementDailyCount(ctx, userID)
	duration := time.Since(start).Seconds()

	monitoring.CacheOperationDuration.WithLabelValues("increment_daily_count").Observe(duration)
	if err != nil {
		monitoring.CacheMissesTotal.WithLabelValues("increment_daily_count").Inc()
	} else {
		monitoring.CacheHitsTotal.WithLabelValues("increment_daily_count").Inc()
	}

	return result, err
}

func (m *MetricsRepository) GetDailyCount(ctx context.Context, userID uint) (int, error) {
	start := time.Now()
	result, err := m.repo.GetDailyCount(ctx, userID)
	duration := time.Since(start).Seconds()

	monitoring.CacheOperationDuration.WithLabelValues("get_daily_count").Observe(duration)
	if err != nil {
		monitoring.CacheMissesTotal.WithLabelValues("get_daily_count").Inc()
	} else {
		monitoring.CacheHitsTotal.WithLabelValues("get_daily_count").Inc()
	}

	return result, err
}

func (m *MetricsRepository) StorePollStats(ctx context.Context, pollID uint, stats interface{}) error {
	start := time.Now()
	err := m.repo.StorePollStats(ctx, pollID, stats)
	duration := time.Since(start).Seconds()

	monitoring.CacheOperationDuration.WithLabelValues("store_poll_stats").Observe(duration)
	if err != nil {
		monitoring.CacheMissesTotal.WithLabelValues("store_poll_stats").Inc()
	} else {
		monitoring.CacheHitsTotal.WithLabelValues("store_poll_stats").Inc()
	}

	return err
}

// GetPollStats gets poll statistics from the cache
func (m *MetricsRepository) GetPollStats(ctx context.Context, pollID uint, result interface{}) (bool, error) {
	start := time.Now()
	found, err := m.repo.GetPollStats(ctx, pollID, result)
	duration := time.Since(start).Seconds()

	monitoring.CacheOperationDuration.WithLabelValues("get_poll_stats").Observe(duration)
	if err != nil || !found {
		monitoring.CacheMissesTotal.WithLabelValues("get_poll_stats").Inc()
	} else {
		monitoring.CacheHitsTotal.WithLabelValues("get_poll_stats").Inc()
	}

	return found, err
}
