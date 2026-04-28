package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/payment/internal/api/controller"
	"github.com/lppduy/ecom-poc/services/payment/internal/api/routes"
	"github.com/lppduy/ecom-poc/services/payment/internal/client"
	"github.com/lppduy/ecom-poc/services/payment/internal/config"
	"github.com/lppduy/ecom-poc/services/payment/internal/repository"
	"github.com/lppduy/ecom-poc/services/payment/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := repository.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("payment: connect postgres: %v", err)
	}

	paymentRepo := repository.NewGormPaymentRepository(db)
	orderClient := client.NewOrderHTTPClient(cfg.OrderBaseURL)
	paymentSvc := service.NewPaymentService(paymentRepo, orderClient)
	paymentCtrl := controller.NewPaymentController(paymentSvc)

	r := gin.Default()
	routes.Register(r, paymentCtrl, cfg.JWTSecret)

	log.Printf("payment service listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("payment: run: %v", err)
	}
}
