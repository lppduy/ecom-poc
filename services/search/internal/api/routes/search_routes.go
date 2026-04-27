package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/search/internal/api/controller"
)

func Register(router *gin.Engine, ctrl *controller.SearchController) {
	router.GET("/health", ctrl.Health)
	router.GET("/search", ctrl.Search)
	router.POST("/search/reindex", ctrl.Reindex)
	router.POST("/search/index", ctrl.IndexOne)
}
