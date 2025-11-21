package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/andro-kes/avito_test/internal/config"
	"github.com/andro-kes/avito_test/internal/http/handlers"
	"github.com/andro-kes/avito_test/internal/http/middleware"
	logger "github.com/andro-kes/avito_test/internal/log"
	"github.com/andro-kes/avito_test/internal/migrations"
)

func main() {
	logger.Init()
	cfg := config.Init()

	router := gin.Default()

	ctx := context.Background()
	pool, err := NewPool(ctx)
	if err != nil {
		logger.Log.Fatal("failed to create pool")
	}

	if err := applyMigrations(ctx, pool); err != nil {
		logger.Log.Fatal("failed to apply migrations", zap.Error(err))
	}

	handlerManager := handlers.NewHandlerManager(pool)

	team := router.Group("/team/")
	team.POST("add/", handlerManager.AddTeam)
	team.GET("get/", handlerManager.GetTeam)

	user := router.Group("/users/")
	user.POST("setIsActive/", middleware.Admin(), handlerManager.SetIsActive)
	user.GET("getReview/", handlerManager.GetUserReview)
	user.GET("countReview/", handlerManager.CountReview)
	user.POST("deactivate/", handlerManager.DeactivateUsers)

	pr := router.Group("/pullRequest/")
	pr.POST("create/", handlerManager.CreatePR)
	pr.POST("merge/", handlerManager.MergePR)
	pr.POST("reassign/", handlerManager.ReassignReviewer)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("server listen error", zap.String("error", err.Error()))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer func() {
		cancel()
		logger.Log.Info("closing database connections")
		pool.Close()
		logger.Close()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("server forced to shutdown", zap.String("error", err.Error()))
	}
}

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dbURL := os.Getenv("DB_URL")
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	attempts := 3
	delay := time.Second
	for i := 0; i < attempts; i++ {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := pool.Ping(pingCtx)
		cancel()
		if err == nil {
			return pool, nil
		}
		time.Sleep(delay)
		delay *= 2
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	return migrations.ApplyMigrations(ctx, pool)
}
