package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/edgekit/edgekit/internal/core/user"
)

const userCacheTTL = 15 * time.Minute

func Connect(redisURL string) *goredis.Client {
	opt, err := goredis.ParseURL(redisURL)
	if err != nil {
		opt = &goredis.Options{Addr: redisURL}
	}
	return goredis.NewClient(opt)
}

type UserCache struct {
	client *goredis.Client
}

func NewUserCache(client *goredis.Client) *UserCache {
	return &UserCache{client: client}
}

func userKey(id string) string {
	return fmt.Sprintf("user:%s", id)
}

func (c *UserCache) Get(ctx context.Context, id string) (*user.User, error) {
	data, err := c.client.Get(ctx, userKey(id)).Bytes()
	if err != nil {
		return nil, err
	}

	var u user.User
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (c *UserCache) Set(ctx context.Context, u *user.User) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, userKey(u.ID), data, userCacheTTL).Err()
}

func (c *UserCache) Invalidate(ctx context.Context, id string) error {
	return c.client.Del(ctx, userKey(id)).Err()
}
