package main

import (
	"context"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/payment/internal/api/controller"
	"github.com/lppduy/ecom-poc/services/payment/internal/api/routes"
	"github.com/lppduy/ecom-poc/services/payment/internal/config"
	"github.com/lppduy/ecom-poc/services/payment/internal/event"
	"github.com/lppduy/ecom-poc/services/payment/internal/repository"
	"github.com/lppduy/ecom-poc/services/payment/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := repository.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("payment: connect postgres: %v", err)
	}

	brokers := strings.Split(cfg.KafkaBrokers, ",")
	publisher := event.NewKafkaPublisher(brokers)
	defer publisher.Close()

	outboxRepo := repository.NewPaymentOutboxRepository(db)
	event.StartRelay(context.Background(), outboxRepo, publisher)

	paymentRepo := repository.NewGormPaymentRepository(db)
	paymentSvc := service.NewPaymentService(paymentRepo)
	paymentCtrl := controller.NewPaymentController(paymentSvc)

	r := gin.Default()
	routes.Register(r, paymentCtrl, cfg.JWTSecret)

	log.Printf("payment service listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("payment: run: %v", err)
	}
}
