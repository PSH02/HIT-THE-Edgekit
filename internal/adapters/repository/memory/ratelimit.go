package memory

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/edgekit/edgekit/pkg/ratelimit"
)

type RateLimiter struct {
	limiters sync.Map
	cfg      ratelimit.Config
}

func NewRateLimiter(cfg ratelimit.Config) *RateLimiter {
	return &RateLimiter{cfg: cfg}
}

func (rl *RateLimiter) Allow(_ context.Context, key string) (ratelimit.Result, error) {
	r := rate.Every(time.Duration(rl.cfg.Window) * time.Second / time.Duration(rl.cfg.Rate))
	val, _ := rl.limiters.LoadOrStore(key, rate.NewLimiter(r, rl.cfg.Rate))
	limiter := val.(*rate.Limiter)

	limit := int64(rl.cfg.Rate)

	if limiter.Allow() {
		return ratelimit.Result{
			Allowed:   true,
			Limit:     limit,
			Remaining: int64(limiter.Burst()) - int64(limiter.Tokens()),
		}, nil
	}

	reservation := limiter.Reserve()
	delay := reservation.Delay()
	reservation.Cancel()

	return ratelimit.Result{
		Allowed:    false,
		Limit:      limit,
		Remaining:  0,
		RetryAfter: int64(delay / time.Second),
	}, nil
}
