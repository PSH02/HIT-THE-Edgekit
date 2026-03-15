package session

import "time"

type Status string

const (
	StatusWaiting  Status = "waiting"
	StatusActive   Status = "active"
	StatusFinished Status = "finished"
)

type Session struct {
	ID         string
	Name       string
	HostUserID string
	MaxPlayers int
	Status     Status
	Players    []string
	Metadata   map[string]string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CreateSessionInput struct {
	Name       string            `json:"name" binding:"required,min=1,max=64"`
	MaxPlayers int               `json:"max_players" binding:"required,min=2,max=100"`
	Metadata   map[string]string `json:"metadata"`
}

type SessionView struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	HostUserID string            `json:"host_user_id"`
	MaxPlayers int               `json:"max_players"`
	Status     Status            `json:"status"`
	Players    []string          `json:"players"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

func (s *Session) ToView() *SessionView {
	return &SessionView{
		ID:         s.ID,
		Name:       s.Name,
		HostUserID: s.HostUserID,
		MaxPlayers: s.MaxPlayers,
		Status:     s.Status,
		Players:    s.Players,
		Metadata:   s.Metadata,
		CreatedAt:  s.CreatedAt,
	}
}

func (s *Session) IsFull() bool {
	return len(s.Players) >= s.MaxPlayers
}

func (s *Session) HasPlayer(userID string) bool {
	for _, p := range s.Players {
		if p == userID {
			return true
		}
	}
	return false
}
