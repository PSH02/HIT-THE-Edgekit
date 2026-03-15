package session_test

import (
	"context"
	"errors"
	"testing"

	"github.com/edgekit/edgekit/internal/core/session"
	"github.com/edgekit/edgekit/pkg/apperror"
)

type mockSessionRepo struct {
	sessions map[string]*session.Session
	updated  map[string]*session.Session
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions: make(map[string]*session.Session),
		updated:  make(map[string]*session.Session),
	}
}

func (m *mockSessionRepo) Create(_ context.Context, s *session.Session) error {
	s.ID = "sess-1"
	m.sessions[s.ID] = s
	return nil
}

func (m *mockSessionRepo) FindByID(_ context.Context, id string) (*session.Session, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return s, nil
}

func (m *mockSessionRepo) Update(_ context.Context, s *session.Session) error {
	m.sessions[s.ID] = s
	m.updated[s.ID] = s
	return nil
}

func (m *mockSessionRepo) ListByStatus(_ context.Context, status session.Status, offset, limit int) ([]*session.Session, error) {
	var result []*session.Session
	for _, s := range m.sessions {
		if s.Status == status {
			result = append(result, s)
		}
	}
	if offset >= len(result) {
		return nil, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (m *mockSessionRepo) Delete(_ context.Context, id string) error {
	delete(m.sessions, id)
	return nil
}

func TestCreate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := newMockSessionRepo()
		svc := session.NewService(repo)
		ctx := context.Background()

		view, err := svc.Create(ctx, "host-1", session.CreateSessionInput{
			Name:       "Test Game",
			MaxPlayers: 4,
			Metadata:   map[string]string{"mode": "ranked"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if view.Name != "Test Game" {
			t.Errorf("name = %q, want %q", view.Name, "Test Game")
		}
		if view.HostUserID != "host-1" {
			t.Errorf("host = %q, want %q", view.HostUserID, "host-1")
		}
		if view.Status != session.StatusWaiting {
			t.Errorf("status = %q, want %q", view.Status, session.StatusWaiting)
		}
		if len(view.Players) != 1 || view.Players[0] != "host-1" {
			t.Errorf("players = %v, want [host-1]", view.Players)
		}
		if view.MaxPlayers != 4 {
			t.Errorf("max_players = %d, want 4", view.MaxPlayers)
		}
		if view.Metadata["mode"] != "ranked" {
			t.Errorf("metadata[mode] = %q, want %q", view.Metadata["mode"], "ranked")
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := newMockSessionRepo()
		svc := session.NewService(repo)
		ctx := context.Background()

		repo.sessions["sess-1"] = &session.Session{
			ID:         "sess-1",
			Name:       "Test Game",
			HostUserID: "host-1",
			MaxPlayers: 4,
			Status:     session.StatusWaiting,
			Players:    []string{"host-1"},
		}

		view, err := svc.Get(ctx, "sess-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if view.ID != "sess-1" {
			t.Errorf("id = %q, want %q", view.ID, "sess-1")
		}
		if view.Name != "Test Game" {
			t.Errorf("name = %q, want %q", view.Name, "Test Game")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		repo := newMockSessionRepo()
		svc := session.NewService(repo)
		ctx := context.Background()

		_, err := svc.Get(ctx, "nonexistent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperror.Is(err, apperror.CodeNotFound) {
			t.Errorf("error code = %v, want %v", err, apperror.CodeNotFound)
		}
	})
}

func TestJoin(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := newMockSessionRepo()
		svc := session.NewService(repo)
		ctx := context.Background()

		repo.sessions["sess-1"] = &session.Session{
			ID:         "sess-1",
			Name:       "Test Game",
			HostUserID: "host-1",
			MaxPlayers: 4,
			Status:     session.StatusWaiting,
			Players:    []string{"host-1"},
		}

		view, err := svc.Join(ctx, "sess-1", "player-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(view.Players) != 2 {
			t.Fatalf("players count = %d, want 2", len(view.Players))
		}
		found := false
		for _, p := range view.Players {
			if p == "player-2" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("player-2 not found in players: %v", view.Players)
		}
	})

	t.Run("SessionNotFound", func(t *testing.T) {
		repo := newMockSessionRepo()
		svc := session.NewService(repo)
		ctx := context.Background()

		_, err := svc.Join(ctx, "nonexistent", "player-2")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperror.Is(err, apperror.CodeNotFound) {
			t.Errorf("error code = %v, want %v", err, apperror.CodeNotFound)
		}
	})

	t.Run("SessionNotWaiting", func(t *testing.T) {
		repo := newMockSessionRepo()
		svc := session.NewService(repo)
		ctx := context.Background()

		repo.sessions["sess-1"] = &session.Session{
			ID:         "sess-1",
			HostUserID: "host-1",
			MaxPlayers: 4,
			Status:     session.StatusActive,
			Players:    []string{"host-1"},
		}

		_, err := svc.Join(ctx, "sess-1", "player-2")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperror.Is(err, apperror.CodeBadRequest) {
			t.Errorf("error code = %v, want %v", err, apperror.CodeBadRequest)
		}
	})

	t.Run("AlreadyInSession", func(t *testing.T) {
		repo := newMockSessionRepo()
		svc := session.NewService(repo)
		ctx := context.Background()

		repo.sessions["sess-1"] = &session.Session{
			ID:         "sess-1",
			HostUserID: "host-1",
			MaxPlayers: 4,
			Status:     session.StatusWaiting,
			Players:    []string{"host-1", "player-2"},
		}

		_, err := svc.Join(ctx, "sess-1", "player-2")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperror.Is(err, apperror.CodeConflict) {
			t.Errorf("error code = %v, want %v", err, apperror.CodeConflict)
		}
	})

	t.Run("SessionFull", func(t *testing.T) {
		repo := newMockSessionRepo()
		svc := session.NewService(repo)
		ctx := context.Background()

		repo.sessions["sess-1"] = &session.Session{
			ID:         "sess-1",
			HostUserID: "host-1",
			MaxPlayers: 2,
			Status:     session.StatusWaiting,
			Players:    []string{"host-1", "player-2"},
		}

		_, err := svc.Join(ctx, "sess-1", "player-3")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperror.Is(err, apperror.CodeBadRequest) {
			t.Errorf("error code = %v, want %v", err, apperror.CodeBadRequest)
		}
	})
}

func TestLeave(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := newMockSessionRepo()
		svc := session.NewService(repo)
		ctx := context.Background()

		repo.sessions["sess-1"] = &session.Session{
			ID:         "sess-1",
			HostUserID: "host-1",
			MaxPlayers: 4,
			Status:     session.StatusWaiting,
			Players:    []string{"host-1", "player-2"},
		}

		view, err := svc.Leave(ctx, "sess-1", "player-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(view.Players) != 1 {
			t.Fatalf("players count = %d, want 1", len(view.Players))
		}
		for _, p := range view.Players {
			if p == "player-2" {
				t.Error("player-2 should have been removed")
			}
		}
		if view.Status != session.StatusWaiting {
			t.Errorf("status = %q, want %q", view.Status, session.StatusWaiting)
		}
	})

	t.Run("NotInSession", func(t *testing.T) {
		repo := newMockSessionRepo()
		svc := session.NewService(repo)
		ctx := context.Background()

		repo.sessions["sess-1"] = &session.Session{
			ID:         "sess-1",
			HostUserID: "host-1",
			MaxPlayers: 4,
			Status:     session.StatusWaiting,
			Players:    []string{"host-1"},
		}

		_, err := svc.Leave(ctx, "sess-1", "player-2")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperror.Is(err, apperror.CodeBadRequest) {
			t.Errorf("error code = %v, want %v", err, apperror.CodeBadRequest)
		}
	})

	t.Run("LastPlayerLeavesStatusFinished", func(t *testing.T) {
		repo := newMockSessionRepo()
		svc := session.NewService(repo)
		ctx := context.Background()

		repo.sessions["sess-1"] = &session.Session{
			ID:         "sess-1",
			HostUserID: "host-1",
			MaxPlayers: 4,
			Status:     session.StatusWaiting,
			Players:    []string{"host-1"},
		}

		view, err := svc.Leave(ctx, "sess-1", "host-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if view.Status != session.StatusFinished {
			t.Errorf("status = %q, want %q", view.Status, session.StatusFinished)
		}
		if len(view.Players) != 0 {
			t.Errorf("players = %v, want empty", view.Players)
		}
	})
}

