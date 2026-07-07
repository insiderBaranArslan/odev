package server

import (
	"net/http"
	"time"

	"banking-api/internal/domain"
	"banking-api/internal/handler"
	"banking-api/internal/middleware"
	"banking-api/internal/service"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Dependencies struct {
	UserService        *service.UserService
	TransactionService *service.TransactionService
	BalanceService     *service.BalanceService
}

func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestLogger)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CORS)
	r.Use(middleware.RateLimit(100, time.Minute))
	r.Use(chimiddleware.Timeout(30 * time.Second))

	authHandler := handler.NewAuthHandler(deps.UserService)
	userHandler := handler.NewUserHandler(deps.UserService)
	txHandler := handler.NewTransactionHandler(deps.TransactionService)
	balanceHandler := handler.NewBalanceHandler(deps.BalanceService)

	r.Get("/health", handler.Health)
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(deps.UserService))

			r.Route("/users", func(r chi.Router) {
				r.With(middleware.RequireRole(domain.RoleAdmin)).Get("/", userHandler.List)
				r.Get("/{id}", userHandler.Get)
				r.With(middleware.RequireRole(domain.RoleAdmin)).Put("/{id}", userHandler.Update)
				r.With(middleware.RequireRole(domain.RoleAdmin)).Delete("/{id}", userHandler.Delete)
			})

			r.Route("/transactions", func(r chi.Router) {
				r.With(middleware.RequireRole(domain.RoleAdmin)).Post("/credit", txHandler.Credit)
				r.Post("/debit", txHandler.Debit)
				r.Post("/transfer", txHandler.Transfer)
				r.Get("/history", txHandler.History)
				r.Get("/stats", txHandler.Stats)
				r.Get("/{id}", txHandler.Get)
			})

			r.Route("/balances", func(r chi.Router) {
				r.Get("/current", balanceHandler.Current)
				r.Get("/historical", balanceHandler.Historical)
				r.Get("/at-time", balanceHandler.AtTime)
			})
		})
	})

	return r
}
