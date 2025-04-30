package domain

import (
	"time"
)

type Vote struct {
	ID          int
	PollID      int
	UserID      int
	OptionIndex int
	CreatedAt   time.Time
}
