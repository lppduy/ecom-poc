package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/pkg/jwtutil"
	"github.com/lppduy/ecom-poc/services/order/internal/api/controller"
)

func RegisterOrderRoutes(router *gin.Engine, orderController *controller.OrderController, jwtSecret string) {
	router.GET("/health", orderController.Health)

	protected := router.Group("/", jwtutil.RequireAuth(jwtSecret))
	{
		protected.POST("/orders", orderController.CreateOrder)
		protected.GET("/orders/:id", orderController.GetOrder)
		protected.PATCH("/orders/:id/confirm", orderController.ConfirmOrder)
		protected.PATCH("/orders/:id/fail", orderController.FailOrder)
	}
}
