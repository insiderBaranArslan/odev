package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"banking-core/internal/config"
	"banking-core/internal/database"
	"banking-core/internal/domain"
	"banking-core/internal/logger"
	"banking-core/internal/repository"
	"banking-core/internal/service"
	"banking-core/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.AppEnv)
	log.Info("starting banking-core", "env", cfg.AppEnv)

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
	log.Info("migrations applied")

	userRepo := repository.NewUserRepo(pool)
	txRepo := repository.NewTransactionRepo(pool)
	balanceRepo := repository.NewBalanceRepo(pool)
	auditRepo := repository.NewAuditRepo(pool)

	workers := worker.NewPool(cfg.Worker.PoolSize, cfg.Worker.QueueSize)
	workers.Start(ctx)
	defer workers.Stop()

	userSvc := service.NewUserService(userRepo, auditRepo)
	balanceSvc := service.NewBalanceService(balanceRepo)
	txSvc := service.NewTransactionService(txRepo, userRepo, balanceSvc, auditRepo, workers)

	if err := demoFlow(ctx, log, userSvc, balanceSvc, txSvc); err != nil {
		log.Error("demo flow failed", "error", err)
		os.Exit(1)
	}

	log.Info("core services ready — waiting for shutdown signal (Ctrl+C)")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("graceful shutdown started")
	cancel()
	log.Info("shutdown complete")
}

func demoFlow(
	ctx context.Context,
	log *slog.Logger,
	users *service.UserService,
	balances *service.BalanceService,
	txs *service.TransactionService,
) error {
	alice, err := users.Register(ctx, domain.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err != nil {
		existing, getErr := users.Authenticate(ctx, domain.LoginRequest{
			Email:    "alice@example.com",
			Password: "password123",
		})
		if getErr != nil {
			return fmt.Errorf("register/login alice: %w", err)
		}
		alice = existing
	}

	bob, err := users.Register(ctx, domain.RegisterRequest{
		Username: "bob",
		Email:    "bob@example.com",
		Password: "password123",
	})
	if err != nil {
		existing, getErr := users.Authenticate(ctx, domain.LoginRequest{
			Email:    "bob@example.com",
			Password: "password123",
		})
		if getErr != nil {
			return fmt.Errorf("register/login bob: %w", err)
		}
		bob = existing
	}

	if err := users.Authorize(alice, domain.RoleUser, domain.RoleAdmin); err != nil {
		return err
	}

	credit, err := txs.Credit(ctx, domain.CreditRequest{UserID: bob.ID, Amount: 100})
	if err != nil {
		return fmt.Errorf("credit: %w", err)
	}

	transfer, err := txs.Transfer(ctx, domain.TransferRequest{
		FromUserID: bob.ID,
		ToUserID:   alice.ID,
		Amount:     25,
	})
	if err != nil {
		return fmt.Errorf("transfer: %w", err)
	}

	aliceBal, _ := balances.GetCurrent(ctx, alice.ID)
	bobBal, _ := balances.GetCurrent(ctx, bob.ID)
	stats := txs.Stats()

	payload, _ := json.MarshalIndent(map[string]any{
		"alice":       alice,
		"bob":         bob,
		"credit":      credit,
		"transfer":    transfer,
		"alice_balance": aliceBal.Snapshot(),
		"bob_balance":   bobBal.Snapshot(),
		"stats":         stats,
	}, "", "  ")

	log.Info("demo completed", "result", string(payload))
	return nil
}
