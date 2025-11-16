package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/errors"
)

func RenderError(c *gin.Context, err error) {
	if de, ok := errors.IsDomain(err); ok {
		c.JSON(statusFromDomain(de.Code), gin.H{
			"error": gin.H{
				"code":    string(de.Code),
				"message": de.Message,
			},
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"error": gin.H{
			"code":    string(errors.CodeValidation),
			"message": "invalid request",
		},
	})
}

func statusFromDomain(code errors.Code) int {
	switch code {
	case errors.CodeNotFound:
		return http.StatusNotFound
	case errors.CodeValidation:
		return http.StatusBadRequest
	case errors.CodePRMerged:
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
}
