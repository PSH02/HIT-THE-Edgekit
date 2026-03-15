package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edgekit/edgekit/internal/core/user"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	u.ID = uuid.New().String()
	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = u.CreatedAt

	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, username, email, password_hash, role, scopes, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		u.ID, u.Username, u.Email, u.PasswordHash, string(u.Role), u.Scopes, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	u := &user.User{}
	var role string
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password_hash, role, scopes, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &role, &u.Scopes, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	u.Role = user.Role(role)
	return u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	u := &user.User{}
	var role string
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password_hash, role, scopes, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &role, &u.Scopes, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	u.Role = user.Role(role)
	return u, nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	u.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET username = $1, email = $2, role = $3, updated_at = $4 WHERE id = $5`,
		u.Username, u.Email, string(u.Role), u.UpdatedAt, u.ID,
	)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*user.User, int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, username, email, password_hash, role, scopes, created_at, updated_at
		 FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		u := &user.User{}
		var role string
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &role, &u.Scopes, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		u.Role = user.Role(role)
		users = append(users, u)
	}

	return users, total, rows.Err()
}
