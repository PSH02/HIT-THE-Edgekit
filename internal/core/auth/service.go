package auth

import (
	"context"

	"github.com/edgekit/edgekit/pkg/jwt"
)

type TokenService struct {
	jwtMgr *jwt.Manager
}

func NewTokenService(jwtMgr *jwt.Manager) *TokenService {
	return &TokenService{jwtMgr: jwtMgr}
}

func (s *TokenService) GenerateTokenPair(_ context.Context, sub Subject) (*jwt.TokenPair, error) {
	return s.jwtMgr.GenerateTokenPair(sub.UserID, sub.Role, sub.Scopes)
}

func (s *TokenService) ValidateToken(_ context.Context, tokenStr string) (*AuthContext, error) {
	claims, err := s.jwtMgr.ValidateToken(tokenStr)
	if err != nil {
		return nil, err
	}
	return &AuthContext{
		UserID: claims.UserID,
		Role:   claims.Role,
		Scopes: claims.Scopes,
	}, nil
}
