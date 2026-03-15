package session

import (
	"context"
	"fmt"

	"github.com/edgekit/edgekit/pkg/apperror"
)

type Service struct {
	repo SessionRepository
}

func NewService(repo SessionRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, hostUserID string, input CreateSessionInput) (*SessionView, error) {
	sess := &Session{
		Name:       input.Name,
		HostUserID: hostUserID,
		MaxPlayers: input.MaxPlayers,
		Status:     StatusWaiting,
		Players:    []string{hostUserID},
		Metadata:   input.Metadata,
	}

	if err := s.repo.Create(ctx, sess); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return sess.ToView(), nil
}

func (s *Service) Get(ctx context.Context, id string) (*SessionView, error) {
	sess, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.New(apperror.CodeNotFound, "session not found")
	}
	return sess.ToView(), nil
}

func (s *Service) Join(ctx context.Context, sessionID, userID string) (*SessionView, error) {
	sess, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, apperror.New(apperror.CodeNotFound, "session not found")
	}

	if sess.Status != StatusWaiting {
		return nil, apperror.New(apperror.CodeBadRequest, "session is not accepting players")
	}

	if sess.HasPlayer(userID) {
		return nil, apperror.New(apperror.CodeConflict, "already in session")
	}

	if sess.IsFull() {
		return nil, apperror.New(apperror.CodeBadRequest, "session is full")
	}

	sess.Players = append(sess.Players, userID)

	if err := s.repo.Update(ctx, sess); err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	return sess.ToView(), nil
}

func (s *Service) Leave(ctx context.Context, sessionID, userID string) (*SessionView, error) {
	sess, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, apperror.New(apperror.CodeNotFound, "session not found")
	}

	if !sess.HasPlayer(userID) {
		return nil, apperror.New(apperror.CodeBadRequest, "not in session")
	}

	filtered := make([]string, 0, len(sess.Players)-1)
	for _, p := range sess.Players {
		if p != userID {
			filtered = append(filtered, p)
		}
	}
	sess.Players = filtered

	if len(sess.Players) == 0 {
		sess.Status = StatusFinished
	}

	if err := s.repo.Update(ctx, sess); err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	return sess.ToView(), nil
}

func (s *Service) ListWaiting(ctx context.Context, offset, limit int) ([]*SessionView, error) {
	sessions, err := s.repo.ListByStatus(ctx, StatusWaiting, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	views := make([]*SessionView, len(sessions))
	for i, sess := range sessions {
		views[i] = sess.ToView()
	}
	return views, nil
}
