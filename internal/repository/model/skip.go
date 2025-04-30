package model

import (
	"time"
)

// Skip represents a user skip on a poll in the repository layer
type Skip struct {
	ID        int
	PollID    int
	UserID    int
	CreatedAt time.Time
}
