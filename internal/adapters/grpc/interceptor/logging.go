package interceptor

import (
	"context"
	"time"

	"github.com/edgekit/edgekit/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type LoggingInterceptor struct {
	log logger.Logger
}

func NewLoggingInterceptor(log logger.Logger) *LoggingInterceptor {
	return &LoggingInterceptor{log: log}
}

func (i *LoggingInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		code := status.Code(err)
		i.log.Info("grpc request",
			"method", info.FullMethod,
			"duration", duration.String(),
			"code", code.String(),
		)

		return resp, err
	}
}

func (i *LoggingInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		err := handler(srv, ss)
		duration := time.Since(start)

		code := status.Code(err)
		i.log.Info("grpc stream",
			"method", info.FullMethod,
			"duration", duration.String(),
			"code", code.String(),
		)

		return err
	}
}
