package domain

import (
	"time"
)

type VoteJob struct {
	ID          string    `json:"id"`
	PollID      uint      `json:"poll_id"`
	UserID      uint      `json:"user_id"`
	OptionIndex int       `json:"option_index"`
	CreatedAt   time.Time `json:"created_at"`
}
