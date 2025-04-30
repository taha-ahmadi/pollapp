package model

import (
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
