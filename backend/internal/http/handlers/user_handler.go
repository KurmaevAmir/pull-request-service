package handlers

import (
	"net/http"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/errors"
	"github.com/gin-gonic/gin"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/dtos"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/services"
)

type UserHandler struct {
	svc services.UserService
}

func NewUserHandler(s services.UserService) *UserHandler { return &UserHandler{svc: s} }

func (h *UserHandler) SetIsActive(c *gin.Context) {
	var in dtos.SetIsActiveRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		RenderError(c, errors.New(errors.CodeValidation, "invalid request"))
		return
	}

	resp, err := h.svc.SetIsActive(c.Request.Context(), in)
	if err != nil {
		RenderError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetReview(c *gin.Context) {
	userID := c.Query("user_id")

	resp, err := h.svc.GetReview(c.Request.Context(), userID)
	if err != nil {
		RenderError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
