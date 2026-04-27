package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/inventory/internal/api/controller"
)

func Register(router *gin.Engine, ctrl *controller.InventoryController) {
	router.GET("/health", ctrl.Health)
	router.GET("/inventory/stock/:productId", ctrl.GetStock)
	router.POST("/inventory/reserve", ctrl.Reserve)
	router.POST("/inventory/release", ctrl.Release)
	router.POST("/inventory/confirm", ctrl.Confirm)
}
