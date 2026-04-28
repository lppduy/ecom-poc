package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/pkg/jwtutil"
	"github.com/lppduy/ecom-poc/pkg/ratelimit"
	"github.com/lppduy/ecom-poc/services/auth/internal/api/controller"
	"github.com/redis/go-redis/v9"
)

func Register(r *gin.Engine, ctrl *controller.AuthController, jwtSecret string, rdb *redis.Client) {
	// 10 attempts per minute per IP on login/register to prevent brute-force
	authLimiter := ratelimit.SlidingWindow(rdb, 10, time.Minute)

	auth := r.Group("/auth")
	{
		auth.POST("/register", authLimiter, ctrl.Register)
		auth.POST("/login", authLimiter, ctrl.Login)
		auth.GET("/me", jwtutil.RequireAuth(jwtSecret), ctrl.Me)
	}
}
