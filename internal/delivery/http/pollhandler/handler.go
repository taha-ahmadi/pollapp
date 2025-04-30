package pollhandler

import (
	"pollapp/internal/service/pollservice"
)

type Handler struct {
	pollService pollservice.IPollService
}

func NewHandler(pollService pollservice.IPollService) *Handler {
	return &Handler{pollService: pollService}
}
