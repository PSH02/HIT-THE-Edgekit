package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/edgekit/edgekit/pkg/ratelimit"
)

type RateLimiter struct {
	client *goredis.Client
	cfg    ratelimit.Config
}

func NewRateLimiter(client *goredis.Client, cfg ratelimit.Config) *RateLimiter {
	return &RateLimiter{client: client, cfg: cfg}
}

func (rl *RateLimiter) Allow(ctx context.Context, key string) (ratelimit.Result, error) {
	window := time.Duration(rl.cfg.Window) * time.Second
	rkey := fmt.Sprintf("rl:%s", key)

	count, err := rl.client.Incr(ctx, rkey).Result()
	if err != nil {
		return ratelimit.Result{}, err
	}

	if count == 1 {
		rl.client.Expire(ctx, rkey, window)
	}

	limit := int64(rl.cfg.Rate)
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	if count > limit {
		ttl, _ := rl.client.TTL(ctx, rkey).Result()
		return ratelimit.Result{
			Allowed:    false,
			Limit:      limit,
			Remaining:  0,
			RetryAfter: int64(ttl / time.Second),
		}, nil
	}

	return ratelimit.Result{
		Allowed:   true,
		Limit:     limit,
		Remaining: remaining,
	}, nil
}
