package model

import (
	"pollapp/internal/service/domain"
	"time"

	"github.com/lib/pq"
)

// Poll represents a poll entity in the repository layer
type Poll struct {
	ID        int
	Title     string
	Options   pq.StringArray
	Tags      pq.StringArray
	CreatedAt time.Time
}

func (p Poll) ToDomain() domain.Poll {
	return domain.Poll{
		ID:        uint(p.ID),
		Title:     p.Title,
		Options:   p.Options,
		Tags:      p.Tags,
		CreatedAt: p.CreatedAt,
	}
}

func ToModel(domainPoll domain.Poll) Poll {
	return Poll{
		ID:        int(domainPoll.ID),
		Title:     domainPoll.Title,
		Options:   domainPoll.Options,
		Tags:      domainPoll.Tags,
		CreatedAt: domainPoll.CreatedAt,
	}
}

// PollStats represents poll statistics in the repository layer
type PollStats struct {
	ID        int
	Title     string
	Options   pq.StringArray
	Tags      pq.StringArray
	CreatedAt time.Time
	Votes     []OptionStat
}

// OptionStat represents vote statistics for a single option
type OptionStat struct {
	Option string
	Count  int
}
