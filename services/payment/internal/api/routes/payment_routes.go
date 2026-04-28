package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/pkg/jwtutil"
	"github.com/lppduy/ecom-poc/services/payment/internal/api/controller"
)

func Register(r *gin.Engine, ctrl *controller.PaymentController, jwtSecret string) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	protected := r.Group("/", jwtutil.RequireAuth(jwtSecret))
	{
		protected.POST("/payments", ctrl.CreatePayment)
		protected.GET("/payments/:id", ctrl.GetPayment)
		protected.POST("/payments/:id/callback", ctrl.Callback)
	}
}
