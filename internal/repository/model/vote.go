package model

import (
	"pollapp/internal/service/domain"
	"time"
)

// Vote represents a user vote on a poll in the repository layer
type Vote struct {
	ID          int
	PollID      int
	UserID      int
	OptionIndex int
	CreatedAt   time.Time
}

func (v Vote) ToDomain() domain.Vote {
	return domain.Vote{
		ID:          v.ID,
		PollID:      v.PollID,
		UserID:      v.UserID,
		OptionIndex: v.OptionIndex,
		CreatedAt:   v.CreatedAt,
	}
}
