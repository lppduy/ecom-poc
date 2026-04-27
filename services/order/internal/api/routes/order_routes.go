package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/order/internal/api/controller"
)

func RegisterOrderRoutes(router *gin.Engine, orderController *controller.OrderController) {
	router.GET("/health", orderController.Health)
	router.POST("/orders", orderController.CreateOrder)
	router.GET("/orders/:id", orderController.GetOrder)
	router.PATCH("/orders/:id/confirm", orderController.ConfirmOrder)
	router.PATCH("/orders/:id/fail", orderController.FailOrder)
}
