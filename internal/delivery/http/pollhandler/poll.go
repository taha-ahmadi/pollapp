package pollhandler

import (
	"errors"
	"net/http"
	"strconv"

	httperror "pollapp/pkg/errors"

	"pollapp/internal/service/domain"

	"github.com/labstack/echo/v4"
)

// @router /polls [post]
// @summary Create a new poll
// @description Create a new poll with title, options, and tags
// @tags Polls
// @accept json
// @produce json
// @param request body CreatePollRequest true "Poll creation request"
// @success 201 "Poll created successfully"
// @failure 400 {object} httperror.ErrorResponse "Invalid request"
// @failure 500 {object} httperror.ErrorResponse "Internal server error"
// @security BearerAuth
func (h *Handler) Create(c echo.Context) error {
	var req CreatePollRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperror.ErrInvalidRequest)
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperror.ErrInvalidRequest)
	}

	poll := domain.Poll{
		Title:   req.Title,
		Options: req.Options,
		Tags:    req.Tags,
	}

	err := h.pollService.Create(c.Request().Context(), poll)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httperror.NewErrorResponse(
			http.StatusInternalServerError,
			"failed to create poll",
			err.Error(),
		))
	}

	return c.NoContent(http.StatusCreated)
}

// @router /polls [get]
// @summary Get polls
// @description Get a list of polls with optional filtering and pagination
// @tags Polls
// @accept json
// @produce json
// @param request query GetPollsRequest true "Poll query parameters"
// @success 200 {array} GetPollsResponse "List of polls"
// @failure 400 {object} httperror.ErrorResponse "Invalid request"
// @failure 500 {object} httperror.ErrorResponse "Internal server error"
// @security BearerAuth
func (h *Handler) Get(c echo.Context) error {
	var req GetPollsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperror.ErrInvalidRequest)
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperror.ErrInvalidPaginationResponse)
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	params := &domain.GetPollsParams{
		UserID: req.UserID,
		Tag:    req.Tag,
		Page:   req.Page,
		Limit:  req.Limit,
	}

	polls, err := h.pollService.Get(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httperror.NewErrorResponse(
			http.StatusInternalServerError,
			"failed to get polls",
			err.Error(),
		))
	}

	response := make([]GetPollsResponse, len(polls))
	for i, poll := range polls {
		options := poll.Options
		tags := poll.Tags

		response[i] = GetPollsResponse{
			ID:        poll.ID,
			Title:     poll.Title,
			Options:   options,
			Tags:      tags,
			CreatedAt: poll.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, response)
}

// @router /polls/{id}/vote [post]
// @summary Vote on a poll
// @description Submit a vote for a specific option in a poll
// @tags Polls
// @accept json
// @produce json
// @param id path int true "Poll ID"
// @param request body VoteRequest true "Vote request"
// @success 200 "Vote recorded successfully"
// @failure 400 {object} httperror.ErrorResponse "Invalid request"
// @failure 404 {object} httperror.ErrorResponse "Poll not found"
// @failure 409 {object} httperror.ErrorResponse "Already voted"
// @failure 500 {object} httperror.ErrorResponse "Internal server error"
// @security BearerAuth
func (h *Handler) Vote(c echo.Context) error {
	pollID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httperror.NewErrorResponse(
			http.StatusBadRequest,
			httperror.ErrInvalidPollID,
			"poll ID must be a valid number",
		))
	}

	var req VoteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperror.ErrInvalidRequest)
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperror.ErrInvalidRequest)
	}

	poll, err := h.pollService.GetByID(c.Request().Context(), uint(pollID))
	if err != nil {
		if httperror.IsNotFoundError(err) {
			return c.JSON(http.StatusNotFound, httperror.ErrPollNotFoundResponse)
		}
		return c.JSON(http.StatusInternalServerError, httperror.NewErrorResponse(
			http.StatusInternalServerError,
			"failed to get poll",
			err.Error(),
		))
	}

	if req.OptionIndex >= len(poll.Options) {
		return c.JSON(http.StatusBadRequest, httperror.ErrInvalidOptionResponse)
	}

	// Use async queue-based vote system for better handling of high volume
	err = h.pollService.QueueVote(c.Request().Context(), uint(pollID), domain.VoteRequest{
		UserID:      req.UserID,
		OptionIndex: req.OptionIndex,
	})
	if err != nil {
		statusCode, errResponse := httperror.HandleVoteError(err)
		return c.JSON(statusCode, errResponse)
	}

	return c.NoContent(http.StatusOK)
}

