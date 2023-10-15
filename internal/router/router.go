package router

import (
	"net/http"

	"car-rental/internal/controller"
	"github.com/gin-gonic/gin"
)

func NewRouter(controller controller.RentalController) *gin.Engine {
	service := gin.Default()

	service.GET("", func(context *gin.Context) {
		context.JSON(http.StatusOK, "welcome home")
	})

	service.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	router := service.Group("/api/v1")
	autoRouter := router.Group("/auto")
	{
		autoRouter.GET("/type/:type", controller.GetAvailableByType)
		autoRouter.POST("/bind", controller.BindAuto)
		autoRouter.GET("/release/:auto_id", controller.ReleaseAuto)
		autoRouter.GET("/commission/:auto_id", controller.GetCurrentCommission)
	}
	return service
}
