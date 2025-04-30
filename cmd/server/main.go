package main

import (
	"context"
	"flag"
	"log"
	stdhttp "net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	_ "pollapp/docs"
	"pollapp/internal/adapter/broker/rabbitmq"
	"pollapp/internal/adapter/redisadapter"
	"pollapp/internal/config"
	"pollapp/internal/delivery/http"
	"pollapp/internal/delivery/http/pollhandler"
	"pollapp/internal/repository"
	"pollapp/internal/repository/postgresql/pollrepo"
	redisdb "pollapp/internal/repository/redis"
	"pollapp/internal/repository/redis/feedrepo"
	"pollapp/internal/repository/redis/voterepo"
	"pollapp/internal/service/pollservice"
	"pollapp/internal/worker"
	"pollapp/pkg/zlog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title Poll API
// @version 1.0
// @description API for managing polls and votes
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Define command-line flags
	migrateUp := flag.Bool("migrate-up", false, "Run database migrations up")
	migrateDown := flag.Bool("migrate-down", false, "Run database migrations down")
	runWorkers := flag.Bool("workers", true, "Run vote processing workers (default: true)")
	workerCount := flag.Int("worker-count", 5, "Number of vote processing workers (default: 5)")
	flag.Parse()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = ".env"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	zlog.New(cfg.Application.RunMode)
	defer zlog.Sync()

	zlog.L.Info("Configuration loaded successfully")

	cwd, err := os.Getwd()
	if err != nil {
		zlog.L.Fatal("Failed to get current working directory", zlog.Error(err))
	}
	migrationsPath := filepath.Join(cwd, "internal", "repository", "postgresql", "migrations")

	dbURL := cfg.Postgres.DSN()

	if *migrateUp {
		zlog.L.Info("Running database migrations up")
		if err := repository.MigrateUp(dbURL, migrationsPath); err != nil {
			zlog.L.Fatal("Failed to apply migrations up", zlog.Error(err))
		}
		zlog.L.Info("Database migrations up completed successfully")
		return
	}

	if *migrateDown {
		zlog.L.Info("Running database migrations down")
		if err := repository.MigrateDown(dbURL, migrationsPath); err != nil {
			zlog.L.Fatal("Failed to apply migrations down", zlog.Error(err))
		}
		zlog.L.Info("Database migrations down completed successfully")
		return
	}

	// Initialize database connections
	db, err := repository.NewPostgresDB(cfg.Postgres)
	if err != nil {
		zlog.L.Fatal("Failed to connect to database", zlog.Error(err))
	}
	defer db.Close()

	// Initialize Redis adapter
	redisAdapter, err := redisadapter.New(*cfg.Redis)
	if err != nil {
		zlog.L.Fatal("Failed to connect to Redis", zlog.Error(err))
	}

	// Initialize RabbitMQ broker
	rabbitMQBroker, err := rabbitmq.New(cfg.RabbitMQ, "api", []string{worker.VoteQueueName})
	if err != nil {
		zlog.L.Fatal("Failed to connect to RabbitMQ", zlog.Error(err))
	}

	// Initialize repositories
	pollRepo := pollrepo.New(db.Pool)
	redisDB := redisdb.New(redisAdapter.Client())
	voteRepo := voterepo.New(redisDB)
	feedRepo := feedrepo.New(redisDB, voteRepo)

	// Initialize service with message broker
	pollService := pollservice.New(pollRepo, voteRepo, feedRepo, rabbitMQBroker)

	// Start worker pool if requested
	var voteWorker *worker.VoteWorker
	if *runWorkers {
		zlog.L.Info("Starting vote processing workers", zlog.Int("count", *workerCount))
		voteWorker = worker.NewVoteWorker(rabbitMQBroker, pollRepo, worker.VoteQueueName, *workerCount)
		if err := voteWorker.Start(context.Background()); err != nil {
			zlog.L.Fatal("Failed to start vote workers", zlog.Error(err))
		}
		zlog.L.Info("Vote processing workers started successfully", zlog.Int("count", *workerCount))
	} else {
		zlog.L.Info("Workers disabled, running in API-only mode")
	}

	// Initialize handler
	pollHandler := pollhandler.NewHandler(pollService)

	// Initialize Echo server
	e := echo.New()

	// Add validator
	e.Validator = http.NewValidator()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Setup routes
	pollHandler.SetupRoutes(e)

	// Start server
	serverAddr := cfg.Server.Host + ":" + strconv.FormatUint(cfg.Server.Port, 10)
	zlog.L.Info("Starting server", zlog.String("address", serverAddr))

	// Start server in a goroutine
	go func() {
		if err := e.Start(serverAddr); err != nil && err != stdhttp.ErrServerClosed {
			zlog.L.Fatal("Failed to start server", zlog.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop workers if running
	if voteWorker != nil {
		zlog.L.Info("Stopping vote workers")
		voteWorker.Stop()
		zlog.L.Info("Vote workers stopped gracefully")
	}

	// Close RabbitMQ connection
	if rabbitMQBroker != nil {
		if err := rabbitMQBroker.Close(); err != nil {
			zlog.L.Error("Error closing RabbitMQ connection", zlog.Error(err))
		}
	}

	// Shutdown the server
	if err := e.Shutdown(ctx); err != nil {
		zlog.L.Fatal("Server forced to shutdown", zlog.Error(err))
	}

	zlog.L.Info("Server exited properly")
}
