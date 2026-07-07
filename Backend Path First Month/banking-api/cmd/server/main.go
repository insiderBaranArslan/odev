package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"banking-api/internal/cache"
	"banking-api/internal/config"
	"banking-api/internal/database"
	"banking-api/internal/repository"
	"banking-api/internal/server"
	"banking-api/internal/service"
	"banking-api/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.Database.URL)
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	migrationSQL, err := os.ReadFile("migrations/001_initial.up.sql")
	if err != nil {
		slog.Error("failed to read migration file", "error", err)
		os.Exit(1)
	}
	if err := database.RunMigrations(ctx, pool, string(migrationSQL)); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	userRepo := repository.NewPostgresUserRepository(pool)
	txRepo := repository.NewPostgresTransactionRepository(pool)
	balanceRepo := repository.NewPostgresBalanceRepository(pool)
	auditRepo := repository.NewPostgresAuditRepository(pool)
	eventRepo := repository.NewPostgresEventRepository(pool)
	cacheStore := cache.NewMemoryCache()

	userService := service.NewUserService(userRepo, auditRepo, cfg.JWT)
	balanceService := service.NewBalanceService(balanceRepo, cacheStore)
	workerPool := worker.NewPool(cfg.Worker.PoolSize, cfg.Worker.QueueSize)
	workerPool.Start(ctx)
	defer workerPool.Stop()

	txService := service.NewTransactionService(txRepo, balanceService, userRepo, auditRepo, eventRepo, workerPool)

	router := server.NewRouter(server.Dependencies{
		UserService:        userService,
		TransactionService: txService,
		BalanceService:     balanceService,
	})

	httpServer := &http.Server{
		Addr:         cfg.ServerAddress(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		slog.Info("server starting", "address", cfg.ServerAddress())
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	slog.Info("server shutting down")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
