package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"banking-api/internal/config"
	"banking-api/internal/database"
	"banking-api/internal/logger"
	"banking-api/internal/repository"
	"banking-api/internal/server"
	"banking-api/internal/service"
	"banking-api/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", "error", err)
		os.Exit(1)
	}
	log := logger.New(cfg.AppEnv)
	log.Info("starting banking-api", "env", cfg.AppEnv, "addr", cfg.Address())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.Connect(ctx, cfg.Database.URL)
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := database.RunMigrations(ctx, pool, "migrations/001_initial.up.sql"); err != nil {
		log.Error("migration failed", "error", err)
		os.Exit(1)
	}

	userRepo := repository.NewUserRepo(pool)
	txRepo := repository.NewTransactionRepo(pool)
	balanceRepo := repository.NewBalanceRepo(pool)
	auditRepo := repository.NewAuditRepo(pool)

	workers := worker.NewPool(cfg.Worker.PoolSize, cfg.Worker.QueueSize)
	workers.Start(ctx)
	defer workers.Stop()

	userSvc := service.NewUserService(userRepo, auditRepo, cfg.JWT)
	balanceSvc := service.NewBalanceService(balanceRepo)
	txSvc := service.NewTransactionService(txRepo, userRepo, balanceSvc, auditRepo, workers)

	router := server.NewRouter(server.Dependencies{
		UserService:        userSvc,
		TransactionService: txSvc,
		BalanceService:     balanceSvc,
	})

	httpServer := &http.Server{
		Addr:         cfg.Address(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Info("http server listening", "address", cfg.Address())
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("graceful shutdown started")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "error", err)
		os.Exit(1)
	}
	log.Info("shutdown complete")
}
