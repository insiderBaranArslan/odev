package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"banking-core/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users domain.UserRepository
	audit domain.AuditRepository
}

func NewUserService(users domain.UserRepository, audit domain.AuditRepository) *UserService {
	return &UserService{users: users, audit: audit}
}

func (s *UserService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.User, error) {
	if len(req.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Role:     domain.RoleUser,
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = string(hash)

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	_ = s.audit.Create(ctx, &domain.AuditLog{
		EntityType: "user",
		EntityID:   user.ID,
		Action:     "register",
		Details:    mustJSON(map[string]any{"email": user.Email, "username": user.Username}),
	})

	return user, nil
}

func (s *UserService) Authenticate(ctx context.Context, req domain.LoginRequest) (*domain.User, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	_ = s.audit.Create(ctx, &domain.AuditLog{
		EntityType: "user",
		EntityID:   user.ID,
		Action:     "login",
		Details:    mustJSON(map[string]any{"email": user.Email}),
	})
	return user, nil
}

func (s *UserService) Authorize(user *domain.User, roles ...domain.Role) error {
	if user == nil {
		return fmt.Errorf("unauthorized")
	}
	for _, role := range roles {
		if user.Role == role {
			return nil
		}
	}
	return fmt.Errorf("forbidden: requires one of %v", roles)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.users.GetByID(ctx, id)
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
