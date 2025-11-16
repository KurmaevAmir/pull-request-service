package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/dtos"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/errors"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/services"
)

type PRHandler struct {
	svc services.PRService
}

func NewPRHandler(s services.PRService) *PRHandler { return &PRHandler{svc: s} }

func (h *PRHandler) Create(c *gin.Context) {
	var req dtos.CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RenderError(c, errors.New(errors.CodeValidation, "invalid request"))
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		RenderError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *PRHandler) Merge(c *gin.Context) {
	var req dtos.MergePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RenderError(c, errors.New(errors.CodeValidation, "invalid request"))
		return
	}

	resp, err := h.svc.Merge(c.Request.Context(), req)
	if err != nil {
		RenderError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *PRHandler) Reassign(c *gin.Context) {
	var req dtos.ReassignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RenderError(c, errors.New(errors.CodeValidation, "invalid request"))
		return
	}

	resp, err := h.svc.Reassign(c.Request.Context(), req)
	if err != nil {
		RenderError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