// @router /polls/{id}/skip [post]
// @summary Skip a poll
// @description Mark a poll as skipped for a user
// @tags Polls
// @accept json
// @produce json
// @param id path int true "Poll ID"
// @param request body SkipRequest true "Skip request"
// @success 200 "Poll skipped successfully"
// @failure 400 {object} httperror.ErrorResponse "Invalid request"
// @failure 404 {object} httperror.ErrorResponse "Poll not found"
// @failure 409 {object} httperror.ErrorResponse "Already skipped"
// @failure 500 {object} httperror.ErrorResponse "Internal server error"
// @security BearerAuth
func (h *Handler) Skip(c echo.Context) error {
	pollID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httperror.NewErrorResponse(
			http.StatusBadRequest,
			httperror.ErrInvalidPollID,
			"poll ID must be a valid number",
		))
	}

	var req SkipRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperror.ErrInvalidRequest)
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperror.NewErrorResponse(
			http.StatusBadRequest,
			"validation failed",
			err.Error(),
		))
	}

	_, err = h.pollService.GetByID(c.Request().Context(), uint(pollID))
	if err != nil {
		if httperror.IsNotFoundError(err) {
			return c.JSON(http.StatusNotFound, httperror.ErrPollNotFoundResponse)
		}
		return c.JSON(http.StatusInternalServerError, httperror.NewErrorResponse(
			http.StatusInternalServerError,
			"failed to get poll",
			err.Error(),
		))
	}

	err = h.pollService.Skip(c.Request().Context(), uint(pollID), domain.SkipRequest{
		UserID: req.UserID,
	})
	if err != nil {
		switch {
		case errors.Is(err, errors.New(httperror.ErrAlreadySkipped)):
			return c.JSON(http.StatusConflict, httperror.ErrAlreadySkippedResponse)
		default:
			return c.JSON(http.StatusInternalServerError, httperror.NewErrorResponse(
				http.StatusInternalServerError,
				"failed to record skip",
				err.Error(),
			))
		}
	}

	return c.NoContent(http.StatusOK)
}

// @router /polls/{id}/stats [get]
// @summary Get poll statistics
// @description Get voting statistics for a specific poll
// @tags Polls
// @accept json
// @produce json
// @param id path int true "Poll ID"
// @success 200 {object} GetPollStatsResponse "Poll statistics"
// @failure 400 {object} httperror.ErrorResponse "Invalid request"
// @failure 404 {object} httperror.ErrorResponse "Poll not found"
// @failure 500 {object} httperror.ErrorResponse "Internal server error"
// @security BearerAuth
func (h *Handler) GetStats(c echo.Context) error {
	pollID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httperror.NewErrorResponse(
			http.StatusBadRequest,
			httperror.ErrInvalidPollID,
			"poll ID must be a valid number",
		))
	}

	_, err = h.pollService.GetByID(c.Request().Context(), uint(pollID))
	if err != nil {
		if httperror.IsNotFoundError(err) {
			return c.JSON(http.StatusNotFound, httperror.ErrPollNotFoundResponse)
		}
		return c.JSON(http.StatusInternalServerError, httperror.NewErrorResponse(
			http.StatusInternalServerError,
			"failed to get poll",
			err.Error(),
		))
	}

	stats, err := h.pollService.GetStats(c.Request().Context(), uint(pollID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httperror.NewErrorResponse(
			http.StatusInternalServerError,
			"failed to get poll statistics",
			err.Error(),
		))
	}

	totalVotes := 0
	for _, vote := range stats.Votes {
		totalVotes += vote.Count
	}

	votes := make([]Vote, len(stats.Votes))
	for i, vote := range stats.Votes {
		votes[i] = Vote{
			Option: vote.Option,
			Count:  vote.Count,
		}
	}

	response := GetPollStatsResponse{
		PollID:     uint(pollID),
		TotalVotes: totalVotes,
		Votes:      votes,
	}

	return c.JSON(http.StatusOK, response)
}
