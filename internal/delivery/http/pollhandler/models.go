package pollhandler

import (
	"time"
)

// @title Poll API
// @version 1.0
// @description API for managing polls and votes
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @tag.name Polls
// @tag.description Poll management endpoints

// GetPollsRequest represents the query parameters for getting polls
// @Description GetPollsRequest represents the query parameters for getting polls
type GetPollsRequest struct {
	UserID uint   `query:"userId" validate:"required" example:"1"`
	Tag    string `query:"tag" example:"technology"`
	Page   int    `query:"page" validate:"min=1" default:"1" example:"1"`
	Limit  int    `query:"limit" validate:"min=1,max=100" default:"20" example:"20"`
}

// GetPollsResponse represents the response for getting polls
// @Description GetPollsResponse represents the response for getting polls
type GetPollsResponse struct {
	ID        uint      `json:"id" example:"1"`
	Title     string    `json:"title" example:"What's your favorite programming language?"`
	Options   []string  `json:"options" example:"['Go', 'Python', 'JavaScript']"`
	Tags      []string  `json:"tags" example:"['technology', 'programming']"`
	CreatedAt time.Time `json:"createdAt" example:"2024-03-20T10:00:00Z"`
}

// CreatePollRequest represents the request for creating a poll
// @Description CreatePollRequest represents the request for creating a poll
type CreatePollRequest struct {
	Title   string   `json:"title" validate:"required,min=3,max=200" example:"What's your favorite programming language?"`
	Options []string `json:"options" validate:"required,min=2,max=10,dive,min=1,max=100" example:"['Go', 'Python', 'JavaScript']"`
	Tags    []string `json:"tags" validate:"required,min=1,max=5,dive,min=2,max=50" example:"['technology', 'programming']"`
}

// VoteRequest represents the request for voting on a poll
// @Description VoteRequest represents the request for voting on a poll
type VoteRequest struct {
	UserID      uint `json:"userId" validate:"required" example:"1"`
	OptionIndex int  `json:"optionIndex" validate:"required,min=0" example:"0"`
}

// SkipRequest represents the request for skipping a poll
// @Description SkipRequest represents the request for skipping a poll
type SkipRequest struct {
	UserID uint `json:"userId" validate:"required" example:"1"`
}

// GetPollStatsResponse represents the response for getting poll statistics
// @Description GetPollStatsResponse represents the response for getting poll statistics
type GetPollStatsResponse struct {
	PollID     uint   `json:"pollId" example:"1"`
	TotalVotes int    `json:"totalVotes" example:"100"`
	Votes      []Vote `json:"votes"`
}

// Vote represents vote statistics for a single option
// @Description Vote represents vote statistics for a single option
type Vote struct {
	Option string `json:"option" example:"Go"`
	Count  int    `json:"count" example:"50"`
}
