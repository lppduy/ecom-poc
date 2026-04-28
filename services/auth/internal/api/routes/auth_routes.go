package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/pkg/jwtutil"
	"github.com/lppduy/ecom-poc/services/auth/internal/api/controller"
)

func Register(r *gin.Engine, ctrl *controller.AuthController, jwtSecret string) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", ctrl.Register)
		auth.POST("/login", ctrl.Login)
		auth.GET("/me", jwtutil.RequireAuth(jwtSecret), ctrl.Me)
	}
}
