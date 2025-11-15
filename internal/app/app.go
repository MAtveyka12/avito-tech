package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"avito-tech/configs"
	httpcontroller "avito-tech/internal/controllers/http"
	"avito-tech/internal/repository/pullrequests"
	"avito-tech/internal/repository/teams"
	"avito-tech/internal/repository/tx"
	"avito-tech/internal/repository/users"
	"avito-tech/internal/service"
	httpserver "avito-tech/pkg/http_server"
)

type App struct {
	httpserver *httpserver.HTTPServer
	pool       *pgxpool.Pool
}

func New() (*App, error) {
	configPath := filepath.Join("configs", "config.yaml")
	envPath := ".env"

	cfg, err := configs.Load(configPath, envPath)
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	pool, err := initPool(cfg, logger)
	if err != nil {
		logger.Error("failed to initialize database connection pool", slog.Any("error", err))
		os.Exit(1)
	}

	ctxManager := tx.NewCtxManager(pool)
	txManager := tx.NewTxManager(ctxManager, logger)

	teamsRepo := teams.NewTeamsRepository(pool, ctxManager)
	usersRepo := users.NewUsersRepository(pool, ctxManager)
	pullRequestsRepo := pullrequests.NewPullRequestsRepository(pool, ctxManager)

	teamsService := service.NewTeamsService(teamsRepo, txManager)
	usersService := service.NewUsersService(usersRepo, txManager)
	pullRequestsService := service.NewPullRequestsService(pullRequestsRepo, usersRepo, teamsRepo, txManager)

	handler := httpcontroller.NewHandler(teamsService, usersService, pullRequestsService)

	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	router.HEAD("/healthcheck", handler.HealthCheck)

	router.POST("/team/add", handler.CreateTeam)
	router.GET("/team/get", handler.GetTeam)

	router.POST("/users/setIsActive", handler.SetIsActive)
	router.GET("/users/getReview", handler.GetReview)
	router.POST("/users/bulkDeactivate", handler.BulkDeactivate)

	router.POST("/pullRequest/create", handler.CreatePullRequest)
	router.POST("/pullRequest/merge", handler.MergePullRequest)
	router.POST("/pullRequest/reassign", handler.ReassignReviewer)

	router.GET("/statistics", handler.GetStatistics)

	srv := httpserver.New(cfg, router)

	return &App{
		httpserver: srv,
		pool:       pool,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer func() {
		if a.pool != nil {
			a.pool.Close()
		}
	}()

	if err := a.httpserver.Start(ctx); err != nil {
		return fmt.Errorf("app run failed: %s", err.Error())
	}

	return nil
}

func initPool(cfg *configs.Config, logger *slog.Logger) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %s", err.Error())
	}

	poolCfg.MinConns = cfg.Database.MinConns
	poolCfg.MaxConns = cfg.Database.MaxConns
	poolCfg.MaxConnIdleTime = time.Duration(cfg.Database.MaxIdleTime) * time.Second
	poolCfg.MaxConnLifetime = time.Duration(cfg.Database.MaxLifeTime) * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("init pgxpool: %s", err.Error())
	}

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if err := pool.Ping(ctx); err == nil {
			return pool, nil
		}

		logger.Warn("Database not ready", slog.Int("attempt", i+1), slog.Int("max", maxRetries), slog.Any("error", err))
		time.Sleep(2 * time.Second)
	}

	if pool != nil {
		pool.Close()
	}

	return nil, fmt.Errorf("pgxpool ping failed after %d retries", maxRetries)
}
