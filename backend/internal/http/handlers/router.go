package handlers

import (
	"github.com/gin-gonic/gin"
)

func NewRouter(
	teamHandler *TeamHandler,
	userHandler *UserHandler,
	prHandler *PRHandler,
) *gin.Engine {
	router := gin.Default()

	team := router.Group("/team")
	team.POST("/add", teamHandler.AddTeam)
	team.GET("/get", teamHandler.GetTeam)

	users := router.Group("/users")
	users.POST("/setIsActive", userHandler.SetIsActive)
	users.GET("/getReview", userHandler.GetReview)

	pr := router.Group("/pullRequest")
	pr.POST("/create", prHandler.Create)
	pr.POST("/merge", prHandler.Merge)
	pr.POST("/reassign", prHandler.Reassign)

	return router
}
