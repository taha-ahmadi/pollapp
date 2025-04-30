package pollhandler

import (
	"github.com/labstack/echo/v4"
)

func (h *Handler) SetupRoutes(e *echo.Echo) {
	e.POST("/polls", h.Create)
	e.GET("/polls", h.Get)
	e.POST("/polls/:id/vote", h.Vote)
	e.POST("/polls/:id/skip", h.Skip)
	e.GET("/polls/:id/stats", h.GetStats)
}
