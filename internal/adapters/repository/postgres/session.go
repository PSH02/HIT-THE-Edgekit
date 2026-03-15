package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edgekit/edgekit/internal/core/session"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, s *session.Session) error {
	s.ID = uuid.New().String()
	s.CreatedAt = time.Now().UTC()
	s.UpdatedAt = s.CreatedAt

	metadata, err := json.Marshal(s.Metadata)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO sessions (id, name, host_user_id, max_players, status, players, metadata, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		s.ID, s.Name, s.HostUserID, s.MaxPlayers, string(s.Status), s.Players, metadata, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*session.Session, error) {
	s := &session.Session{}
	var status string
	var metadata []byte

	err := r.pool.QueryRow(ctx,
		`SELECT id, name, host_user_id, max_players, status, players, metadata, created_at, updated_at
		 FROM sessions WHERE id = $1`, id,
	).Scan(&s.ID, &s.Name, &s.HostUserID, &s.MaxPlayers, &status, &s.Players, &metadata, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}

	s.Status = session.Status(status)
	if err := json.Unmarshal(metadata, &s.Metadata); err != nil {
		return nil, err
	}

	return s, nil
}

func (r *SessionRepository) Update(ctx context.Context, s *session.Session) error {
	s.UpdatedAt = time.Now().UTC()

	metadata, err := json.Marshal(s.Metadata)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx,
		`UPDATE sessions SET name = $1, max_players = $2, status = $3, players = $4, metadata = $5, updated_at = $6
		 WHERE id = $7`,
		s.Name, s.MaxPlayers, string(s.Status), s.Players, metadata, s.UpdatedAt, s.ID,
	)
	return err
}

func (r *SessionRepository) ListByStatus(ctx context.Context, status session.Status, offset, limit int) ([]*session.Session, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, host_user_id, max_players, status, players, metadata, created_at, updated_at
		 FROM sessions WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		string(status), limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*session.Session
	for rows.Next() {
		s := &session.Session{}
		var st string
		var metadata []byte

		if err := rows.Scan(&s.ID, &s.Name, &s.HostUserID, &s.MaxPlayers, &st, &s.Players, &metadata, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}

		s.Status = session.Status(st)
		if err := json.Unmarshal(metadata, &s.Metadata); err != nil {
			return nil, err
		}

		sessions = append(sessions, s)
	}

	return sessions, rows.Err()
}

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}
