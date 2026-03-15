package user

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*User, int64, error)
}

type UserCache interface {
	Get(ctx context.Context, id string) (*User, error)
	Set(ctx context.Context, user *User) error
	Invalidate(ctx context.Context, id string) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}
