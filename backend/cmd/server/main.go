// 口述历史采集工具后端入口。
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oralhistory/oralhistory/internal/config"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/database"
	"github.com/oralhistory/oralhistory/internal/router"
	"github.com/oralhistory/oralhistory/internal/util"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := slog.Default()
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config failed", "error", err)
		os.Exit(1)
	}
	logger = util.NewLogger(cfg.RunMode)

	db, err := database.New(cfg, logger)
	if err != nil {
		logger.Error("connect database failed", "error", err)
		os.Exit(1)
	}
	if err := database.SeedAdmin(db, logger); err != nil {
		logger.Error("seed admin failed", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Warn("redis ping failed, rate limit disabled", "error", err)
		rdb = nil
	} else {
		logger.Info(fmt.Sprintf(constants.LogRedisConnected, cfg.RedisAddr))
	}

	engine, err := router.Setup(cfg, db, rdb, logger)
	if err != nil {
		logger.Error("setup router failed", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: engine,
	}

	go func() {
		logger.Info(fmt.Sprintf(constants.LogStartServer, cfg.ServerPort))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server listen failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("server shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", "error", err)
	}
}
