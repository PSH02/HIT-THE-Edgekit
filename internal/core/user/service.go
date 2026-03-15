package user

import (
	"context"
	"fmt"

	"github.com/edgekit/edgekit/internal/core/auth"
	"github.com/edgekit/edgekit/pkg/apperror"
)

type Service struct {
	repo     UserRepository
	cache    UserCache
	hasher   PasswordHasher
	tokenSvc *auth.TokenService
}

func NewService(
	repo UserRepository,
	cache UserCache,
	hasher PasswordHasher,
	tokenSvc *auth.TokenService,
) *Service {
	return &Service{
		repo:     repo,
		cache:    cache,
		hasher:   hasher,
		tokenSvc: tokenSvc,
	}
}

func (s *Service) Register(ctx context.Context, input CreateUserInput) (*User, error) {
	existing, _ := s.repo.FindByEmail(ctx, input.Email)
	if existing != nil {
		return nil, apperror.New(apperror.CodeConflict, "email already registered")
	}

	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: hash,
		Role:         RoleUser,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	user, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, apperror.New(apperror.CodeUnauthorized, "invalid credentials")
	}

	if err := s.hasher.Verify(input.Password, user.PasswordHash); err != nil {
		return nil, apperror.New(apperror.CodeUnauthorized, "invalid credentials")
	}

	pair, err := s.tokenSvc.GenerateTokenPair(ctx, auth.Subject{
		UserID: user.ID,
		Role:   string(user.Role),
		Scopes: user.Scopes,
	})
	if err != nil {
		return nil, fmt.Errorf("generate token pair: %w", err)
	}

	return &AuthResult{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
	}, nil
}

func (s *Service) GetProfile(ctx context.Context, id string) (*UserProfile, error) {
	if cached, err := s.cache.Get(ctx, id); err == nil && cached != nil {
		return cached.ToProfile(), nil
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.New(apperror.CodeNotFound, "user not found")
	}

	_ = s.cache.Set(ctx, user)
	return user.ToProfile(), nil
}

func (s *Service) UpdateProfile(ctx context.Context, id string, input UpdateProfileInput) (*UserProfile, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.New(apperror.CodeNotFound, "user not found")
	}

	if input.Username != "" {
		user.Username = input.Username
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	_ = s.cache.Invalidate(ctx, id)
	return user.ToProfile(), nil
}
