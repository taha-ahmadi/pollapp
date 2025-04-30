package domain

import (
	"time"
)

type Skip struct {
	ID        int
	PollID    int
	UserID    int
	CreatedAt time.Time
}
