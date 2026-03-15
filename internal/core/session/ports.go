package session

import "context"

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	FindByID(ctx context.Context, id string) (*Session, error)
	Update(ctx context.Context, session *Session) error
	ListByStatus(ctx context.Context, status Status, offset, limit int) ([]*Session, error)
	Delete(ctx context.Context, id string) error
}
