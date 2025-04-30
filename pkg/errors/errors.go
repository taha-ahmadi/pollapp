package pollhandler

import (
	"errors"
	"net/http"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Common error messages
const (
	ErrInvalidPollID     = "invalid poll ID"
	ErrPollNotFound      = "poll not found"
	ErrAlreadyVoted      = "you have already voted on this poll"
	ErrAlreadySkipped    = "you have already skipped this poll"
	ErrDailyVoteLimit    = "daily vote limit exceeded (maximum 100 votes per day)"
	ErrInvalidOption     = "invalid option index"
	ErrInvalidPagination = "invalid pagination parameters"
)

// NewErrorResponse creates a new error response
func NewErrorResponse(code int, message string, details string) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Common error responses
var (
	ErrInvalidRequest = NewErrorResponse(
		http.StatusBadRequest,
		"invalid request",
		"please check your request parameters",
	)

	ErrPollNotFoundResponse = NewErrorResponse(
		http.StatusNotFound,
		ErrPollNotFound,
		"the requested poll does not exist",
	)

	ErrAlreadyVotedResponse = NewErrorResponse(
		http.StatusConflict,
		ErrAlreadyVoted,
		"you can only vote once per poll",
	)

	ErrAlreadySkippedResponse = NewErrorResponse(
		http.StatusConflict,
		ErrAlreadySkipped,
		"you have already skipped this poll",
	)

	ErrDailyVoteLimitResponse = NewErrorResponse(
		http.StatusTooManyRequests,
		ErrDailyVoteLimit,
		"please try again tomorrow",
	)

	ErrInvalidOptionResponse = NewErrorResponse(
		http.StatusBadRequest,
		ErrInvalidOption,
		"the option index must be between 0 and the number of options minus 1",
	)

	ErrInvalidPaginationResponse = NewErrorResponse(
		http.StatusBadRequest,
		ErrInvalidPagination,
		"page must be greater than 0 and limit must be between 1 and 100",
	)
)

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, errors.New(ErrPollNotFound))
}

// HandleVoteError handles vote-related errors and returns appropriate HTTP response
func HandleVoteError(err error) (int, ErrorResponse) {
	switch {
	case errors.Is(err, errors.New(ErrAlreadyVoted)):
		return http.StatusConflict, ErrAlreadyVotedResponse
	case errors.Is(err, errors.New(ErrDailyVoteLimit)):
		return http.StatusTooManyRequests, ErrDailyVoteLimitResponse
	default:
		return http.StatusInternalServerError, NewErrorResponse(
			http.StatusInternalServerError,
			"failed to record vote",
			err.Error(),
		)
	}
}
