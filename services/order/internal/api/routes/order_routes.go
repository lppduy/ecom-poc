package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/pkg/jwtutil"
	"github.com/lppduy/ecom-poc/pkg/ratelimit"
	"github.com/lppduy/ecom-poc/services/order/internal/api/controller"
	"github.com/redis/go-redis/v9"
)

func RegisterOrderRoutes(router *gin.Engine, orderController *controller.OrderController, jwtSecret string, rdb *redis.Client) {
	router.GET("/health", orderController.Health)

	// 5 orders per minute per user - prevents Black Friday order flooding
	orderLimiter := ratelimit.SlidingWindow(rdb, 5, time.Minute)

	protected := router.Group("/", jwtutil.RequireAuth(jwtSecret))
	{
		protected.POST("/orders", orderLimiter, orderController.CreateOrder)
		protected.GET("/orders/:id", orderController.GetOrder)
	}

	// Internal routes for service-to-service calls (no JWT required)
	internal := router.Group("/internal")
	{
		internal.PATCH("/orders/:id/confirm", orderController.ConfirmOrder)
		internal.PATCH("/orders/:id/fail", orderController.FailOrder)
	}
}
