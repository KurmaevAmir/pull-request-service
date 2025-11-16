package handlers

import (
	"net/http"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	service *services.StatsService
}

func NewStatsHandler(service *services.StatsService) *StatsHandler {
	return &StatsHandler{service: service}
}

func (h *StatsHandler) GetAssignments(c *gin.Context) {
	stats, err := h.service.GetAssignmentStats(c.Request.Context())

	if err != nil {
		RenderError(c, err)
	}

	c.JSON(http.StatusOK, stats)
}
