package model

import (
	"pollapp/internal/service/domain"
	"time"
)

// Skip represents a user skip on a poll in the repository layer
type Skip struct {
	ID        int
	PollID    int
	UserID    int
	CreatedAt time.Time
}

func (s Skip) ToDomain() domain.Skip {
	return domain.Skip{
		ID:        s.ID,
		PollID:    s.PollID,
		UserID:    s.UserID,
		CreatedAt: s.CreatedAt,
	}
}
