package interceptor

import (
	"context"

	"github.com/edgekit/edgekit/internal/core/auth"
	"github.com/edgekit/edgekit/pkg/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type RateLimitInterceptor struct {
	limiter ratelimit.Limiter
}

func NewRateLimitInterceptor(limiter ratelimit.Limiter) *RateLimitInterceptor {
	return &RateLimitInterceptor{limiter: limiter}
}

func (i *RateLimitInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if i.limiter == nil {
			return handler(ctx, req)
		}

		key := extractRateLimitKey(ctx)
		result, err := i.limiter.Allow(ctx, key)
		if err != nil {
			return handler(ctx, req)
		}
		if !result.Allowed {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded, retry after %ds", result.RetryAfter)
		}

		return handler(ctx, req)
	}
}

func extractRateLimitKey(ctx context.Context) string {
	if ac, ok := auth.FromContext(ctx); ok && ac.UserID != "" {
		return "user:" + ac.UserID
	}

	if p, ok := peer.FromContext(ctx); ok {
		return "ip:" + p.Addr.String()
	}

	return "unknown"
}
