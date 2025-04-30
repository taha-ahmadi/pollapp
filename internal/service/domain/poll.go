package domain

import (
	"time"
)

type Poll struct {
	ID        uint
	Title     string
	Options   []string
	Tags      []string
	CreatedAt time.Time
}

type Tag struct {
	ID   uint
	Name string
}

type GetPollsParams struct {
	UserID uint
	Tag    string
	Page   int
	Limit  int
}

type VoteRequest struct {
	UserID      uint
	OptionIndex int
}

type SkipRequest struct {
	UserID uint
}

type PollFilter struct {
	UserID    uint
	Tags      []string
	Status    string
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}
