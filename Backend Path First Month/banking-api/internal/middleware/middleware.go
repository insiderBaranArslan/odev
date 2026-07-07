package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"banking-api/internal/domain"
	"banking-api/internal/service"
	"banking-api/pkg/response"

	"github.com/go-chi/chi/v5/middleware"
)

type contextKey string

const (
	ContextUserID contextKey = "userID"
	ContextRole   contextKey = "role"
)

func RequestLogger(next http.Handler) http.Handler {
	return middleware.Logger(next)
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	rl := &rateLimiter{requests: make(map[string][]time.Time), limit: limit, window: window}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			now := time.Now()
			rl.mu.Lock()
			times := rl.requests[ip]
			filtered := times[:0]
			for _, t := range times {
				if now.Sub(t) <= rl.window {
					filtered = append(filtered, t)
				}
			}
			if len(filtered) >= rl.limit {
				rl.mu.Unlock()
				response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			rl.requests[ip] = append(filtered, now)
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}

func Auth(userService *service.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")
			userID, role, err := userService.ParseAccessToken(token)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), ContextUserID, userID)
			ctx = context.WithValue(ctx, ContextRole, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	allowed := make(map[domain.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(ContextRole).(domain.Role)
			if !ok {
				response.Error(w, http.StatusForbidden, "forbidden")
				return
			}
			if _, ok := allowed[role]; !ok {
				response.Error(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Recoverer(next http.Handler) http.Handler {
	return middleware.Recoverer(next)
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(ContextUserID).(int64)
	return userID, ok
}
