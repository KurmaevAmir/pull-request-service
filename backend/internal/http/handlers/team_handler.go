package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/dtos"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/errors"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/services"
)

type TeamHandler struct {
	svc services.TeamService
}

func NewTeamHandler(s services.TeamService) *TeamHandler {
	return &TeamHandler{svc: s}
}

func (h *TeamHandler) AddTeam(c *gin.Context) {
	var in dtos.AddTeamRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		RenderError(c, errors.New(errors.CodeValidation, "invalid request"))
		return
	}
	resp, err := h.svc.AddTeam(c.Request.Context(), in)
	if err != nil {
		RenderError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	name := c.Query("team_name")
	if name == "" {
		RenderError(c, errors.New(errors.CodeValidation, "team_name required"))
		return
	}
	out, err := h.svc.GetTeam(c.Request.Context(), name)
	if err != nil {
		RenderError(c, err)
		return
	}
	c.JSON(http.StatusOK, dtos.TeamResponse{Team: out})
}
