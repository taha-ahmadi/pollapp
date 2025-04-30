package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"pollapp/internal/adapter/redisadapter"
	"pollapp/internal/config"
	"pollapp/internal/repository"
	"pollapp/internal/repository/postgresql/pollrepo"
	redisdb "pollapp/internal/repository/redis"
	"pollapp/internal/repository/redis/feedrepo"
	"pollapp/internal/repository/redis/voterepo"
	"pollapp/pkg/zlog"
)

var (
	interval   = flag.Int("interval", 60, "Sync interval in seconds")
	once       = flag.Bool("once", false, "Run sync only once and exit")
	configPath = flag.String("config", ".env", "Path to config file")
)

func main() {
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	zlog.New(cfg.Application.RunMode)
	defer zlog.Sync()

	zlog.L.Info("Configuration loaded successfully")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		zlog.L.Info("Received shutdown signal, gracefully shutting down...")
		cancel()
	}()

	postgresRepo, pgPool, err := setupPostgresRepo(cfg.Postgres)
	if err != nil {
		zlog.L.Fatal("Failed to setup Postgres repository", zlog.Error(err))
	}
	defer pgPool.Close()

	redisRepo, _, err := setupRedisRepo(cfg.Redis)
	if err != nil {
		zlog.L.Fatal("Failed to setup Redis repository", zlog.Error(err))
	}

	syncer := NewSyncer(postgresRepo, redisRepo)

	if *once {
		if err := syncer.SyncAll(ctx); err != nil {
			zlog.L.Fatal("Sync failed", zlog.Error(err))
		}
		zlog.L.Info("Sync completed successfully")
		return
	}

	ticker := time.NewTicker(time.Duration(*interval) * time.Second)
	defer ticker.Stop()

	if err := syncer.SyncAll(ctx); err != nil {
		zlog.L.Error("Initial sync failed", zlog.Error(err))
	} else {
		zlog.L.Info("Initial sync completed successfully")
	}

	for {
		select {
		case <-ticker.C:
			if err := syncer.SyncAll(ctx); err != nil {
				zlog.L.Error("Sync failed", zlog.Error(err))
			} else {
				zlog.L.Info("Sync completed successfully")
			}
		case <-ctx.Done():
			zlog.L.Info("Syncer shutting down...")
			return
		}
	}
}

// Syncer handles synchronization between Postgres and Redis
type Syncer struct {
	pgRepo    pollrepo.IPollRepository
	redisRepo *feedrepo.Repository
	mu        sync.Mutex
}

// NewSyncer creates a new Syncer
func NewSyncer(pgRepo pollrepo.IPollRepository, redisRepo *feedrepo.Repository) *Syncer {
	return &Syncer{
		pgRepo:    pgRepo,
		redisRepo: redisRepo,
	}
}

// SyncAll synchronizes all data from Postgres to Redis
func (s *Syncer) SyncAll(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	zlog.L.Info("Starting synchronization from Postgres to Redis...")

	if err := s.syncPollTags(ctx); err != nil {
		return fmt.Errorf("failed to sync poll tags: %w", err)
	}

	zlog.L.Info("Synchronization completed")
	return nil
}

// syncPollTags synchronizes poll tags from Postgres to Redis
func (s *Syncer) syncPollTags(ctx context.Context) error {
	polls, err := s.pgRepo.GetAllWithTags(ctx)
	if err != nil {
		return fmt.Errorf("failed to get polls with tags: %w", err)
	}

	zlog.L.Info("Syncing tags for polls", zlog.Int("count", len(polls)))

	for _, poll := range polls {
		for _, tag := range poll.Tags {
			err := s.redisRepo.AddPollToTag(ctx, tag, int(poll.ID), poll.CreatedAt)
			if err != nil {
				zlog.L.Warn("Failed to add poll to tag in Redis",
					zlog.Error(err),
					zlog.Any("poll_id", poll.ID),
					zlog.String("tag", tag))
			}
		}
	}

	return nil
}

// setupPostgresRepo sets up the Postgres repository
func setupPostgresRepo(pgConfig *config.Postgres) (pollrepo.IPollRepository, *repository.PostgresDB, error) {
	db, err := repository.NewPostgresDB(pgConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to postgres: %w", err)
	}

	pollRepo := pollrepo.New(db.Pool)

	return pollRepo, db, nil
}

// setupRedisRepo sets up the Redis repository
func setupRedisRepo(redisConfig *redisadapter.Config) (*feedrepo.Repository, *redisadapter.Adapter, error) {
	redisAdapter, err := redisadapter.New(*redisConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to redis: %w", err)
	}

	redisDB := redisdb.New(redisAdapter.Client())
	voteRepo := voterepo.New(redisDB)
	feedRepo := feedrepo.New(redisDB, voteRepo)

	return feedRepo, &redisAdapter, nil
}
