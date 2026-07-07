package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"banking-api/internal/config"
	"banking-api/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users domain.UserRepository
	audit domain.AuditRepository
	jwt   config.JWTConfig
}

func NewUserService(users domain.UserRepository, audit domain.AuditRepository, jwtCfg config.JWTConfig) *UserService {
	return &UserService{users: users, audit: audit, jwt: jwtCfg}
}

func (s *UserService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.User, error) {
	user := &domain.User{
		Username: req.Username,
		Email:    strings.ToLower(strings.TrimSpace(req.Email)),
		Role:     domain.RoleUser,
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	if len(req.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = string(hash)
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	_ = s.audit.Create(ctx, &domain.AuditLog{
		EntityType: "user",
		EntityID:   user.ID,
		Action:     "register",
		Details:    mustJSON(map[string]any{"email": user.Email}),
	})
	return user, nil
}

func (s *UserService) Authenticate(ctx context.Context, req domain.LoginRequest) (*domain.User, error) {
	user, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	return s.users.GetByID(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.users.List(ctx)
}

func (s *UserService) UpdateUser(ctx context.Context, id int64, req domain.UpdateUserRequest) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = strings.ToLower(req.Email)
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	return s.users.Delete(ctx, id)
}

func (s *UserService) IssueTokens(_ context.Context, user *domain.User) (*domain.AuthResponse, error) {
	access, err := s.signToken(user, "access", s.jwt.AccessTokenTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := s.signToken(user, "refresh", s.jwt.RefreshTokenTTL)
	if err != nil {
		return nil, err
	}
	return &domain.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(s.jwt.AccessTokenTTL.Seconds()),
	}, nil
}

func (s *UserService) RefreshTokens(_ context.Context, refreshToken string) (*domain.AuthResponse, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil || claims["type"] != "refresh" {
		return nil, fmt.Errorf("invalid refresh token")
	}
	userID, _ := strconv.ParseInt(fmt.Sprint(claims["sub"]), 10, 64)
	user := &domain.User{ID: userID, Role: domain.Role(claims["role"].(string))}
	return s.IssueTokens(context.Background(), user)
}

func (s *UserService) signToken(user *domain.User, tokenType string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": string(user.Role),
		"type": tokenType,
		"exp":  time.Now().Add(ttl).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwt.Secret))
}

func (s *UserService) ParseAccessToken(token string) (int64, domain.Role, error) {
	claims, err := s.parseToken(token)
	if err != nil || claims["type"] != "access" {
		return 0, "", fmt.Errorf("invalid access token")
	}
	userID, _ := strconv.ParseInt(fmt.Sprint(claims["sub"]), 10, 64)
	return userID, domain.Role(claims["role"].(string)), nil
}

func (s *UserService) parseToken(token string) (jwt.MapClaims, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.jwt.Secret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
