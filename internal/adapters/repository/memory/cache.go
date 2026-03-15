package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/edgekit/edgekit/internal/core/user"
)

var ErrCacheMiss = errors.New("cache miss")

type UserCache struct {
	store sync.Map
}

func NewUserCache() *UserCache {
	return &UserCache{}
}

func (c *UserCache) Get(_ context.Context, id string) (*user.User, error) {
	val, ok := c.store.Load(id)
	if !ok {
		return nil, ErrCacheMiss
	}
	u := val.(user.User)
	return &u, nil
}

func (c *UserCache) Set(_ context.Context, u *user.User) error {
	c.store.Store(u.ID, *u)
	return nil
}

func (c *UserCache) Invalidate(_ context.Context, id string) error {
	c.store.Delete(id)
	return nil
}
