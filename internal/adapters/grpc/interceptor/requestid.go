package interceptor

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type requestIDKey struct{}

type RequestIDInterceptor struct{}

func NewRequestIDInterceptor() *RequestIDInterceptor {
	return &RequestIDInterceptor{}
}

func (i *RequestIDInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		requestID := uuid.New().String()
		ctx = context.WithValue(ctx, requestIDKey{}, requestID)

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		md = md.Copy()
		md.Set("x-request-id", requestID)
		ctx = metadata.NewIncomingContext(ctx, md)

		_ = grpc.SetHeader(ctx, metadata.Pairs("x-request-id", requestID))

		return handler(ctx, req)
	}
}

func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}
