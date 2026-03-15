package interceptor

import (
	"context"
	"strings"

	"github.com/edgekit/edgekit/internal/core/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var skipAuthMethods = map[string]bool{
	"/grpc.health.v1.Health/Check":       true,
	"/grpc.health.v1.Health/Watch":        true,
	"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo":   true,
	"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo": true,
	"/user.v1.UserService/Register": true,
	"/user.v1.UserService/Login":    true,
}

type AuthInterceptor struct {
	tokenSvc *auth.TokenService
}

func NewAuthInterceptor(tokenSvc *auth.TokenService) *AuthInterceptor {
	return &AuthInterceptor{tokenSvc: tokenSvc}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if skipAuthMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		newCtx, err := i.authenticate(ctx)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

func (i *AuthInterceptor) authenticate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}

	token := values[0]
	if !strings.HasPrefix(token, "Bearer ") {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
	}
	token = strings.TrimPrefix(token, "Bearer ")

	ac, err := i.tokenSvc.ValidateToken(ctx, token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return auth.WithAuth(ctx, ac), nil
}
