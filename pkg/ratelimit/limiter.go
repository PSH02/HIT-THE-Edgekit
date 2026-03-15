package ratelimit

import "context"

type Limiter interface {
	Allow(ctx context.Context, key string) (Result, error)
}

type Result struct {
	Allowed    bool
	Limit      int64
	Remaining  int64
	RetryAfter int64
}

type Config struct {
	Rate   int
	Window int
}

type KeyFunc func(ctx context.Context) string