func TestListWaiting(t *testing.T) {
	t.Run("ReturnsSessionViews", func(t *testing.T) {
		repo := newMockSessionRepo()
		svc := session.NewService(repo)
		ctx := context.Background()

		repo.sessions["sess-1"] = &session.Session{
			ID:         "sess-1",
			Name:       "Game 1",
			HostUserID: "host-1",
			MaxPlayers: 4,
			Status:     session.StatusWaiting,
			Players:    []string{"host-1"},
		}
		repo.sessions["sess-2"] = &session.Session{
			ID:         "sess-2",
			Name:       "Game 2",
			HostUserID: "host-2",
			MaxPlayers: 2,
			Status:     session.StatusWaiting,
			Players:    []string{"host-2"},
		}
		repo.sessions["sess-3"] = &session.Session{
			ID:         "sess-3",
			Name:       "Active Game",
			HostUserID: "host-3",
			MaxPlayers: 4,
			Status:     session.StatusActive,
			Players:    []string{"host-3"},
		}

		views, err := svc.ListWaiting(ctx, 0, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(views) != 2 {
			t.Fatalf("count = %d, want 2", len(views))
		}
		for _, v := range views {
			if v.Status != session.StatusWaiting {
				t.Errorf("session %s status = %q, want %q", v.ID, v.Status, session.StatusWaiting)
			}
		}
	})

	t.Run("EmptyList", func(t *testing.T) {
		repo := newMockSessionRepo()
		svc := session.NewService(repo)
		ctx := context.Background()

		views, err := svc.ListWaiting(ctx, 0, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if views != nil && len(views) != 0 {
			t.Errorf("count = %d, want 0", len(views))
		}
	})
}
