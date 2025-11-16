package handlers

import (
	"github.com/gin-gonic/gin"
)

func NewRouter(
	teamHandler *TeamHandler,
	userHandler *UserHandler,
	prHandler *PRHandler,
	statsHandler *StatsHandler,
) *gin.Engine {
	router := gin.Default()

	team := router.Group("/team")
	team.POST("/add", teamHandler.AddTeam)
	team.GET("/get", teamHandler.GetTeam)
	team.POST("/bulkDeactivate", teamHandler.BulkDeactivate)

	users := router.Group("/users")
	users.POST("/setIsActive", userHandler.SetIsActive)
	users.GET("/getReview", userHandler.GetReview)

	pr := router.Group("/pullRequest")
	pr.POST("/create", prHandler.Create)
	pr.POST("/merge", prHandler.Merge)
	pr.POST("/reassign", prHandler.Reassign)

	stats := router.Group("/stats")
	stats.GET("/assignments", statsHandler.GetAssignments)

	return router
}
